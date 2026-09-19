package agents

import (
	"fmt"
	"sort"
)

// Role describes the work and constraints of a named dispatched sub-agent.
type Role struct {
	Name             string
	Purpose          string
	Capability       string
	Optional         bool
	Reasoning        string
	Isolation        string
	Concurrent       bool
	EstimatedTokens  TokenRange
	TokenBudget      TokenRange
	ConcurrencyLimit int
	AccountQuota     CostStatus
	MonetaryCost     CostStatus
	ReservedBudget   ReservedBudget
}

// TokenRange represents an estimate or approved budget without false precision.
type TokenRange struct{ Min, Max int }

// CostStatus keeps quota and billing status distinct from token estimates.
type CostStatus struct{ Status, Detail string }

// ReservedBudget records capacity held for later workflow phases.
type ReservedBudget struct{ Review, Fixes, Integration TokenRange }

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
	Role             Role
	ModelProfile     string
	ResolvedModel    string
	Reasoning        string
	TokenBudget      TokenRange
	ConcurrencyLimit int
	AccountQuota     CostStatus
	MonetaryCost     CostStatus
	ReservedBudget   ReservedBudget
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
		if !supportedReasoning(assignment.Reasoning) {
			return nil, fmt.Errorf("role %q: reasoning effort %q is unsupported; choose low, medium, high, xhigh, or max", role.Name, assignment.Reasoning)
		}
		result = append(result, DispatchAssignment{Role: role, ModelProfile: assignment.ModelProfile, ResolvedModel: resolved, Reasoning: assignment.Reasoning, TokenBudget: role.TokenBudget, ConcurrencyLimit: role.ConcurrencyLimit, AccountQuota: role.AccountQuota, MonetaryCost: role.MonetaryCost, ReservedBudget: role.ReservedBudget})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Role.Name < result[j].Role.Name })
	return result, nil
}

func assignmentFor(role Role, plan Plan, custom map[string]Assignment) (Assignment, bool) {
	switch plan {
	case Economy:
		profile := "balanced"
		return Assignment{ModelProfile: profile, Reasoning: "low"}, true
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

func supportedReasoning(value string) bool {
	switch value {
	case "low", "medium", "high", "xhigh", "max":
		return true
	}
	return false
}
