---
name: propose
description: Create or update a structured proposal for a Colign change. Use when the user wants to define a new feature, write a problem statement, scope work, says "let's write a proposal", "create a new change", "I want to build X", or any request that involves capturing requirements and saving them to Colign.
---

# Propose a Change

Create a structured proposal and save it to the Colign platform.

## Workflow

1. If no change exists yet, ask the user for the project and change name
2. Gather context from the user about what they want to build
3. Call `mcp__colign__read_spec` with `doc_type: proposal` to check for existing content
4. Draft a structured proposal in this format:

```markdown
## Problem
Why is this change needed? What user pain does it solve?

## Scope
What specifically will change? Be concrete about deliverables.

## Out of Scope
What is explicitly NOT included in this change?

## Approach
Technical direction and key design decisions.
```

5. Present the draft to the user for review
6. After approval, call `mcp__colign__write_spec` with `doc_type: proposal` to save

## Guidelines

- Keep proposals concise — focus on the "why" and "what", not implementation details
- Always check existing content before overwriting
- Ask clarifying questions if the scope is unclear
- Use the project's memory context if available (call `read_spec` with `doc_type: spec` for project conventions)

## Next Step

After the proposal is saved, suggest running `/colign:design` to generate the detailed design and implementation tasks.
