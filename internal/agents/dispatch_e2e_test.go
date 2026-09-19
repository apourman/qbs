package agents_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/trues/qbs/internal/agents"
)

func TestEndToEndTicketPlansCoverEveryPlanChoiceAndHarness(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name           string
		harness        agents.Harness
		plan           agents.Plan
		custom         map[string]agents.Assignment
		resolvedModels map[string]string
		reasoning      map[string]string
	}{
		{
			name:    "economy codex uses Luna",
			harness: agents.Codex,
			plan:    agents.Economy,
			resolvedModels: map[string]string{
				"01": "gpt-5.6-luna", "02": "gpt-5.6-luna", "03": "gpt-5.6-luna", "04": "gpt-5.6-luna",
			},
			reasoning: map[string]string{"01": "low", "02": "low", "03": "low", "04": "low"},
		},
		{
			name:    "recommended codex varies by capability",
			harness: agents.Codex,
			plan:    agents.Recommended,
			resolvedModels: map[string]string{
				"01": "gpt-5.6-luna", "02": "gpt-5.6-sol", "03": "gpt-5.6-sol", "04": "gpt-5.6-luna",
			},
			reasoning: map[string]string{"01": "low", "02": "medium", "03": "low", "04": "low"},
		},
		{
			name:           "deep opencode resolves every row",
			harness:        agents.OpenCode,
			plan:           agents.Deep,
			resolvedModels: map[string]string{"01": "openai/gpt-5.6-sol", "02": "openai/gpt-5.6-sol", "03": "openai/gpt-5.6-sol", "04": "openai/gpt-5.6-sol"},
			reasoning:      map[string]string{"01": "high", "02": "high", "03": "high", "04": "high"},
		},
		{
			name:    "customize claude resolves per ticket",
			harness: agents.Claude,
			plan:    agents.Customize,
			custom: map[string]agents.Assignment{
				"01": {ModelProfile: "balanced", Reasoning: "low"},
				"02": {ModelProfile: "capable", Reasoning: "high"},
				"03": {ModelProfile: "balanced", Reasoning: "medium"},
				"04": {ModelProfile: "balanced", Reasoning: "medium"},
			},
			resolvedModels: map[string]string{"01": "sonnet", "02": "opus", "03": "sonnet", "04": "sonnet"},
			reasoning:      map[string]string{"01": "low", "02": "high", "03": "medium", "04": "medium"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			plan, err := catalog.ResolveTicketPlanWithLimits(
				e2eTickets(), testCase.harness, testCase.plan, testCase.custom, 2,
				map[string]int{"mechanical": 1, "implementer": 1, "architect": 1},
				map[string]int{"implementation": 2},
			)
			if err != nil {
				t.Fatal(err)
			}
			if plan.TransitionPolicy != "serialized" || plan.GlobalConcurrencyLimit != 2 {
				t.Fatalf("plan limits were not preserved: %+v", plan)
			}
			if len(plan.Tickets) != 4 {
				t.Fatalf("got %d ticket rows, want 4: %+v", len(plan.Tickets), plan.Tickets)
			}

			for _, assignment := range plan.Tickets {
				assertCompleteTicketRow(t, assignment)
				if got := assignment.ResolvedModel; got != testCase.resolvedModels[assignment.TicketID] {
					t.Errorf("ticket %s resolved model = %q, want %q", assignment.TicketID, got, testCase.resolvedModels[assignment.TicketID])
				}
				if got := assignment.Reasoning; got != testCase.reasoning[assignment.TicketID] {
					t.Errorf("ticket %s reasoning = %q, want %q", assignment.TicketID, got, testCase.reasoning[assignment.TicketID])
				}
			}

			approval := e2eApproval(testCase.plan, plan)
			output, err := agents.FormatDispatchApproval(approval)
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range []string{
				"| 01 | mechanical dispatch coverage |",
				"| 02 | cross-cutting approval coverage |",
				"| 03 | high-risk enforcement coverage |",
				"| 04 | handoff preservation coverage |",
			} {
				if got := strings.Count(output, row); got != 1 {
					t.Errorf("approval table contains %d rows matching %q, want 1", got, row)
				}
			}
			for _, expected := range []string{
				"Plan: " + string(testCase.plan),
				"Approval: approved",
				"Estimate",
				"Approved budget",
				"Fix reserve",
				"advisory: measured",
				"unavailable",
				"Global implementation concurrency: 2",
				"### Workflow overhead",
				"### Optional roles",
				"| explorer | omitted |",
				"| reviewer | conditional |",
			} {
				if !strings.Contains(output, expected) {
					t.Errorf("approval output missing %q:\n%s", expected, output)
				}
			}
		})
	}
}

