package templates_test

import (
	"testing"

	"github.com/trues/qbs/internal/templates"
)

func TestReadReturnsEmbeddedCanonicalTemplate(t *testing.T) {
	data, err := templates.Read("AGENTS.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("embedded template is empty")
	}
	if string(data) == "" || !contains(string(data), "CLAUDE.md") {
		t.Errorf("unexpected template: %q", data)
	}
}

func TestReadReportsMissingTemplate(t *testing.T) {
	_, err := templates.Read("missing.md")
	if err == nil {
		t.Fatal("Read(missing.md) succeeded")
	}
}

func TestEmbeddedQBSSkillHasDiscoverableFrontmatter(t *testing.T) {
	data, err := templates.Read("skills/qbs/SKILL.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"---\nname: qbs\n", "description:", "qbs init", "qbs skills import"} {
		if !contains(string(data), want) {
			t.Errorf("QBS skill is missing %q: %q", want, data)
		}
	}
}

func contains(s, want string) bool {
	for i := 0; i+len(want) <= len(s); i++ {
		if s[i:i+len(want)] == want {
			return true
		}
	}
	return false
}
