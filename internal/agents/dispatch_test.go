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
	if assignments[0].Role.Name != "explorer" || assignments[0].ResolvedModel != "gpt-5.6-terra" {
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
	if assignments[1].ModelProfile != "capable" || assignments[1].ResolvedModel != "opus" || assignments[1].Reasoning != "low" {
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
