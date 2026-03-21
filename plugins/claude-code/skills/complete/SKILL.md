---
name: complete
description: Finalize a Colign change after all tasks are done. Use when implementation is finished, all tasks are marked done, and the change is ready to advance to the next workflow stage. Also use when the user says they're done with a change, want to wrap up, or asks to move a change forward.
---

# Complete a Change

Verify all work is done and advance the change to the next workflow stage.

## Workflow

1. Call `mcp__colign__list_tasks` to verify all tasks are `done`
2. If any tasks remain `todo` or `in_progress`, warn the user and list them
3. Call `mcp__colign__get_change` to check the current stage
4. Summarize what was accomplished:
   - Tasks completed
   - Files modified (from local git status)
   - Key decisions made
5. Ask the user to confirm they want to advance the stage
6. If confirmed, the user can advance the stage via the Colign web UI

## Completion Summary Format

```
Change: [name]
Stage: [current] → [next]
Tasks: [X/Y completed]
Files changed: [count]

Key deliverables:
- [deliverable 1]
- [deliverable 2]
```

## Guidelines

- Never advance a stage without user confirmation
- If tasks are incomplete, suggest running `/colign:implement` first
- Include a brief summary of what was built for team visibility
