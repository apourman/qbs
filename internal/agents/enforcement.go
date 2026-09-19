package agents

import "fmt"

// ActiveTicket describes an implementation ticket that is currently running.
// The execution state is supplied by the orchestration layer so the dispatch
// gate can enforce aggregate limits without owning agent lifecycle state.
type ActiveTicket struct {
	TicketID         string
	RoleName         string
	ConcurrencyClass string
}

// TicketDispatchState is the execution state consulted before starting a
// ticket. Completed dependencies are represented by ID; active tickets are
// used for global, role, and class concurrency accounting.
type TicketDispatchState struct {
	Active               []ActiveTicket
	Completed            map[string]bool
	TransitionInProgress bool
	MergeInProgress      bool
}

// TicketDispatchRequest is the concrete invocation the harness is about to
// start. Every constrained field must match the approved ticket row exactly.
// TokenUsage is optional at dispatch time and can be checked incrementally by
// ValidateTicketUsage.
type TicketDispatchRequest struct {
	TicketID         string
	RoleName         string
	ModelProfile     string
	ResolvedModel    string
	Reasoning        string
	TokenBudget      TokenRange
	TokenUsage       int
	ConcurrencyClass string
	Isolation        string
	OptionalRole     string
	ReviewScope      string
}

// TicketDispatchAuthorization is the approved row returned for an allowed
// invocation. The token limit is the row's approved maximum, not an inferred
// or silently upgraded budget.
type TicketDispatchAuthorization struct {
	Assignment TicketDispatchAssignment
	TokenLimit int
}

// DispatchGateError indicates that execution must stop at an approval gate.
// RenewalRequired distinguishes ordinary frontier blocking from a material
// change that needs renewed user approval.
type DispatchGateError struct {
	TicketID        string
	Constraint      string
	Detail          string
	RenewalRequired bool
}

func (e *DispatchGateError) Error() string {
	prefix := "dispatch blocked"
	if e.RenewalRequired {
		prefix = "renewed approval required"
	}
	if e.TicketID == "" {
		return fmt.Sprintf("%s: %s: %s", prefix, e.Constraint, e.Detail)
	}
	return fmt.Sprintf("%s for ticket %q: %s: %s", prefix, e.TicketID, e.Constraint, e.Detail)
}

func gateError(ticketID, constraint, detail string, renewal bool) error {
	return &DispatchGateError{TicketID: ticketID, Constraint: constraint, Detail: detail, RenewalRequired: renewal}
}

// AuthorizeTicketDispatch checks the approved table and execution frontier
// before returning the exact row that must be passed to the harness.
func (a DispatchApproval) AuthorizeTicketDispatch(ticketID string, state TicketDispatchState) (TicketDispatchAuthorization, error) {
	if !a.Approved {
		return TicketDispatchAuthorization{}, gateError(ticketID, "approval", "the dispatch table is not approved", false)
	}
	if err := a.TicketPlan.validateGraph(); err != nil {
		return TicketDispatchAuthorization{}, err
	}
	assignment, ok := a.TicketPlan.assignment(ticketID)
	if !ok {
		return TicketDispatchAuthorization{}, gateError(ticketID, "ticket row", "no approved row exists for this ticket", true)
	}
	if assignment.Role.Name == "" || assignment.ModelProfile == "" || assignment.ResolvedModel == "" || assignment.Reasoning == "" || assignment.Isolation == "" || assignment.TokenBudget.Min < 0 || assignment.TokenBudget.Max <= 0 || assignment.TokenBudget.Min > assignment.TokenBudget.Max {
		return TicketDispatchAuthorization{}, gateError(ticketID, "ticket row", "the approved row is incomplete", true)
	}
	if assignment.Role.Optional && !a.optionalRoleSelected(assignment.Role.Name) {
		return TicketDispatchAuthorization{}, gateError(ticketID, "optional role", fmt.Sprintf("role %q is not selected", assignment.Role.Name), true)
	}

	if a.TicketPlan.transitionPolicy() == "serialized" && (state.TransitionInProgress || state.MergeInProgress) {
		return TicketDispatchAuthorization{}, gateError(ticketID, "serialized transition", "a merge or dependency transition is already in progress", false)
	}

	activeIDs := make(map[string]bool, len(state.Active))
	roleCount := make(map[string]int)
	classCount := make(map[string]int)
	for _, active := range state.Active {
		if activeIDs[active.TicketID] {
			return TicketDispatchAuthorization{}, gateError(ticketID, "execution state", fmt.Sprintf("ticket %q is already active more than once", active.TicketID), false)
		}
		activeIDs[active.TicketID] = true
		activeAssignment, found := a.TicketPlan.assignment(active.TicketID)
		if !found {
			return TicketDispatchAuthorization{}, gateError(active.TicketID, "ticket row", "an active ticket has no approved row", true)
		}
		if active.RoleName != "" && active.RoleName != activeAssignment.Role.Name {
			return TicketDispatchAuthorization{}, gateError(active.TicketID, "role", "active state does not match the approved row", true)
		}
		if active.ConcurrencyClass != "" && active.ConcurrencyClass != activeAssignment.ConcurrencyClass {
			return TicketDispatchAuthorization{}, gateError(active.TicketID, "concurrency class", "active state does not match the approved row", true)
		}
		roleCount[activeAssignment.Role.Name]++
		classCount[activeAssignment.ConcurrencyClass]++
	}
	if activeIDs[ticketID] {
		return TicketDispatchAuthorization{}, gateError(ticketID, "execution state", "ticket is already active", false)
	}
	if len(state.Active) >= a.TicketPlan.GlobalConcurrencyLimit {
		return TicketDispatchAuthorization{}, gateError(ticketID, "global concurrency", fmt.Sprintf("%d active ticket(s) already use the limit of %d", len(state.Active), a.TicketPlan.GlobalConcurrencyLimit), false)
	}
	if limit := a.TicketPlan.roleLimit(assignment); limit > 0 && roleCount[assignment.Role.Name] >= limit {
		return TicketDispatchAuthorization{}, gateError(ticketID, "role concurrency", fmt.Sprintf("role %q already has %d active ticket(s); limit is %d", assignment.Role.Name, roleCount[assignment.Role.Name], limit), false)
	}
	if limit := a.TicketPlan.classLimit(assignment); limit > 0 && classCount[assignment.ConcurrencyClass] >= limit {
		return TicketDispatchAuthorization{}, gateError(ticketID, "class concurrency", fmt.Sprintf("class %q already has %d active ticket(s); limit is %d", assignment.ConcurrencyClass, classCount[assignment.ConcurrencyClass], limit), false)
	}
	for _, dependency := range assignment.Dependencies {
		if !state.Completed[dependency] {
			return TicketDispatchAuthorization{}, gateError(ticketID, "dependency frontier", fmt.Sprintf("dependency %q is not complete", dependency), false)
		}
	}

	return TicketDispatchAuthorization{Assignment: assignment, TokenLimit: assignment.TokenBudget.Max}, nil
}

