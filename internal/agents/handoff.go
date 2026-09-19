package agents

import (
	"fmt"
	"reflect"
	"strings"
)

// ApprovalEvent records the approval artifact used to authorize a workflow
// phase. Approval is copied into the record so later changes cannot rewrite
// what the phase originally received.
type ApprovalEvent struct {
	Name     string
	Approval DispatchApproval
}

// ApprovalRenewal records a material change to an approved plan and the
// approval event that authorized the new plan.
type ApprovalRenewal struct {
	Name     string
	Reason   string
	Previous DispatchApproval
	Updated  DispatchApproval
}

// ImplementationContext is the handoff payload shared by implementation,
// merger, reviewer, fixes, and integration. Current contains the exact plan
// those roles must enforce; Renewals preserves the approval history.
type ImplementationContext struct {
	Initial  ApprovalEvent
	Current  DispatchApproval
	Renewals []ApprovalRenewal
}

// NewImplementationContext snapshots an approved dispatch plan for later
// workflow handoffs.
func NewImplementationContext(approval DispatchApproval, eventName string) (ImplementationContext, error) {
	if err := validateApprovedPlan(approval); err != nil {
		return ImplementationContext{}, err
	}
	if strings.TrimSpace(eventName) == "" {
		return ImplementationContext{}, fmt.Errorf("create implementation context: approval event name is required")
	}
	snapshot := cloneDispatchApproval(approval)
	return ImplementationContext{
		Initial:  ApprovalEvent{Name: eventName, Approval: cloneDispatchApproval(snapshot)},
		Current:  snapshot,
		Renewals: nil,
	}, nil
}

// RecordRenewedApproval adds a material, explicitly named approval change to
// the context. It returns an error when the new plan is not approved or does
// not differ from the current plan.
func (c *ImplementationContext) RecordRenewedApproval(approval DispatchApproval, eventName, reason string) error {
	if c == nil {
		return fmt.Errorf("record renewed approval: implementation context is nil")
	}
	if err := validateApprovedPlan(approval); err != nil {
		return err
	}
	if strings.TrimSpace(eventName) == "" || strings.TrimSpace(reason) == "" {
		return fmt.Errorf("record renewed approval: event name and reason are required")
	}
	if reflect.DeepEqual(c.Current, approval) {
		return fmt.Errorf("record renewed approval: updated plan is unchanged")
	}
	c.Renewals = append(c.Renewals, ApprovalRenewal{
		Name:     eventName,
		Reason:   reason,
		Previous: cloneDispatchApproval(c.Current),
		Updated:  cloneDispatchApproval(approval),
	})
	c.Current = cloneDispatchApproval(approval)
	return nil
}

func validateApprovedPlan(approval DispatchApproval) error {
	if err := approval.Validate(); err != nil {
		return err
	}
	if !approval.Approved {
		return fmt.Errorf("approved dispatch plan is required")
	}
	return nil
}

// Evidence is deliberately typed so postflight reports do not blur verified
// checks with conclusions inferred from the implementation.
type Evidence struct {
	Statement string
	Pointer   string
}

// RequirementResult records the status of one originating requirement.
type RequirementResult struct {
	Requirement string
	Status      string
	Evidence    string
}

// ReviewHandoff is the final, context-complete handoff for review and
// postflight. It carries the same plan that merger and reviewer received.
type ReviewHandoff struct {
	SpecPath            string
	SpecBranch          string
	FinalRevision       string
	Context             ImplementationContext
	RequirementCoverage []RequirementResult
	VerifiedChecks      []Evidence
	InferredConclusions []Evidence
	ReviewFindings      []Evidence
	RemainingGates      []string
	Risks               []string
}

// FormatImplementationContext renders the complete current plan and its
// approval history for merger, reviewer, fixes, and integration prompts.
func FormatImplementationContext(context ImplementationContext) (string, error) {
	if err := validateContext(context); err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "## Implementation context\n\nInitial approval: %s\n\n", context.Initial.Name)
	initial, err := FormatDispatchApproval(context.Initial.Approval)
	if err != nil {
		return "", err
	}
	b.WriteString(initial)
	if len(context.Renewals) > 0 {
		b.WriteString("\n### Renewed approvals\n\n")
		for _, renewal := range context.Renewals {
			fmt.Fprintf(&b, "- %s: %s\n", markdownCell(renewal.Name), markdownCell(renewal.Reason))
		}
	}
	b.WriteString("\n### Current approved plan\n\n")
	current, err := FormatDispatchApproval(context.Current)
	if err != nil {
		return "", err
	}
	b.WriteString(current)
	return b.String(), nil
}

