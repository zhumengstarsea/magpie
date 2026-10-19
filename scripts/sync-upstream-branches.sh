#!/usr/bin/env bash
# Mirror upstream branches, keeping the customized master and fork-only branches.
set -euo pipefail
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
upstream=https://github.com/yetone/magpie.git
git ls-remote --heads "$upstream" > "$RUNNER_TEMP/upstream-heads.txt"
git ls-remote --heads origin > "$RUNNER_TEMP/fork-heads.txt"
changed=false
while read -r upstream_sha ref; do
  branch=${ref#refs/heads/}
  [ "$branch" != master ] || continue
  fork_sha=$(awk -v ref="$ref" '$2 == ref { print $1 }' "$RUNNER_TEMP/fork-heads.txt")
  [ "$upstream_sha" != "$fork_sha" ] || continue
  : "${SYNC_TOKEN:?Configure SYNC_TOKEN for branch synchronization}"
  tracking_ref="refs/remotes/upstream/branches/$branch"
  git fetch --no-tags "$upstream" "+$ref:$tracking_ref"
  # Fail if the destination changed since the snapshot; never overwrite a race.
  git push "--force-with-lease=$ref:$fork_sha" origin "$tracking_ref:$ref"
  changed=true
  echo "已同步上游分支：$branch" | tee -a "${GITHUB_STEP_SUMMARY:-/dev/null}"
done < "$RUNNER_TEMP/upstream-heads.txt"
if [ -n "${GITHUB_OUTPUT:-}" ]; then
  printf 'changed=%s\n' "$changed" >> "$GITHUB_OUTPUT"
fi