// AuthorizeTicketDispatchRequest is the single preflight gate for starting a
// harness invocation. It checks both the execution frontier and the concrete
// invocation, so callers cannot accidentally authorize a row and then run a
// different model or budget.
func (a DispatchApproval) AuthorizeTicketDispatchRequest(request TicketDispatchRequest, state TicketDispatchState) (TicketDispatchAuthorization, error) {
	authorization, err := a.AuthorizeTicketDispatch(request.TicketID, state)
	if err != nil {
		return TicketDispatchAuthorization{}, err
	}
	if err := a.ValidateTicketDispatchRequest(request); err != nil {
		return TicketDispatchAuthorization{}, err
	}
	return authorization, nil
}

// ValidateTicketDispatchRequest prevents the harness invocation from silently
// changing any approved model, reasoning, budget, isolation, role, or scope.
func (a DispatchApproval) ValidateTicketDispatchRequest(request TicketDispatchRequest) error {
	if !a.Approved {
		return gateError(request.TicketID, "approval", "the dispatch table is not approved", false)
	}
	assignment, ok := a.TicketPlan.assignment(request.TicketID)
	if !ok {
		return gateError(request.TicketID, "ticket row", "no approved row exists for this ticket", true)
	}
	checks := []struct {
		constraint string
		approved   string
		requested  string
	}{
		{"role", assignment.Role.Name, request.RoleName},
		{"model profile", assignment.ModelProfile, request.ModelProfile},
		{"resolved model", assignment.ResolvedModel, request.ResolvedModel},
		{"reasoning", assignment.Reasoning, request.Reasoning},
		{"concurrency class", assignment.ConcurrencyClass, request.ConcurrencyClass},
		{"isolation", assignment.Isolation, request.Isolation},
	}
	for _, check := range checks {
		if check.approved != check.requested {
			return gateError(request.TicketID, check.constraint, fmt.Sprintf("approved %q, requested %q", check.approved, check.requested), true)
		}
	}
	if assignment.TokenBudget != request.TokenBudget {
		return gateError(request.TicketID, "token budget", fmt.Sprintf("approved %v, requested %v", assignment.TokenBudget, request.TokenBudget), true)
	}
	if request.TokenUsage < 0 {
		return gateError(request.TicketID, "token budget", "token usage cannot be negative", false)
	}
	if request.TokenUsage > assignment.TokenBudget.Max {
		return gateError(request.TicketID, "token budget", fmt.Sprintf("usage %d exceeds approved maximum %d", request.TokenUsage, assignment.TokenBudget.Max), true)
	}
	if request.OptionalRole != "" {
		if !assignment.Role.Optional || request.OptionalRole != assignment.Role.Name || !a.optionalRoleSelected(request.OptionalRole) {
			return gateError(request.TicketID, "optional role", "the requested optional role is not selected in the approval", true)
		}
	}
	if request.ReviewScope != a.ReviewScope {
		return gateError(request.TicketID, "review scope", fmt.Sprintf("approved %q, requested %q", a.ReviewScope, request.ReviewScope), true)
	}
	return nil
}

