---
name: explore
description: Browse Colign projects and changes. Use when the user asks about specs, wants to see what's in a project, checks change status, asks "what's the current spec", "show me the proposal", "what stage is this at", or any request that requires reading data from the Colign platform before doing other work.
---

# Explore Colign

Browse projects, changes, and specs on the Colign platform.

## Workflow

1. Call `mcp__colign__list_projects` to see all accessible projects
2. Ask the user which project or change they want to work on
3. Call `mcp__colign__get_change` with the change ID to see its current stage and metadata
4. Call `mcp__colign__read_spec` to read the relevant documents (proposal, design, spec, tasks)
5. Summarize the current state concisely:
   - What stage is the change in?
   - What has been written so far?
   - What tasks exist and their status?

## Output Format

Present findings as:
```
Project: [name]
Change: [name] (Stage: [stage])
Proposal: [exists/empty] — [one-line summary]
Design: [exists/empty] — [one-line summary]
Tasks: [X done / Y total]
```

## Tips
- If the user gives a project name instead of ID, use `list_projects` to find it
- Read all document types to give a complete picture
- Highlight any blockers or pending reviews
