package agents_test

import (
	"strings"
	"testing"

	"github.com/trues/qbs/internal/agents"
)

func TestResolvePlansByRoleAndHarness(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	roles := []agents.Role{
		{Name: "explorer", Capability: "exploration", Reasoning: "medium"},
		{Name: "implementer", Capability: "implementation", Reasoning: "high"},
	}
	assignments, err := catalog.ResolvePlan(roles, agents.Codex, agents.Recommended, nil)
	if err != nil {
		t.Fatal(err)
	}
	if assignments[0].Role.Name != "explorer" || assignments[0].ResolvedModel != "gpt-5.6-luna" {
		t.Fatalf("unexpected explorer assignment: %+v", assignments[0])
	}
	if assignments[1].ModelProfile != "capable" || assignments[1].ResolvedModel != "gpt-5.6-sol" {
		t.Fatalf("unexpected implementer assignment: %+v", assignments[1])
	}

	assignments, err = catalog.ResolvePlan(roles, agents.Claude, agents.Economy, nil)
	if err != nil {
		t.Fatal(err)
	}
	if assignments[0].ModelProfile != "balanced" || assignments[0].ResolvedModel != "sonnet" || assignments[0].Reasoning != "low" {
		t.Fatalf("unexpected economy explorer assignment: %+v", assignments[0])
	}
	if assignments[1].ModelProfile != "balanced" || assignments[1].ResolvedModel != "sonnet" || assignments[1].Reasoning != "low" {
		t.Fatalf("unexpected economy implementer assignment: %+v", assignments[1])
	}
}

func TestResolveDeepAndCustomize(t *testing.T) {
	catalog, _ := agents.Canonical()
	roles := []agents.Role{{Name: "merge", Capability: "integration"}}
	deep, err := catalog.ResolvePlan(roles, agents.OpenCode, agents.Deep, nil)
	if err != nil {
		t.Fatal(err)
	}
	if deep[0].ResolvedModel != "openai/gpt-5.6-sol" || deep[0].Reasoning != "high" {
		t.Fatalf("unexpected deep assignment: %+v", deep[0])
	}
	custom, err := catalog.ResolvePlan(roles, agents.Claude, agents.Customize, map[string]agents.Assignment{"merge": {ModelProfile: "balanced", Reasoning: "medium"}})
	if err != nil {
		t.Fatal(err)
	}
	if custom[0].ResolvedModel != "sonnet" || custom[0].Reasoning != "medium" {
		t.Fatalf("unexpected custom assignment: %+v", custom[0])
	}
}

func TestDispatchCarriesBudgetsAndSeparateCostStatuses(t *testing.T) {
	catalog, _ := agents.Canonical()
	role := agents.Role{Name: "review", Capability: "adversarial-review", TokenBudget: agents.TokenRange{Min: 100, Max: 200}, ConcurrencyLimit: 1, AccountQuota: agents.CostStatus{Status: "advisory"}, MonetaryCost: agents.CostStatus{Status: "unavailable"}, ReservedBudget: agents.ReservedBudget{Fixes: agents.TokenRange{Min: 20, Max: 40}}}
	assignments, err := catalog.ResolvePlan([]agents.Role{role}, agents.Codex, agents.Economy, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := assignments[0]
	if got.TokenBudget != role.TokenBudget || got.ConcurrencyLimit != 1 || got.AccountQuota.Status != "advisory" || got.MonetaryCost.Status != "unavailable" || got.ReservedBudget.Fixes.Max != 40 {
		t.Fatalf("budget metadata not preserved: %+v", got)
	}
}

func TestCustomReasoningMustBeSupported(t *testing.T) {
	catalog, _ := agents.Canonical()
	_, err := catalog.ResolvePlan([]agents.Role{{Name: "worker"}}, agents.Codex, agents.Customize, map[string]agents.Assignment{"worker": {ModelProfile: "balanced", Reasoning: "turbo"}})
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("error = %v", err)
	}
}

func TestResolveUnavailableProfileFailsWithoutFallback(t *testing.T) {
	catalog, _ := agents.Canonical()
	_, err := catalog.ResolvePlan([]agents.Role{{Name: "worker"}}, agents.Codex, agents.Customize, map[string]agents.Assignment{"worker": {ModelProfile: "fast", Reasoning: "low"}})
	if err == nil || !strings.Contains(err.Error(), `profile "fast" for harness "codex": profile is unavailable`) {
		t.Fatalf("error = %v", err)
	}
}

