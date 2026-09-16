#!/usr/bin/env bash
# Fails when a downstream brand leaks into the tree.
#
# The fork is meant to stay mergeable with upstream casdoor/casdoor: every product
# name, wordmark and marketing URL is deployment configuration (CASDOOR_BRAND_*,
# see docs/branding.md), never source. This is the CI gate for that rule — run it
# before pushing the fork's `develop` and before opening an upstream PR.
#
# Usage:
#   scripts/check-no-brand.sh                # check the tracked tree
#   BRAND_PATTERNS='acme|acme-logo' scripts/check-no-brand.sh
set -euo pipefail

cd "$(dirname "$0")/.."

# Extended regular expression, case-insensitive, matched against tracked files.
# Add a downstream brand here when a new deployment starts using this fork.
BRAND_PATTERNS="${BRAND_PATTERNS:-receipt.?hunter|receipt-hunter|rh-logo|receipt scanner}"

# The only file allowed to name a brand: this script, which carries the patterns.
ALLOWLIST_REGEX='^(scripts/check-no-brand\.sh)$'

matches=$(git grep --untracked -I -n -i -E "$BRAND_PATTERNS" -- . ':!*.bundle' || true)

if [ -n "$matches" ]; then
  matches=$(printf '%s\n' "$matches" | grep -vE "^($(echo "$ALLOWLIST_REGEX" | sed 's/^\^//; s/\$$//')):" || true)
fi

if [ -n "$matches" ]; then
  echo "check-no-brand: the tree mentions a downstream brand; move it into CASDOOR_BRAND_* (docs/branding.md):" >&2
  printf '%s\n' "$matches" >&2
  exit 1
fi

# Brand assets must not be committed either: they are copied in at image build
# time from BRAND_ASSETS_DIR (see Dockerfile).
assets=$(git ls-files 'web/public/brand/*' || true)
if [ -n "$assets" ]; then
  echo "check-no-brand: brand assets are committed, they belong to BRAND_ASSETS_DIR:" >&2
  printf '%s\n' "$assets" >&2
  exit 1
fi

echo "check-no-brand: OK, no downstream brand in the tree"
