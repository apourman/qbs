#!/bin/sh
set -eu

release_dir=${QBS_RELEASE_DIR:-dist}
version=${VERSION:-dev}
binary="$release_dir/qbs_${version}_linux_amd64"
if [ ! -f "$binary" ]; then
  echo "no Linux amd64 release binary for $version found in $release_dir" >&2
  exit 1
fi
binary=$(cd "$(dirname "$binary")" && pwd)/$(basename "$binary")

root=$(mktemp -d)
repo="$root/repository"
skill_source="$root/skill-source/demo"
mkdir -p "$repo" "$skill_source"
trap 'rm -rf "$root"' EXIT

git -C "$repo" init -q
git -C "$repo" config user.email qbs-smoke@example.invalid
git -C "$repo" config user.name qbs-smoke
printf 'fixture\n' > "$repo/README.md"
printf '%s\n' '---' 'name: demo' 'description: smoke test skill' '---' > "$skill_source/SKILL.md"
git -C "$repo" add README.md
git -C "$repo" commit -qm fixture

version_output=$("$binary" --version)
test "$version_output" = "qbs $version"
(cd "$repo" && "$binary" init)
QBS_HOME="$root/qbs-home" \
QBS_SKILL_TARGETS="$root/codex-skills:$root/claude-skills" \
  "$binary" skills import "$skill_source" </dev/null

test -f "$repo/AGENTS.md"
test -f "$repo/.agents/skills/qbs/SKILL.md"
test -f "$repo/.claude/skills/qbs/SKILL.md"
test -f "$repo/.opencode/skills/qbs/SKILL.md"
test -d "$repo/.research"
test -d "$repo/.specs"
grep -q 'CLAUDE.md' "$repo/AGENTS.md"
test -f "$root/qbs-home/skills/demo/SKILL.md"
test -f "$root/codex-skills/demo/SKILL.md"
test -f "$root/claude-skills/demo/SKILL.md"
printf 'release smoke test passed\n'
