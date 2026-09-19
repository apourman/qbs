package agents

import "fmt"

// ImplementSpecInput contains the concrete inputs needed to turn a spec's
// ticket graph into an approval artifact. It is the workflow boundary between
// implement-spec's ticket interpretation and the dispatch gate.
type ImplementSpecInput struct {
	Catalog           Catalog
	Harness           Harness
	Plan              Plan
	Tickets           []Ticket
	CustomAssignments map[string]Assignment
	GlobalLimit       int
	RoleLimits        map[string]int
	ClassLimits       map[string]int
	WorkflowBudget    WorkflowBudget
	OptionalRoles     []OptionalRoleStatus
	ReviewScope       string
}

// ImplementSpecPreflight is the pending, rendered approval shown before any
// implementation ticket is dispatched.
type ImplementSpecPreflight struct {
	Approval DispatchApproval
	Table    string
}

// PrepareImplementSpec resolves every known ticket, renders its approval
// table, and leaves the approval pending until the user accepts it.
func PrepareImplementSpec(input ImplementSpecInput) (ImplementSpecPreflight, error) {
	if input.GlobalLimit == 0 {
		input.GlobalLimit = 2
	}
	plan, err := input.Catalog.ResolveTicketPlanWithLimits(input.Tickets, input.Harness, input.Plan, input.CustomAssignments, input.GlobalLimit, input.RoleLimits, input.ClassLimits)
	if err != nil {
		return ImplementSpecPreflight{}, fmt.Errorf("prepare implement-spec: %w", err)
	}
	approval := DispatchApproval{
		Plan:           input.Plan,
		TicketPlan:     plan,
		WorkflowBudget: input.WorkflowBudget,
		OptionalRoles:  append([]OptionalRoleStatus(nil), input.OptionalRoles...),
		ReviewScope:    input.ReviewScope,
	}
	table, err := FormatDispatchApproval(approval)
	if err != nil {
		return ImplementSpecPreflight{}, fmt.Errorf("prepare implement-spec approval: %w", err)
	}
	return ImplementSpecPreflight{Approval: approval, Table: table}, nil
}

// ApproveImplementSpec records the user's approval and snapshots the exact
// plan that later dispatch, merge, review, and handoff phases must use.
func ApproveImplementSpec(preflight ImplementSpecPreflight, eventName string) (ImplementationContext, error) {
	approval := cloneDispatchApproval(preflight.Approval)
	approval.Approved = true
	return NewImplementationContext(approval, eventName)
}

// AuthorizeImplementSpecDispatch is the workflow's single dispatch seam. It
// prevents callers from bypassing the approved row or changing invocation
// parameters between approval and harness execution.
func AuthorizeImplementSpecDispatch(context ImplementationContext, request TicketDispatchRequest, state TicketDispatchState) (TicketDispatchAuthorization, error) {
	return context.Current.AuthorizeTicketDispatchRequest(request, state)
}

// FormatImplementSpecHandoff renders the current approved plan for the final
// merger, review, fixes, integration, and postflight handoff.
func FormatImplementSpecHandoff(context ImplementationContext, specPath, specBranch, finalRevision string, coverage []RequirementResult, checks, conclusions, findings []Evidence, gates, risks []string) (string, error) {
	return FormatReviewHandoff(ReviewHandoff{
		SpecPath:            specPath,
		SpecBranch:          specBranch,
		FinalRevision:       finalRevision,
		Context:             context,
		RequirementCoverage: coverage,
		VerifiedChecks:      checks,
		InferredConclusions: conclusions,
		ReviewFindings:      findings,
		RemainingGates:      gates,
		Risks:               risks,
	})
}