// FormatReviewHandoff renders a deterministic final handoff with plan data,
// requirement coverage, evidence, remaining gates, and risks.
func FormatReviewHandoff(handoff ReviewHandoff) (string, error) {
	if err := validateContext(handoff.Context); err != nil {
		return "", err
	}
	if strings.TrimSpace(handoff.SpecPath) == "" || strings.TrimSpace(handoff.SpecBranch) == "" || strings.TrimSpace(handoff.FinalRevision) == "" {
		return "", fmt.Errorf("format review handoff: spec path, branch, and final revision are required")
	}
	context, err := FormatImplementationContext(handoff.Context)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Implementation review handoff\n\n- Spec: %s\n- Branch: %s\n- Final revision: %s\n\n", markdownCell(handoff.SpecPath), markdownCell(handoff.SpecBranch), markdownCell(handoff.FinalRevision))
	b.WriteString(context)
	b.WriteString("\n### Requirement coverage\n\n| Requirement | Status | Evidence |\n|---|---|---|\n")
	for _, result := range handoff.RequirementCoverage {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", markdownCell(result.Requirement), markdownCell(result.Status), markdownCell(result.Evidence))
	}
	writeEvidenceSection(&b, "Verified checks", handoff.VerifiedChecks)
	writeEvidenceSection(&b, "Inferred conclusions", handoff.InferredConclusions)
	writeEvidenceSection(&b, "Review findings", handoff.ReviewFindings)
	writeStringSection(&b, "Remaining gates", handoff.RemainingGates)
	writeStringSection(&b, "Risks", handoff.Risks)
	return b.String(), nil
}

func validateContext(context ImplementationContext) error {
	if strings.TrimSpace(context.Initial.Name) == "" {
		return fmt.Errorf("validate implementation context: initial approval event is required")
	}
	if err := validateApprovedPlan(context.Initial.Approval); err != nil {
		return fmt.Errorf("validate implementation context: initial approval: %w", err)
	}
	if err := validateApprovedPlan(context.Current); err != nil {
		return fmt.Errorf("validate implementation context: current approval: %w", err)
	}
	if len(context.Renewals) > 0 {
		previous := context.Initial.Approval
		for _, renewal := range context.Renewals {
			if strings.TrimSpace(renewal.Name) == "" || strings.TrimSpace(renewal.Reason) == "" {
				return fmt.Errorf("validate implementation context: renewed approval name and reason are required")
			}
			if !reflect.DeepEqual(previous, renewal.Previous) {
				return fmt.Errorf("validate implementation context: approval history is not contiguous")
			}
			if reflect.DeepEqual(renewal.Previous, renewal.Updated) {
				return fmt.Errorf("validate implementation context: renewed approval %q is unchanged", renewal.Name)
			}
			previous = renewal.Updated
		}
		if !reflect.DeepEqual(previous, context.Current) {
			return fmt.Errorf("validate implementation context: current approval does not match renewal history")
		}
	}
	return nil
}

func writeEvidenceSection(b *strings.Builder, title string, evidence []Evidence) {
	fmt.Fprintf(b, "\n### %s\n\n", title)
	if len(evidence) == 0 {
		b.WriteString("- none\n")
		return
	}
	for _, item := range evidence {
		fmt.Fprintf(b, "- %s (%s)\n", markdownCell(item.Statement), markdownCell(item.Pointer))
	}
}

func writeStringSection(b *strings.Builder, title string, values []string) {
	fmt.Fprintf(b, "\n### %s\n\n", title)
	if len(values) == 0 {
		b.WriteString("- none\n")
		return
	}
	for _, value := range values {
		fmt.Fprintf(b, "- %s\n", markdownCell(value))
	}
}

func cloneDispatchApproval(approval DispatchApproval) DispatchApproval {
	clone := approval
	clone.TicketPlan.Tickets = append([]TicketDispatchAssignment(nil), approval.TicketPlan.Tickets...)
	for i := range clone.TicketPlan.Tickets {
		clone.TicketPlan.Tickets[i].Dependencies = append([]string(nil), approval.TicketPlan.Tickets[i].Dependencies...)
	}
	clone.TicketPlan.RoleConcurrencyLimits = copyLimits(approval.TicketPlan.RoleConcurrencyLimits)
	clone.TicketPlan.ClassConcurrencyLimits = copyLimits(approval.TicketPlan.ClassConcurrencyLimits)
	clone.OptionalRoles = append([]OptionalRoleStatus(nil), approval.OptionalRoles...)
	return clone
}
