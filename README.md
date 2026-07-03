# Nimbus

A local execution agent for Azure consultants, with centralized observability. Nimbus runs on your machine and drives your own already-authenticated `az` CLI session — it never stores or proxies your credentials. Built with Go (Wails) and React/TypeScript.

In **Cloud Ops mode** (the default), the agent is a cautious cloud engineer: it answers questions by querying Azure directly, but any command that changes state is proposed to you first — with a plain-English explanation, a rollback hint, and a risk tier — before it runs. A **General mode** is also available for local shell/file automation unrelated to Azure.

## Features

- **Azure context awareness**: detects your `az` login, active tenant/subscription/user, and lets you switch subscriptions from the header
- **Command risk classifier**: every `az` command is classified as READ / MUTATE / DESTRUCTIVE before it's allowed to run
- **Propose-then-approve workflow**: state-changing commands render as an approval card (command, explanation, rollback hint, blast radius) — nothing executes without your Approve
- **Audit trail**: every execution and denial is appended to `~/.agent_desktop/audit.jsonl`, viewable in-app
- **Subscription overview dashboard**: resource groups, resource counts, VM power states, and this month's cost (when available)
- **General mode**: local shell commands and file management for non-Azure tasks
- **Multiple LLM Providers**: works with OpenAI, LM Studio, OpenRouter, and any OpenAI-compatible API
- **Real-time Progress**: watch the agent's thinking, tool calls, and approvals step-by-step
- **Cross-platform**: runs on Windows, macOS, and Linux

## Supported LLM Providers

| Provider | Endpoint | Notes |
|----------|----------|-------|
| OpenAI | `https://api.openai.com/v1` | GPT-4o, GPT-4, etc. |
| LM Studio | `http://localhost:1234/v1` | Local models |
| OpenRouter | `https://openrouter.ai/api/v1` | Multiple providers |
| Custom | Any URL | Any OpenAI-compatible API |

## Prerequisites

- [Go 1.21+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)
- [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli) (`az`), logged in via `az login` — required for Cloud Ops mode

Install Wails CLI:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## Building

### Development Mode

Run with hot-reload for development:
```bash
wails dev
```

The frontend dev server runs at `http://localhost:5173/` for debugging.

### Production Build

Build the production executable:
```bash
wails build
```

The built application will be at:
- Windows: `build/bin/nimbus.exe`
- macOS: `build/bin/nimbus.app`
- Linux: `build/bin/nimbus`

## Running

### From Built Executable

```bash
# Windows
.\build\bin\nimbus.exe

# macOS/Linux
./build/bin/nimbus
```

### Configuration

On first run, configure your LLM provider in the sidebar:

| Field | Description | Example |
|-------|-------------|---------|
| Provider Preset | Quick select for common providers | OpenAI, LM Studio, OpenRouter, Custom |
| Endpoint URL | API base URL | `https://api.openai.com/v1` |
| API Key | Your API key | `sk-...` |
| Model | Model name | `gpt-4o`, `deepseek-chat` |
| Timeout | Execution timeout in seconds | `60` |
| Agent mode | Cloud Ops (Azure) or General (local shell/files) | `CLOUD_OPS` (default) |

Configuration is saved to `~/.agent_desktop/config.json`. The audit trail is saved separately to `~/.agent_desktop/audit.jsonl`.

## Usage

1. **Configure LLM** - Select a provider preset or enter a custom endpoint, then add your API key and model
2. **Log in to Azure** - Run `az login` in a terminal; Nimbus picks up your session automatically and shows tenant/subscription/user in the header
3. **Ask a question** - e.g. "What VMs do I have in this subscription?" — answered directly via a READ-tier query
4. **Approve or deny changes** - Anything that mutates or deletes resources shows up as an approval card before it runs
5. **Review the audit trail** - Open the Audit panel to see every command executed or denied, filterable by risk tier

### Example Prompts (Cloud Ops mode)

