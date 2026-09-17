package agents_test

import (
	"strings"
	"testing"

	"github.com/trues/qbs/internal/agents"
)

func TestCanonicalCatalogParsesAndPreservesAllRoles(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	if got := len(catalog.Agents); got != 13 {
		t.Fatalf("agent count = %d, want 13", got)
	}
	wants := map[string]string{
		"architecture-explorer": "Walk the codebase organically",
		"implementer":           "Follow the repository's /implement skill instructions",
		"merger":                "Do not rewrite unrelated history or push",
		"review-spec":           "If no specification is available",
		"task-agent":            "Complete only the tightly scoped task",
	}
	for _, agent := range catalog.Agents {
		if want, ok := wants[agent.Name]; ok {
			if !strings.Contains(agent.Prompt, want) {
				t.Errorf("%s prompt does not contain %q: %q", agent.Name, want, agent.Prompt)
			}
			delete(wants, agent.Name)
		}
	}
	for name := range wants {
		t.Errorf("canonical catalog is missing %s", name)
	}
}

func TestParseRejectsUnknownAndInvalidFields(t *testing.T) {
	valid := `version: 1
models:
  balanced: {codex: codex-model, opencode: provider/model, claude: sonnet}
agents:
  - name: reader
    description: Reads things
    model: balanced
    reasoning: medium
    permission: read-only
    tools: [read]
    isolation: shared
    prompt: Read carefully.
`
	for name, input := range map[string]string{
		"unknown field":    strings.Replace(valid, "    prompt:", "    surprise: true\n    prompt:", 1),
		"unknown model":    strings.Replace(valid, "model: balanced", "model: missing", 1),
		"write mismatch":   strings.Replace(valid, "tools: [read]", "tools: [read, write]", 1),
		"empty tools":      strings.Replace(valid, "tools: [read]", "tools: []", 1),
		"multiple document": valid + "---\nversion: 1\n",
		"duplicate name": valid + `  - name: reader
    description: Also reads
    model: balanced
    reasoning: medium
    permission: read-only
    tools: [read]
    isolation: shared
    prompt: Read again.
`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := agents.Parse([]byte(input)); err == nil {
				t.Fatal("Parse succeeded")
			}
		})
	}
}
