#!/usr/bin/env bash
set -euo pipefail

TAG="${1:-${TAG:-}}"
REMOTE="${REMOTE:-origin}"

usage() {
  cat <<'EOF'
Usage:
  scripts/release.sh <tag>

Or:
  TAG=<tag> scripts/release.sh

Optional:
  REMOTE=<remote> scripts/release.sh <tag>

Examples:
  scripts/release.sh v0.0.1
  TAG=v0.0.1 scripts/release.sh
  REMOTE=upstream scripts/release.sh v0.0.1
EOF
}

if [[ -z "${TAG}" ]]; then
  echo "Error: missing tag." >&2
  usage >&2
  exit 2
fi

if ! [[ "${TAG}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z]+)*$ ]]; then
  echo "Error: tag must look like a semver tag starting with 'v' (e.g. v0.0.1). Got: ${TAG}" >&2
  exit 2
fi

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Error: not inside a git repository." >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Error: working tree is not clean. Commit or stash changes before releasing." >&2
  git status --porcelain >&2
  exit 1
fi

if git rev-parse -q --verify "refs/tags/${TAG}" >/dev/null; then
  echo "Error: tag ${TAG} already exists." >&2
  exit 1
fi

echo "Creating annotated tag ${TAG}..."
git tag -a "${TAG}" -m "Release ${TAG}"

echo "Pushing tag ${TAG} to ${REMOTE}..."
git push "${REMOTE}" "${TAG}"

echo "Done. GitHub Actions should build and publish release assets for ${TAG}."
