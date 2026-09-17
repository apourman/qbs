#!/bin/sh
set -eu

cat <<'EOF' >&2
scripts/release.sh is deprecated. Releases are created by release-please from
Conventional Commits; use the generated release PR and tag instead.
EOF
exit 1
