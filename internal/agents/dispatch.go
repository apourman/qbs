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

// Ticket describes one implementation unit in a ticket-level dispatch plan.
// Role remains the stable workflow declaration; the ticket supplies the
// concrete scope and dependency frontier for that role.
type Ticket struct {
	ID               string
	Scope            string
	Dependencies     []string
	Role             Role
	ConcurrencyClass string
}

// TicketDispatchAssignment is the approved, harness-resolved assignment for
// one implementation ticket.
type TicketDispatchAssignment struct {
	TicketID            string
	Scope               string
	Dependencies        []string
	Role                Role
	ModelProfile        string
	ResolvedModel       string
	Reasoning           string
	EstimatedTokens     TokenRange
	TokenBudget         TokenRange
	ConcurrencyClass    string
	ConcurrencyLimit    int
	Isolation           string
	ReservedFixCapacity TokenRange
	AccountQuota        CostStatus
	MonetaryCost        CostStatus
}

// TicketDispatchPlan is the complete approved implementation dispatch plan.
// Global, role, and class limits are intentionally separate so their
// constraints cannot be mistaken for one another.
type TicketDispatchPlan struct {
	Tickets                []TicketDispatchAssignment
	GlobalConcurrencyLimit int
	RoleConcurrencyLimits  map[string]int
	ClassConcurrencyLimits map[string]int
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

// ResolveTicketPlan resolves one assignment for every implementation ticket.
// Custom assignments are keyed by ticket ID, making the ticket the cost and
// approval boundary while retaining role-level defaults for other plans.
func (c Catalog) ResolveTicketPlan(tickets []Ticket, harness Harness, plan Plan, custom map[string]Assignment) (TicketDispatchPlan, error) {
	return c.ResolveTicketPlanWithLimits(tickets, harness, plan, custom, 2, nil, nil)
}

// ResolveTicketPlanWithLimits is ResolveTicketPlan with explicit concurrency
// limits for the approved plan.
func (c Catalog) ResolveTicketPlanWithLimits(tickets []Ticket, harness Harness, plan Plan, custom map[string]Assignment, globalLimit int, roleLimits, classLimits map[string]int) (TicketDispatchPlan, error) {
	if len(tickets) == 0 {
		return TicketDispatchPlan{}, fmt.Errorf("resolve %s ticket plan: ticket list is empty", plan)
	}
	if globalLimit <= 0 {
		return TicketDispatchPlan{}, fmt.Errorf("resolve %s ticket plan: global concurrency limit must be positive", plan)
	}
	if plan != Recommended && plan != Economy && plan != Deep && plan != Customize {
		return TicketDispatchPlan{}, fmt.Errorf("resolve ticket plan: unsupported plan %q", plan)
	}
	result := TicketDispatchPlan{
		GlobalConcurrencyLimit: globalLimit,
		RoleConcurrencyLimits:  copyLimits(roleLimits),
		ClassConcurrencyLimits: copyLimits(classLimits),
		Tickets:                make([]TicketDispatchAssignment, 0, len(tickets)),
	}
	seen := make(map[string]bool, len(tickets))
	for _, ticket := range tickets {
		if ticket.ID == "" || seen[ticket.ID] {
			return TicketDispatchPlan{}, fmt.Errorf("resolve %s ticket plan: ticket IDs must be non-empty and unique", plan)
		}
		seen[ticket.ID] = true
		if ticket.Role.Name == "" {
			return TicketDispatchPlan{}, fmt.Errorf("resolve %s ticket plan: ticket %q has no role", plan, ticket.ID)
		}
		if _, ok := result.RoleConcurrencyLimits[ticket.Role.Name]; !ok && ticket.Role.ConcurrencyLimit > 0 {
			result.RoleConcurrencyLimits[ticket.Role.Name] = ticket.Role.ConcurrencyLimit
		}
		assignment, ok := assignmentForTicket(ticket, plan, custom)
		if !ok {
			return TicketDispatchPlan{}, fmt.Errorf("resolve %s ticket plan: missing custom assignment for ticket %q", plan, ticket.ID)
		}
		resolved, err := c.ResolveProfile(assignment.ModelProfile, harness)
		if err != nil {
			return TicketDispatchPlan{}, fmt.Errorf("ticket %q: %w", ticket.ID, err)
		}
		if assignment.Reasoning == "" {
			return TicketDispatchPlan{}, fmt.Errorf("ticket %q: reasoning effort is required", ticket.ID)
		}
		if !supportedReasoning(assignment.Reasoning) {
			return TicketDispatchPlan{}, fmt.Errorf("ticket %q: reasoning effort %q is unsupported; choose low, medium, high, xhigh, or max", ticket.ID, assignment.Reasoning)
		}
		if limit, ok := result.RoleConcurrencyLimits[ticket.Role.Name]; ok && limit <= 0 {
			return TicketDispatchPlan{}, fmt.Errorf("ticket %q: role %q concurrency limit must be positive", ticket.ID, ticket.Role.Name)
		}
		if limit, ok := result.ClassConcurrencyLimits[ticket.ConcurrencyClass]; ok && limit <= 0 {
			return TicketDispatchPlan{}, fmt.Errorf("ticket %q: concurrency class %q limit must be positive", ticket.ID, ticket.ConcurrencyClass)
		}
		result.Tickets = append(result.Tickets, TicketDispatchAssignment{
			TicketID:            ticket.ID,
			Scope:               ticket.Scope,
			Dependencies:        append([]string(nil), ticket.Dependencies...),
			Role:                ticket.Role,
			ModelProfile:        assignment.ModelProfile,
			ResolvedModel:       resolved,
			Reasoning:           assignment.Reasoning,
			EstimatedTokens:     ticket.Role.EstimatedTokens,
			TokenBudget:         ticket.Role.TokenBudget,
			ConcurrencyClass:    ticket.ConcurrencyClass,
			ConcurrencyLimit:    ticket.Role.ConcurrencyLimit,
			Isolation:           ticket.Role.Isolation,
			ReservedFixCapacity: ticket.Role.ReservedBudget.Fixes,
			AccountQuota:        ticket.Role.AccountQuota,
			MonetaryCost:        ticket.Role.MonetaryCost,
		})
	}
	sort.SliceStable(result.Tickets, func(i, j int) bool { return result.Tickets[i].TicketID < result.Tickets[j].TicketID })
	return result, nil
}

func assignmentForTicket(ticket Ticket, plan Plan, custom map[string]Assignment) (Assignment, bool) {
	if assignment, ok := custom[ticket.ID]; ok {
		return assignment, true
	}
	if plan == Customize {
		return Assignment{}, false
	}
	return assignmentFor(ticket.Role, plan, custom)
}

func copyLimits(limits map[string]int) map[string]int {
	if len(limits) == 0 {
		return map[string]int{}
	}
	copy := make(map[string]int, len(limits))
	for name, limit := range limits {
		copy[name] = limit
	}
	return copy
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