func TestResolveTicketPlanCarriesPerTicketApprovalData(t *testing.T) {
	catalog, _ := agents.Canonical()
	role := agents.Role{
		Name:             "implementer",
		Capability:       "implementation",
		Isolation:        "worktree",
		EstimatedTokens:  agents.TokenRange{Min: 100, Max: 150},
		TokenBudget:      agents.TokenRange{Min: 200, Max: 300},
		ConcurrencyLimit: 1,
		AccountQuota:     agents.CostStatus{Status: "advisory", Detail: "measured"},
		MonetaryCost:     agents.CostStatus{Status: "unavailable"},
		ReservedBudget:   agents.ReservedBudget{Fixes: agents.TokenRange{Min: 30, Max: 50}},
	}
	plan, err := catalog.ResolveTicketPlanWithLimits([]agents.Ticket{
		{ID: "02", Scope: "approval table", Dependencies: []string{"01"}, Role: role, ConcurrencyClass: "implementation"},
		{ID: "01", Scope: "dispatch model", Role: role, ConcurrencyClass: "implementation"},
	}, agents.Codex, agents.Customize, map[string]agents.Assignment{
		"01": {ModelProfile: "balanced", Reasoning: "medium"},
		"02": {ModelProfile: "capable", Reasoning: "high"},
	}, 2, map[string]int{"implementer": 1}, map[string]int{"implementation": 2})
	if err != nil {
		t.Fatal(err)
	}
	if plan.GlobalConcurrencyLimit != 2 || plan.RoleConcurrencyLimits["implementer"] != 1 || plan.ClassConcurrencyLimits["implementation"] != 2 {
		t.Fatalf("unexpected concurrency limits: %+v", plan)
	}
	if len(plan.Tickets) != 2 || plan.Tickets[0].TicketID != "01" || plan.Tickets[1].TicketID != "02" {
		t.Fatalf("tickets were not returned in stable order: %+v", plan.Tickets)
	}
	got := plan.Tickets[1]
	if got.ResolvedModel != "gpt-5.6-sol" || got.ModelProfile != "capable" || got.Reasoning != "high" || got.Scope != "approval table" || len(got.Dependencies) != 1 || got.Dependencies[0] != "01" {
		t.Fatalf("ticket assignment not resolved: %+v", got)
	}
	if got.EstimatedTokens != role.EstimatedTokens || got.TokenBudget != role.TokenBudget || got.Isolation != "worktree" || got.ConcurrencyClass != "implementation" || got.ReservedFixCapacity.Max != 50 || got.AccountQuota.Status != "advisory" || got.MonetaryCost.Status != "unavailable" {
		t.Fatalf("ticket approval metadata not preserved: %+v", got)
	}
}

func TestResolveTicketPlanFailsUnavailableProfileWithoutFallback(t *testing.T) {
	catalog, _ := agents.Canonical()
	_, err := catalog.ResolveTicketPlan([]agents.Ticket{{ID: "01", Role: agents.Role{Name: "implementer"}}}, agents.Codex, agents.Customize, map[string]agents.Assignment{"01": {ModelProfile: "fast", Reasoning: "medium"}})
	if err == nil || !strings.Contains(err.Error(), `profile "fast" for harness "codex": profile is unavailable`) {
		t.Fatalf("error = %v", err)
	}
}

func TestFormatDispatchApprovalShowsTicketRowsAndWorkflowOverhead(t *testing.T) {
	catalog, _ := agents.Canonical()
	role := agents.Role{Name: "implementer", Isolation: "worktree", ConcurrencyLimit: 1, TokenBudget: agents.TokenRange{Min: 200, Max: 300}, EstimatedTokens: agents.TokenRange{Min: 100, Max: 150}, ReservedBudget: agents.ReservedBudget{Fixes: agents.TokenRange{Min: 30, Max: 50}}}
	plan, err := catalog.ResolveTicketPlanWithLimits([]agents.Ticket{
		{ID: "02", Scope: "approval | table", Dependencies: []string{"01"}, Role: role, ConcurrencyClass: "implementation"},
		{ID: "01", Scope: "dispatch model", Role: role, ConcurrencyClass: "implementation"},
	}, agents.Codex, agents.Customize, map[string]agents.Assignment{
		"01": {ModelProfile: "balanced", Reasoning: "medium"},
		"02": {ModelProfile: "capable", Reasoning: "high"},
	}, 2, map[string]int{"implementer": 1}, map[string]int{"implementation": 2})
	if err != nil {
		t.Fatal(err)
	}
	output, err := agents.FormatDispatchApproval(agents.DispatchApproval{
		Plan:       agents.Customize,
		TicketPlan: plan,
		WorkflowBudget: agents.WorkflowBudget{
			Merger:      agents.TokenRange{Min: 100, Max: 200},
			Reviewer:    agents.TokenRange{Min: 150, Max: 250},
			Fixes:       agents.TokenRange{Min: 50, Max: 100},
			Integration: agents.TokenRange{Min: 75, Max: 125},
		},
		OptionalRoles: []agents.OptionalRoleStatus{{Name: "explorer", Status: "omitted", Detail: "not required"}},
		Approved:      true,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"Plan: Customize",
		"Approval: approved",
		"| 01 | dispatch model | none |",
		"| 02 | approval \\| table | 01 |",
		"gpt-5.6-sol",
		"medium",
		"high",
		"100–200",
		"### Workflow overhead",
		"| explorer | 0–0 |",
		"| merger | 100–200 |",
		"### Optional roles",
		"| explorer | omitted | not required |",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("approval output missing %q:\n%s", expected, output)
		}
	}
}

func TestFormatDispatchApprovalRejectsInvalidOptionalRoleStatus(t *testing.T) {
	_, err := agents.FormatDispatchApproval(agents.DispatchApproval{
		Plan:       agents.Economy,
		TicketPlan: agents.TicketDispatchPlan{Tickets: []agents.TicketDispatchAssignment{{TicketID: "01"}}},
		OptionalRoles: []agents.OptionalRoleStatus{{
			Name:   "explorer",
			Status: "maybe",
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported status") {
		t.Fatalf("error = %v", err)
	}
}
