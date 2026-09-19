package agents_test

import (
	"strings"
	"testing"

	"github.com/trues/qbs/internal/agents"
)

func TestImplementationContextPreservesApprovedPlanForHandoffs(t *testing.T) {
	approval := handoffApproval("gpt-5.6-luna", "medium", 1400, 2600)
	context, err := agents.NewImplementationContext(approval, "preflight-approved")
	if err != nil {
		t.Fatal(err)
	}

	// The context is a snapshot, not an alias to mutable approval input.
	approval.TicketPlan.Tickets[0].Dependencies[0] = "changed-after-dispatch"
	approval.TicketPlan.ClassConcurrencyLimits["implementation"] = 99

	output, err := agents.FormatImplementationContext(context)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"Initial approval: preflight-approved",
		"gpt-5.6-luna",
		"medium",
		"1400–2600",
		"global <= 2",
		"class <= 1",
		"### Workflow overhead",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("context missing %q:\n%s", expected, output)
		}
	}
	if strings.Contains(output, "changed-after-dispatch") || strings.Contains(output, "99") {
		t.Fatalf("context changed after input mutation:\n%s", output)
	}
}

func TestImplementationContextRecordsRenewedApprovalAndFinalHandoff(t *testing.T) {
	initial := handoffApproval("gpt-5.6-luna", "medium", 1400, 2600)
	context, err := agents.NewImplementationContext(initial, "preflight-approved")
	if err != nil {
		t.Fatal(err)
	}
	updated := handoffApproval("gpt-5.6-luna", "high", 1600, 3000)
	if err := context.RecordRenewedApproval(updated, "risk-renewal-01", "cross-cutting review required stronger reasoning"); err != nil {
		t.Fatal(err)
	}

	handoff, err := agents.FormatReviewHandoff(agents.ReviewHandoff{
		SpecPath:      ".specs/example/spec.md",
		SpecBranch:    "codex/example",
		FinalRevision: "abc1234",
		Context:       context,
		RequirementCoverage: []agents.RequirementResult{{
			Requirement: "preserve handoff plan",
			Status:      "complete",
			Evidence:    "internal/agents/handoff.go",
		}},
		VerifiedChecks:      []agents.Evidence{{Statement: "focused tests passed", Pointer: "go test ./internal/agents"}},
		InferredConclusions: []agents.Evidence{{Statement: "reviewer can enforce current plan", Pointer: "handoff context"}},
		RemainingGates:      []string{"review approval"},
		Risks:               []string{"runtime adapter is outside this package"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"# Implementation review handoff",
		"abc1234",
		"risk-renewal-01: cross-cutting review required stronger reasoning",
		"Reasoning | Assignment note | Estimate",
		"high",
		"### Verified checks",
		"### Inferred conclusions",
		"### Remaining gates",
		"### Risks",
	} {
		if !strings.Contains(handoff, expected) {
			t.Errorf("handoff missing %q:\n%s", expected, handoff)
		}
	}
}

func TestImplementationContextRejectsUnchangedRenewal(t *testing.T) {
	approval := handoffApproval("gpt-5.6-luna", "medium", 1400, 2600)
	context, err := agents.NewImplementationContext(approval, "preflight-approved")
	if err != nil {
		t.Fatal(err)
	}
	if err := context.RecordRenewedApproval(approval, "duplicate", "no material change"); err == nil || !strings.Contains(err.Error(), "unchanged") {
		t.Fatalf("expected unchanged renewal error, got %v", err)
	}
}

func handoffApproval(model, reasoning string, min, max int) agents.DispatchApproval {
	return agents.DispatchApproval{
		Plan: agents.Economy,
		TicketPlan: agents.TicketDispatchPlan{
			Tickets: []agents.TicketDispatchAssignment{{
				TicketID:            "04",
				Scope:               "carry plan through handoff",
				Dependencies:        []string{"03"},
				Role:                agents.Role{Name: "implementer", Isolation: "worktree"},
				ModelProfile:        "balanced",
				ResolvedModel:       model,
				Reasoning:           reasoning,
				EstimatedTokens:     agents.TokenRange{Min: 1000, Max: 2000},
				TokenBudget:         agents.TokenRange{Min: min, Max: max},
				ConcurrencyClass:    "implementation",
				ConcurrencyLimit:    1,
				Isolation:           "worktree",
				ReservedFixCapacity: agents.TokenRange{Min: 400, Max: 600},
				AccountQuota:        agents.CostStatus{Status: "advisory"},
				MonetaryCost:        agents.CostStatus{Status: "unavailable"},
			}},
			GlobalConcurrencyLimit: 2,
			RoleConcurrencyLimits:  map[string]int{"implementer": 1},
			ClassConcurrencyLimits: map[string]int{"implementation": 1},
			TransitionPolicy:       "serialized",
		},
		WorkflowBudget: agents.WorkflowBudget{
			Merger:      agents.TokenRange{Min: 500, Max: 800},
			Reviewer:    agents.TokenRange{Min: 600, Max: 900},
			Fixes:       agents.TokenRange{Min: 400, Max: 600},
			CI:          agents.TokenRange{Min: 100, Max: 200},
			Integration: agents.TokenRange{Min: 300, Max: 500},
		},
		OptionalRoles: []agents.OptionalRoleStatus{{Name: "explorer", Status: "omitted"}},
		ReviewScope:   "full-review",
		Approved:      true,
	}
}
