#!/usr/bin/env python3
"""Build main's upstream tree with push triggers excluding main."""

import json
import os
from pathlib import PurePosixPath
import re
import subprocess
import sys
import tempfile

import yaml
from yaml.nodes import MappingNode, ScalarNode, SequenceNode


def pair(node, name):
    if isinstance(node, MappingNode):
        return next((value for key, value in node.value if key.value == name), None)
    return None


def values(node):
    if isinstance(node, SequenceNode):
        return [item.value for item in node.value]
    if isinstance(node, ScalarNode):
        return [node.value]
    raise ValueError("Expected a branch name or a list of branch names")


def replace_node(text, node, replacement):
    start, end = node.start_mark.index, node.end_mark.index
    whitespace = re.search(r"\s*$", text[start:end]).group()
    return text[:start] + replacement + whitespace + text[end:]


def exclude_main(text):
    root = yaml.compose(text, Loader=yaml.BaseLoader)
    events = pair(root, "on")
    if events is None:
        return text
    if not isinstance(events, MappingNode):
        names = values(events)
        if "push" not in names:
            return text
        mapping = {name: None for name in names}
        mapping["push"] = {"branches-ignore": ["main"]}
        replacement = yaml.safe_dump(mapping, sort_keys=False).rstrip()
        return replace_node(
            text, events, "\n" + "\n".join("  " + line for line in replacement.splitlines())
        )
    push = pair(events, "push")
    if push is None:
        return text
    if isinstance(push, MappingNode):
        branches = pair(push, "branches")
        ignored = pair(push, "branches-ignore")
        if branches is not None and ignored is not None:
            raise ValueError("A push cannot set both branches and branches-ignore")
        if branches is not None or ignored is not None:
            node = branches if branches is not None else ignored
            names = values(node)
            exclusion = "!main" if branches is not None else "main"
            # A last negative pattern wins over earlier positive patterns.
            if branches is not None:
                names = [name for name in names if name != exclusion] + [exclusion]
            elif exclusion not in names:
                names.append(exclusion)
            if names == values(node):
                return text
            return replace_node(text, node, json.dumps(names, ensure_ascii=False))
        if push.flow_style:
            mapping = yaml.load(
                text[push.start_mark.index : push.end_mark.index], Loader=yaml.BaseLoader
            )
            mapping["branches-ignore"] = ["main"]
            return replace_node(text, push, json.dumps(mapping, ensure_ascii=False))
        # Insert before the mapping's first key, keeping all tags/paths intact.
        mark = push.start_mark
        return (
            text[: mark.index]
            + "branches-ignore: [main]\n"
            + " " * mark.column
            + text[mark.index :]
        )
    if isinstance(push, ScalarNode) and push.style is None and push.value.lower() in ("", "null", "~"):
        if events.flow_style:
            return replace_node(text, push, "{branches-ignore: [main]}")
        # A null node has no text; put the mapping after the push key's line.
        return replace_node(text, push, "\n    branches-ignore: [main]")
    raise ValueError("Unsupported push trigger")


def git(*args, data=None, env=None):
    return subprocess.run(
        ["git", *args], input=data, env=env, check=True, stdout=subprocess.PIPE
    ).stdout


def guarded_tree(ref):
    entries = git("ls-tree", "-rz", ref, "--", ".github/workflows").split(b"\0")
    changes = []
    for entry in entries:
        if not entry:
            continue
        metadata, name = entry.split(b"\t", 1)
        mode, kind, oid = metadata.decode().split()
        path = name.decode("utf-8")
        if PurePosixPath(path).suffix not in (".yml", ".yaml"):
            continue
        if mode != "100644" or kind != "blob":
            raise ValueError(f"Workflow must be a regular file: {path}")
        original = git("cat-file", "blob", oid)
        try:
            updated = exclude_main(original.decode("utf-8")).encode("utf-8")
        except (ValueError, yaml.YAMLError) as error:
            raise ValueError(f"{path}: {error}") from error
        if updated != original:
            # Validate the result before it can be committed or pushed.
            yaml.compose(updated.decode("utf-8"), Loader=yaml.BaseLoader)
            changes.append((mode, path, updated))
    if not changes:
        return git("rev-parse", f"{ref}^{{tree}}").decode().strip()
    with tempfile.TemporaryDirectory(prefix="main-workflows-") as directory:
        env = dict(os.environ, GIT_INDEX_FILE=os.path.join(directory, "index"))
        git("read-tree", ref, env=env)
        for mode, path, updated in changes:
            oid = git("hash-object", "-w", "--stdin", data=updated).decode().strip()
            git("update-index", "--add", "--cacheinfo", mode, oid, path, env=env)
        return git("write-tree", env=env).decode().strip()


if __name__ == "__main__":
    if len(sys.argv) != 2:
        sys.exit("Usage: guard-main-workflows.py <upstream-ref>")
    print(guarded_tree(sys.argv[1]))