func TestEndToEndUnavailableTicketProfileFailsWithoutFallback(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	custom := map[string]agents.Assignment{
		"01": {ModelProfile: "balanced", Reasoning: "medium"},
		"02": {ModelProfile: "does-not-exist", Reasoning: "high"},
		"03": {ModelProfile: "balanced", Reasoning: "medium"},
		"04": {ModelProfile: "balanced", Reasoning: "medium"},
	}
	_, err = catalog.ResolveTicketPlan(e2eTickets(), agents.Codex, agents.Customize, custom)
	if err == nil || !strings.Contains(err.Error(), `profile "does-not-exist" for harness "codex": profile is unavailable`) {
		t.Fatalf("unavailable ticket profile should stop preflight without fallback, error = %v", err)
	}
}

func TestEndToEndDispatchGatesAndExactHandoffPreservation(t *testing.T) {
	catalog, err := agents.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	plan, err := catalog.ResolveTicketPlanWithLimits(
		e2eTickets(), agents.Codex, agents.Economy, nil, 2,
		map[string]int{"mechanical": 1, "implementer": 1, "architect": 1},
		map[string]int{"implementation": 2},
	)
	if err != nil {
		t.Fatal(err)
	}
	approval := e2eApproval(agents.Economy, plan)
	approval.ReviewScope = "full-review"

	if _, err := approval.AuthorizeTicketDispatch("01", agents.TicketDispatchState{}); err != nil {
		t.Fatalf("ready root ticket rejected: %v", err)
	}
	if _, err := approval.AuthorizeTicketDispatch("02", agents.TicketDispatchState{}); err == nil || !strings.Contains(err.Error(), "dependency frontier") {
		t.Fatalf("blocked dependency was not enforced: %v", err)
	}
	if _, err := approval.AuthorizeTicketDispatch("04", agents.TicketDispatchState{Active: []agents.ActiveTicket{{TicketID: "01", RoleName: "mechanical", ConcurrencyClass: "implementation"}}}); err == nil || !strings.Contains(err.Error(), "role concurrency") {
		t.Fatalf("per-role concurrency was not enforced: %v", err)
	}
	if _, err := approval.AuthorizeTicketDispatch("01", agents.TicketDispatchState{MergeInProgress: true}); err == nil || !strings.Contains(err.Error(), "serialized transition") {
		t.Fatalf("serialized merge was not enforced: %v", err)
	}

	classLimited := approval
	classLimited.TicketPlan.GlobalConcurrencyLimit = 3
	classLimited.TicketPlan.RoleConcurrencyLimits = map[string]int{"mechanical": 2, "implementer": 2, "architect": 2}
	classLimited.TicketPlan.ClassConcurrencyLimits = map[string]int{"implementation": 1}
	if _, err := classLimited.AuthorizeTicketDispatch("03", agents.TicketDispatchState{
		Active: []agents.ActiveTicket{
			{TicketID: "01", RoleName: "mechanical", ConcurrencyClass: "implementation"},
			{TicketID: "02", RoleName: "implementer", ConcurrencyClass: "implementation"},
		},
		Completed: map[string]bool{"01": true},
	}); err == nil || !strings.Contains(err.Error(), "class concurrency") {
		t.Fatalf("per-class concurrency was not enforced: %v", err)
	}
	if _, err := approval.AuthorizeTicketDispatch("03", agents.TicketDispatchState{
		Active: []agents.ActiveTicket{
			{TicketID: "01", RoleName: "mechanical", ConcurrencyClass: "implementation"},
			{TicketID: "02", RoleName: "implementer", ConcurrencyClass: "implementation"},
		},
		Completed: map[string]bool{"01": true},
	}); err == nil || !strings.Contains(err.Error(), "global concurrency") {
		t.Fatalf("global concurrency was not enforced: %v", err)
	}

	if _, err := approval.AuthorizeTicketDispatch("99", agents.TicketDispatchState{}); err == nil || !requiresRenewal(err) {
		t.Fatalf("newly discovered ticket should require renewal: %v", err)
	}
	if err := approval.AuthorizeOptionalRole("explorer"); err == nil || !requiresRenewal(err) {
		t.Fatalf("omitted optional role should require renewal: %v", err)
	}
	request := agents.TicketDispatchRequest{
		TicketID:         "01",
		RoleName:         "mechanical",
		ModelProfile:     "balanced",
		ResolvedModel:    "gpt-5.6-luna",
		Reasoning:        "low",
		TokenBudget:      agents.TokenRange{Min: 600, Max: 900},
		TokenUsage:       700,
		ConcurrencyClass: "implementation",
		Isolation:        "worktree",
		ReviewScope:      "full-review",
	}
	if _, err := approval.AuthorizeTicketDispatchRequest(request, agents.TicketDispatchState{}); err != nil {
		t.Fatalf("exact approved invocation rejected: %v", err)
	}
	if err := approval.ValidateTicketUsage("01", 901); err == nil || !requiresRenewal(err) {
		t.Fatalf("budget overrun should require renewal: %v", err)
	}

	updated := addRenewedTicket(approval)
	updated.OptionalRoles = []agents.OptionalRoleStatus{{Name: "explorer", Status: "selected", Detail: "renewed for architectural context"}, {Name: "reviewer", Status: "conditional"}}
	updated.WorkflowBudget.Fixes = agents.TokenRange{Min: 700, Max: 1000}
	context, err := agents.NewImplementationContext(approval, "preflight-approved")
	if err != nil {
		t.Fatal(err)
	}
	if err := context.RecordRenewedApproval(updated, "renewal-01", "new ticket and selected explorer require broader approved coverage"); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(context.Current, updated) || len(context.Renewals) != 1 {
		t.Fatalf("renewed plan was not recorded exactly: current=%+v renewals=%+v", context.Current, context.Renewals)
	}

	expectedCurrent, err := agents.FormatDispatchApproval(updated)
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"merger", "reviewer", "fixes", "integration"} {
		output, err := agents.FormatImplementationContext(context)
		if err != nil {
			t.Fatalf("%s handoff rejected: %v", phase, err)
		}
		if !strings.Contains(output, expectedCurrent) {
			t.Fatalf("%s handoff did not preserve exact current plan:\n%s", phase, output)
		}
	}

	handoff, err := agents.FormatReviewHandoff(agents.ReviewHandoff{
		SpecPath:      ".specs/ticket-level-economy-dispatch-approvals/spec.md",
		SpecBranch:    "codex/implement-ticket-level-economy",
		FinalRevision: "ticket-05-test",
		Context:       context,
		RequirementCoverage: []agents.RequirementResult{{
			Requirement: "preserve exact approved ticket rows",
			Status:      "complete",
			Evidence:    "dispatch_e2e_test.go",
		}},
		VerifiedChecks: []agents.Evidence{{Statement: "all workflow phases preserve the current plan", Pointer: "dispatch_e2e_test.go"}},
		RemainingGates: []string{"review approval"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		expectedCurrent,
		"renewal-01: new ticket and selected explorer require broader approved coverage",
		"| 99 | renewed coverage | 04 |",
		"| explorer | selected | renewed for architectural context |",
		"| fixes | 700–1000 |",
		"### Requirement coverage",
	} {
		if !strings.Contains(handoff, expected) {
			t.Errorf("final handoff missing %q:\n%s", expected, handoff)
		}
	}
}

func e2eTickets() []agents.Ticket {
	mechanical := e2eRole("mechanical", "testing", "low", 600, 900, 300, 450)
	return []agents.Ticket{
		{ID: "01", Scope: "mechanical dispatch coverage", Role: mechanical, ConcurrencyClass: "implementation"},
		{ID: "02", Scope: "cross-cutting approval coverage", Dependencies: []string{"01"}, Role: e2eRole("implementer", "implementation", "medium", 900, 1400, 400, 600), ConcurrencyClass: "implementation"},
		{ID: "03", Scope: "high-risk enforcement coverage", Role: e2eRole("architect", "architecture", "low", 1200, 1800, 500, 750), ConcurrencyClass: "implementation"},
		{ID: "04", Scope: "handoff preservation coverage", Dependencies: []string{"02", "03"}, Role: mechanical, ConcurrencyClass: "implementation"},
	}
}

func e2eRole(name, capability, reasoning string, min, max, fixMin, fixMax int) agents.Role {
	return agents.Role{
		Name:             name,
		Capability:       capability,
		Reasoning:        reasoning,
		Isolation:        "worktree",
		EstimatedTokens:  agents.TokenRange{Min: min / 2, Max: max - 200},
		TokenBudget:      agents.TokenRange{Min: min, Max: max},
		ConcurrencyLimit: 1,
		AccountQuota:     agents.CostStatus{Status: "advisory", Detail: "measured"},
		MonetaryCost:     agents.CostStatus{Status: "unavailable", Detail: "billing not exposed"},
		ReservedBudget:   agents.ReservedBudget{Fixes: agents.TokenRange{Min: fixMin, Max: fixMax}},
	}
}

func e2eApproval(plan agents.Plan, ticketPlan agents.TicketDispatchPlan) agents.DispatchApproval {
	return agents.DispatchApproval{
		Plan:       plan,
		TicketPlan: ticketPlan,
		WorkflowBudget: agents.WorkflowBudget{
			Merger:      agents.TokenRange{Min: 300, Max: 500},
			Reviewer:    agents.TokenRange{Min: 500, Max: 800},
			Fixes:       agents.TokenRange{Min: 400, Max: 700},
			CI:          agents.TokenRange{Min: 100, Max: 200},
			Integration: agents.TokenRange{Min: 300, Max: 500},
		},
		OptionalRoles: []agents.OptionalRoleStatus{
			{Name: "explorer", Status: "omitted", Detail: "not required"},
			{Name: "reviewer", Status: "conditional", Detail: "selected only after implementation"},
		},
		Approved: true,
	}
}

func assertCompleteTicketRow(t *testing.T, assignment agents.TicketDispatchAssignment) {
	t.Helper()
	if assignment.TicketID == "" || assignment.Scope == "" || assignment.Role.Name == "" || assignment.ModelProfile == "" || assignment.ResolvedModel == "" || assignment.Reasoning == "" || assignment.ConcurrencyClass == "" || assignment.Isolation != "worktree" {
		t.Errorf("incomplete ticket row: %+v", assignment)
	}
	if assignment.TokenBudget.Min <= 0 || assignment.TokenBudget.Max < assignment.TokenBudget.Min || assignment.EstimatedTokens.Min <= 0 || assignment.ReservedFixCapacity.Max <= 0 {
		t.Errorf("invalid token ranges in ticket row: %+v", assignment)
	}
	if assignment.AccountQuota.Status == "" || assignment.MonetaryCost.Status == "" {
		t.Errorf("cost statuses missing from ticket row: %+v", assignment)
	}
}

func requiresRenewal(err error) bool {
	var gate *agents.DispatchGateError
	return errors.As(err, &gate) && gate.RenewalRequired
}

func addRenewedTicket(approval agents.DispatchApproval) agents.DispatchApproval {
	updated := approval
	updated.TicketPlan.Tickets = append([]agents.TicketDispatchAssignment(nil), approval.TicketPlan.Tickets...)
	updated.TicketPlan.Tickets = append(updated.TicketPlan.Tickets, agents.TicketDispatchAssignment{
		TicketID:            "99",
		Scope:               "renewed coverage",
		Dependencies:        []string{"04"},
		Role:                agents.Role{Name: "implementer", Capability: "implementation", Isolation: "worktree"},
		ModelProfile:        "balanced",
		ResolvedModel:       "gpt-5.6-luna",
		Reasoning:           "high",
		EstimatedTokens:     agents.TokenRange{Min: 1200, Max: 2000},
		TokenBudget:         agents.TokenRange{Min: 1800, Max: 3200},
		ConcurrencyClass:    "implementation",
		ConcurrencyLimit:    1,
		Isolation:           "worktree",
		ReservedFixCapacity: agents.TokenRange{Min: 500, Max: 800},
		AccountQuota:        agents.CostStatus{Status: "advisory", Detail: "measured"},
		MonetaryCost:        agents.CostStatus{Status: "unavailable", Detail: "billing not exposed"},
	})
	return updated
}
