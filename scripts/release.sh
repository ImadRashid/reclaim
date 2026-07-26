#!/usr/bin/env bash
# scripts/release.sh — cut a release end to end.
#
#   scripts/release.sh v0.1.2
#
# What it does:
#   1. Verifies you're on main with a clean tree, up to date with origin.
#   2. Renames the "## Unreleased" CHANGELOG section to the new version
#      (with today's date) and commits it.
#   3. Tags and pushes. The release workflow then builds the binaries,
#      publishes the GitHub Release, and pushes the updated formula to
#      ImadRashid/homebrew-tap (needs the TAP_GITHUB_TOKEN repo secret).
set -euo pipefail

usage() {
  echo "usage: scripts/release.sh vX.Y.Z" >&2
  exit 1
}

[ $# -eq 1 ] || usage
ver="$1"
[[ "$ver" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || usage

cd "$(git rev-parse --show-toplevel)"

branch="$(git rev-parse --abbrev-ref HEAD)"
if [ "$branch" != "main" ]; then
  echo "release: must be on main (currently on $branch)" >&2
  exit 1
fi
if [ -n "$(git status --porcelain)" ]; then
  echo "release: working tree not clean — commit or stash first" >&2
  exit 1
fi
git pull --ff-only
if git rev-parse -q --verify "refs/tags/$ver" >/dev/null; then
  echo "release: tag $ver already exists" >&2
  exit 1
fi

# Rename the Unreleased section so release.yml can extract the notes, bump
# the versioned asset names in the README download commands, and sync the
# fallback version string used by source builds.
if grep -q '^## Unreleased' CHANGELOG.md; then
  today="$(date +%Y-%m-%d)"
  sed -i '' "s/^## Unreleased/## $ver — $today/" CHANGELOG.md
  sed -i '' "s/reclaim_v[0-9][0-9.]*_/reclaim_${ver}_/g" README.md
  sed -i '' "s/^var version = \".*\"/var version = \"${ver#v}\"/" main.go
  git add CHANGELOG.md README.md main.go
  git commit -m "Release $ver"
elif ! grep -q "^## $ver" CHANGELOG.md; then
  echo "release: CHANGELOG.md has no '## Unreleased' or '## $ver' section" >&2
  exit 1
fi

git tag "$ver" -m "$ver"
git push origin main "$ver"

echo
echo "✓ Pushed $ver. The release workflow is now building:"
echo "  https://github.com/ImadRashid/reclaim/actions/workflows/release.yml"
echo "It will publish the GitHub Release and update the Homebrew tap."
