package agents

import (
	"fmt"
	"sort"
	"strings"
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

// WorkflowBudget keeps work around ticket implementation visible in its own
// approval section. These ranges must not be folded into ticket budgets.
type WorkflowBudget struct {
	Explorer    TokenRange
	Merger      TokenRange
	Reviewer    TokenRange
	RedTeam     TokenRange
	Fixes       TokenRange
	CI          TokenRange
	Integration TokenRange
}

// OptionalRoleStatus records whether an optional workflow role is omitted,
// selected, or conditionally available in the approved plan.
type OptionalRoleStatus struct {
	Name   string
	Status string
	Detail string
}

// DispatchApproval is the complete user-facing approval artifact. It retains
// the resolved ticket plan rather than only the role defaults so the approval
// can be checked before every implementation dispatch.
type DispatchApproval struct {
	Plan           Plan
	TicketPlan     TicketDispatchPlan
	WorkflowBudget WorkflowBudget
	OptionalRoles  []OptionalRoleStatus
	ReviewScope    string
	Approved       bool
}

// Validate checks that an approval contains the information required to show
// and record a complete ticket-level dispatch decision.
func (a DispatchApproval) Validate() error {
	if a.Plan != Recommended && a.Plan != Economy && a.Plan != Deep && a.Plan != Customize {
		return fmt.Errorf("validate dispatch approval: unsupported plan %q", a.Plan)
	}
	if len(a.TicketPlan.Tickets) == 0 {
		return fmt.Errorf("validate dispatch approval: ticket plan is empty")
	}
	for _, role := range a.OptionalRoles {
		if strings.TrimSpace(role.Name) == "" {
			return fmt.Errorf("validate dispatch approval: optional role name is empty")
		}
		if role.Status != "omitted" && role.Status != "selected" && role.Status != "conditional" {
			return fmt.Errorf("validate dispatch approval: optional role %q has unsupported status %q", role.Name, role.Status)
		}
	}
	return nil
}

// FormatDispatchApproval renders the approval artifact as deterministic
// Markdown suitable for the user prompt and implementation context.
func FormatDispatchApproval(approval DispatchApproval) (string, error) {
	if err := approval.Validate(); err != nil {
		return "", err
	}
	var b strings.Builder
	status := "pending"
	if approval.Approved {
		status = "approved"
	}
	fmt.Fprintf(&b, "Plan: %s\nApproval: %s\n\n", approval.Plan, status)
	b.WriteString("### Ticket dispatch\n\n")
	b.WriteString("| Ticket | Scope | Dependencies | Role | Profile | Resolved model | Reasoning | Estimate | Approved budget | Concurrency | Isolation | Fix reserve | Account quota | Monetary cost |\n")
	b.WriteString("|---|---|---|---|---|---|---|---:|---:|---|---|---:|---|---|\n")
	for _, ticket := range approval.TicketPlan.Tickets {
		dependencies := strings.Join(ticket.Dependencies, ", ")
		if dependencies == "" {
			dependencies = "none"
		}
		eligibility := "eligible"
		if len(ticket.Dependencies) > 0 {
			eligibility = "after " + dependencies
		}
		concurrency := fmt.Sprintf("%s; global <= %d", eligibility, approval.TicketPlan.GlobalConcurrencyLimit)
		if ticket.ConcurrencyLimit > 0 {
			concurrency += fmt.Sprintf("; role <= %d", ticket.ConcurrencyLimit)
		}
		if classLimit, ok := approval.TicketPlan.ClassConcurrencyLimits[ticket.ConcurrencyClass]; ok {
			concurrency += fmt.Sprintf("; class <= %d", classLimit)
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			markdownCell(ticket.TicketID), markdownCell(ticket.Scope), markdownCell(dependencies), markdownCell(ticket.Role.Name), markdownCell(ticket.ModelProfile), markdownCell(ticket.ResolvedModel), markdownCell(ticket.Reasoning), formatRange(ticket.EstimatedTokens), formatRange(ticket.TokenBudget), markdownCell(concurrency), markdownCell(ticket.Isolation), formatRange(ticket.ReservedFixCapacity), formatStatus(ticket.AccountQuota), formatStatus(ticket.MonetaryCost))
	}
	fmt.Fprintf(&b, "\nGlobal implementation concurrency: %d\n", approval.TicketPlan.GlobalConcurrencyLimit)
	b.WriteString("\n### Workflow overhead\n\n")
	b.WriteString("| Work | Approved budget |\n|---|---:|\n")
	workflow := []struct {
		name  string
		value TokenRange
	}{
		{"explorer", approval.WorkflowBudget.Explorer},
		{"merger", approval.WorkflowBudget.Merger},
		{"reviewer", approval.WorkflowBudget.Reviewer},
		{"red-team", approval.WorkflowBudget.RedTeam},
		{"fixes", approval.WorkflowBudget.Fixes},
		{"CI", approval.WorkflowBudget.CI},
		{"integration", approval.WorkflowBudget.Integration},
	}
	for _, item := range workflow {
		fmt.Fprintf(&b, "| %s | %s |\n", item.name, formatRange(item.value))
	}
	if len(approval.OptionalRoles) > 0 {
		b.WriteString("\n### Optional roles\n\n| Role | Status | Detail |\n|---|---|---|\n")
		for _, role := range approval.OptionalRoles {
			fmt.Fprintf(&b, "| %s | %s | %s |\n", markdownCell(role.Name), markdownCell(role.Status), markdownCell(role.Detail))
		}
	}
	return b.String(), nil
}

func formatRange(value TokenRange) string { return fmt.Sprintf("%d–%d", value.Min, value.Max) }

func formatStatus(value CostStatus) string {
	if value.Detail == "" {
		return markdownCell(value.Status)
	}
	return markdownCell(value.Status + ": " + value.Detail)
}

func markdownCell(value string) string { return strings.ReplaceAll(value, "|", "\\|") }

// TicketDispatchPlan is the complete approved implementation dispatch plan.
// Global, role, and class limits are intentionally separate so their
// constraints cannot be mistaken for one another.
type TicketDispatchPlan struct {
	Tickets                []TicketDispatchAssignment
	GlobalConcurrencyLimit int
	RoleConcurrencyLimits  map[string]int
	ClassConcurrencyLimits map[string]int
	// TransitionPolicy defaults to serialized. A parallel policy must be
	// explicitly approved because merges and dependency transitions are
	// otherwise serialized workflow boundaries.
	TransitionPolicy string
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
		TransitionPolicy:       "serialized",
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
