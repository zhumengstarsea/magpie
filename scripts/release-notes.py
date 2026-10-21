#!/usr/bin/env python3
"""Build bilingual release notes using free AI and web translation."""

import argparse
import json
import os
import re
import sys
import time
from pathlib import Path
from urllib.parse import urlencode
from urllib.error import HTTPError
from urllib.request import Request, urlopen


# Keep source attributions and technical literals intact. A Chinese name or
# a parenthesized Discord attribution must not make an otherwise English
# subject look like an already translated subject.
LITERALS = re.compile(
    r"`+[^`]+`+|https?://[^\s<>\]\)]+|\b[0-9a-f]{7,40}\b|"
    r"\([^()\n]*[\u3400-\u9fff][^()\n]*\)"
)


def mostly_chinese(text):
    """Return true only when Chinese is the subject's main language."""
    text = LITERALS.sub("", text)
    chinese = len(re.findall(r"[\u3400-\u9fff]", text))
    latin = len(re.findall(r"[A-Za-z]", text))
    return chinese >= 2 and chinese >= latin


def protect(text):
    literals = []

    def replace(match):
        token = f"ZXQMP{len(literals):04d}QXZ"
        literals.append((token, match.group()))
        return token

    return LITERALS.sub(replace, text), literals


def restore(text, literals):
    if not isinstance(text, str) or not text.strip() or "\n" in text.strip():
        raise ValueError("invalid translated subject")
    for token, value in literals:
        if text.count(token) != 1:
            raise ValueError("translation changed a protected literal")
        text = text.replace(token, value)
    return text.strip()


def request_json(request, timeout=15):
    with urlopen(request, timeout=timeout) as response:
        return json.loads(response.read(2 * 1024 * 1024))


def parse_zhipu_content(content, expected):
    """Accept the JSON object we ask for and common model variants."""
    if isinstance(content, list):
        values = content
    elif isinstance(content, dict):
        values = content.get("translations")
    elif isinstance(content, str):
        text = content.strip()
        text = re.sub(r"^```(?:json)?\s*|\s*```$", "", text, flags=re.I)
        try:
            parsed = json.loads(text)
        except json.JSONDecodeError:
            values = [text] if expected == 1 and text else None
        else:
            if isinstance(parsed, dict):
                values = parsed.get("translations")
            elif isinstance(parsed, list):
                values = parsed
            elif expected == 1 and isinstance(parsed, str):
                values = [parsed]
            else:
                values = None
    else:
        values = None
    if isinstance(values, str):
        values = [line.strip() for line in values.splitlines() if line.strip()]
    if isinstance(values, list):
        converted = []
        for value in values:
            if isinstance(value, str):
                converted.append(value)
            elif isinstance(value, dict):
                text_value = value.get("translation", value.get("text"))
                if isinstance(text_value, str):
                    converted.append(text_value)
                else:
                    converted = None
                    break
            else:
                converted = None
                break
        values = converted
    if values is None and isinstance(content, str) and expected > 1:
        lines = []
        for line in content.splitlines():
            line = re.sub(r"^\s*(?:[-*•]|\d+[.)])\s*", "", line).strip()
            if line:
                lines.append(line)
        if len(lines) == expected:
            values = lines
    if not isinstance(values, list) or len(values) != expected:
        raise ValueError("Zhipu returned an invalid translations array")
    if not all(isinstance(value, str) and value.strip() for value in values):
        raise ValueError("Zhipu returned an invalid translations array")
    return values


ZHIPU_MODELS = (
    # Prefer the newest/general model at the largest context, then the free
    # 128K text models, and finally the 64K V-thinking fallback.
    "glm-4.7-flash",
    "glm-4-flash-250414",
    "glm-z1-flash",
    "glm-4.1v-thinking-flash",
)

ZHIPU_CONTEXT_TOKENS = {
    "glm-4.7-flash": 200_000,
    "glm-4-flash-250414": 128_000,
    "glm-z1-flash": 128_000,
    "glm-4.1v-thinking-flash": 64_000,
}
ZHIPU_MAX_OUTPUT_TOKENS = 4096
ZHIPU_SAFE_INPUT_TOKENS = 8_000

