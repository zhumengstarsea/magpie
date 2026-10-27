#!/usr/bin/env bash
# Keep Actions history short and disable scheduled sync after repeated failures.
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

# Count the current scheduled sync failure together with the immediately
# preceding scheduled sync failures. Manual runs and successful runs reset the
# streak. Stop the schedule after three consecutive failures.
if [ "${GITHUB_EVENT_NAME:-}" = schedule ] && [ "${CURRENT_SYNC_OUTCOME:-}" = failure ]; then
  consecutive=1
  gh api --paginate "repos/$GITHUB_REPOSITORY/actions/workflows/release.yml/runs?event=schedule&branch=master&status=completed&per_page=100" \
    --jq '.workflow_runs[] | [.id, .conclusion, .updated_at] | @tsv' > "$RUNNER_TEMP/scheduled-sync-history.txt"
  while IFS=$'\t' read -r run_id conclusion updated; do
    [ "$run_id" != "${GITHUB_RUN_ID:-}" ] || continue
    if [ "$conclusion" = failure ]; then
      consecutive=$((consecutive + 1))
    else
      break
    fi
  done < "$RUNNER_TEMP/scheduled-sync-history.txt"
  if [ "$consecutive" -ge 3 ]; then
    if [ "$dry_run" = true ]; then
      echo "Would disable scheduled release workflow after $consecutive consecutive failures"
    else
      gh api --method PUT "repos/$GITHUB_REPOSITORY/actions/workflows/release.yml/disable"
      echo "连续 $consecutive 次定时同步失败，已停止发布工作流"
    fi
  fi
fi

# Delete completed successful and skipped runs at this scan. Keep failed or
# cancelled runs for one day. GitHub also deletes a run's Actions artifacts;
# the published GitHub Release and its assets remain available.
gh api --paginate "repos/$GITHUB_REPOSITORY/actions/runs?status=completed&per_page=100" \
  --jq '.workflow_runs[] | select(.status == "completed") | [.id, .conclusion, .updated_at] | @tsv' > "$RUNNER_TEMP/action-run-candidates.txt"
now=$(date -u +%s)
while IFS=$'\t' read -r run_id conclusion finished; do
  [[ "$run_id" =~ ^[0-9]+$ ]] || { echo "Invalid run ID" >&2; exit 1; }
  [ "$run_id" != "${GITHUB_RUN_ID:-}" ] || continue
  case "$conclusion" in
    success|skipped) keep_seconds=0; label='成功或跳过记录' ;;
    failure|cancelled|timed_out|action_required|neutral|stale|startup_failure) keep_seconds=$((24 * 3600)); label='失败记录超过 1 天' ;;
    *) continue ;;
  esac
  finished_epoch=$(date -u -d "$finished" +%s)
  if [ "$((now - finished_epoch))" -ge "$keep_seconds" ]; then
    delete_run "$run_id" "$label"
  fi
done < "$RUNNER_TEMP/action-run-candidates.txt"
