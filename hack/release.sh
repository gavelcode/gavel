#!/usr/bin/env bash
# Safe release driver for the gavel CLI.
#
#   make release VERSION=0.1.1
#
# All checks run locally and fail fast. The single irreversible public action
# (pushing the tag, which triggers the release workflow) happens last, only
# after every check passes and you confirm. The git tag is the single source of
# truth for the CLI version (injected into the binary via ldflags), so this
# script edits no files — there is no version to forget to bump.
#
# Set RELEASE_YES=1 to skip the confirmation prompt (e.g. from automation).
set -euo pipefail

REPO="gavelcode/gavel"
CI_WORKFLOW="ci.yml"

die() {
	echo "release: $*" >&2
	exit 1
}

VERSION="${1:-}"
[ -n "$VERSION" ] || die "usage: make release VERSION=X.Y.Z"

# 1. Version format.
echo "$VERSION" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$' ||
	die "VERSION must be semver X.Y.Z (got '$VERSION')"
TAG="v$VERSION"

# 2. Required tooling.
command -v git >/dev/null || die "git is required"
command -v gh >/dev/null || die "gh (GitHub CLI) is required"
command -v goreleaser >/dev/null || die "goreleaser is required (brew install goreleaser)"

# 3. On the default branch.
branch="$(git rev-parse --abbrev-ref HEAD)"
[ "$branch" = "main" ] || die "must release from 'main' (on '$branch')"

# 4. Clean working tree.
git diff --quiet && git diff --cached --quiet ||
	die "working tree is dirty — commit or stash first"

# 5. In sync with origin/main (tagging exactly what is published).
git fetch --quiet origin main
local_sha="$(git rev-parse @)"
remote_sha="$(git rev-parse '@{u}')"
[ "$local_sha" = "$remote_sha" ] ||
	die "local main is not in sync with origin/main — push or pull first"

# 6. Tag must not already exist.
git rev-parse -q --verify "refs/tags/$TAG" >/dev/null 2>&1 &&
	die "tag $TAG already exists locally"
git ls-remote --exit-code --tags origin "$TAG" >/dev/null 2>&1 &&
	die "tag $TAG already exists on origin"

# 7. CI for this exact commit must be green — never release an unverified commit.
echo "release: checking CI status for ${local_sha} ..."
ci_status="$(gh run list -R "$REPO" --workflow "$CI_WORKFLOW" --commit "$local_sha" \
	--json conclusion,status --jq '.[0] | "\(.status)/\(.conclusion)"' 2>/dev/null || true)"
case "$ci_status" in
	completed/success) ;;
	"" | "/") die "no completed CI run found for ${local_sha} — wait for CI to finish" ;;
	*) die "CI for ${local_sha} is '${ci_status}' — must be completed/success" ;;
esac

# 8. goreleaser config valid.
goreleaser check

# 9. Confirm the single irreversible step.
echo
echo "release: ready to tag ${TAG} at ${local_sha}"
echo "release:   branch=main  tree=clean  origin=in-sync  ci=green  goreleaser=ok"
if [ "${RELEASE_YES:-}" != "1" ]; then
	printf "release: push %s and trigger the release workflow? [y/N] " "$TAG"
	read -r reply
	case "$reply" in
		y | Y) ;;
		*) die "aborted" ;;
	esac
fi

# 10. The only public action.
git tag -a "$TAG" -m "Release ${TAG}"
git push origin "$TAG"
echo "release: pushed ${TAG}"
echo "release: watch it with  gh run watch -R ${REPO} --workflow release.yml"
