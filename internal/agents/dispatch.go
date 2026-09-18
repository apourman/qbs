package agents

import (
	"fmt"
	"sort"
)

// Role describes the work and constraints of a named dispatched sub-agent.
type Role struct {
	Name       string
	Purpose    string
	Capability string
	Optional   bool
	Reasoning  string
	Isolation  string
	Concurrent bool
}

// Plan selects how model assignments are chosen for a roster.
type Plan string

const (
	Recommended Plan = "Recommended"
	Economy     Plan = "Economy"
	Deep        Plan = "Deep"
	Customize   Plan = "Customize"
)

// Assignment is a neutral model profile and reasoning selection for one role.
type Assignment struct {
	ModelProfile string
	Reasoning    string
}

// DispatchAssignment is an assignment resolved to the active harness model.
type DispatchAssignment struct {
	Role          Role
	ModelProfile  string
	ResolvedModel string
	Reasoning     string
}

// ResolveProfile resolves a neutral profile without falling back to another
// profile when the requested profile or harness is unavailable.
func (c Catalog) ResolveProfile(profile string, harness Harness) (string, error) {
	model, ok := c.Models[profile]
	if !ok {
		return "", fmt.Errorf("resolve model profile %q for harness %q: profile is unavailable", profile, harness)
	}
	var resolved string
	switch harness {
	case Codex:
		resolved = model.Codex
	case OpenCode:
		resolved = model.OpenCode
	case Claude:
		resolved = model.Claude
	default:
		return "", fmt.Errorf("resolve model profile %q: unsupported harness %q", profile, harness)
	}
	if resolved == "" {
		return "", fmt.Errorf("resolve model profile %q for harness %q: model is unavailable", profile, harness)
	}
	return resolved, nil
}

// ResolvePlan assigns models to every role and resolves profiles for harness.
// Custom plans must provide an explicit assignment for every role.
func (c Catalog) ResolvePlan(roles []Role, harness Harness, plan Plan, custom map[string]Assignment) ([]DispatchAssignment, error) {
	if len(roles) == 0 {
		return nil, fmt.Errorf("resolve %s plan: role roster is empty", plan)
	}
	if plan != Recommended && plan != Economy && plan != Deep && plan != Customize {
		return nil, fmt.Errorf("resolve plan: unsupported plan %q", plan)
	}
	seen := make(map[string]bool, len(roles))
	result := make([]DispatchAssignment, 0, len(roles))
	for _, role := range roles {
		if role.Name == "" || seen[role.Name] {
			return nil, fmt.Errorf("resolve %s plan: role names must be non-empty and unique", plan)
		}
		seen[role.Name] = true
		assignment, ok := assignmentFor(role, plan, custom)
		if !ok {
			return nil, fmt.Errorf("resolve %s plan: missing custom assignment for role %q", plan, role.Name)
		}
		resolved, err := c.ResolveProfile(assignment.ModelProfile, harness)
		if err != nil {
			return nil, fmt.Errorf("role %q: %w", role.Name, err)
		}
		if assignment.Reasoning == "" {
			return nil, fmt.Errorf("role %q: reasoning effort is required", role.Name)
		}
		result = append(result, DispatchAssignment{Role: role, ModelProfile: assignment.ModelProfile, ResolvedModel: resolved, Reasoning: assignment.Reasoning})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Role.Name < result[j].Role.Name })
	return result, nil
}

func assignmentFor(role Role, plan Plan, custom map[string]Assignment) (Assignment, bool) {
	switch plan {
	case Economy:
		return Assignment{ModelProfile: "balanced", Reasoning: "low"}, true
	case Deep:
		return Assignment{ModelProfile: "capable", Reasoning: "high"}, true
	case Recommended:
		profile, reasoning := "balanced", role.Reasoning
		if role.Capability == "implementation" || role.Capability == "architecture" || role.Capability == "adversarial-review" {
			profile = "capable"
		}
		if reasoning == "" {
			reasoning = "medium"
		}
		return Assignment{ModelProfile: profile, Reasoning: reasoning}, true
	case Customize:
		assignment, ok := custom[role.Name]
		return assignment, ok
	default:
		return Assignment{}, false
	}
}
