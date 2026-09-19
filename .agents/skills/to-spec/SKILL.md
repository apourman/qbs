---
name: to-spec
description: "Turn the current conversation into a spec and publish it to the project issue tracker: no interview, just synthesis of what you've already discussed."
---

This skill takes the current conversation context and codebase understanding and produces a spec. Do NOT interview the user; just synthesize what you already know.

This skill produces exactly one specification artifact. Ticket creation is a
separate workflow owned by `to-tickets`; run that skill only when the user
explicitly requests tickets. A successful `to-spec` run leaves implementation
ticket creation to a later, explicit `to-tickets` run.

Use the repository's local Markdown tracker convention and canonical triage
statuses when no repository-specific tracker configuration is present.

## Process

1. Determine whether this run will dispatch any sub-agents. If it will not, continue without a model-selection or dispatch-approval prompt. If it will, apply the shared `orchestration-gate` contract before the first dispatch.

   For delegated exploration or validation, declare the complete roster before dispatch. Each role must include its name, purpose, neutral capability, optionality, expected reasoning, isolation, concurrency eligibility, and estimated token range. Include reserved review/fix/integration capacity and any separately selectable red-team role.

   Show one preflight dispatch summary with the request, confidence score (1–5), rationale, assumptions, unresolved questions, roster, and separate token-estimate, account-quota, and monetary-cost fields. For each role show the recommended neutral profile, harness-resolved model when available, reasoning, concurrency, and token budget. Offer exactly `Recommended`, `Economy`, `Deep`, and `Customize`; customization assigns a profile and reasoning effort to every role. Apply the shared confidence gates and obtain clarification and plan approval before dispatch. An unavailable resolution stops with an actionable error; never silently fall back, upgrade, add optional agents, or exceed the approved plan.

   Resolve the roster through the repository catalog before dispatch and carry
   the approved role/model assignments and constraints into each handoff. The
   final workflow report must be generated from the written spec and checks
   performed, with requirement coverage and evidence pointers; describing the
   intended gate without performing these actions is insufficient.

2. Explore the repo to understand the current state of the codebase, if you haven't already. Use the project's domain glossary vocabulary throughout the spec, and respect any ADRs in the area you're touching.

3. Sketch out the seams at which you're going to test the feature. Existing seams should be preferred to new ones. Use the highest seam possible. If new seams are needed, propose them at the highest point you can. The fewer seams across the codebase, the better - the ideal number is one.

Check with the user that these seams match their expectations.

4. Write exactly one spec using the template below under `.specs/<domain-or-feature>/<spec-slug>.md`, then publish that spec record to the project issue tracker. Choose a stable, lowercase kebab-case domain or feature directory and reuse an existing directory when one applies. Apply the `ready-for-agent` triage label - no need for additional triage. The completion output is the spec itself; do not create, derive, or publish implementation tickets during this workflow. If tickets are wanted, stop after the spec and ask the user to run `to-tickets` separately.

5. After the spec is written, produce the shared postflight report. Compare
   every original requirement against the completed spec and classify it as
   `complete`, `partial`, `missing`, `ambiguous`, or `unrequested`; cite
   verified evidence separately from inferred conclusions; and list
   assumptions, low-confidence areas, risks, unnecessary complexity, and the
   smallest confidence-raising checks. If the workflow selected a red-team
   role, give it only the requirements and spec with minimal framing and
   include its concrete findings; otherwise leave `red_team_findings` empty.
   Complexity suggestions must preserve every required behavior. Carry the
   approved plan, preflight confidence, assumptions, unresolved risks, and
   final evidence into the handoff or final report.

## Spec paths

`.specs/` is the canonical home for specifications. Every spec has one
domain-or-feature directory and one descriptive Markdown file beneath it. Do
not put new specs directly in `.specs/`; the directory name is part of the
spec's identity and groups related research, specs, and implementation tickets.

<spec-template>

## Problem Statement

The problem that the user is facing, from the user's perspective.

## Solution

The solution to the problem, from the user's perspective.

## User Stories

A LONG, numbered list of user stories. Each user story should be in the format of:

1. As an <actor>, I want a <feature>, so that <benefit>

<user-story-example>
1. As a mobile bank customer, I want to see balance on my accounts, so that I can make better informed decisions about my spending
</user-story-example>

This list of user stories should be extremely extensive and cover all aspects of the feature.

## Implementation Decisions

A list of implementation decisions that were made. This can include:

- The modules that will be built/modified
- The interfaces of those modules that will be modified
- Technical clarifications from the developer
- Architectural decisions
- Schema changes
- API contracts
- Specific interactions

Do NOT include specific file paths or code snippets. They may end up being outdated very quickly.

Exception: if a prototype produced a snippet that encodes a decision more precisely than prose can (state machine, reducer, schema, type shape), inline it within the relevant decision and note briefly that it came from a prototype. Trim to the decision-rich parts, not a working demo, just the important bits.

## Testing Decisions

A list of testing decisions that were made. Include:

- A description of what makes a good test (only test external behavior, not implementation details)
- Which modules will be tested
- Prior art for the tests (i.e. similar types of tests in the codebase)

## Out of Scope

A description of the things that are out of scope for this spec.

## Further Notes

Any further notes about the feature.

</spec-template>
