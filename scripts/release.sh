#!/bin/sh
set -eu

repo_root=$(git rev-parse --show-toplevel)
version_file=${VERSION_FILE:-$repo_root/VERSION}
output=${OUTPUT:-$repo_root/qbs}
module='github.com/trues/qbs/internal/cli'

if [ ! -f "$version_file" ]; then
	printf 'version file not found: %s\n' "$version_file" >&2
	exit 1
fi

version=$(tr -d '[:space:]' < "$version_file")
case "$version" in
	[0-9]*.[0-9]*.[0-9]*) ;;
	*)
		printf 'invalid semantic version in %s: %s\n' "$version_file" "$version" >&2
		exit 1
		;;
esac

major=${version%%.*}
rest=${version#*.}
minor=${rest%%.*}
patch=${rest#*.}
case "$major:$minor:$patch" in
	*[!0-9:]*|*:*:*:*)
		printf 'invalid semantic version in %s: %s\n' "$version_file" "$version" >&2
		exit 1
		;;
esac

next_version="$major.$minor.$((patch + 1))"
printf '%s\n' "$next_version" > "$version_file"

printf 'building qbs %s\n' "$next_version"
go build -trimpath -ldflags "-s -w -X ${module}.Version=${next_version}" -o "$output" ./cmd/qbs
printf 'released qbs %s: %s\n' "$next_version" "$output"
