package agents

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const schemaVersion = 1

var (
	//go:embed definitions.yaml
	canonicalFiles embed.FS
	namePattern    = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
)

// Catalog is the harness-neutral source model for QBS agents.
type Catalog struct {
	Version int                     `yaml:"version"`
	Models  map[string]ModelProfile `yaml:"models"`
	Agents  []Agent                 `yaml:"agents"`
}

// ModelProfile maps a neutral model profile to each harness's native model ID.
type ModelProfile struct {
	Codex    string `yaml:"codex"`
	OpenCode string `yaml:"opencode"`
	Claude   string `yaml:"claude"`
}

// Agent describes behavior and capabilities independently of a harness format.
type Agent struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Model       string   `yaml:"model"`
	Reasoning   string   `yaml:"reasoning"`
	Permission  string   `yaml:"permission"`
	Tools       []string `yaml:"tools"`
	Skills      []string `yaml:"skills,omitempty"`
	Isolation   string   `yaml:"isolation"`
	Prompt      string   `yaml:"prompt"`
}

// Canonical loads the embedded, tracked agent catalog.
func Canonical() (Catalog, error) {
	data, err := canonicalFiles.ReadFile("definitions.yaml")
	if err != nil {
		return Catalog{}, fmt.Errorf("read canonical agent definitions: %w", err)
	}
	return Parse(data)
}

// Parse strictly decodes and validates an agent catalog.
func Parse(data []byte) (Catalog, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	var catalog Catalog
	if err := decoder.Decode(&catalog); err != nil {
		return Catalog{}, fmt.Errorf("parse agent definitions: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return Catalog{}, errors.New("parse agent definitions: multiple YAML documents are not allowed")
		}
		return Catalog{}, fmt.Errorf("parse agent definitions: %w", err)
	}
	if err := catalog.Validate(); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}

// Validate enforces the portable schema before any harness files are written.
func (c Catalog) Validate() error {
	if c.Version != schemaVersion {
		return fmt.Errorf("validate agent definitions: version must be %d", schemaVersion)
	}
	if len(c.Models) == 0 {
		return errors.New("validate agent definitions: at least one model profile is required")
	}
	profileNames := make([]string, 0, len(c.Models))
	for name := range c.Models {
		profileNames = append(profileNames, name)
	}
	sort.Strings(profileNames)
	for _, name := range profileNames {
		profile := c.Models[name]
		if !namePattern.MatchString(name) {
			return fmt.Errorf("validate agent definitions: invalid model profile name %q", name)
		}
		if strings.TrimSpace(profile.Codex) == "" || strings.TrimSpace(profile.OpenCode) == "" || strings.TrimSpace(profile.Claude) == "" {
			return fmt.Errorf("validate agent definitions: model profile %q requires codex, opencode, and claude values", name)
		}
	}
	if len(c.Agents) == 0 {
		return errors.New("validate agent definitions: at least one agent is required")
	}
	seen := make(map[string]bool, len(c.Agents))
	for i, agent := range c.Agents {
		label := fmt.Sprintf("agent %d", i+1)
		if agent.Name != "" {
			label = fmt.Sprintf("agent %q", agent.Name)
		}
		if !namePattern.MatchString(agent.Name) {
			return fmt.Errorf("validate agent definitions: %s has an invalid name", label)
		}
		if seen[agent.Name] {
			return fmt.Errorf("validate agent definitions: duplicate agent name %q", agent.Name)
		}
		seen[agent.Name] = true
		if strings.TrimSpace(agent.Description) == "" || strings.TrimSpace(agent.Prompt) == "" {
			return fmt.Errorf("validate agent definitions: %s requires description and prompt", label)
		}
		if _, ok := c.Models[agent.Model]; !ok {
			return fmt.Errorf("validate agent definitions: %s references unknown model profile %q", label, agent.Model)
		}
		if !oneOf(agent.Reasoning, "low", "medium", "high", "xhigh", "max") {
			return fmt.Errorf("validate agent definitions: %s has invalid reasoning %q", label, agent.Reasoning)
		}
		if !oneOf(agent.Permission, "read-only", "workspace-write") {
			return fmt.Errorf("validate agent definitions: %s has invalid permission %q", label, agent.Permission)
		}
		if !oneOf(agent.Isolation, "shared", "worktree") {
			return fmt.Errorf("validate agent definitions: %s has invalid isolation %q", label, agent.Isolation)
		}
		if agent.Isolation == "worktree" && agent.Permission != "workspace-write" {
			return fmt.Errorf("validate agent definitions: %s cannot use worktree isolation without workspace-write permission", label)
		}
		if len(agent.Tools) == 0 {
			return fmt.Errorf("validate agent definitions: %s requires at least one tool capability", label)
		}
		toolSeen := make(map[string]bool, len(agent.Tools))
		for _, tool := range agent.Tools {
			if !oneOf(tool, "read", "search", "shell", "web", "skills", "write") {
				return fmt.Errorf("validate agent definitions: %s has unknown tool capability %q", label, tool)
			}
			if toolSeen[tool] {
				return fmt.Errorf("validate agent definitions: %s repeats tool capability %q", label, tool)
			}
			toolSeen[tool] = true
		}
		if agent.Permission == "read-only" && toolSeen["write"] {
			return fmt.Errorf("validate agent definitions: %s grants write tools with read-only permission", label)
		}
		skillSeen := make(map[string]bool, len(agent.Skills))
		for _, skill := range agent.Skills {
			if !namePattern.MatchString(skill) {
				return fmt.Errorf("validate agent definitions: %s has invalid skill name %q", label, skill)
			}
			if skillSeen[skill] {
				return fmt.Errorf("validate agent definitions: %s repeats skill %q", label, skill)
			}
			skillSeen[skill] = true
		}
	}
	return nil
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
