---
name: implement
description: Implement code based on a Colign spec. Use when the user says "implement this", "start coding", "work on the next task", "build this feature", or any request to write code that has a corresponding spec or task list on the Colign platform. Reads specs and tasks from Colign, writes code locally, and updates task status as work progresses.
---

# Implement from Spec

Read the spec from Colign, implement the code locally, and update task progress on the platform.

## Workflow

1. Call `mcp__colign__read_spec` with `doc_type: design` to read the design
2. Call `mcp__colign__read_spec` with `doc_type: proposal` for additional context
3. Call `mcp__colign__list_tasks` to see the task list and their current status
4. Pick the next `todo` task (or let the user choose)
5. Call `mcp__colign__update_task` to set it to `in_progress`
6. Implement the code locally following the spec
7. Verify the implementation (run tests, build, etc.)
8. Call `mcp__colign__update_task` to set the task to `done`
9. Move to the next task or ask the user

## Guidelines

- Always read the spec before coding — don't guess requirements
- Update task status in real-time so the platform reflects actual progress
- If the spec is unclear or needs changes, suggest running `/colign:propose` or `/colign:design` first
- Follow existing code conventions found in the local codebase
- Write tests alongside implementation
- If a task is too large, suggest breaking it down via `/colign:design`

## Verification Checklist

Before marking a task as `done`:
- [ ] Code compiles without errors
- [ ] Tests pass
- [ ] Implementation matches the spec requirements

## Next Step

When all tasks are complete, suggest running `/colign:complete` to finalize the change.
