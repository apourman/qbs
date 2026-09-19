package cli_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/trues/qbs/internal/cli"
	"github.com/trues/qbs/internal/templates"
)

func TestMain(m *testing.M) {
	if marker := os.Getenv("QBS_TEST_GH_HELPER_MARKER"); marker != "" {
		_ = os.WriteFile(marker, []byte("invoked\n"), 0o644)
		os.Exit(99)
	}
	os.Exit(m.Run())
}

func TestInitProvisionsIgnoredAIWorkspaceIdempotently(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })

	var firstOut, firstErr bytes.Buffer
	if err := cli.Run([]string{"init"}, &firstOut, &firstErr); err != nil {
		t.Fatal(err)
	}

	wants := []string{
		"AGENTS.md",
		"CLAUDE.md",
		".research",
		".specs",
		filepath.Join("docs", "agents", "domain.md"),
		filepath.Join("docs", "agents", "issue-tracker.md"),
	}
	for _, name := range wants {
		if _, err := os.Stat(filepath.Join(repo, name)); err != nil {
			t.Errorf("init did not create %s: %v", name, err)
		}
	}
	for _, name := range wants {
		if out := runGit(t, repo, "check-ignore", "--", name); strings.TrimSpace(out) == "" {
			t.Errorf("%s is not ignored", name)
		}
	}
	if status := runGit(t, repo, "status", "--short"); status != "" {
		t.Fatalf("initialized repository has Git changes: %q", status)
	}
	for _, name := range []string{".agents/skills", ".claude/skills", ".opencode/skills"} {
		if _, err := os.Stat(filepath.Join(repo, name)); !os.IsNotExist(err) {
			t.Errorf("init created local skill directory %s: %v", name, err)
		}
	}

	localSkill := filepath.Join(repo, ".agents", "skills", "local", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(localSkill), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localSkill, []byte("local skill\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	agents := filepath.Join(repo, "AGENTS.md")
	if err := os.WriteFile(agents, []byte("local instructions\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var secondOut, secondErr bytes.Buffer
	if err := cli.Run([]string{"init"}, &secondOut, &secondErr); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(agents)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "local instructions\n" {
		t.Fatalf("existing AGENTS.md was overwritten: %q", data)
	}
	if data, err := os.ReadFile(localSkill); err != nil || string(data) != "local skill\n" {
		t.Fatalf("existing local skill was changed: %q, %v", data, err)
	}
	existingAgent := filepath.Join(repo, ".codex", "agents", "existing.toml")
	if err := os.MkdirAll(filepath.Dir(existingAgent), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existingAgent, []byte("existing agent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := cli.Run([]string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	generatedData, err := os.ReadFile(existingAgent)
	if err != nil {
		t.Fatal(err)
	}
	if string(generatedData) != "existing agent\n" {
		t.Fatalf("init changed existing agent: %q", generatedData)
	}
	unmanaged := filepath.Join(repo, ".opencode", "agents", "implementer.md")
	if err := os.MkdirAll(filepath.Dir(unmanaged), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(unmanaged, []byte("---\ndescription: local agent\n---\n\nKeep me.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	unrelated := filepath.Join(repo, ".opencode", "agents", "local-only.md")
	if err := os.WriteFile(unrelated, []byte("local only\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var preserveErr bytes.Buffer
	if err := cli.Run([]string{"init"}, &bytes.Buffer{}, &preserveErr); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]string{unmanaged: "Keep me.", unrelated: "local only"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), want) {
			t.Errorf("init changed unmanaged agent %s: %q", path, data)
		}
	}
	if data, err := os.ReadFile(localSkill); err != nil || string(data) != "local skill\n" {
		t.Fatalf("existing local skill changed after repeated init: %q, %v", data, err)
	}
	if strings.Contains(preserveErr.String(), "local agent") {
		t.Errorf("init unexpectedly wrote agent warnings: %q", preserveErr.String())
	}
	excludeData, err := os.ReadFile(filepath.Join(repo, ".git", "info", "exclude"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{"AGENTS.md", "CLAUDE.md", ".agents/skills/", ".claude/agents/", ".claude/skills/", ".codex/agents/", ".opencode/agents/", ".opencode/skills/", ".research/", ".specs/", "docs/agents/"} {
		if count := strings.Count(string(excludeData), entry); count != 1 {
			t.Errorf("exclude entry %q occurs %d times", entry, count)
		}
	}
}

func TestInitUsesLocalIssueTrackerRegardlessOfRemoteAndPreservesLocalRecords(t *testing.T) {
	const issueContent = "# Existing issue\n\n**Status:** ready-for-agent\n\n## Triage notes\n\nVerified locally.\n\n## Agent brief\n\nImplement from this record.\n"
	var trackerWithoutRemote []byte
	var triageWithoutRemote []byte
	for _, test := range []struct {
		name         string
		githubRemote bool
	}{
		{name: "without remote"},
		{name: "with GitHub remote", githubRemote: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			runGit(t, repo, "init", "-q")
			writeSkill(t, filepath.Join(repo, ".agents", "skills"), "triage")
			const remoteURL = "git@github.com:example/project.git"
			if test.githubRemote {
				runGit(t, repo, "remote", "add", "origin", remoteURL)
			}

			research := filepath.Join(repo, ".research", "local-first", "background.md")
			spec := filepath.Join(repo, ".specs", "local-first", "spec.md")
			issue := filepath.Join(repo, ".specs", "local-first", "issues", "01-existing.md")
			if err := os.MkdirAll(filepath.Dir(research), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Dir(issue), 0o755); err != nil {
				t.Fatal(err)
			}
			for path, content := range map[string]string{
				research: "existing local research\n",
				spec:     "existing local spec\n",
				issue:    issueContent,
			} {
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			tracker := filepath.Join(repo, "docs", "agents", "issue-tracker.md")
			data, err := os.ReadFile(tracker)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), "authoritative source") {
				t.Fatalf("tracker template = %q", data)
			}
			triageData, err := os.ReadFile(filepath.Join(repo, "docs", "agents", "triage.md"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(triageData), "authoritative Markdown ticket record") ||
				!strings.Contains(string(triageData), "`Status` field") {
				t.Fatalf("triage template = %q", triageData)
			}
			for _, forbidden := range []string{"gh issue ", "This repository uses GitHub Issues"} {
				if strings.Contains(string(data), forbidden) {
					t.Errorf("local tracker template contains hosted instruction %q", forbidden)
				}
			}
			if test.githubRemote {
				if !bytes.Equal(data, trackerWithoutRemote) {
					t.Errorf("GitHub remote changed tracker guidance\nwithout remote:\n%s\nwith remote:\n%s", trackerWithoutRemote, data)
				}
				if !bytes.Equal(triageData, triageWithoutRemote) {
					t.Errorf("GitHub remote changed triage guidance\nwithout remote:\n%s\nwith remote:\n%s", triageWithoutRemote, triageData)
				}
				if got := strings.TrimSpace(runGit(t, repo, "remote", "get-url", "origin")); got != remoteURL {
					t.Errorf("init changed origin remote: got %q, want %q", got, remoteURL)
				}
			} else {
				trackerWithoutRemote = append([]byte(nil), data...)
				triageWithoutRemote = append([]byte(nil), triageData...)
			}

			if err := os.WriteFile(tracker, []byte("custom tracker\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			data, err = os.ReadFile(tracker)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "custom tracker\n" {
				t.Fatalf("existing tracker was overwritten: %q", data)
			}
			for path, want := range map[string]string{
				research: "existing local research\n",
				spec:     "existing local spec\n",
				issue:    issueContent,
			} {
				record, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if string(record) != want {
					t.Fatalf("existing local record %s was overwritten: %q", path, record)
				}
			}
		})
	}
}

func TestInitWithGitHubRemoteDoesNotInvokeHostedTrackerCLI(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "remote", "add", "origin", "git@github.com:example/project.git")

	binDir := t.TempDir()
	marker := filepath.Join(t.TempDir(), "gh-invoked")
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	fakeGHName := "gh"
	if runtime.GOOS == "windows" {
		fakeGHName += ".exe"
	}
	fakeGH := filepath.Join(binDir, fakeGHName)
	executableData, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fakeGH, executableData, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("QBS_TEST_GH_HELPER_MARKER", marker)
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("qbs init invoked the hosted-tracker CLI: %v", err)
	}
}

func TestInitPreservesCustomizedEngineeringGuidance(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	writeSkill(t, filepath.Join(repo, ".agents", "skills"), "triage")
	if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	custom := map[string][]byte{
		"domain.md":        []byte("custom domain guidance\n"),
		"issue-tracker.md": []byte("custom tracker guidance\n"),
		"triage.md":        []byte("custom triage guidance\n"),
	}
	for name, content := range custom {
		if err := os.WriteFile(filepath.Join(repo, "docs", "agents", name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for name, want := range custom {
		data, err := os.ReadFile(filepath.Join(repo, "docs", "agents", name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, want) {
			t.Errorf("custom %s was overwritten: %q", name, data)
		}
	}
}

func TestInitKeepsLocalContextWorktreeLocalAndUsesSharedExcludes(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.email", "qbs-tests@example.invalid")
	runGit(t, repo, "config", "user.name", "QBS Tests")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("test repository\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "README.md")
	runGit(t, repo, "commit", "-qm", "initialize repository")

	mainResearch := filepath.Join(repo, ".research", "main-only.md")
	mainSpec := filepath.Join(repo, ".specs", "main-only.md")
	if err := os.MkdirAll(filepath.Dir(mainResearch), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(mainSpec), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mainResearch, []byte("main research\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mainSpec, []byte("main spec\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	linked := filepath.Join(t.TempDir(), "linked")
	runGit(t, repo, "worktree", "add", "-q", "-b", "linked-test", linked)
	if err := runInDirectory(linked, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{".research", ".specs", filepath.Join("docs", "agents", "issue-tracker.md")} {
		if _, err := os.Stat(filepath.Join(linked, name)); err != nil {
			t.Errorf("linked worktree did not receive %s: %v", name, err)
		}
	}
	for _, path := range []string{mainResearch, mainSpec} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("initializing linked worktree changed main-worktree context %s: %v", path, err)
		}
	}
	for _, path := range []string{
		filepath.Join(linked, ".research", "main-only.md"),
		filepath.Join(linked, ".specs", "main-only.md"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("worktree-local context leaked into linked worktree %s: %v", path, err)
		}
	}
	exclude := filepath.Join(repo, ".git", "info", "exclude")
	data, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{".research/", ".specs/", "docs/agents/"} {
		if count := strings.Count(string(data), entry); count != 1 {
			t.Errorf("shared exclude entry %q occurs %d times", entry, count)
		}
	}
}

func TestInitProvisionsLocalTriageStatusesWhenLocalTriageSkillExists(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	skill := filepath.Join(repo, ".agents", "skills", "triage")
	if err := os.MkdirAll(skill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("triage\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(repo, "docs", "agents", "triage.md"))
	if err != nil {
		t.Fatal(err)
	}
	for state, definition := range map[string]string{
		"needs-triage":    "`needs-triage`: the issue has not been assessed.",
		"needs-info":      "`needs-info`: more information is required before the issue can proceed.",
		"ready-for-agent": "`ready-for-agent`: the issue is sufficiently specified for implementation.",
		"ready-for-human": "`ready-for-human`: the issue needs a maintainer decision or action.",
		"wontfix":         "`wontfix`: the issue will not be pursued.",
	} {
		if !strings.Contains(string(data), definition) {
			t.Errorf("local triage guidance is missing %q and its meaning", state)
		}
	}
	for _, want := range []string{
		"authoritative Markdown ticket record",
		"`Status` field",
		"`## Triage notes`",
		"`## Agent brief`",
		"Do not create or manage hosted labels, comments, or tickets",
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("local triage guidance is missing %q", want)
		}
	}
}

func TestInitProvisionsLocalTriageGuidanceFromGlobalSkillLocations(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, qbsHome string)
	}{
		{
			name: "catalog",
			setup: func(t *testing.T, qbsHome string) {
				t.Setenv("QBS_SKILL_TARGETS", filepath.Join(t.TempDir(), "empty-target"))
				writeSkill(t, filepath.Join(qbsHome, "skills"), "triage")
			},
		},
		{
			name: "configured global target",
			setup: func(t *testing.T, qbsHome string) {
				target := filepath.Join(t.TempDir(), "global-skills")
				t.Setenv("QBS_SKILL_TARGETS", target)
				writeSkill(t, target, "triage")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := t.TempDir()
			qbsHome := t.TempDir()
			t.Setenv("QBS_HOME", qbsHome)
			test.setup(t, qbsHome)
			runGit(t, repo, "init", "-q")
			if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(filepath.Join(repo, "docs", "agents", "triage.md")); err != nil {
				t.Fatalf("global triage skill did not provision local guidance: %v", err)
			}
		})
	}
}

func TestInitUpgradesLegacyInstructionsAndPreservesCustomizedFiles(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	legacyAgents, err := templates.Read("legacy/qbs-agents-v1.md")
	if err != nil {
		t.Fatal(err)
	}
	legacyClaude, err := templates.Read("legacy/qbs-claude-v1.md")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), legacyAgents, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "CLAUDE.md"), legacyClaude, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"AGENTS.md", "CLAUDE.md"} {
		data, err := os.ReadFile(filepath.Join(repo, name))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "qbs-managed-instruction: v2") {
			t.Errorf("%s was not marked as current QBS instruction", name)
		}
	}
	agentsData, err := os.ReadFile(filepath.Join(repo, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(agentsData), "read and\nfollow that file") {
		t.Errorf("AGENTS.md does not point to CLAUDE.md: %q", agentsData)
	}
	claudeData, err := os.ReadFile(filepath.Join(repo, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(claudeData), "## Agent skills") {
		t.Errorf("CLAUDE.md is missing shared project instructions")
	}

	customAgents := []byte("# User-managed instructions\n\nKeep this exact text.\n")
	customClaude := []byte("# User-managed Claude instructions\n")
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), customAgents, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "CLAUDE.md"), customClaude, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string][]byte{"AGENTS.md": customAgents, "CLAUDE.md": customClaude} {
		data, err := os.ReadFile(filepath.Join(repo, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, want) {
			t.Errorf("custom %s was overwritten: %q", name, data)
		}
	}
}

func TestInitLocalIssueTrackerTemplateDescribesFilesystemWorkflow(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	if err := runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(repo, "docs", "agents", "issue-tracker.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Local Markdown ticket records",
		"authoritative source",
		"List tickets",
		"Create a ticket",
		"Update the ticket file",
		"separate, explicit request",
	} {
		if !strings.Contains(string(data), want) {
			t.Errorf("local tracker template is missing %q", want)
		}
	}
}

func TestInitPreflightsFilesystemConflictsBeforeChangingExcludes(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	exclude := filepath.Join(repo, ".git", "info", "exclude")
	before, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(repo, "AGENTS.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	err = runInDirectory(repo, []string{"init"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "not a regular file, but a file is required") {
		t.Fatalf("conflict error = %v", err)
	}
	after, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("init changed excludes despite preflight failure")
	}
}

func TestInitRejectsTrackedAIFileBeforeChangingExcludes(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.email", "qbs-tests@example.invalid")
	runGit(t, repo, "config", "user.name", "QBS Tests")
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), []byte("tracked instructions\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", "AGENTS.md")
	runGit(t, repo, "commit", "-qm", "track agent instructions")
	exclude := filepath.Join(repo, ".git", "info", "exclude")
	before, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}

	err = runInDirectory(repo, []string{"init"}, &bytes.Buffer{}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "tracked") || !strings.Contains(err.Error(), "AGENTS.md") {
		t.Fatalf("error = %v", err)
	}
	after, err := os.ReadFile(exclude)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("init changed excludes despite tracked AI file")
	}
}

func runInDirectory(dir string, args []string, stdout, stderr io.Writer) error {
	old, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := os.Chdir(dir); err != nil {
		return err
	}
	defer func() { _ = os.Chdir(old) }()
	return cli.Run(args, stdout, stderr)
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func writeSkill(t *testing.T, root, name string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("triage\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
