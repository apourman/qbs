package agents_test

import (
	"strings"
	"testing"

	"github.com/trues/qbs/internal/agents"
)

func TestImplementSpecWorkflowUsesApprovalGateAndHandoff(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	role := agents.Role{
		Name: "implementer", Capability: "implementation", Isolation: "worktree",
		Reasoning: "low", ConcurrencyLimit: 2,
		EstimatedTokens: agents.TokenRange{Min: 100, Max: 200}, TokenBudget: agents.TokenRange{Min: 200, Max: 400},
		ReservedBudget: agents.ReservedBudget{Fixes: agents.TokenRange{Min: 50, Max: 80}},
		AccountQuota:   agents.CostStatus{Status: "advisory"}, MonetaryCost: agents.CostStatus{Status: "unavailable"},
	}
	preflight, err := agents.PrepareImplementSpec(agents.ImplementSpecInput{
		Catalog: catalog, Harness: agents.Codex, Plan: agents.Economy,
		Tickets:       []agents.Ticket{{ID: "01", Scope: "implementation", Role: role, Risk: "high", ConcurrencyClass: "implementation"}},
		ClassLimits:   map[string]int{"implementation": 1},
		OptionalRoles: []agents.OptionalRoleStatus{{Name: "explorer", Status: "omitted", Detail: "not required"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if preflight.Approval.Approved || !strings.Contains(preflight.Table, "Assignment note") {
		t.Fatalf("preflight was not pending or did not render assignment data: %+v\n%s", preflight.Approval, preflight.Table)
	}
	assignment := preflight.Approval.TicketPlan.Tickets[0]
	if assignment.ResolvedModel != "gpt-5.6-luna" || assignment.Reasoning != "medium" || assignment.ExceptionReason == "" || assignment.ConcurrencyLimit != 1 {
		t.Fatalf("economy assignment did not reflect ticket risk and effective limit: %+v", assignment)
	}
	if !strings.Contains(preflight.Table, "high-risk ticket requires at least medium reasoning") || !strings.Contains(preflight.Table, "effective <= 1; global <= 2; role <= 2; class <= 1") {
		t.Fatalf("approval table omitted resolved exception or effective limits:\n%s", preflight.Table)
	}

	context, err := agents.ApproveImplementSpec(preflight, "preflight-approved")
	if err != nil {
		t.Fatal(err)
	}
	request := agents.TicketDispatchRequest{
		TicketID: "01", RoleName: "implementer", ModelProfile: "balanced", ResolvedModel: "gpt-5.6-luna",
		Reasoning: "medium", TokenBudget: agents.TokenRange{Min: 200, Max: 400},
		ConcurrencyClass: "implementation", Isolation: "worktree",
	}
	if _, err := agents.AuthorizeImplementSpecDispatch(context, request, agents.TicketDispatchState{}); err != nil {
		t.Fatalf("workflow dispatch seam rejected approved request: %v", err)
	}
	handoff, err := agents.FormatImplementSpecHandoff(context, "spec.md", "codex/spec", "abc123", nil, nil, nil, nil, nil, nil)
	if err != nil || !strings.Contains(handoff, "gpt-5.6-luna") {
		t.Fatalf("workflow handoff did not preserve approved plan: %v\n%s", err, handoff)
	}
}

func TestImplementSpecRejectsParallelTransitions(t *testing.T) {
	approval := agents.DispatchApproval{
		Plan:     agents.Economy,
		Approved: true,
		TicketPlan: agents.TicketDispatchPlan{
			GlobalConcurrencyLimit: 1,
			TransitionPolicy:       "parallel",
			Tickets:                []agents.TicketDispatchAssignment{{TicketID: "01", Role: agents.Role{Name: "implementer"}}},
		},
	}
	if _, err := approval.AuthorizeTicketDispatch("01", agents.TicketDispatchState{}); err == nil || !strings.Contains(err.Error(), "unsupported transition policy") {
		t.Fatalf("parallel transition policy was accepted: %v", err)
	}
}
