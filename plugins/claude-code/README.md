# Colign Plugin for Claude Code

Connect Claude Code to the Colign spec management platform. Read, write, and manage specs directly from your terminal.

## Setup

### 1. Generate an API Token

Go to Colign Settings > AI & API Keys and generate a new API token.

### 2. Set Environment Variables

```bash
export COLIGN_API_TOKEN=col_your_token_here
export COLIGN_API_URL=https://your-colign-instance.com  # defaults to http://localhost:8080
```

### 3. Install the Plugin

```bash
# Local testing
claude --plugin-dir ./plugins/claude-code

# Or install from marketplace
claude plugin install colign
```

## Workflow

The plugin follows Colign's change lifecycle:

```
explore → propose → plan → implement → complete
```

| Skill | Stage | Description |
|-------|-------|-------------|
| `/colign:explore` | Any | Browse projects, read specs, check change status |
| `/colign:propose` | Draft → Problem | Define the problem, scope, and write a structured proposal |
| `/colign:plan` | Problem → Solution | Break proposal into architecture, implementation steps, and tasks |
| `/colign:implement` | Solution → Review | Code against the spec, update task progress |
| `/colign:complete` | Review → Done | Verify all tasks are done, advance the workflow |

Skills also auto-trigger based on context. For example, saying "implement the next task" will trigger `/colign:implement` automatically.

## MCP Tools

| Tool | Description |
|------|-------------|
| `list_projects` | List all accessible projects |
| `get_change` | Get change details including stage and artifacts |
| `read_spec` | Read a spec document |
| `write_spec` | Write or update a spec document |
| `list_tasks` | List implementation tasks for a change |
| `update_task` | Update a task's status |
| `suggest_spec` | Get AI suggestions for improving a spec |

## Example Usage

```
# Explicit skill invocation
> /colign:explore
> Show me the current spec for change 42

# Auto-triggered by context
> I want to build a new notification system
  (triggers: propose)

> Break this down into implementation tasks
  (triggers: plan)

> Start coding the first task
  (triggers: implement)

> All tasks are done, let's wrap up
  (triggers: complete)
```
