#!/bin/sh
set -eu

release_dir=${QBS_RELEASE_DIR:-dist}
binary=$(find "$release_dir" -maxdepth 1 -type f -name 'qbs_*_linux_amd64' -print -quit)
if [ -z "$binary" ]; then
  echo "no Linux amd64 release binary found in $release_dir" >&2
  exit 1
fi
binary=$(cd "$(dirname "$binary")" && pwd)/$(basename "$binary")

root=$(mktemp -d)
tmux_bin="$root/bin"
repo="$root/repository"
skill_source="$root/skill-source/demo"
mkdir -p "$tmux_bin" "$repo" "$skill_source"
trap 'rm -rf "$root"' EXIT

cat > "$tmux_bin/tmux" <<'EOF'
#!/bin/sh
exit 0
EOF
chmod +x "$tmux_bin/tmux"

git -C "$repo" init -q
git -C "$repo" config user.email qbs-smoke@example.invalid
git -C "$repo" config user.name qbs-smoke
printf 'fixture\n' > "$repo/README.md"
printf '%s\n' '---' 'name: demo' 'description: smoke test skill' '---' > "$skill_source/SKILL.md"
git -C "$repo" add README.md
git -C "$repo" commit -qm fixture

version=${VERSION:-dev}
version_output=$(PATH="$tmux_bin:$PATH" "$binary" --version)
test "$version_output" = "qbs $version"
(cd "$repo" && PATH="$tmux_bin:$PATH" "$binary" init)
(cd "$repo" && PATH="$tmux_bin:$PATH" "$binary" task smoke)
QBS_HOME="$root/qbs-home" \
QBS_SKILL_TARGETS="$root/codex-skills:$root/claude-skills" \
  "$binary" skills import "$skill_source"

test -f "$repo/AGENTS.md"
test -f "$repo/.agents/skills/qbs/SKILL.md"
test -f "$repo/.claude/skills/qbs/SKILL.md"
test -f "$repo/.opencode/skills/qbs/SKILL.md"
test -d "$repo/.research"
test -d "$repo/.specs"
test -f "${repo}-smoke/.agents/skills/qbs/SKILL.md"
test -f "${repo}-smoke/.claude/skills/qbs/SKILL.md"
test -f "${repo}-smoke/.opencode/skills/qbs/SKILL.md"
test -d "${repo}-smoke/.research"
test -d "${repo}-smoke/.specs"
grep -q 'CLAUDE.md' "$repo/AGENTS.md"
grep -q 'CLAUDE.md' "${repo}-smoke/AGENTS.md"
test -f "$root/qbs-home/skills/demo/SKILL.md"
test -f "$root/codex-skills/demo/SKILL.md"
test -f "$root/claude-skills/demo/SKILL.md"
printf 'release smoke test passed\n'
