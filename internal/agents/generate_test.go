package agents_test

import (
	"bytes"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/trues/qbs/internal/agents"
	"gopkg.in/yaml.v3"
)

func TestGenerateIsDeterministicAndSorted(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	for _, harness := range []agents.Harness{agents.Codex, agents.OpenCode, agents.Claude} {
		t.Run(string(harness), func(t *testing.T) {
			first, err := agents.Generate(catalog, harness)
			if err != nil {
				t.Fatal(err)
			}
			reversed := catalog
			reversed.Agents = append([]agents.Agent(nil), catalog.Agents...)
			for left, right := 0, len(reversed.Agents)-1; left < right; left, right = left+1, right-1 {
				reversed.Agents[left], reversed.Agents[right] = reversed.Agents[right], reversed.Agents[left]
			}
			second, err := agents.Generate(reversed, harness)
			if err != nil {
				t.Fatal(err)
			}
			if len(first) != 13 || len(second) != 13 {
				t.Fatalf("generated counts = %d, %d", len(first), len(second))
			}
			for i := range first {
				if first[i].Path != second[i].Path || !bytes.Equal(first[i].Data, second[i].Data) {
					t.Fatalf("generation differs at index %d", i)
				}
				if i > 0 && first[i-1].Path >= first[i].Path {
					t.Fatalf("paths are not sorted: %q before %q", first[i-1].Path, first[i].Path)
				}
			}
		})
	}
}

func TestGeneratedMarkdownFrontmatterIsValidYAML(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	for _, harness := range []agents.Harness{agents.OpenCode, agents.Claude} {
		files, err := agents.Generate(catalog, harness)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			if !agents.IsGenerated(file.Data) {
				t.Errorf("%s does not have the QBS generated marker", file.Path)
			}
			parts := bytes.SplitN(file.Data, []byte("---\n"), 3)
			if len(parts) != 3 || len(parts[0]) != 0 {
				t.Errorf("%s has malformed frontmatter delimiters", file.Path)
				continue
			}
			var frontmatter map[string]any
			if err := yaml.Unmarshal(parts[1], &frontmatter); err != nil {
				t.Errorf("%s has invalid YAML frontmatter: %v", file.Path, err)
			}
			required := []string{"description", "model"}
			if harness == agents.OpenCode {
				required = append(required, "mode", "reasoningEffort", "permission")
			} else {
				required = append(required, "name", "tools", "permissionMode", "effort")
			}
			for _, field := range required {
				if frontmatter[field] == nil {
					t.Errorf("%s is missing required frontmatter field %s: %#v", file.Path, field, frontmatter)
				}
			}
		}
	}
}

func TestGeneratedCodexTOMLHasRequiredFields(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	files, err := agents.Generate(catalog, agents.Codex)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if !agents.IsGenerated(file.Data) {
			t.Errorf("%s does not have the QBS generated marker", file.Path)
		}
		fields := parseGeneratedTOML(t, file.Data)
		for _, field := range []string{"name", "description", "model", "model_reasoning_effort", "sandbox_mode", "developer_instructions"} {
			if strings.TrimSpace(fields[field]) == "" {
				t.Errorf("%s is missing required field %s", file.Path, field)
			}
		}
	}
}

func TestRepresentativeNativeOutputs(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}

	codex := generatedFile(t, catalog, agents.Codex, path.Join(".codex", "agents", "implementer.toml"))
	for _, want := range []string{
		`name = "implementer"`,
		`model = "gpt-5.6-sol"`,
		`model_reasoning_effort = "high"`,
		`sandbox_mode = "workspace-write"`,
		`developer_instructions = "Implement only the assigned ticket`,
	} {
		if !strings.Contains(codex, want) {
			t.Errorf("Codex output is missing %q:\n%s", want, codex)
		}
	}

	openCode := generatedFile(t, catalog, agents.OpenCode, path.Join(".opencode", "agents", "review-spec.md"))
	for _, want := range []string{
		"mode: subagent",
		`model: "openai/gpt-5.6-terra"`,
		`reasoningEffort: "high"`,
		`"edit":`,
		`"read": allow`,
		`"task": deny`,
		"Review only the supplied diff",
	} {
		if !strings.Contains(openCode, want) {
			t.Errorf("OpenCode output is missing %q:\n%s", want, openCode)
		}
	}
	if strings.Contains(openCode, `"edit": allow`) {
		t.Errorf("read-only OpenCode agent allows edits:\n%s", openCode)
	}

	claude := generatedFile(t, catalog, agents.Claude, path.Join(".claude", "agents", "implementer.md"))
	for _, want := range []string{
		`name: "implementer"`,
		`tools: Read, Grep, Glob, Bash, WebFetch, WebSearch, Skill, Edit, Write`,
		`model: "opus"`,
		`permissionMode: acceptEdits`,
		`effort: "high"`,
		`isolation: worktree`,
	} {
		if !strings.Contains(claude, want) {
			t.Errorf("Claude output is missing %q:\n%s", want, claude)
		}
	}
}

func TestGenerateTranslatesSkillsForNativeFormats(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	catalog.Agents = []agents.Agent{catalog.Agents[0]}
	catalog.Agents[0].Skills = []string{"codebase-design"}

	codex := generatedFile(t, catalog, agents.Codex, path.Join(".codex", "agents", "architecture-explorer.toml"))
	if !strings.Contains(codex, "Load and follow these skills before starting: codebase-design.") {
		t.Fatalf("Codex skill mapping:\n%s", codex)
	}
	openCode := generatedFile(t, catalog, agents.OpenCode, path.Join(".opencode", "agents", "architecture-explorer.md"))
	if !strings.Contains(openCode, `"skill": allow`) {
		t.Fatalf("OpenCode skill mapping:\n%s", openCode)
	}
	claude := generatedFile(t, catalog, agents.Claude, path.Join(".claude", "agents", "architecture-explorer.md"))
	if !strings.Contains(claude, "skills:\n  - \"codebase-design\"") {
		t.Fatalf("Claude skill mapping:\n%s", claude)
	}
}

func parseGeneratedTOML(t *testing.T, data []byte) map[string]string {
	t.Helper()
	fields := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			t.Fatalf("invalid generated TOML line %q", line)
		}
		value, err := strconv.Unquote(strings.TrimSpace(parts[1]))
		if err != nil {
			t.Fatalf("invalid generated TOML string %q: %v", line, err)
		}
		fields[strings.TrimSpace(parts[0])] = value
	}
	return fields
}

func generatedFile(t *testing.T, catalog agents.Catalog, harness agents.Harness, path string) string {
	t.Helper()
	files, err := agents.Generate(catalog, harness)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.Path == path {
			return string(file.Data)
		}
	}
	t.Fatalf("generated output is missing %s", path)
	return ""
}
