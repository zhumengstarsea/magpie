#!/usr/bin/env bash
# Mirror moepie (except its up) and yetone/main as up; master stays custom.
set -euo pipefail
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
git ls-remote --heads origin > "$RUNNER_TEMP/fork-heads.txt"
changed=false
sync_branch() {
  local source=$1 source_branch=$2 branch=$3 upstream_sha=$4
  local ref="refs/heads/$branch" fork_sha tracking_ref push_ref guarded
  [ "$branch" != master ] || { echo "::error::master is not a mirror destination"; exit 1; }
  fork_sha=$(awk -v ref="$ref" '$2 == ref { print $1 }' "$RUNNER_TEMP/fork-heads.txt")
  : "${SYNC_TOKEN:?Configure SYNC_TOKEN for branch synchronization}"
  tracking_ref="refs/remotes/branch-sync/$branch"
  git fetch --no-tags "$source" "+refs/heads/$source_branch:$tracking_ref"
  [ "$(git rev-parse "$tracking_ref")" = "$upstream_sha" ] || {
    echo "::error::Source $source_branch changed during synchronization; retry"; exit 1;
  }
  push_ref=$tracking_ref
  guarded=$(python3 "$script_dir/guard-main-workflows.py" --manual-only "$tracking_ref")
  if [ "$guarded" != "$(git rev-parse "$tracking_ref^{tree}")" ]; then
    bash "$script_dir/configure-release-gpg.sh"
    if [ -n "$fork_sha" ]; then
      git fetch --no-tags origin "+$ref:refs/remotes/origin/$branch"
      if [ "$(git rev-parse "$fork_sha^{tree}")" = "$guarded" ] &&
         [ "$(git rev-list --parents -n 1 "$fork_sha")" = "$fork_sha $upstream_sha" ]; then
        git verify-commit "$fork_sha"
        return
      fi
    fi
    push_ref=$(git commit-tree -S "$guarded" -p "$upstream_sha" -m "暂停 $branch 分支自动工作流")
    git verify-commit "$push_ref"
  fi
  [ "$(git rev-parse "$push_ref")" != "$fork_sha" ] || return
  git push "--force-with-lease=$ref:$fork_sha" origin "$push_ref:$ref"
  changed=true
  echo "已同步分支：$source/$source_branch → $branch" | tee -a "${GITHUB_STEP_SUMMARY:-/dev/null}"
}

moepie=https://github.com/lingqiqi5211/moepie.git
git ls-remote --heads "$moepie" > "$RUNNER_TEMP/moepie-heads.txt"
while read -r upstream_sha ref; do
  branch=${ref#refs/heads/}
  case "$branch" in up|master) continue ;; esac
  sync_branch "$moepie" "$branch" "$branch" "$upstream_sha"
done < "$RUNNER_TEMP/moepie-heads.txt"

yetone=https://github.com/yetone/magpie.git
git ls-remote --heads "$yetone" refs/heads/main > "$RUNNER_TEMP/yetone-main.txt"
read -r upstream_sha ref < "$RUNNER_TEMP/yetone-main.txt"
[ "$ref" = refs/heads/main ] || { echo "::error::yetone/main was not found"; exit 1; }
sync_branch "$yetone" main up "$upstream_sha"
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  printf 'changed=%s\n' "$changed" >> "$GITHUB_OUTPUT"
fi
