#!/usr/bin/env bash
# Mirror upstream branches; main alone keeps a signed workflow-trigger override.
set -euo pipefail
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
upstream=https://github.com/yetone/magpie.git
git ls-remote --heads "$upstream" > "$RUNNER_TEMP/upstream-heads.txt"
git ls-remote --heads origin > "$RUNNER_TEMP/fork-heads.txt"
changed=false
while read -r upstream_sha ref; do
  branch=${ref#refs/heads/}
  [ "$branch" != master ] || continue
  fork_sha=$(awk -v ref="$ref" '$2 == ref { print $1 }' "$RUNNER_TEMP/fork-heads.txt")
  if [ "$branch" != main ] && [ "$upstream_sha" = "$fork_sha" ]; then continue; fi
  : "${SYNC_TOKEN:?Configure SYNC_TOKEN for branch synchronization}"
  tracking_ref="refs/remotes/upstream/branches/$branch"
  git fetch --no-tags "$upstream" "+$ref:$tracking_ref"
  [ "$(git rev-parse "$tracking_ref")" = "$upstream_sha" ] || {
    echo "::error::Upstream $branch changed during synchronization; retry"; exit 1;
  }
  push_ref=$tracking_ref
  if [ "$branch" = main ]; then
    main_tree=$(python3 "$script_dir/guard-main-workflows.py" "$tracking_ref")
    if [ "$main_tree" != "$(git rev-parse "$tracking_ref^{tree}")" ]; then
      bash "$script_dir/configure-release-gpg.sh"
      if [ -n "$fork_sha" ]; then
        git fetch --no-tags origin "+$ref:refs/remotes/origin/main"
        # Reuse the existing signed override when upstream and its tree match.
        if [ "$(git rev-parse "$fork_sha^{tree}")" = "$main_tree" ] &&
           [ "$(git rev-list --parents -n 1 "$fork_sha")" = "$fork_sha $upstream_sha" ]; then
          git verify-commit "$fork_sha"
          continue
        fi
      fi
      push_ref=$(git commit-tree -S "$main_tree" -p "$upstream_sha" -m "禁止 main 分支自动触发工作流")
      git verify-commit "$push_ref"
    fi
  fi
  [ "$(git rev-parse "$push_ref")" != "$fork_sha" ] || continue
  # Fail if the destination changed since the snapshot; never overwrite a race.
  git push "--force-with-lease=$ref:$fork_sha" origin "$push_ref:$ref"
  changed=true
  echo "已同步上游分支：$branch" | tee -a "${GITHUB_STEP_SUMMARY:-/dev/null}"
done < "$RUNNER_TEMP/upstream-heads.txt"
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  printf 'changed=%s\n' "$changed" >> "$GITHUB_OUTPUT"
fi
