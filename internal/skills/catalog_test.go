package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportCollectionSyncUpdateAndRemove(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeSkill(t, source, "alpha", "alpha v1\n")
	writeSkill(t, source, "beta", "beta v1\n")
	targetOne := filepath.Join(root, "codex")
	targetTwo := filepath.Join(root, "claude")
	catalog := Catalog{
		Root:    filepath.Join(root, "catalog"),
		Targets: []string{targetOne, targetTwo},
	}

	names, err := catalog.Import(source, false)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(names, ","); got != "alpha,beta" {
		t.Fatalf("imported names = %q", got)
	}
	for _, target := range catalog.Targets {
		for _, name := range names {
			assertFile(t, filepath.Join(target, name, "SKILL.md"))
			assertFile(t, filepath.Join(target, name, managedMarker))
		}
	}

	if _, err := catalog.Import(source, false); err == nil || !strings.Contains(err.Error(), "--force") {
		t.Fatalf("duplicate import error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "alpha", "SKILL.md"), []byte("alpha v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.Import(filepath.Join(source, "alpha"), true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(targetTwo, "alpha", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "alpha v2\n" {
		t.Fatalf("synchronized content = %q", data)
	}

	if err := catalog.Remove("alpha"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(catalog.Root, "alpha"),
		filepath.Join(targetOne, "alpha"),
		filepath.Join(targetTwo, "alpha"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("removed skill remains at %s: %v", path, err)
		}
	}
}

func TestImportPreservesUnmanagedTargetBeforeChangingCatalog(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeSkill(t, source, "alpha", "catalog copy\n")
	target := filepath.Join(root, "target")
	writeSkill(t, target, "alpha", "unmanaged copy\n")
	catalog := Catalog{Root: filepath.Join(root, "catalog"), Targets: []string{target}}

	_, err := catalog.Import(source, false)
	if err == nil || !strings.Contains(err.Error(), "preserving unmanaged skill") {
		t.Fatalf("import error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(catalog.Root, "alpha")); !os.IsNotExist(err) {
		t.Fatalf("failed import changed catalog: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(target, "alpha", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "unmanaged copy\n" {
		t.Fatalf("unmanaged skill changed: %q", data)
	}
}

func TestSyncInteractiveAsksBeforeReplacingUnmanagedTarget(t *testing.T) {
	root := t.TempDir()
	catalog := Catalog{
		Root:    filepath.Join(root, "catalog"),
		Targets: []string{filepath.Join(root, "target")},
	}
	writeSkill(t, catalog.Root, "alpha", "qbs copy\n")
	writeSkill(t, catalog.Targets[0], "alpha", "personal copy\n")

	var asked string
	names, err := catalog.SyncInteractive(func(name, target string) (bool, error) {
		asked = name + "@" + target
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || asked == "" {
		t.Fatalf("names = %#v, asked = %q", names, asked)
	}
	data, err := os.ReadFile(filepath.Join(catalog.Targets[0], "alpha", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "personal copy\n" {
		t.Fatalf("declined replacement changed skill: %q", data)
	}

	if _, err := catalog.SyncInteractive(func(string, string) (bool, error) { return true, nil }); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(filepath.Join(catalog.Targets[0], "alpha", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "qbs copy\n" || !isManaged(filepath.Join(catalog.Targets[0], "alpha")) {
		t.Fatalf("accepted replacement did not install managed skill: %q", data)
	}
}

func TestImportInteractiveDoesNotPromptForNewTarget(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	writeSkill(t, source, "alpha", "new skill\n")
	catalog := Catalog{
		Root:    filepath.Join(root, "catalog"),
		Targets: []string{filepath.Join(root, "target")},
	}

	_, err := catalog.ImportInteractive(source, false, func(string, string) (bool, error) {
		t.Fatal("asked to replace a skill at an empty target")
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(catalog.Targets[0], "alpha", "SKILL.md"))
}

func TestDefaultCatalogUsesOverrides(t *testing.T) {
	root := t.TempDir()
	targetOne := filepath.Join(root, "one")
	targetTwo := filepath.Join(root, "two")
	t.Setenv("QBS_HOME", filepath.Join(root, "home"))
	t.Setenv("QBS_SKILL_TARGETS", strings.Join([]string{targetOne, targetTwo}, string(os.PathListSeparator)))

	catalog, err := DefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Root != filepath.Join(root, "home", "skills") {
		t.Fatalf("catalog root = %s", catalog.Root)
	}
	if len(catalog.Targets) != 2 || catalog.Targets[0] != targetOne || catalog.Targets[1] != targetTwo {
		t.Fatalf("catalog targets = %#v", catalog.Targets)
	}
}

func TestDefaultCatalogUsesSharedAgentSkillsDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("QBS_HOME", "")
	t.Setenv("QBS_SKILL_TARGETS", "")
	chdirForTest(t, home)

	catalog, err := DefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		filepath.Join(home, ".agents", "skills"),
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".config", "opencode", "skills"),
	}
	if len(catalog.Targets) != len(want) {
		t.Fatalf("skill targets = %#v, want %#v", catalog.Targets, want)
	}
	for i := range want {
		if catalog.Targets[i] != want[i] {
			t.Fatalf("skill targets = %#v, want %#v", catalog.Targets, want)
		}
	}
}

func TestDefaultCatalogIncludesProjectAgentSkillsDirectory(t *testing.T) {
	home := t.TempDir()
	project := filepath.Join(home, "project")
	if err := os.MkdirAll(filepath.Join(project, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(project, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("QBS_HOME", "")
	t.Setenv("QBS_SKILL_TARGETS", "")
	chdirForTest(t, filepath.Join(project, "nested"))

	catalog, err := DefaultCatalog()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(project, ".agents", "skills")
	if len(catalog.Targets) != 4 || catalog.Targets[3] != want {
		t.Fatalf("skill targets = %#v, want project target %s", catalog.Targets, want)
	}
}

func writeSkill(t *testing.T, parent, name, contents string) {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func chdirForTest(t *testing.T, directory string) {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previous); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

func assertFile(t *testing.T, path string) {
	t.Helper()
	if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("expected file %s: %v", path, err)
	}
}