GOOGLE_AI_MODELS = (
    "gemini-3.8-flash",
    "gemini-3.6-flash",
    "gemini-3.5-flash-lite",
    "gemini-3.1-flash-lite",
    "gemini-3-flash-preview",
)
GOOGLE_AI_CONTEXT_TOKENS = {model: 1_048_576 for model in GOOGLE_AI_MODELS}
GOOGLE_AI_MAX_OUTPUT_TOKENS = 4096


def zhipu_once(texts, model):
    key = os.environ.get("ZHIPU_API_KEY")
    if not key:
        raise RuntimeError("ZHIPU_API_KEY is not configured")
    prompt = (
        "Translate each English commit subject to Simplified Chinese. "
        "Return only a JSON object with a translations array of strings with the same length and order. "
        "Preserve tokens such as ZXQMP0000QXZ, code, URLs, names and issue numbers. "
        "Do not add explanations.\n\n"
        + json.dumps(texts, ensure_ascii=False)
    )
    request = Request(
        "https://open.bigmodel.cn/api/paas/v4/chat/completions",
        data=json.dumps({
            "model": model,
            "messages": [{"role": "user", "content": prompt}],
            "temperature": 0,
            "do_sample": False,
            "thinking": {"type": "disabled"},
            "stream": False,
            "max_tokens": ZHIPU_MAX_OUTPUT_TOKENS,
            "response_format": {"type": "json_object"},
        }, ensure_ascii=False).encode("utf-8"),
        headers={
            "Content-Type": "application/json",
            "User-Agent": "magpie-release-notes",
            "Authorization": "Bearer " + key,
        },
        method="POST",
    )
    result = request_json(request, timeout=60)
    text = result["choices"][0]["message"]["content"]
    return parse_zhipu_content(text, len(texts))


def zhipu_business_code(error):
    """Read Zhipu's inner business code without exposing its response."""
    cached = getattr(error, "zhipu_business_code", None)
    if cached is not None:
        return cached
    try:
        raw = error.read()
        error.zhipu_body = raw
        payload = json.loads(raw.decode("utf-8", "replace"))
        code = int(payload.get("error", {}).get("code", 0))
        error.zhipu_business_code = code
        return code
    except (AttributeError, ValueError, TypeError, json.JSONDecodeError):
        return 0


def zhipu_retryable(error):
    """Retry temporary transport/model throttles, not quota exhaustion."""
    if not isinstance(error, HTTPError):
        return False
    business = zhipu_business_code(error)
    return error.code in (500, 502, 503, 529) or business in (1302, 1305)


def zhipu_switch_model(error):
    """Allow model fallback for model-specific or temporary failures."""
    if not isinstance(error, HTTPError):
        return False
    business = getattr(error, "zhipu_business_code", None)
    if business is None:
        business = zhipu_business_code(error)
    return (
        error.code in (500, 502, 503, 529)
        or business in (1211, 1212, 1220, 1221, 1302, 1305, 1311)
    )


def zhipu_batch(texts, model):
    """Translate one batch, splitting only when the response cannot fit."""
    for attempt in range(3):
        try:
            return zhipu_once(texts, model)
        except ValueError:
            if len(texts) == 1:
                raise
            # Keep the largest successful request. A malformed or truncated
            # response usually means this batch is too large for the model.
            middle = len(texts) // 2
            return (zhipu_batch(texts[:middle], model)
                    + zhipu_batch(texts[middle:], model))
        except HTTPError as error:
            business = zhipu_business_code(error)
            if len(texts) > 1 and business == 1261:
                middle = len(texts) // 2
                return (zhipu_batch(texts[:middle], model)
                        + zhipu_batch(texts[middle:], model))
            if not zhipu_retryable(error) or attempt == 2:
                raise
            time.sleep(1 + attempt)
    raise RuntimeError("Zhipu translation retry loop ended unexpectedly")


