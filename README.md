# Agent Desktop

A desktop AI agent application built with Go (Wails) and React/TypeScript. The agent can execute shell commands, manage files, and complete tasks autonomously using any OpenAI-compatible LLM.

## Features

- **Agent Mode**: AI assistant that can execute commands and manage files
- **Multiple LLM Providers**: Works with OpenAI, LM Studio, OpenRouter, and any OpenAI-compatible API
- **Safe Command Execution**: Built-in blocklist prevents dangerous commands
- **Real-time Progress**: Watch the agent work step-by-step
- **Token Usage Tracking**: Monitor API usage
- **Cross-platform**: Runs on Windows, macOS, and Linux

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
- Windows: `build/bin/agent-desktop.exe`
- macOS: `build/bin/agent-desktop.app`
- Linux: `build/bin/agent-desktop`

## Running

### From Built Executable

```bash
# Windows
.\build\bin\agent-desktop.exe

# macOS/Linux
./build/bin/agent-desktop
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

Configuration is saved to `~/.agent_desktop/config.json`.

## Usage

1. **Configure LLM** - Select a provider preset or enter a custom endpoint, then add your API key and model
2. **Test Connection** - Click "Test" to verify your configuration
3. **Enter a Task** - Type what you want the agent to do
4. **Run Task** - Click "Run Task" or press Ctrl+Enter
5. **Watch Progress** - See the agent's thinking, tool calls, and results in real-time

### Example Tasks

- "List all Python files in my Documents folder"
- "Create a new folder called 'project' and add a README.md file"
- "Find files larger than 10MB in the current directory"
- "Show me the contents of config.json"

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
agent-desktop-go/
├── main.go                 # Wails entry point
├── app.go                  # App struct with bound methods
├── internal/
│   ├── config/            # Configuration management
│   ├── llm/               # OpenAI-compatible client
│   ├── tools/             # Tool implementations (10 tools)
│   └── agent/             # Agent loop and prompts
├── frontend/
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── App.tsx        # Main app component
│   │   └── style.css      # Tailwind CSS styles
│   └── wailsjs/           # Generated Wails bindings
├── cmd/
│   └── testapi/           # API testing utility
├── build/                 # Build output
└── python-reference/      # Original Python app for reference
```

## Available Tools

The agent has access to these tools:

| Tool | Description |
|------|-------------|
| `run_command` | Execute shell commands |
| `read_file` | Read file contents |
| `write_file` | Create or modify files |
| `list_directory` | List directory contents |
| `delete_file` | Delete files |
| `copy_file` | Copy files |
| `move_file` | Move/rename files |
| `get_current_directory` | Get current working directory |
| `change_directory` | Change working directory |
| `task_complete` | Signal task completion |

## Safety

The agent includes safety features:
- **Command Blocklist**: Prevents dangerous commands like `rm -rf /`, `format`, `del /s /q`
- **Path Validation**: Validates and expands file paths safely
- **Timeout Protection**: Commands timeout after configured duration

## Tech Stack

- **Backend**: Go 1.21+
- **Desktop Framework**: Wails v2
- **Frontend**: React 18 + TypeScript
- **Styling**: Tailwind CSS v3
- **LLM**: Any OpenAI-compatible API (OpenAI, LM Studio, OpenRouter, etc.)

## License

MIT
