#!/bin/sh
set -eu

repo_root=$(git rev-parse --show-toplevel)
expected=${1:-}
actual=$(tr -d '[:space:]' < "$repo_root/VERSION")

if [ -z "$expected" ]; then
  printf 'usage: %s VERSION\n' "$0" >&2
  exit 2
fi

case "$expected" in
  v*) expected=${expected#v} ;;
esac

if [ "$actual" != "$expected" ]; then
  printf 'VERSION (%s) does not match release version (%s)\n' "$actual" "$expected" >&2
  exit 1
fi

printf 'release version %s matches VERSION\n' "$expected"