// ValidateTicketUsage enforces the approved maximum after dispatch as work
// reports additional token usage.
func (a DispatchApproval) ValidateTicketUsage(ticketID string, used int) error {
	if !a.Approved {
		return gateError(ticketID, "approval", "the dispatch table is not approved", false)
	}
	if used < 0 {
		return gateError(ticketID, "token budget", "token usage cannot be negative", false)
	}
	assignment, ok := a.TicketPlan.assignment(ticketID)
	if !ok {
		return gateError(ticketID, "ticket row", "no approved row exists for this ticket", true)
	}
	if used > assignment.TokenBudget.Max {
		return gateError(ticketID, "token budget", fmt.Sprintf("usage %d exceeds approved maximum %d", used, assignment.TokenBudget.Max), true)
	}
	return nil
}

// AuthorizeOptionalRole ensures an optional role cannot be introduced after
// approval without a renewed approval event.
func (a DispatchApproval) AuthorizeOptionalRole(roleName string) error {
	if !a.Approved {
		return gateError("", "approval", "the dispatch table is not approved", false)
	}
	if !a.optionalRoleSelected(roleName) {
		return gateError("", "optional role", fmt.Sprintf("role %q was not selected in the approved plan", roleName), true)
	}
	return nil
}

func (a DispatchApproval) optionalRoleSelected(roleName string) bool {
	for _, role := range a.OptionalRoles {
		if role.Name == roleName {
			return role.Status == "selected"
		}
	}
	return false
}

func (p TicketDispatchPlan) assignment(ticketID string) (TicketDispatchAssignment, bool) {
	for _, assignment := range p.Tickets {
		if assignment.TicketID == ticketID {
			return assignment, true
		}
	}
	return TicketDispatchAssignment{}, false
}

func (p TicketDispatchPlan) roleLimit(assignment TicketDispatchAssignment) int {
	if limit := p.RoleConcurrencyLimits[assignment.Role.Name]; limit > 0 {
		return limit
	}
	return assignment.ConcurrencyLimit
}

func (p TicketDispatchPlan) classLimit(assignment TicketDispatchAssignment) int {
	return p.ClassConcurrencyLimits[assignment.ConcurrencyClass]
}

func (p TicketDispatchPlan) transitionPolicy() string {
	if p.TransitionPolicy == "parallel" {
		return "parallel"
	}
	return "serialized"
}

func (p TicketDispatchPlan) validateGraph() error {
	if p.GlobalConcurrencyLimit <= 0 {
		return fmt.Errorf("validate ticket dispatch plan: global concurrency limit must be positive")
	}
	if p.TransitionPolicy != "" && p.TransitionPolicy != "serialized" && p.TransitionPolicy != "parallel" {
		return fmt.Errorf("validate ticket dispatch plan: unsupported transition policy %q", p.TransitionPolicy)
	}
	ids := make(map[string]bool, len(p.Tickets))
	for _, assignment := range p.Tickets {
		if assignment.TicketID == "" || ids[assignment.TicketID] {
			return fmt.Errorf("validate ticket dispatch plan: ticket IDs must be non-empty and unique")
		}
		ids[assignment.TicketID] = true
		if limit, ok := p.RoleConcurrencyLimits[assignment.Role.Name]; ok && limit <= 0 {
			return fmt.Errorf("validate ticket dispatch plan: role %q concurrency limit must be positive", assignment.Role.Name)
		}
		if limit, ok := p.ClassConcurrencyLimits[assignment.ConcurrencyClass]; ok && limit <= 0 {
			return fmt.Errorf("validate ticket dispatch plan: class %q concurrency limit must be positive", assignment.ConcurrencyClass)
		}
	}
	for _, assignment := range p.Tickets {
		for _, dependency := range assignment.Dependencies {
			if dependency == assignment.TicketID {
				return fmt.Errorf("validate ticket dispatch plan: ticket %q cannot depend on itself", assignment.TicketID)
			}
			if !ids[dependency] {
				return fmt.Errorf("validate ticket dispatch plan: ticket %q depends on unknown ticket %q", assignment.TicketID, dependency)
			}
		}
	}
	if cycle := dependencyCycle(p.Tickets); cycle != "" {
		return fmt.Errorf("validate ticket dispatch plan: dependency cycle includes ticket %q", cycle)
	}
	return nil
}

func dependencyCycle(tickets []TicketDispatchAssignment) string {
	dependencies := make(map[string][]string, len(tickets))
	for _, ticket := range tickets {
		dependencies[ticket.TicketID] = ticket.Dependencies
	}
	visiting := make(map[string]bool, len(tickets))
	visited := make(map[string]bool, len(tickets))
	var visit func(string) string
	visit = func(ticketID string) string {
		if visiting[ticketID] {
			return ticketID
		}
		if visited[ticketID] {
			return ""
		}
		visiting[ticketID] = true
		for _, dependency := range dependencies[ticketID] {
			if cycle := visit(dependency); cycle != "" {
				return cycle
			}
		}
		delete(visiting, ticketID)
		visited[ticketID] = true
		return ""
	}
	for ticketID := range dependencies {
		if cycle := visit(ticketID); cycle != "" {
			return cycle
		}
	}
	return ""
}
