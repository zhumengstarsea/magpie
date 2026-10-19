#!/usr/bin/env bash
# Follow published upstream versions, preserving this fork's commits.
set -euo pipefail
: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
: "${RUNNER_TEMP:?RUNNER_TEMP is required}"
: "${TARGET_BRANCH:=master}"
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
say() { echo "$*"; echo "$*" >> "${GITHUB_STEP_SUMMARY:-/dev/null}"; }
die() { echo "::error::$*"; exit 1; }
output() { if [ -n "${GITHUB_OUTPUT:-}" ]; then printf '%s=%s\n' "$1" "$2" >> "$GITHUB_OUTPUT"; fi; }

gh api repos/yetone/magpie-releases/releases/latest > "$RUNNER_TEMP/upstream-release.json"
tag=$(jq -er '.tag_name' "$RUNNER_TEMP/upstream-release.json")
[[ "$tag" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]] || die "Unexpected upstream version: $tag"
jq -e '.draft == false and .prerelease == false' "$RUNNER_TEMP/upstream-release.json" >/dev/null
version="${tag#v}"

gh api "repos/$GITHUB_REPOSITORY/releases?per_page=100" > "$RUNNER_TEMP/fork-releases.json"
latest=$(jq -r '.[] | select(.draft == false and .prerelease == false) | .tag_name | select(test("^v[0-9]+\\.[0-9]+\\.[0-9]+$")) | ltrimstr("v")' "$RUNNER_TEMP/fork-releases.json" | sort -V | tail -1)
if [ -n "$latest" ] && { [ "$latest" = "$version" ] || [ "$(printf '%s\n' "$latest" "$version" | sort -V | tail -1)" != "$version" ]; }; then
  say "No new upstream release: upstream $tag, fork v$latest."
  output result no-update
  output build false
  exit 0
fi

# Only complete, public upstream releases are eligible.
assets=(SHA256SUMS magpie-cli-android-arm64)
for arch in amd64 arm64; do
  assets+=("magpie-darwin-$arch.dmg" "magpie-darwin-$arch.zip" "magpie-windows-$arch.exe" "magpie-linux-$arch")
  for os in darwin linux windows; do
    ext=""; [ "$os" != windows ] || ext=.exe
    assets+=("magpie-cli-$os-$arch$ext")
  done
done
for asset in "${assets[@]}"; do
  if ! jq -e --arg name "$asset" '.assets[] | select(.name == $name and .size > 0)' "$RUNNER_TEMP/upstream-release.json" >/dev/null; then
    say "Upstream $tag is still missing $asset; retry at the next check."
    output result incomplete
    output build false
    exit 0
  fi
done
: "${SYNC_TOKEN:?Configure SYNC_TOKEN with Contents, Workflows and Actions read/write access}"
: "${GPG_PRIVATE_KEY:?Configure GPG_PRIVATE_KEY for signed automatic commits}"
[ -z "$(git status --porcelain)" ] || die "Run only in a clean checkout"
[ "$(git branch --show-current)" = "$TARGET_BRANCH" ] || die "Check out $TARGET_BRANCH first"
git fetch --no-tags origin "refs/heads/$TARGET_BRANCH:refs/remotes/origin/$TARGET_BRANCH"
git merge --ff-only "refs/remotes/origin/$TARGET_BRANCH"
# Keep upstream tags out of refs/tags: this fork's vX.Y.Z tags refer to its builds.
upstream_ref="refs/remotes/upstream/releases/$tag"
git fetch --no-tags https://github.com/yetone/magpie.git "refs/tags/$tag:$upstream_ref"
upstream_commit=$(git rev-parse "$upstream_ref^{commit}")

remote_tag=$(git ls-remote --tags origin "refs/tags/$tag" "refs/tags/$tag^{}")
build_ref=""
if [ -n "$remote_tag" ]; then
  fork_tag_ref="refs/remotes/origin/releases/$tag"
  git fetch --no-tags origin "refs/tags/$tag:$fork_tag_ref"
  build_ref=$(git rev-parse "$fork_tag_ref^{commit}")
  [ "$build_ref" != "$upstream_commit" ] || die "$tag already points to an inherited upstream tag; resolve it manually before publishing this version"
  git merge-base --is-ancestor "$upstream_commit" "$build_ref" || die "$tag does not contain this upstream release"
  git merge-base --is-ancestor "$build_ref" HEAD || die "$tag is not part of $TARGET_BRANCH"
  say "Retrying unpublished fork tag $tag at $build_ref."
else
  if ! git merge-base --is-ancestor "$upstream_commit" HEAD; then
    bash "$script_dir/configure-release-gpg.sh"
    if ! git merge --no-ff --no-commit "$upstream_commit"; then
      git merge --abort
      die "Merge conflict with $tag; $TARGET_BRANCH was not pushed"
    fi
    git commit -S -m "同步上游 $version 版本"
    git verify-commit HEAD
    # A concurrent push fails normally; never force-update the user's branch.
    git push origin "HEAD:refs/heads/$TARGET_BRANCH"
  fi
  build_ref=$(git rev-parse HEAD)
fi

# Continue building inside this same workflow; no second dispatch/run is needed.
output ref "$build_ref"
output version "$version"
output result update
output build true
say "Prepared fork release $tag from $build_ref."
