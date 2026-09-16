#!/usr/bin/env bash
# Fails when deployment-private data leaks into the tree: host names of a private
# stand, personal identifiers, private-range addresses, credential-looking blobs
# or local artifacts that should never be committed.
#
# Product branding is a separate concern with its own gate: scripts/check-no-brand.sh.
#
# The pattern scan runs over *our* delta (the lines this fork adds on top of
# upstream/master) when that ref exists, and over the whole tracked tree
# otherwise; upstream's own code is never our finding.
#
# Usage:
#   scripts/check-no-private.sh
#   PRIVATE_PATTERNS='acme\.internal|jenkins-01' scripts/check-no-private.sh
set -euo pipefail

cd "$(dirname "$0")/.."

# Host names and personal identifiers of the deployments that use this fork, plus
# credential shapes. Extend when a new deployment starts using the fork.
PRIVATE_PATTERNS="${PRIVATE_PATTERNS:-corprightline|demid-auth|maxs\.pro|vitalik|backend_supergraph|harness-(platform|external|staging)}"
# Private-range addresses and credential blobs; these are never legitimate here.
SECRET_PATTERNS='(^|[^0-9.])(10\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}|192\.168\.[0-9]{1,3}\.[0-9]{1,3}|172\.(1[6-9]|2[0-9]|3[01])\.[0-9]{1,3}\.[0-9]{1,3})|-----BEGIN [A-Z ]*PRIVATE KEY-----|ghp_[A-Za-z0-9]{30,}|xox[baprs]-[A-Za-z0-9-]{10,}|AKIA[0-9A-Z]{16}'

# This script carries the patterns, so it is never its own finding.
self='scripts/check-no-private.sh'

if git rev-parse --verify --quiet upstream/master >/dev/null; then
  scope="the delta against upstream/master"
  added=$(git diff upstream/master...HEAD -- . ":!$self" | grep '^+' | grep -v '^+++' || true)
else
  scope="the tracked tree"
  added=$(git grep -I -n -E '' -- . ":!$self" || true)
fi

findings=$(printf '%s\n' "$added" | grep -I -i -E "$PRIVATE_PATTERNS" || true)
findings="$findings$(printf '%s\n' "$added" | grep -I -E "$SECRET_PATTERNS" || true)"

if [ -n "$(printf '%s' "$findings" | tr -d '[:space:]')" ]; then
  echo "check-no-private: $scope carries deployment-private data:" >&2
  printf '%s\n' "$findings" | grep -v '^$' | head -40 >&2
  echo "(move it into the deployment's environment; see docs/branding.md for how branding does it)" >&2
  exit 1
fi

# Local artifacts that must never be committed, whatever they contain. Only files
# this fork adds count: upstream's own fixtures (test certificates and keys) are
# upstream's business.
if git rev-parse --verify --quiet upstream/master >/dev/null; then
  tracked=$(git diff upstream/master...HEAD --diff-filter=A --name-only)
else
  tracked=$(git ls-files)
fi
artifacts=$(printf '%s\n' "$tracked" | grep -E '(^|/)(\.env(\..*)?|bun\.lock|bun\.lockb|.*\.log|.*\.orig|.*\.bak|.*\.pem|.*\.key)$' || true)
if [ -n "$artifacts" ]; then
  echo "check-no-private: local artifacts are committed:" >&2
  printf '%s\n' "$artifacts" >&2
  exit 1
fi

echo "check-no-private: OK, no private data in $scope"
