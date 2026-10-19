#!/usr/bin/env python3
"""Build bilingual release notes using keyless Edge/Google translation."""

import argparse
import json
import re
import sys
from pathlib import Path
from urllib.parse import urlencode
from urllib.request import Request, urlopen


LITERALS = re.compile(r"`+[^`]+`+|https?://[^\s<>\]\)]+|\b[0-9a-f]{7,40}\b")


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


def request_json(request):
    with urlopen(request, timeout=15) as response:
        return json.loads(response.read(2 * 1024 * 1024))


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
    providers = providers if providers is not None else (edge, google)
    output = list(subjects)
    # Already-Chinese subjects need no request. Bound each batch and preserve
    # code, hashes and links instead of asking a translator to rewrite Markdown.
    pending = [i for i, text in enumerate(subjects) if not re.search(r"[\u3400-\u9fff]", text)]
    while pending:
        batch = [pending.pop(0)]
        size = len(subjects[batch[0]])
        while pending and len(batch) < 20 and size + len(subjects[pending[0]]) < 2000:
            batch.append(pending.pop(0))
            size += len(subjects[batch[-1]])
        if size > 5000:
            continue
        protected = [protect(subjects[i]) for i in batch]
        for provider in providers:
            try:
                values = provider([text for text, _ in protected])
                if len(values) != len(batch):
                    raise ValueError("translation returned a different number of subjects")
                restored = [restore(value, literals) for value, (_, literals) in zip(values, protected)]
                for i, value in zip(batch, restored):
                    output[i] = value
                break
            except Exception as error:
                print(f"Translation via {provider.__name__} failed ({type(error).__name__}); trying fallback.", file=sys.stderr)
        else:
            print("Translation unavailable; keeping the original commit subjects.", file=sys.stderr)
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