- "What resource groups do I have?"
- "Create a resource group called rg-demo in eastus"
- "Delete rg-demo, I'm done with it"
- "Give foo@bar.com Owner on this subscription"
- "Clean up the dev resource group" (ambiguous — the agent will enumerate resources and ask before proposing any deletion)

### Example Prompts (General mode)

- "List all Python files in my Documents folder"
- "Create a new folder called 'project' and add a README.md file"
- "Find files larger than 10MB in the current directory"

## Testing

### Run All Go Tests
```bash
go test ./...
```

### Run Tests with Verbose Output
```bash
go test -v ./...
```

### Test API Connection
```bash
# Create .env file with your credentials
echo "LLM_ENDPOINT=https://api.openai.com/v1" > .env
echo "LLM_API_KEY=your-api-key" >> .env
echo "LLM_MODEL=gpt-4o" >> .env

# Run connection test
go run ./cmd/testapi
```

## Project Structure

```
nimbus/
├── main.go                 # Wails entry point
├── app.go                  # App struct with bound methods
├── internal/
│   ├── config/             # Configuration management (LLM + agent mode)
│   ├── llm/                # OpenAI-compatible client
│   ├── tools/               # Tool implementations (general + cloud ops)
│   ├── agent/               # Agent loop, prompts, and personas
│   ├── azure/               # az CLI context, subscriptions, dashboard data
│   ├── safety/               # az command risk classifier
│   └── audit/                # Append-only JSONL audit trail
├── frontend/
│   ├── src/
│   │   ├── components/     # React components (chat, approval card, audit/dashboard panels, etc.)
│   │   ├── App.tsx         # Main app component
│   │   └── style.css       # Tailwind CSS styles
│   └── wailsjs/             # Generated Wails bindings
├── cmd/
│   └── testapi/             # API testing utility
├── build/                    # Build output
└── python-reference/         # Original Python app for reference
```

## Available Tools

### Cloud Ops mode (default)

| Tool | Description |
|------|-------------|
| `az_query` | Execute a READ-tier `az` command directly and return the output |
| `az_propose` | Propose a MUTATE/DESTRUCTIVE `az` command as an approval card; never executes on its own |
| `read_file` / `write_file` / `list_directory` | Save and organize reports locally |
| `copy_file` / `move_file` | Local file management |
| `get_current_directory` / `change_directory` | Local filesystem navigation |
| `task_complete` | Signal task completion |

### General mode

| Tool | Description |
|------|-------------|
| `run_command` | Execute shell commands |
| `delete_file` | Delete files (requires confirmation) |
| `read_file` / `write_file` / `list_directory` / `copy_file` / `move_file` | File management |
| `get_current_directory` / `change_directory` | Filesystem navigation |
| `task_complete` | Signal task completion |

## Safety

- **Command risk classifier**: every `az` command is classified READ / MUTATE / DESTRUCTIVE (table-driven, in `internal/safety`); unrecognized verbs fail closed to MUTATE (human review required)
- **Propose-then-approve**: MUTATE and DESTRUCTIVE commands never execute without an explicit Approve; DESTRUCTIVE proposals render in a distinct danger color with the resource name echoed back
- **Audit trail**: every execution and denial is recorded to `~/.agent_desktop/audit.jsonl`, including tier, decision, exit code, duration, and a truncated output hash
- **Local shell blocklist** (General mode): prevents catastrophic commands like `rm -rf /`, `format`, `del /s /q`
- **Path Validation**: validates and expands file paths safely
- **Timeout Protection**: commands time out after a configured duration

## Tech Stack

- **Backend**: Go 1.21+
- **Desktop Framework**: Wails v2
- **Frontend**: React 18 + TypeScript
- **Styling**: Tailwind CSS v3
- **LLM**: Any OpenAI-compatible API (OpenAI, LM Studio, OpenRouter, etc.)
- **Cloud**: Azure CLI (`az`), driven locally — no credentials stored or proxied

## License

MIT
