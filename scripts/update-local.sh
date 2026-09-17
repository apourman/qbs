#!/bin/sh
set -eu

repo_root=$(git rev-parse --show-toplevel)
branch=$(git -C "$repo_root" branch --show-current)
target_branch=${QBS_UPDATE_BRANCH:-main}

if [ "$branch" != "$target_branch" ]; then
	printf 'skipping local QBS update: current branch is %s, target branch is %s\n' "$branch" "$target_branch"
	exit 0
fi

prefix=${PREFIX:-${HOME}/.local}
export PATH="$prefix/bin:${HOME}/.local/bin:$PATH"

printf 'updating local QBS from %s\n' "$repo_root"
make -C "$repo_root" install PREFIX="$prefix"

qbs_bin="$prefix/bin/qbs"
if [ -x "$qbs_bin" ]; then
	"$qbs_bin" skills sync
fi

printf 'local QBS update complete: %s\n' "$qbs_bin"
