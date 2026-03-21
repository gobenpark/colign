---
name: plan
description: Generate a detailed implementation plan and tasks from a Colign proposal. Use when the user says "plan this out", "break this down into tasks", "how should we implement this", "create an implementation plan", or after a proposal exists and needs to be turned into actionable architecture decisions and ordered tasks on the platform.
---

# Plan Implementation

Read the proposal from Colign and generate an implementation plan with tasks.

## Workflow

1. Call `mcp__colign__read_spec` with `doc_type: proposal` to read the proposal
2. Optionally call `mcp__colign__suggest_spec` to get AI improvement suggestions
3. Analyze the codebase locally to understand existing patterns and constraints
4. Draft an implementation plan covering:
   - Architecture decisions
   - Key interfaces and data models
   - Implementation steps in order
   - Edge cases and error handling
   - Testing strategy
5. Break the plan into ordered implementation tasks
6. Present the plan + task list to the user for review
7. After approval:
   - Call `mcp__colign__write_spec` with `doc_type: design` to save the plan

## Plan Format

```markdown
## Architecture
High-level approach and key decisions.

## Data Model
New or modified data structures.

## Implementation Steps
1. Step one — description
2. Step two — description
...

## Testing Strategy
How to verify the implementation.
```

## Guidelines

- Always read the proposal first — don't plan in a vacuum
- Reference existing code patterns found in the local codebase
- Keep tasks small and independently verifiable
- Each task should be completable in a single session

## Next Step

After the plan is saved, suggest running `/colign:implement` to start coding against the first task.
