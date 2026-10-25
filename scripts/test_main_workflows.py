#!/usr/bin/env python3
"""Check main's trigger override and the isolated Git-tree transformation."""

import importlib.util
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

import yaml

SCRIPT = Path(__file__).with_name("guard-main-workflows.py")
spec = importlib.util.spec_from_file_location("main_workflows", SCRIPT)
guard = importlib.util.module_from_spec(spec)
spec.loader.exec_module(guard)


def events(text):
    return yaml.load(text, Loader=yaml.BaseLoader)["on"]


class MainWorkflowTests(unittest.TestCase):
    def check_trigger(self, before):
        after = guard.exclude_main(before)
        push = events(after)["push"]
        if "branches" in push:
            self.assertEqual(push["branches"][-1], "!main")
        else:
            self.assertIn("main", push["branches-ignore"])
        self.assertEqual(guard.exclude_main(after), after)
        return after

    def test_main_only_and_existing_paths(self):
        before = "name: Test\non:\n  push:\n    branches: [main]\n    paths: ['src/**']\n  workflow_dispatch:\njobs: {}\n"
        after = self.check_trigger(before)
        self.assertEqual(events(after)["push"]["paths"], ["src/**"])
        self.assertIn("workflow_dispatch", events(after))
        self.assertTrue(after.startswith("name: Test\n"))
        self.assertTrue(after.endswith("jobs: {}\n"))

    def test_block_lists_and_tags_keep_their_boundaries(self):
        before = "on:\n  push:\n    branches:\n      - main\n      - pf/**\n    tags:\n      - pf-v*\n  pull_request:\n    paths: [src]\njobs: {}\n"
        after = self.check_trigger(before)
        self.assertEqual(events(after)["push"]["branches"], ["main", "pf/**", "!main"])
        self.assertEqual(events(after)["push"]["tags"], ["pf-v*"])
        self.assertEqual(events(after)["pull_request"], {"paths": ["src"]})

    def test_tags_without_branch_filter(self):
        before = "on:\n  push:\n    tags: ['v*']\n  workflow_dispatch:\njobs: {}\n"
        after = self.check_trigger(before)
        self.assertEqual(events(after)["push"]["tags"], ["v*"])

    def test_existing_exclusions(self):
        self.check_trigger("on:\n  push:\n    branches-ignore: [other]\njobs: {}\n")

    def test_negative_rule_must_come_last(self):
        before = "on:\n  push:\n    branches: [main, '!main', '**']\njobs: {}\n"
        self.check_trigger(before)

    def test_null_push(self):
        for value in ("", "null", "~"):
            with self.subTest(value=value):
                self.check_trigger(f"on:\n  push: {value}\n  workflow_dispatch:\njobs: {{}}\n")

    def test_short_event_forms(self):
        for trigger in ("push", "[push, pull_request]", "[push]"):
            with self.subTest(trigger=trigger):
                self.check_trigger(f"on: {trigger}\njobs: {{}}\n")
        self.check_trigger("on:\n  - push\n  - pull_request\njobs: {}\n")

    def test_flow_mapping(self):
        self.check_trigger("on: {push: {tags: ['v*']}, workflow_dispatch: null}\njobs: {}\n")
        self.check_trigger("on: {push: {branches: [main]}, workflow_dispatch: null}\njobs: {}\n")
        self.check_trigger("on: {push: null, workflow_dispatch: null}\njobs: {}\n")

    def test_non_push_workflows_are_unchanged(self):
        for before in (
            "on:\n  workflow_dispatch:\njobs: {}\n",
            "on: pull_request\njobs: {}\n",
            "on: [pull_request, workflow_dispatch]\njobs: {}\n",
        ):
            self.assertEqual(guard.exclude_main(before), before)

    def test_invalid_filters_fail_closed(self):
        with self.assertRaises(ValueError):
            guard.exclude_main("on:\n  push:\n    branches: [main]\n    branches-ignore: [other]\n")
        with self.assertRaises(yaml.YAMLError):
            guard.exclude_main("on: [push\n")

    def test_tree_changes_only_workflows_without_touching_checkout(self):
        with tempfile.TemporaryDirectory(prefix="main-workflows-test-") as directory:
            root = Path(directory)
            env = dict(os.environ, GIT_CONFIG_NOSYSTEM="1", GIT_CONFIG_GLOBAL=os.devnull)

            def git(*args, data=None):
                return subprocess.check_output(["git", *args], cwd=root, env=env, input=data)

            git("init", "-q", "-b", "main")
            git("config", "user.name", "Workflow test")
            git("config", "user.email", "workflow@example.invalid")
            workflows = root / ".github/workflows"
            workflows.mkdir(parents=True)
            original = "name: Test\non:\n  push:\n    branches: [main]\n  workflow_dispatch:\njobs: {}\n"
            (workflows / "test.yml").write_text(original, encoding="utf-8")
            (root / "source.txt").write_text("upstream source\n", encoding="utf-8")
            git("add", ".")
            git("-c", "commit.gpgsign=false", "commit", "-qm", "upstream")
            tip = git("rev-parse", "HEAD").decode().strip()
            source = git("show", "HEAD:source.txt")
            # Keep a dirty tracked file and a separately staged file intact.
            (root / "source.txt").write_text("local work\n", encoding="utf-8")
            (root / "staged.txt").write_text("staged work\n", encoding="utf-8")
            git("add", "staged.txt")
            status = git("status", "--porcelain")
            index_tree = git("write-tree")
            tree = subprocess.check_output(
                [os.sys.executable, str(SCRIPT), tip], cwd=root, env=env
            ).decode().strip()
            changed = git("diff-tree", "--no-commit-id", "--name-only", "-r", tip, tree)
            self.assertEqual(changed.decode().strip(), ".github/workflows/test.yml")
            self.assertEqual(git("show", f"{tree}:source.txt"), source)
            self.assertIn("!main", events(git("show", f"{tree}:.github/workflows/test.yml"))["push"]["branches"])
            self.assertEqual(git("rev-parse", "HEAD").decode().strip(), tip)
            self.assertEqual(git("write-tree"), index_tree)
            self.assertEqual(git("status", "--porcelain"), status)
            self.assertEqual((workflows / "test.yml").read_text(encoding="utf-8"), original)
            self.assertEqual((root / "source.txt").read_text(encoding="utf-8"), "local work\n")


if __name__ == "__main__":
    unittest.main()
