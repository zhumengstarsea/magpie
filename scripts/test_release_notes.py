import importlib.util
import io
import json
import unittest
from contextlib import redirect_stderr
from pathlib import Path
from unittest.mock import patch


spec = importlib.util.spec_from_file_location(
    "release_notes", Path(__file__).with_name("release-notes.py")
)
notes = importlib.util.module_from_spec(spec)
spec.loader.exec_module(notes)


class ReleaseNotesTest(unittest.TestCase):
    def test_english_echo_uses_fallback(self):
        subject = "providers: a switched-off provider says how to turn it on"
        calls = []

        def echo(texts):
            return texts

        def fallback(texts):
            calls.append(texts)
            return ["提供商：关闭时提示如何启用"]

        with redirect_stderr(io.StringIO()):
            result = notes.translate([subject], [echo, fallback])
        self.assertEqual(result, ["提供商：关闭时提示如何启用"])
        self.assertEqual(len(calls), 1)

    def test_chinese_attribution_does_not_validate_english_echo(self):
        subject = "gui: the gateway status stays visible (未寒 on Discord)"
        with redirect_stderr(io.StringIO()), self.assertRaises(RuntimeError):
            notes.translate([subject], [lambda texts: texts])

    def test_chinese_title_with_long_product_name_is_not_translated_again(self):
        subject = "接入 Cloudflare 更新服务"
        def fail_if_called(texts):
            self.fail("Chinese subject was sent for translation")
        self.assertEqual(notes.translate([subject], [fail_if_called]), [subject])

    def test_english_title_with_inline_chinese_is_still_translated(self):
        subject = 'gui: show the 中文 label after opening the settings page'
        self.assertFalse(notes.mostly_chinese(subject))

    def test_web_retry_does_not_retry_successful_batch_items(self):
        calls = []
        def successful(texts):
            calls.append(list(texts))
            return ["界面：刷新模型列表", "网关：修复连接中断"]
        with patch.object(notes, "edge", successful), redirect_stderr(io.StringIO()):
            notes.translate(["gui: refresh the model list", "gateway: fix a disconnected stream"], [notes.edge])
        self.assertEqual(len(calls), 1)

    def test_protected_literal_cannot_count_as_chinese_translation(self):
        subject = "gui: refresh the model list `中文`"
        with redirect_stderr(io.StringIO()), self.assertRaises(RuntimeError):
            notes.translate([subject], [lambda texts: texts])

    def test_all_providers_failed_does_not_publish_english_as_chinese(self):
        def unavailable(texts):
            raise OSError("offline")

        with redirect_stderr(io.StringIO()), self.assertRaises(RuntimeError):
            notes.translate(["gui: refresh the model list"], [unavailable])

    def test_partial_success_only_retries_missing_subject(self):
        calls = []

        def first(texts):
            return ["界面：刷新模型列表", texts[1]]

        def second(texts):
            calls.extend(texts)
            return ["网关：修复连接中断"]

        with redirect_stderr(io.StringIO()):
            result = notes.translate(
                ["gui: refresh the model list", "gateway: fix a disconnected stream"],
                [first, second],
            )
        self.assertEqual(result, ["界面：刷新模型列表", "网关：修复连接中断"])
        self.assertEqual(calls, ["gateway: fix a disconnected stream"])

    def test_web_translation_retries_a_rejected_line_alone(self):
        calls = []

        def flaky(texts):
            calls.append(list(texts))
            if len(texts) > 1:
                return ["界面：刷新模型列表", texts[1]]
            return ["网关：修复连接中断"]

        with patch.object(notes, "edge", flaky), redirect_stderr(io.StringIO()):
            result = notes.translate(
                ["gui: refresh the model list", "gateway: fix a disconnected stream"],
                [notes.edge],
            )
        self.assertEqual(result, ["界面：刷新模型列表", "网关：修复连接中断"])
        self.assertEqual(calls, [
            ["gui: refresh the model list", "gateway: fix a disconnected stream"],
            ["gateway: fix a disconnected stream"],
        ])

    def test_chinese_and_protected_literals_are_kept(self):
        subjects = ["修复模型名称", "gui: refresh `model_catalog_json` (未寒 on Discord)"]

        def provider(texts):
            self.assertEqual(len(texts), 1)
            return [texts[0].replace("gui: refresh", "界面：刷新")]

        self.assertEqual(
            notes.translate(subjects, [provider]),
            ["修复模型名称", "界面：刷新 `model_catalog_json` (未寒 on Discord)"],
        )

    def test_changed_literal_falls_back(self):
        def broken(texts):
            return ["界面：刷新模型列表"]

        def good(texts):
            return [texts[0].replace("gui: refresh", "界面：刷新")]

        with redirect_stderr(io.StringIO()):
            result = notes.translate(["gui: refresh `models_cache.json`"], [broken, good])
        self.assertEqual(result, ["界面：刷新 `models_cache.json`"])

    def test_provider_error_does_not_leak_details(self):
        secret = "test-secret-not-for-logs"

        def broken(texts):
            raise OSError(secret)

        log = io.StringIO()
        with redirect_stderr(log), self.assertRaises(RuntimeError):
            notes.translate(["gui: refresh the list"], [broken])
        self.assertNotIn(secret, log.getvalue())

    def test_release_metadata_filtered_without_dropping_product_changes(self):
        subjects = [
            "发布 0.1.1326 版本",
            "同步上游 0.1.1485 版本",
            "同步上游并发布 0.1.1294",
            "更新 Homebrew 发布至 v0.1.1294",
            "gui: fix the release notes button",
            "修复发布流程的翻译",
            "同步上游文件的路径检查",
        ]
        self.assertEqual(notes.change_subjects(subjects), subjects[4:])

    def test_context_batches_are_not_limited_to_five(self):
        subjects = ["gui: refresh the list"] * 80
        batches = list(notes.zhipu_context_batches(subjects, "glm-4.7-flash"))
        self.assertEqual(sum(map(len, batches)), 80)
        self.assertGreater(len(batches[0]), 5)

    def test_publication_added_after_release_and_idempotent(self):
        body = notes.render(["gui: refresh the list"], ["界面：刷新列表"],
                            "owner/repo", "0.1.1485")
        sha = "a" * 40
        updated = notes.publication_notes(body, "owner/repo", "0.1.1485", sha)
        self.assertIn(f"- [Release 0.1.1485](https://github.com/owner/repo/commit/{sha})", updated)
        self.assertIn(f"- [发布 0.1.1485 版本](https://github.com/owner/repo/commit/{sha})", updated)
        self.assertEqual(notes.publication_notes(updated, "owner/repo", "0.1.1485", sha),
                         updated)
        changed = notes.publication_notes(updated, "owner/repo", "0.1.1485", "b" * 40)
        self.assertEqual(changed.count("- [发布 0.1.1485 版本]"), 1)
        self.assertNotIn(sha, changed)

    def test_publication_does_not_accept_missing_chinese_notes(self):
        with self.assertRaises(ValueError):
            notes.publication_notes("### Changes\n\n- test\n", "owner/repo",
                                    "0.1.1485", "a" * 40)

    def test_google_ai_discards_thought_parts(self):
        translated = "界面：刷新模型列表"
        response = {"candidates": [{
            "finishReason": "STOP",
            "content": {"parts": [
                {"thought": True, "text": "I will translate the subjects."},
                {"text": json.dumps({"translations": [translated]})},
            ]},
        }]}
        with patch.dict("os.environ", {"GOOGLE_AI_API_KEY": "test-key"}), \
                patch.object(notes, "request_json", return_value=response):
            self.assertEqual(notes.google_ai_once(["gui: refresh the list"], "test-model"),
                             [translated])

    def test_google_ai_truncated_output_is_rejected(self):
        response = {"candidates": [{
            "finishReason": "MAX_TOKENS",
            "content": {"parts": [{"text": '{"translations":["unfinished'}]},
        }]}
        with patch.dict("os.environ", {"GOOGLE_AI_API_KEY": "test-key"}), \
                patch.object(notes, "request_json", return_value=response), \
                self.assertRaises(ValueError):
            notes.google_ai_once(["gui: refresh the list"], "test-model")


if __name__ == "__main__":
    unittest.main()
