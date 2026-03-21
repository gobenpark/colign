---
name: onboard
description: Set up and verify the Colign MCP connection. Use when the user first installs the plugin, says "set up colign", "connect to colign", asks how to get started, or when any colign MCP tool call fails with a connection or authentication error.
---

# Onboard to Colign

Guide the user through setting up the Colign MCP server connection.

## Workflow

### Step 1: Check if colign-mcp binary is available

```bash
which colign-mcp
```

If not found, guide the user to install it:
- **From source**: `go install github.com/gobenpark/colign/cmd/mcp@latest`
- **From release**: download the binary from GitHub releases

### Step 2: Check environment variables

```bash
echo $COLIGN_API_TOKEN
echo $COLIGN_API_URL
```

If `COLIGN_API_TOKEN` is not set:
1. Tell the user to go to **Colign Settings > AI & API Keys**
2. Click **Generate Token** and name it (e.g., "Claude Code")
3. Copy the token (it's only shown once)
4. Set it: `export COLIGN_API_TOKEN=col_...`

If `COLIGN_API_URL` is not set, that's OK — it defaults to `http://localhost:8080`.

### Step 3: Verify the connection

Try calling `mcp__colign__list_projects` to verify the MCP server is working.

- **Success**: Show the project list and confirm everything is connected
- **Auth error**: Token is invalid or expired — regenerate in Colign Settings
- **Connection error**: Check if the Colign server is running at the configured URL

### Step 4: Confirm setup

```
Colign MCP: Connected
API URL: [url]
Projects: [count] accessible
Token: col_xxxx... (valid)

You're all set! Here's how to get started:
- /colign:explore — Browse your projects and specs
- /colign:propose — Start a new change with a proposal
```

## Error Recovery

This skill should also trigger when other colign skills fail due to:
- "COLIGN_API_TOKEN environment variable is required"
- Connection refused errors
- Authentication failures

In these cases, guide the user through the relevant fix step.
