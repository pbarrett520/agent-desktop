# 🤖 Agent Desktop

A simple tool for non-technical users to describe tasks in plain English and have an AI generate and execute scripts for them.

## The Problem

You're a graphic designer. You asked ChatGPT to write a Python script to rename all your JPG files to lowercase. It gave you a script. Now what? You don't have Python installed, you don't know what a terminal is, and the script just sits there useless.

## The Solution

Agent Desktop closes that gap. Describe what you want in plain English, and the app:

1. **Generates** a script using your preferred AI provider
2. **Explains** what the script will do in plain English
3. **Shows** which files will be affected
4. **Executes** the script with your confirmation

## Quick Start

```bash
# Install dependencies
pip install -r requirements.txt

# Run the app
streamlit run agent_desktop.py
```

The app will open in your browser at `http://localhost:8501`.

## First-Time Setup

1. Click **"+ Add Provider"** in the sidebar
2. Choose your AI provider type (OpenAI, OpenRouter, LM Studio, or Azure)
3. Enter your API key
4. Add the models you want to use
5. Click **Save**

## Supported Providers

| Provider | Type | Description |
|----------|------|-------------|
| **OpenAI** | `openai` | Direct OpenAI API access |
| **OpenRouter** | `openrouter` | Multi-model aggregator with many providers |
| **LM Studio** | `lmstudio` | Local models running on your machine |
| **Azure OpenAI** | `azure` | Microsoft Azure-hosted OpenAI models |

### Example Configurations

#### OpenAI
- **Type:** `openai`
- **Base URL:** `https://api.openai.com/v1`
- **API Key:** Your OpenAI API key
- **Models:** `gpt-4o`, `gpt-4o-mini`

#### OpenRouter
- **Type:** `openrouter`
- **Base URL:** `https://openrouter.ai/api/v1`
- **API Key:** Your OpenRouter API key
- **Models:** `anthropic/claude-3.5-sonnet`, `openai/gpt-4o`

#### LM Studio (Local)
- **Type:** `lmstudio`
- **Base URL:** `http://localhost:1234/v1`
- **API Key:** `not-needed`
- **Models:** Use the 🔄 Refresh button to auto-discover

## Usage

1. **Describe your task** in the main text area
   - Example: "Rename all .jpg files in ~/Downloads to lowercase"
   
2. **Add context** (optional)
   - Specify working directory
   - Add constraints or requirements

3. **Click Generate**
   - Review the explanation
   - Check the script
   - See affected files

4. **Click Execute**
   - Watch the output
   - See results

## Safety Features

- **Confirmation required** before execution (configurable)
- **Timeout protection** prevents runaway scripts (default: 60 seconds)
- **Clear file impact display** shows exactly what will be modified
- **Script preview** lets you review before running

## Configuration

Configuration is stored in `~/.agent_desktop/config.json`:

```json
{
  "providers": [
    {
      "name": "OpenAI",
      "type": "openai",
      "base_url": "https://api.openai.com/v1",
      "api_key": "sk-...",
      "models": ["gpt-4o", "gpt-4o-mini"]
    }
  ],
  "active_provider": "OpenAI",
  "active_model": "gpt-4o",
  "execution_timeout": 60,
  "confirm_before_execute": true
}
```

## Tips

- **Be specific** in your task descriptions
- **Mention the operating system** if it matters (Windows vs. Mac/Linux)
- **Start small** - test with non-destructive tasks first
- **Use local models** (LM Studio) for privacy-sensitive tasks

## Requirements

- Python 3.9+
- An AI provider API key (or local LM Studio)
- Dependencies: `streamlit`, `openai`, `requests`

## License

MIT

