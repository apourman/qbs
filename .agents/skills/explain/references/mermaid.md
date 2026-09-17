# Mermaid Diagram Examples

Use Mermaid when the target Markdown renderer supports it, especially GitHub issues, GitHub PRs, README files, and docs sites. Keep diagrams renderer-friendly: short labels, common Mermaid primitives, explicit boundaries, and enough structure to teach without prose.

## 1. Flowchart With Boundaries

Use for process flow, architecture ownership, or handoffs between systems.

```mermaid
flowchart LR
  user[User asks for work] --> planner[Planner clarifies behavior]

  subgraph issue[Issue Body]
    explain[High-level explanation]
    paths[Affected code paths]
    steps[Step-by-step implementation]
    checks[Validation]
  end

  planner --> explain
  explain --> paths
  paths --> steps
  steps --> checks

  checks --> reviewer{Reviewer can verify?}
  reviewer -->|yes| done[Ready for agent]
  reviewer -->|no| revise[Split or clarify issue]
  revise --> explain
```

This shape works when readers need to see where one artifact or system owns each part of the flow.

## 2. Sequence Diagram

Use for request/response, actor interaction, retries, or external calls over time.

```mermaid
sequenceDiagram
  participant Builder
  participant Issue
  participant Codebase
  participant Tests

  Builder->>Issue: Read behavior contract
  Builder->>Issue: Read high-level explanation
  Builder->>Codebase: Modify named seam
  Builder->>Tests: Run focused validation

  alt validation passes
    Tests-->>Builder: Green result
    Builder-->>Issue: Acceptance criteria satisfied
  else validation fails
    Tests-->>Builder: Failure details
    Builder->>Codebase: Fix named seam only
  end
```

This shape works when order and responsibility matter more than static structure.

## 3. State Diagram

Use for lifecycle, queue state, status labels, verdicts, or anything where transitions are the concept.

```mermaid
stateDiagram-v2
  [*] --> Drafted
  Drafted --> Explained: add diagram and plain-language summary
  Explained --> Scoped: name behavior contract
  Scoped --> Implementable: add steps, contracts, validation

  Implementable --> ReadyForAgent: quality checklist passes
  Implementable --> NeedsSplit: too many files or contracts
  NeedsSplit --> Drafted: create smaller issue

  ReadyForAgent --> [*]
```

This shape works when the main question is "what states exist, and what changes them?"

## 4. Architecture Boundary

Use for modules, ownership, dependencies, and data crossing boundaries.

```mermaid
flowchart TB
  subgraph input[Input Boundary]
    source[PRD or parent issue]
    glossary[Glossary and ADRs]
  end

  subgraph planning[Planning Boundary]
    contract[Behavior contract]
    diagram[High-level explanation]
    seams[Mutation seams]
  end

  subgraph execution[Execution Boundary]
    builder[Builder agent]
    tests[Focused tests]
  end

  source --> contract
  glossary --> contract
  contract --> diagram
  contract --> seams
  diagram --> builder
  seams --> builder
  builder --> tests
```

This shape works when the explanation needs to separate "what defines the work" from "who executes it."

## 5. Decision Matrix As Flow

Use when a reader needs to understand how branches are chosen.

```mermaid
flowchart TD
  start[Need a diagram] --> target{Where will it live?}

  target -->|terminal or plain text| ascii[Use ASCII]
  target -->|GitHub Markdown| shape{What shape is the idea?}
  target -->|unknown renderer| fallback[Use ASCII, mention Mermaid option]

  shape -->|process or dependency| flow[Mermaid flowchart]
  shape -->|actors over time| sequence[Mermaid sequenceDiagram]
  shape -->|lifecycle or status| state[Mermaid stateDiagram-v2]

  ascii --> done[Diagram first, prose second]
  flow --> done
  sequence --> done
  state --> done
  fallback --> done
```

This shape works when the explanation itself is about choosing between options.
