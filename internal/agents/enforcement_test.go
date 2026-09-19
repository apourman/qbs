package agents_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/trues/qbs/internal/agents"
)

func TestAuthorizeTicketDispatchRequiresApprovedRowAndUsesExactAssignment(t *testing.T) {
	approval := testEnforcementApproval()
	if _, err := approval.AuthorizeTicketDispatch("01", agents.TicketDispatchState{}); err == nil || !strings.Contains(err.Error(), "not approved") {
		t.Fatalf("unapproved dispatch error = %v", err)
	}

	approval.Approved = true
	authorization, err := approval.AuthorizeTicketDispatch("01", agents.TicketDispatchState{})
	if err != nil {
		t.Fatal(err)
	}
	if authorization.Assignment.ResolvedModel != "gpt-5.6-luna" || authorization.Assignment.Reasoning != "high" || authorization.Assignment.Isolation != "worktree" || authorization.TokenLimit != 300 {
		t.Fatalf("authorization did not preserve the approved row: %+v", authorization)
	}

	request := agents.TicketDispatchRequest{
		TicketID:         "01",
		RoleName:         "implementer",
		ModelProfile:     "balanced",
		ResolvedModel:    "gpt-5.6-luna",
		Reasoning:        "high",
		TokenBudget:      agents.TokenRange{Min: 200, Max: 300},
		ConcurrencyClass: "implementation",
		Isolation:        "worktree",
	}
	if err := approval.ValidateTicketDispatchRequest(request); err != nil {
		t.Fatalf("exact approved request rejected: %v", err)
	}

	request.Reasoning = "low"
	var gate *agents.DispatchGateError
	if err := approval.ValidateTicketDispatchRequest(request); !errors.As(err, &gate) || !gate.RenewalRequired || gate.Constraint != "reasoning" {
		t.Fatalf("reasoning change should require renewal, error = %v", err)
	}
}

func TestAuthorizeTicketDispatchEnforcesDependenciesAndConcurrency(t *testing.T) {
	approval := testEnforcementApproval()
	approval.Approved = true

	if _, err := approval.AuthorizeTicketDispatch("02", agents.TicketDispatchState{}); err == nil || !strings.Contains(err.Error(), "dependency frontier") {
		t.Fatalf("incomplete dependency should block dispatch, error = %v", err)
	}

	state := agents.TicketDispatchState{
		Active:    []agents.ActiveTicket{{TicketID: "01", RoleName: "implementer", ConcurrencyClass: "implementation"}},
		Completed: map[string]bool{"01": true},
	}
	if _, err := approval.AuthorizeTicketDispatch("02", state); err == nil || !strings.Contains(err.Error(), "role concurrency") {
		t.Fatalf("role limit should block dispatch, error = %v", err)
	}

	approval.TicketPlan.RoleConcurrencyLimits["implementer"] = 2
	approval.TicketPlan.ClassConcurrencyLimits["implementation"] = 1
	if _, err := approval.AuthorizeTicketDispatch("02", state); err == nil || !strings.Contains(err.Error(), "class concurrency") {
		t.Fatalf("class limit should block dispatch, error = %v", err)
	}

	approval.TicketPlan.ClassConcurrencyLimits["implementation"] = 2
	approval.TicketPlan.GlobalConcurrencyLimit = 1
	if _, err := approval.AuthorizeTicketDispatch("02", state); err == nil || !strings.Contains(err.Error(), "global concurrency") {
		t.Fatalf("global limit should block dispatch, error = %v", err)
	}
}

func TestAuthorizeTicketDispatchSerializesTransitionsAndRequiresRenewalForNewTicket(t *testing.T) {
	approval := testEnforcementApproval()
	approval.Approved = true
	state := agents.TicketDispatchState{Completed: map[string]bool{"01": true}, TransitionInProgress: true}
	if _, err := approval.AuthorizeTicketDispatch("02", state); err == nil || !strings.Contains(err.Error(), "serialized transition") {
		t.Fatalf("transition should be serialized, error = %v", err)
	}

	if _, err := approval.AuthorizeTicketDispatch("03", agents.TicketDispatchState{}); err == nil || !strings.Contains(err.Error(), "renewed approval required") {
		t.Fatalf("newly discovered ticket should require renewed approval, error = %v", err)
	}
}

func TestDispatchPlanRejectsBudgetOptionalRoleAndReviewScopeChanges(t *testing.T) {
	approval := testEnforcementApproval()
	approval.Approved = true
	approval.ReviewScope = "ticket-only"
	request := agents.TicketDispatchRequest{
		TicketID:         "01",
		RoleName:         "implementer",
		ModelProfile:     "balanced",
		ResolvedModel:    "gpt-5.6-luna",
		Reasoning:        "high",
		TokenBudget:      agents.TokenRange{Min: 200, Max: 300},
		TokenUsage:       301,
		ConcurrencyClass: "implementation",
		Isolation:        "worktree",
		ReviewScope:      "full-review",
	}
	if err := approval.ValidateTicketDispatchRequest(request); err == nil || !strings.Contains(err.Error(), "token budget") {
		t.Fatalf("budget overrun should be gated, error = %v", err)
	}
	if err := approval.ValidateTicketUsage("01", 301); err == nil || !strings.Contains(err.Error(), "token budget") {
		t.Fatalf("reported budget overrun should be gated, error = %v", err)
	}
	request.TokenUsage = 1
	request.ReviewScope = "ticket-only"
	if err := approval.ValidateTicketDispatchRequest(request); err != nil {
		t.Fatalf("valid budget and review scope rejected: %v", err)
	}
	request.ReviewScope = "full-review"
	if err := approval.ValidateTicketDispatchRequest(request); err == nil || !strings.Contains(err.Error(), "review scope") {
		t.Fatalf("review expansion should require renewal, error = %v", err)
	}
	if err := approval.AuthorizeOptionalRole("red-team"); err == nil || !strings.Contains(err.Error(), "optional role") {
		t.Fatalf("unapproved optional role should require renewal, error = %v", err)
	}
}

func testEnforcementApproval() agents.DispatchApproval {
	role := agents.Role{Name: "implementer", Capability: "implementation", Isolation: "worktree", ConcurrencyLimit: 1}
	return agents.DispatchApproval{
		Plan: agents.Economy,
		TicketPlan: agents.TicketDispatchPlan{
			Tickets: []agents.TicketDispatchAssignment{
				{
					TicketID: "01", Scope: "dispatch model", Role: role, ModelProfile: "balanced", ResolvedModel: "gpt-5.6-luna", Reasoning: "high",
					TokenBudget: agents.TokenRange{Min: 200, Max: 300}, ConcurrencyClass: "implementation", ConcurrencyLimit: 1, Isolation: "worktree",
				},
				{
					TicketID: "02", Scope: "approval table", Dependencies: []string{"01"}, Role: role, ModelProfile: "balanced", ResolvedModel: "gpt-5.6-luna", Reasoning: "medium",
					TokenBudget: agents.TokenRange{Min: 100, Max: 200}, ConcurrencyClass: "implementation", ConcurrencyLimit: 1, Isolation: "worktree",
				},
			},
			GlobalConcurrencyLimit: 2,
			RoleConcurrencyLimits:  map[string]int{"implementer": 1},
			ClassConcurrencyLimits: map[string]int{"implementation": 2},
		},
	}
}
