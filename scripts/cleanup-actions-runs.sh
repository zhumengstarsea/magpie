#!/usr/bin/env bash
# Clean no-update checks and expired auxiliary CI records, preserving releases.
set -euo pipefail
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
dry_run=false
case "${1:-}" in
  '') ;;
  --dry-run) dry_run=true ;;
  *) echo "Usage: cleanup-actions-runs.sh [--dry-run]" >&2; exit 1 ;;
esac
delete_run() {
  if [ "$dry_run" = true ]; then
    echo "Would delete run $1: $2"
  else
    gh api --method DELETE "repos/$GITHUB_REPOSITORY/actions/runs/$1"
    echo "已清理运行记录 $1：$2"
  fi
}

gh api --paginate "repos/$GITHUB_REPOSITORY/actions/workflows/release.yml/runs?status=completed&per_page=100" \
  --jq '.workflow_runs[] | select(.status == "completed" and .conclusion == "success") | .id' > "$RUNNER_TEMP/no-update-candidates.txt"
while IFS= read -r run_id; do
  [[ "$run_id" =~ ^[0-9]+$ ]] || { echo "Invalid run ID" >&2; exit 1; }
  [ "$run_id" != "${GITHUB_RUN_ID:-}" ] || continue
  marker=$(gh api --paginate "repos/$GITHUB_REPOSITORY/actions/runs/$run_id/jobs?per_page=100" \
    --jq '.jobs[].steps[]? | select((.name == "No upstream update" or .name == "未发现上游更新") and .conclusion == "success") | .number')
  if [ -n "$marker" ]; then
    delete_run "$run_id" '未发现上游更新'
  fi
done < "$RUNNER_TEMP/no-update-candidates.txt"

# Test/Docker/privacy-filter runs come from upstream branch synchronization.
# Keep successful auxiliary runs for 7 days and unsuccessful ones for 30 days.
gh api --paginate "repos/$GITHUB_REPOSITORY/actions/runs?status=completed&per_page=100" \
  --jq '.workflow_runs[] | select(.status == "completed") | select(.name == "Test" or .name == "代码测试" or .name == "Docker" or .name == "Docker 镜像构建" or .name == "privacy-filter" or .name == "隐私过滤检查") | [.id, .conclusion, .updated_at] | @tsv' > "$RUNNER_TEMP/auxiliary-run-candidates.txt"
now=$(date -u +%s)
while IFS=$'\t' read -r run_id conclusion finished; do
  [[ "$run_id" =~ ^[0-9]+$ ]] || { echo "Invalid run ID" >&2; exit 1; }
  [ "$run_id" != "${GITHUB_RUN_ID:-}" ] || continue
  case "$conclusion" in
    success) keep_days=7 ;;
    failure|cancelled|timed_out|action_required|neutral|skipped|stale|startup_failure) keep_days=30 ;;
    *) continue ;;
  esac
  finished_epoch=$(date -u -d "$finished" +%s)
  if [ "$((now - finished_epoch))" -ge "$((keep_days * 86400))" ]; then
    delete_run "$run_id" "辅助检查记录已超过 $keep_days 天"
  fi
done < "$RUNNER_TEMP/auxiliary-run-candidates.txt"