def zhipu_estimate_tokens(text):
    """Conservatively estimate tokens for mixed English/Chinese subjects."""
    ascii_count = sum(ord(char) < 128 for char in text)
    other_count = len(text) - ascii_count
    return max(1, (ascii_count + 2) // 3 + (other_count + 1) // 2)


def translation_context_batches(texts, context, max_output):
    """Pack as many subjects as the model context and output can hold."""
    # Leave room for the instruction, JSON framing, and a safety margin.
    input_budget = context - 1024
    output_budget = int(max_output * 0.8)
    batch = []
    input_tokens = 0
    output_tokens = 0
    for text in texts:
        estimated = zhipu_estimate_tokens(text)
        input_cost = estimated + 8
        output_cost = max(16, int(estimated * 1.3) + 8)
        too_large = (
            batch and (
                input_tokens + input_cost + output_tokens + output_cost > input_budget
                or output_tokens + output_cost > output_budget
            )
        )
        if too_large:
            yield batch
            batch = []
            input_tokens = 0
            output_tokens = 0
        batch.append(text)
        input_tokens += input_cost
        output_tokens += output_cost
    if batch:
        yield batch


def zhipu_context_batches(texts, model):
    return translation_context_batches(
        texts,
        min(ZHIPU_CONTEXT_TOKENS[model], ZHIPU_SAFE_INPUT_TOKENS),
        ZHIPU_MAX_OUTPUT_TOKENS,
    )


def zhipu(texts):
    """Translate with context-sized batches and free-model fallback."""
    last_error = None
    for model in ZHIPU_MODELS:
        try:
            output = []
            for batch in zhipu_context_batches(texts, model):
                output.extend(zhipu_batch(batch, model))
            return output
        except (HTTPError, ValueError) as error:
            last_error = error
            if isinstance(error, HTTPError) and not zhipu_switch_model(error):
                raise
            print(f"Zhipu model {model} failed; trying next model.", file=sys.stderr)
    if last_error is not None:
        raise last_error
    raise RuntimeError("No Zhipu models configured")


def google_ai_business_status(error):
    """Read Google's status without exposing the response or API key."""
    cached = getattr(error, "google_status", None)
    if cached is not None:
        return cached
    try:
        raw = error.read()
        error.google_body = raw
        payload = json.loads(raw.decode("utf-8", "replace"))
        status = payload.get("error", {}).get("status", "")
        error.google_status = status
        return status
    except (AttributeError, ValueError, TypeError, json.JSONDecodeError):
        return ""


def google_ai_retryable(error):
    if not isinstance(error, HTTPError):
        return isinstance(error, OSError)
    status = google_ai_business_status(error)
    return error.code in (429, 500, 502, 503, 504) or status in (
        "RESOURCE_EXHAUSTED", "UNAVAILABLE", "INTERNAL",
    )


def google_ai_switch_model(error):
    if not isinstance(error, HTTPError):
        return isinstance(error, OSError)
    status = google_ai_business_status(error)
    return error.code in (400, 404) or status in (
        "INVALID_ARGUMENT", "FAILED_PRECONDITION", "NOT_FOUND",
    ) or google_ai_retryable(error)


def google_ai_once(texts, model):
    key = os.environ.get("GOOGLE_AI_API_KEY")
    if not key:
        raise RuntimeError("GOOGLE_AI_API_KEY is not configured")
    prompt = (
        "Translate each English commit subject to Simplified Chinese. "
        "Return only a JSON object with a translations array of strings with the same length and order. "
        "Preserve tokens such as ZXQMP0000QXZ, code, URLs, names and issue numbers. "
        "Do not add explanations.\n\n"
        + json.dumps(texts, ensure_ascii=False)
    )
    request = Request(
        "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":generateContent",
        data=json.dumps({
            "contents": [{"parts": [{"text": prompt}]}],
            "generationConfig": {
                "temperature": 0,
                "maxOutputTokens": GOOGLE_AI_MAX_OUTPUT_TOKENS,
                "responseMimeType": "application/json",
            },
        }, ensure_ascii=False).encode("utf-8"),
        headers={
            "Content-Type": "application/json",
            "User-Agent": "magpie-release-notes",
            "x-goog-api-key": key,
        },
        method="POST",
    )
    result = request_json(request, timeout=60)
    parts = result["candidates"][0]["content"]["parts"]
    text = "".join(part.get("text", "") for part in parts)
    return parse_zhipu_content(text, len(texts))


def google_ai_batch(texts, model):
    """Translate one Gemini batch, splitting context/response overflows."""
    for attempt in range(3):
        try:
            return google_ai_once(texts, model)
        except ValueError:
            if len(texts) == 1:
                raise
            middle = len(texts) // 2
            return (google_ai_batch(texts[:middle], model)
                    + google_ai_batch(texts[middle:], model))
        except HTTPError as error:
            business = google_ai_business_status(error)
            message = getattr(error, "google_body", b"").decode("utf-8", "replace").lower()
            if len(texts) > 1 and business == "INVALID_ARGUMENT" and any(
                phrase in message for phrase in ("too long", "token limit", "context length")
            ):
                middle = len(texts) // 2
                return (google_ai_batch(texts[:middle], model)
                        + google_ai_batch(texts[middle:], model))
            if not google_ai_retryable(error) or attempt == 2:
                raise
            time.sleep(1 + attempt)
        except OSError:
            if attempt == 2:
                raise
            time.sleep(1 + attempt)
    raise RuntimeError("Google AI translation retry loop ended unexpectedly")


def google_ai(texts):
    """Translate with Google's free Gemini models and model fallback."""
    last_error = None
    for model in GOOGLE_AI_MODELS:
        try:
            output = []
            for batch in translation_context_batches(
                texts, GOOGLE_AI_CONTEXT_TOKENS[model], GOOGLE_AI_MAX_OUTPUT_TOKENS
            ):
                output.extend(google_ai_batch(batch, model))
            return output
        except (HTTPError, OSError, ValueError) as error:
            last_error = error
            if isinstance(error, HTTPError) and not google_ai_switch_model(error):
                raise
            print(f"Google AI model {model} failed; trying next model.", file=sys.stderr)
    if last_error is not None:
        raise last_error
    raise RuntimeError("No Google AI models configured")


def edge(texts):
    query = urlencode({"from": "", "to": "zh-Hans", "isEnterpriseClient": "false"})
    result = request_json(Request(
        "https://edge.microsoft.com/translate/translatetext?" + query,
        data=json.dumps(texts, ensure_ascii=False).encode("utf-8"),
        headers={"Content-Type": "application/json", "User-Agent": "Mozilla/5.0"},
        method="POST",
    ))
    return [row["translations"][0]["text"] for row in result]


def google(texts):
    query = urlencode({
        "client": "gtx", "dt": "t", "dj": "1", "ie": "UTF-8",
        "sl": "auto", "tl": "zh-CN", "q": "\n".join(texts),
    })
    url = "https://translate.googleapis.com/translate_a/single?" + query
    if len(url) > 7500:
        raise ValueError("Google request is too long")
    result = request_json(Request(url, headers={"User-Agent": "Mozilla/5.0"}))
    return "".join(row.get("trans", "") for row in result["sentences"]).strip().splitlines()


def translate(subjects, providers=None):
    providers = providers if providers is not None else tuple(
        ([google_ai] if os.environ.get("GOOGLE_AI_API_KEY") else [])
        + ([zhipu] if os.environ.get("ZHIPU_API_KEY") else [])
        + [edge, google]
    )
    output = list(subjects)
    # Already-Chinese subjects need no request. A mixed subject such as
    # "gateway: ... (新成员 on Discord)" still needs translating. Bound each
    # batch and preserve code, hashes, links and source attributions.
    pending = [i for i, text in enumerate(subjects) if not mostly_chinese(text)]

    def translate_one(subject):
        protected, literals = protect(subject)
        for provider in providers:
            try:
                values = provider([protected])
                if len(values) != 1:
                    raise ValueError("translation returned a different number of subjects")
                return restore(values[0], literals)
            except Exception as error:
                detail = f"HTTP {error.code}" if isinstance(error, HTTPError) else type(error).__name__
                print(f"Translation via {provider.__name__} failed for one subject ({detail}); trying fallback.", file=sys.stderr)
        return subject

    while pending:
        if google_ai in providers or zhipu in providers:
            # Let the AI provider pack the entire pending set using the
            # selected model's context and output budgets.
            batch = pending
            pending = []
        else:
            batch = [pending.pop(0)]
            size = len(subjects[batch[0]])
            while pending and len(batch) < 20 and size + len(subjects[pending[0]]) < 2000:
                batch.append(pending.pop(0))
                size += len(subjects[batch[-1]])
        if len(batch) == 1 and len(subjects[batch[0]]) > 5000:
            for i in batch:
                output[i] = translate_one(subjects[i])
            continue
        protected = [protect(subjects[i]) for i in batch]
        translated = False
        for provider in providers:
            try:
                if provider is google_ai or provider is zhipu:
                    ranges = [(0, len(protected))]
                else:
                    ranges = []
                    start = 0
                    size = 0
                    for offset, (text, _) in enumerate(protected):
                        if offset > start and (offset - start >= 20 or size + len(text) >= 2000):
                            ranges.append((start, offset))
                            start = offset
                            size = 0
                        size += len(text)
                    if start < len(protected):
                        ranges.append((start, len(protected)))
                values = []
                for start, end in ranges:
                    values.extend(provider([text for text, _ in protected[start:end]]))
                if len(values) != len(batch):
                    raise ValueError("translation returned a different number of subjects")
                restored = [restore(value, literals) for value, (_, literals) in zip(values, protected)]
                for i, value in zip(batch, restored):
                    output[i] = value
                translated = True
                break
            except Exception as error:
                detail = f"HTTP {error.code}" if isinstance(error, HTTPError) else type(error).__name__
                print(f"Translation via {provider.__name__} failed ({detail}); trying fallback.", file=sys.stderr)
        if not translated:
            # One malformed or unusually long subject must not discard the
            # translations for the other subjects in its batch.
            for i in batch:
                output[i] = translate_one(subjects[i])
    return output


def render(subjects, translated, repository, version, compare=None, total=None):
    # GitHub rejects release bodies over 125,000 characters. Bound UTF-8 bytes
    # too, so bilingual output and unusually long subjects stay below the limit.
    total = len(subjects) if total is None else total
    count = min(len(subjects), len(translated), 100)
    while True:
        body = render_body(subjects[:count], translated[:count], repository, version)
        if total > count:
            body += f"\nShowing {count} of {total} commits. / 展示 {total} 条提交中的 {count} 条。\n"
        if compare:
            body += f"\n[Full changes / 完整变更]({compare})\n"
        if len(body.encode('utf-8')) <= 120000:
            return body
        count -= 1
        if count < 0:
            raise ValueError("release metadata exceeds the body limit")


def render_body(subjects, translated, repository, version):
    url = f"https://github.com/{repository}/releases/tag/v{version}"
    english = "\n".join(f"- {text}" for text in subjects) or "- Improvements and bug fixes"
    chinese = "\n".join(f"- {text}" for text in translated) or "- 改进与问题修复"
    return f"""### Changes

{english}

### Install

Download from [this fork's GitHub Releases]({url}). In-app updates use this repository.

- **macOS:** `magpie-darwin-arm64.dmg` / `magpie-darwin-amd64.dmg`
- **Windows:** `magpie-windows-amd64.exe` / `magpie-windows-arm64.exe`
- **Linux:** `magpie-linux-amd64` / `magpie-linux-arm64` (GTK 3 and WebKitGTK 4.1)
- **CLI:** `magpie-cli-<os>-<arch>`

<!-- lang:zh -->

### 更新内容

{chinese}

### 安装

从[本仓库的 Releases]({url}) 下载。应用内版本检查、更新下载和更新日志均使用本仓库的发布。
"""


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--commits", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    parser.add_argument("--repository", required=True)
    parser.add_argument("--version", required=True)
    parser.add_argument("--compare", help="GitHub compare link for the complete changelog")
    args = parser.parse_args()
    if not re.fullmatch(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+", args.repository):
        parser.error("invalid repository")
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", args.version):
        parser.error("version must be X.Y.Z")
    if args.compare and not re.fullmatch(r"https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/compare/[0-9a-f]{40}\.\.\.[0-9a-f]{40}", args.compare):
        parser.error("invalid compare link")
    subjects = [line.strip() for line in args.commits.read_text(encoding="utf-8").splitlines() if line.strip()]
    recent = subjects[:100]
    body = render(recent, translate(recent), args.repository, args.version, args.compare, len(subjects))
    args.output.write_text(body, encoding="utf-8", newline="\n")


if __name__ == "__main__":
    main()
