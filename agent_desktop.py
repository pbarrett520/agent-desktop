"""
Agent Desktop - A simple tool for non-technical users to describe tasks
in plain English and have an AI generate and execute scripts for them.

Supports two modes:
- Script Mode: Generate a script and execute it (original behavior)
- Agent Mode: AI agent with full shell access that iteratively completes tasks
"""

import json
import subprocess
import sys
import tempfile
from pathlib import Path

import requests
import streamlit as st
from openai import OpenAI, AzureOpenAI

# Import agent tools and loop (optional - for Agent Mode)
try:
    from tools import get_session_info, reset_session
    from agent_loop import run_agent_loop, AgentStep

    AGENT_MODE_AVAILABLE = True
except ImportError:
    AGENT_MODE_AVAILABLE = False
    # Stub for type hints when imports fail
    AgentStep = None  # type: ignore

# ============================================================================
# Configuration Management
# ============================================================================

CONFIG_DIR = Path.home() / ".agent_desktop"
CONFIG_FILE = CONFIG_DIR / "config.json"

DEFAULT_URLS = {
    "openai": "https://api.openai.com/v1",
    "openrouter": "https://openrouter.ai/api/v1",
    "lmstudio": "http://localhost:1234/v1",
    "azure": "https://your-resource.openai.azure.com",
}

PROVIDER_TYPES = ["openai", "openrouter", "lmstudio", "azure"]


def load_config() -> dict:
    """Load configuration from disk, creating default if not exists."""
    if not CONFIG_FILE.exists():
        CONFIG_DIR.mkdir(parents=True, exist_ok=True)
        default_config = {
            "providers": [],
            "active_provider": None,
            "active_model": None,
            "execution_timeout": 60,
            "confirm_before_execute": True,
        }
        save_config(default_config)
        return default_config
    return json.loads(CONFIG_FILE.read_text())


def save_config(config: dict):
    """Save configuration to disk."""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    CONFIG_FILE.write_text(json.dumps(config, indent=2))


def get_provider_by_name(config: dict, name: str) -> dict | None:
    """Get a provider configuration by name."""
    for p in config.get("providers", []):
        if p["name"] == name:
            return p
    return None


# ============================================================================
# LLM Client Management
# ============================================================================


def get_client(provider: dict):
    """Create an OpenAI-compatible client for the given provider."""
    if provider["type"] == "azure":
        return AzureOpenAI(
            azure_endpoint=provider["base_url"],
            api_key=provider["api_key"],
            api_version=provider.get("api_version", "2024-02-15-preview"),
        )
    else:
        return OpenAI(
            base_url=provider["base_url"],
            api_key=provider["api_key"],
        )


def test_connection(provider: dict) -> tuple[bool, str]:
    """Test connection to a provider. Returns (success, message)."""
    try:
        if provider["type"] == "lmstudio":
            # For LM Studio, just try to get models list
            response = requests.get(
                f"{provider['base_url'].rstrip('/')}/models", timeout=10
            )
            if response.status_code == 200:
                return True, "Connected successfully!"
            return False, f"HTTP {response.status_code}: {response.text}"
        else:
            # For other providers, make a minimal chat completion
            client = get_client(provider)
            models = provider.get("models", [])
            model = models[0] if models else "gpt-3.5-turbo"
            client.chat.completions.create(
                model=model,
                messages=[{"role": "user", "content": "Hi"}],
                max_tokens=1,
            )
            return True, "Connected successfully!"
    except Exception as e:
        return False, str(e)


def fetch_lmstudio_models(base_url: str) -> tuple[list[str], str]:
    """Fetch available models from LM Studio endpoint."""
    try:
        response = requests.get(f"{base_url.rstrip('/')}/models", timeout=10)
        if response.status_code == 200:
            data = response.json()
            models = [m["id"] for m in data.get("data", [])]
            return models, ""
        return [], f"HTTP {response.status_code}"
    except Exception as e:
        return [], str(e)


# ============================================================================
# Script Generation
# ============================================================================

SYSTEM_PROMPT = """You are a helpful assistant that generates scripts to accomplish user tasks.

When the user describes a task, you must:
1. Generate a Python or bash script to accomplish it
2. Provide a plain-English explanation of what the script does
3. List any files/directories that will be affected

IMPORTANT: Respond ONLY with valid JSON in this exact format (no markdown, no code fences):
{
    "explanation": "Plain English description of what the script will do",
    "script": "The actual script code here",
    "affected_paths": ["list", "of", "paths", "that", "will", "be", "modified"],
    "script_type": "python or bash"
}

Guidelines:
- Prefer Python for cross-platform compatibility
- Use bash only for simple file operations on Unix-like systems
- Be conservative - ask for confirmation paths when unsure
- Include error handling in scripts
- Never delete files without explicit user request
- For file operations, always show which files will be affected"""


def generate_script(client, model: str, task: str, context: str = "") -> dict:
    """Generate a script for the given task."""
    user_message = task
    if context:
        user_message += f"\n\nAdditional context:\n{context}"

    response = client.chat.completions.create(
        model=model,
        messages=[
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": user_message},
        ],
        temperature=0.7,
    )

    content = response.choices[0].message.content.strip()

    # Try to parse as JSON, handling potential markdown code fences
    if content.startswith("```"):
        # Remove markdown code fences
        lines = content.split("\n")
        content = "\n".join(lines[1:-1]) if lines[-1] == "```" else "\n".join(lines[1:])

    try:
        return json.loads(content)
    except json.JSONDecodeError:
        return {
            "explanation": "Failed to parse LLM response as JSON",
            "script": content,
            "affected_paths": [],
            "script_type": "unknown",
            "raw_response": True,
        }


# ============================================================================
# Script Execution
# ============================================================================


def execute_script(
    script: str, script_type: str, timeout: int
) -> tuple[bool, str, str]:
    """Execute a script and return (success, stdout, stderr)."""
    try:
        with tempfile.NamedTemporaryFile(
            mode="w",
            suffix=".py" if script_type == "python" else ".sh",
            delete=False,
        ) as f:
            f.write(script)
            script_path = f.name

        if script_type == "python":
            cmd = [sys.executable, script_path]
        else:
            cmd = ["bash", script_path]

        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            timeout=timeout,
            cwd=Path.home(),  # Run from user's home directory
        )

        # Clean up temp file
        Path(script_path).unlink(missing_ok=True)

        return result.returncode == 0, result.stdout, result.stderr

    except subprocess.TimeoutExpired:
        Path(script_path).unlink(missing_ok=True)
        return False, "", f"Script timed out after {timeout} seconds"
    except Exception as e:
        return False, "", str(e)


# ============================================================================
# Streamlit UI
# ============================================================================


def init_session_state():
    """Initialize session state variables."""
    if "config" not in st.session_state:
        st.session_state.config = load_config()
    if "generated_result" not in st.session_state:
        st.session_state.generated_result = None
    if "execution_result" not in st.session_state:
        st.session_state.execution_result = None
    if "show_add_provider" not in st.session_state:
        st.session_state.show_add_provider = False
    if "editing_provider" not in st.session_state:
        st.session_state.editing_provider = None


def save_and_sync_config():
    """Save config to disk and sync session state."""
    save_config(st.session_state.config)


def render_sidebar():
    """Render the sidebar with settings."""
    st.sidebar.title("Settings")

    config = st.session_state.config
    providers = config.get("providers", [])
    provider_names = [p["name"] for p in providers]

    # ---------- Provider Selection ----------
    st.sidebar.subheader("Provider")

    if provider_names:
        current_provider = config.get("active_provider")
        if current_provider not in provider_names:
            current_provider = provider_names[0] if provider_names else None

        selected_provider = st.sidebar.selectbox(
            "Select Provider",
            options=provider_names,
            index=(
                provider_names.index(current_provider)
                if current_provider in provider_names
                else 0
            ),
            key="provider_select",
        )

        if selected_provider != config.get("active_provider"):
            config["active_provider"] = selected_provider
            config["active_model"] = None
            save_and_sync_config()
            st.rerun()

        # Model selection for active provider
        provider = get_provider_by_name(config, selected_provider)
        if provider:
            models = provider.get("models", [])

            col1, col2 = st.sidebar.columns([3, 1])
            with col1:
                if models:
                    current_model = config.get("active_model")
                    if current_model not in models:
                        current_model = models[0]

                    selected_model = st.selectbox(
                        "Model",
                        options=models,
                        index=(
                            models.index(current_model)
                            if current_model in models
                            else 0
                        ),
                    )

                    if selected_model != config.get("active_model"):
                        config["active_model"] = selected_model
                        save_and_sync_config()
                else:
                    st.info("No models configured")

            with col2:
                if provider.get("type") == "lmstudio":
                    if st.button("🔄", help="Refresh models from LM Studio"):
                        models, error = fetch_lmstudio_models(provider["base_url"])
                        if models:
                            provider["models"] = models
                            save_and_sync_config()
                            st.success(f"Found {len(models)} models")
                            st.rerun()
                        else:
                            st.error(f"Failed: {error}")

            # Edit provider button
            if st.sidebar.button("Edit Provider", key="edit_provider_btn"):
                st.session_state.editing_provider = selected_provider
                st.session_state.show_add_provider = True
                st.rerun()
    else:
        st.sidebar.info("No providers configured. Add one below.")

    # ---------- Add Provider Button ----------
    if st.sidebar.button("+ Add Provider"):
        st.session_state.show_add_provider = True
        st.session_state.editing_provider = None
        st.rerun()

    # ---------- Add/Edit Provider Form ----------
    if st.session_state.show_add_provider:
        render_provider_form()

    # ---------- Execution Settings ----------
    st.sidebar.markdown("---")
    st.sidebar.subheader("Execution")

    timeout = st.sidebar.number_input(
        "Timeout (seconds)",
        min_value=5,
        max_value=300,
        value=config.get("execution_timeout", 60),
    )
    if timeout != config.get("execution_timeout"):
        config["execution_timeout"] = timeout
        save_and_sync_config()

    confirm = st.sidebar.checkbox(
        "Confirm before executing",
        value=config.get("confirm_before_execute", True),
    )
    if confirm != config.get("confirm_before_execute"):
        config["confirm_before_execute"] = confirm
        save_and_sync_config()


def render_provider_form():
    """Render the add/edit provider form in sidebar."""
    config = st.session_state.config
    editing = st.session_state.editing_provider
    existing_provider = get_provider_by_name(config, editing) if editing else None

    st.sidebar.markdown("---")
    st.sidebar.subheader("Edit Provider" if editing else "Add Provider")

    # Provider name
    provider_name = st.sidebar.text_input(
        "Provider Name",
        value=existing_provider["name"] if existing_provider else "",
        placeholder="My OpenAI Provider",
    )

    # Provider type
    provider_types = PROVIDER_TYPES
    default_type_idx = (
        provider_types.index(existing_provider["type"])
        if existing_provider and existing_provider["type"] in provider_types
        else 0
    )
    provider_type = st.sidebar.selectbox(
        "Provider Type",
        options=provider_types,
        index=default_type_idx,
    )

    # Base URL (auto-populated based on type)
    default_url = (
        existing_provider["base_url"]
        if existing_provider
        else DEFAULT_URLS.get(provider_type, "")
    )
    base_url = st.sidebar.text_input(
        "Base URL",
        value=default_url,
        placeholder=DEFAULT_URLS.get(provider_type, ""),
    )

    # API Key
    api_key = st.sidebar.text_input(
        "API Key",
        value=existing_provider["api_key"] if existing_provider else "",
        type="password",
        placeholder="sk-..." if provider_type != "lmstudio" else "not-needed",
    )

    # API Version (Azure only)
    api_version = ""
    if provider_type == "azure":
        api_version = st.sidebar.text_input(
            "API Version",
            value=(
                existing_provider.get("api_version", "2024-02-15-preview")
                if existing_provider
                else "2024-02-15-preview"
            ),
        )

    # Models
    models_str = st.sidebar.text_area(
        "Models (comma-separated)",
        value=(
            ", ".join(existing_provider.get("models", [])) if existing_provider else ""
        ),
        placeholder="gpt-4o, gpt-4o-mini",
        help="For LM Studio, you can use the Refresh button after saving",
    )

    # Parse models
    models = [m.strip() for m in models_str.split(",") if m.strip()]

    # Action buttons - Save and Test on first row
    col1, col2 = st.sidebar.columns(2)

    with col1:
        if st.button("Save", use_container_width=True):
            if not provider_name:
                st.sidebar.error("Provider name required")
            elif not base_url:
                st.sidebar.error("Base URL required")
            else:
                new_provider = {
                    "name": provider_name,
                    "type": provider_type,
                    "base_url": base_url,
                    "api_key": api_key,
                    "models": models,
                }
                if provider_type == "azure" and api_version:
                    new_provider["api_version"] = api_version

                # Remove existing provider with same name if editing
                if editing:
                    config["providers"] = [
                        p for p in config["providers"] if p["name"] != editing
                    ]

                # Check for duplicate name
                if any(p["name"] == provider_name for p in config["providers"]):
                    st.sidebar.error("Provider name already exists")
                else:
                    config["providers"].append(new_provider)
                    config["active_provider"] = provider_name
                    save_and_sync_config()
                    st.session_state.show_add_provider = False
                    st.session_state.editing_provider = None
                    st.rerun()

    with col2:
        if st.button("Test", use_container_width=True):
            test_provider = {
                "name": provider_name,
                "type": provider_type,
                "base_url": base_url,
                "api_key": api_key,
                "models": models,
            }
            if provider_type == "azure":
                test_provider["api_version"] = api_version

            with st.spinner("Testing..."):
                success, message = test_connection(test_provider)
            if success:
                st.sidebar.success(message)
            else:
                st.sidebar.error(message)

    # Delete and Cancel buttons on second row
    if editing:
        col3, col4 = st.sidebar.columns(2)
        with col3:
            if st.button("Delete", use_container_width=True, type="secondary"):
                config["providers"] = [
                    p for p in config["providers"] if p["name"] != editing
                ]
                if config.get("active_provider") == editing:
                    config["active_provider"] = (
                        config["providers"][0]["name"] if config["providers"] else None
                    )
                    config["active_model"] = None
                save_and_sync_config()
                st.session_state.show_add_provider = False
                st.session_state.editing_provider = None
                st.rerun()
        with col4:
            if st.button("Cancel", use_container_width=True):
                st.session_state.show_add_provider = False
                st.session_state.editing_provider = None
                st.rerun()
    else:
        if st.sidebar.button("Cancel", use_container_width=True):
            st.session_state.show_add_provider = False
            st.session_state.editing_provider = None
            st.rerun()


def render_main_area():
    """Render the main content area."""
    st.title("Agent Desktop")
    st.markdown(
        "Automate tasks by describing them in plain English. The AI will generate and execute scripts for you."
    )

    config = st.session_state.config

    # Check if we have a configured provider
    if not config.get("active_provider") or not config.get("active_model"):
        st.warning(
            "⚠️ Please configure a provider and select a model in the sidebar to get started."
        )
        return

    # Mode toggle
    if AGENT_MODE_AVAILABLE:
        mode = st.radio(
            "Mode",
            ["Script Mode", "Agent Mode"],
            horizontal=True,
            help="Script Mode: Generate and review a script before running. Agent Mode: AI executes commands directly.",
            label_visibility="collapsed",
        )

        if mode == "Agent Mode":
            render_agent_mode(config)
            return

    # Task description
    st.subheader("What do you want to do?")
    task = st.text_area(
        "Describe your task",
        placeholder="Example: Rename all .jpg files in ~/Downloads to lowercase",
        height=120,
        label_visibility="collapsed",
    )

    # Optional context (file path, etc.)
    with st.expander("Additional Context (optional)"):
        context = st.text_input(
            "Working directory or file path",
            placeholder="e.g., C:\\Users\\Me\\Documents or ~/projects",
        )
        extra_info = st.text_area(
            "Any other details",
            placeholder="Additional requirements or constraints...",
            height=80,
        )
        full_context = (
            f"Working directory: {context}\n{extra_info}"
            if context or extra_info
            else ""
        )

    # Generate button
    col1, col2 = st.columns([1, 4])
    with col1:
        generate_clicked = st.button(
            "Generate Script", type="primary", use_container_width=True
        )

    if generate_clicked and task:
        provider = get_provider_by_name(config, config["active_provider"])
        if not provider:
            st.error("Selected provider not found")
            return

        try:
            with st.spinner("🧠 Thinking..."):
                client = get_client(provider)
                result = generate_script(
                    client, config["active_model"], task, full_context
                )
                st.session_state.generated_result = result
                st.session_state.execution_result = None
        except Exception as e:
            st.error(f"Generation failed: {e}")
            return

    # Display generated result
    if st.session_state.generated_result:
        result = st.session_state.generated_result
        st.markdown("---")

        # Explanation
        with st.expander("Description", expanded=True):
            st.markdown(result.get("explanation", "No explanation provided"))

        # Script
        with st.expander("Generated Script", expanded=True):
            script_type = result.get("script_type", "python")
            lang = "python" if script_type == "python" else "bash"
            st.code(result.get("script", ""), language=lang)

        # Affected paths
        affected_paths = result.get("affected_paths", [])
        if affected_paths:
            with st.expander("Affected Paths", expanded=True):
                for path in affected_paths:
                    st.markdown(f"- `{path}`")

        # Warning for raw response
        if result.get("raw_response"):
            st.warning(
                "⚠️ The LLM response could not be parsed as structured JSON. Review the script carefully."
            )

        # Execute section
        st.markdown("---")

        if config.get("confirm_before_execute", True):
            st.warning(
                "⚠️ **Review the script carefully before executing.** This will run code on your computer."
            )

            if affected_paths:
                st.info(
                    f"**Files that may be modified:** {', '.join(affected_paths[:5])}"
                    + (
                        f" and {len(affected_paths) - 5} more..."
                        if len(affected_paths) > 5
                        else ""
                    )
                )

        col1, col2 = st.columns([1, 4])
        with col1:
            execute_clicked = st.button(
                "Execute Script",
                type=(
                    "primary"
                    if not config.get("confirm_before_execute")
                    else "secondary"
                ),
                use_container_width=True,
            )

        if execute_clicked:
            script = result.get("script", "")
            script_type = result.get("script_type", "python")

            with st.spinner("⚡ Executing..."):
                success, stdout, stderr = execute_script(
                    script, script_type, config.get("execution_timeout", 60)
                )
                st.session_state.execution_result = {
                    "success": success,
                    "stdout": stdout,
                    "stderr": stderr,
                }

    # Display execution result
    if st.session_state.execution_result:
        result = st.session_state.execution_result
        st.markdown("---")
        st.subheader("📤 Output")

        if result["success"]:
            st.success("✅ Script executed successfully!")
        else:
            st.error("❌ Script execution failed")

        # Terminal-style output
        output = ""
        if result["stdout"]:
            output += result["stdout"]
        if result["stderr"]:
            if output:
                output += "\n\n--- STDERR ---\n"
            output += result["stderr"]

        if output:
            st.code(output, language="text")
        else:
            st.info("No output produced")


def render_agent_mode(config: dict):
    """Render the Agent Mode interface with full shell access."""
    st.markdown("---")
    st.markdown("### Agent Mode")
    st.caption(
        "The AI assistant can execute commands, manage files, and navigate your system to complete tasks."
    )

    # Initialize agent state
    if "agent_steps" not in st.session_state:
        st.session_state.agent_steps = []
    if "agent_running" not in st.session_state:
        st.session_state.agent_running = False
    if "agent_task" not in st.session_state:
        st.session_state.agent_task = ""

    # Show session info
    session_info = get_session_info()
    st.caption(f"Working directory: `{session_info['cwd']}`")

    # Task input
    task = st.text_area(
        "What would you like me to do?",
        placeholder="Example: Find all Python files larger than 1MB in my Documents folder and list them",
        height=100,
        key="agent_task_input",
    )

    # Additional context
    with st.expander("Additional Context (optional)"):
        working_dir = st.text_input(
            "Starting directory",
            value=session_info["cwd"],
            placeholder="e.g., C:\\Users\\Me\\Documents",
        )
        extra_context = st.text_area(
            "Additional instructions",
            placeholder="Any specific requirements or constraints...",
            height=60,
        )

    # Control buttons
    col1, col2, col3 = st.columns([1, 1, 3])

    with col1:
        start_clicked = st.button(
            "Run Task",
            type="primary",
            use_container_width=True,
            disabled=st.session_state.agent_running or not task,
        )

    with col2:
        if st.button("Reset", use_container_width=True):
            reset_session()
            st.session_state.agent_steps = []
            st.session_state.agent_running = False
            st.rerun()

    # Run agent when start is clicked
    if start_clicked and task:
        st.session_state.agent_steps = []
        st.session_state.agent_running = True
        st.session_state.agent_task = task

        # Build context
        context_parts = []
        if working_dir and working_dir != session_info["cwd"]:
            context_parts.append(f"Please start in this directory: {working_dir}")
        if extra_context:
            context_parts.append(extra_context)
        full_context = "\n".join(context_parts)

        # Get provider and client
        provider = get_provider_by_name(config, config["active_provider"])
        if not provider:
            st.error("Selected provider not found")
            st.session_state.agent_running = False
            return

        client = get_client(provider)

        # Create a placeholder for live updates
        steps_container = st.container()

        try:
            # Run the agent loop
            for step in run_agent_loop(
                client=client,
                model=config["active_model"],
                task=task,
                context=full_context,
                max_steps=20,
            ):
                st.session_state.agent_steps.append(step)

                # Display the step
                with steps_container:
                    display_agent_step(step)

                # Check if complete or error
                if step.type in ("complete", "error"):
                    st.session_state.agent_running = False
                    break

        except Exception as e:
            st.error(f"Agent error: {e}")
            st.session_state.agent_running = False

    # Display previous steps if any
    if st.session_state.agent_steps and not start_clicked:
        st.markdown("---")
        st.markdown("### Execution Log")

        for step in st.session_state.agent_steps:
            display_agent_step(step)


def display_agent_step(step: AgentStep):
    """Display a single agent step in the UI."""
    if step.type == "thinking":
        with st.chat_message("assistant"):
            st.markdown(step.content)

    elif step.type == "tool_call":
        with st.expander(f"🔧 {step.tool_name}", expanded=True):
            st.json(step.tool_args)

    elif step.type == "tool_result":
        success = step.tool_result.success if step.tool_result else True
        icon = "✅" if success else "❌"

        with st.expander(f"{icon} Result: {step.tool_name}", expanded=True):
            if step.content:
                # Determine if output looks like code/terminal
                if step.tool_name == "run_command" or "\n" in step.content:
                    st.code(step.content, language="text")
                else:
                    st.markdown(step.content)

    elif step.type == "complete":
        st.success(step.content)

    elif step.type == "error":
        st.error(step.content)


def main():
    """Main application entry point."""
    st.set_page_config(
        page_title="Agent Desktop",
        page_icon="⚡",
        layout="wide",
        initial_sidebar_state="expanded",
    )

    # Corporate-friendly CSS theme
    st.markdown(
        """
    <style>
    /* Import professional font */
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
    
    /* Root variables for easy theming */
    :root {
        --primary-color: #0066CC;
        --primary-hover: #0052A3;
        --secondary-color: #00A3A3;
        --bg-dark: #0D1117;
        --bg-card: #161B22;
        --bg-sidebar: #0D1117;
        --border-color: #30363D;
        --text-primary: #F0F6FC;
        --text-secondary: #C9D1D9;
        --text-muted: #A8B2BD;
        --success-color: #238636;
        --error-color: #DA3633;
        --warning-color: #D29922;
    }
    
    /* Global text readability improvements */
    body, .stApp, .stMarkdown, p, span, label, div {
        color: var(--text-primary);
    }
    
    /* Ensure all paragraph text is readable */
    .stMarkdown p, .stCaption, [data-testid="stCaptionContainer"] {
        color: var(--text-secondary) !important;
    }
    
    /* Base app styling */
    .stApp {
        background: var(--bg-dark);
        font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
    }
    
    /* Sidebar styling */
    [data-testid="stSidebar"] {
        background: var(--bg-sidebar);
        border-right: 1px solid var(--border-color);
    }
    
    [data-testid="stSidebar"] .stMarkdown h1 {
        color: var(--text-primary) !important;
        font-size: 1.25rem;
        font-weight: 600;
        letter-spacing: -0.02em;
    }
    
    /* Main title styling */
    h1 {
        color: var(--text-primary) !important;
        font-weight: 700;
        letter-spacing: -0.03em;
        background: none !important;
        -webkit-text-fill-color: var(--text-primary) !important;
    }
    
    h3 {
        color: var(--text-primary) !important;
        font-weight: 600;
        letter-spacing: -0.02em;
    }
    
    /* Subtle accent for headers */
    .main h1::before {
        content: '';
        display: inline-block;
        width: 4px;
        height: 1.2em;
        background: linear-gradient(180deg, var(--primary-color) 0%, var(--secondary-color) 100%);
        margin-right: 12px;
        border-radius: 2px;
        vertical-align: middle;
    }
    
    /* Primary buttons - professional blue */
    .stButton button[kind="primary"],
    .stButton button[data-testid="stBaseButton-primary"] {
        background: linear-gradient(135deg, var(--primary-color) 0%, #0077B6 100%);
        border: none;
        border-radius: 6px;
        font-weight: 500;
        letter-spacing: 0.01em;
        transition: all 0.2s ease;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
    }
    
    .stButton button[kind="primary"]:hover,
    .stButton button[data-testid="stBaseButton-primary"]:hover {
        background: linear-gradient(135deg, var(--primary-hover) 0%, #005F8A 100%);
        transform: translateY(-1px);
        box-shadow: 0 4px 12px rgba(0, 102, 204, 0.3);
    }
    
    /* Secondary buttons */
    .stButton button[kind="secondary"],
    .stButton button[data-testid="stBaseButton-secondary"] {
        background: transparent;
        border: 1px solid var(--border-color);
        border-radius: 6px;
        color: var(--text-primary);
        font-weight: 500;
        transition: all 0.2s ease;
    }
    
    .stButton button[kind="secondary"]:hover,
    .stButton button[data-testid="stBaseButton-secondary"]:hover {
        background: var(--bg-card);
        border-color: var(--text-secondary);
    }
    
    /* All buttons base styling */
    .stButton button {
        border-radius: 6px;
        font-weight: 500;
        font-family: 'Inter', sans-serif;
        transition: all 0.2s ease;
    }
    
    /* Text areas and inputs */
    .stTextArea textarea,
    .stTextInput input {
        border-radius: 6px;
        border: 1px solid var(--border-color);
        background: var(--bg-card);
        color: var(--text-primary);
        font-family: 'Inter', sans-serif;
    }
    
    .stTextArea textarea:focus,
    .stTextInput input:focus {
        border-color: var(--primary-color);
        box-shadow: 0 0 0 3px rgba(0, 102, 204, 0.2);
    }
    
    /* Code blocks - clean terminal style */
    .stCodeBlock {
        border: 1px solid var(--border-color);
        border-radius: 6px;
        background: var(--bg-card);
    }
    
    /* Expanders */
    .streamlit-expanderHeader {
        background: var(--bg-card);
        border-radius: 6px;
        border: 1px solid var(--border-color);
        font-weight: 500;
    }
    
    /* Select boxes */
    .stSelectbox > div > div {
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        border-radius: 6px;
    }
    
    /* Radio buttons */
    .stRadio > div {
        background: var(--bg-card);
        padding: 8px 16px;
        border-radius: 6px;
        border: 1px solid var(--border-color);
    }
    
    /* Success/error/warning messages */
    .stSuccess {
        background: rgba(35, 134, 54, 0.15);
        border: 1px solid var(--success-color);
        border-radius: 6px;
    }
    
    .stError {
        background: rgba(218, 54, 51, 0.15);
        border: 1px solid var(--error-color);
        border-radius: 6px;
    }
    
    .stWarning {
        background: rgba(210, 153, 34, 0.15);
        border: 1px solid var(--warning-color);
        border-radius: 6px;
    }
    
    .stInfo {
        background: rgba(0, 102, 204, 0.15);
        border: 1px solid var(--primary-color);
        border-radius: 6px;
    }
    
    /* Separator lines */
    hr {
        border-color: var(--border-color);
        opacity: 0.5;
    }
    
    /* Checkbox styling */
    .stCheckbox label span {
        color: var(--text-primary);
    }
    
    /* Form labels - ensure high contrast */
    .stTextInput label, .stTextArea label, .stSelectbox label,
    .stNumberInput label, .stRadio label, .stCheckbox label,
    [data-testid="stWidgetLabel"] {
        color: var(--text-secondary) !important;
    }
    
    /* Placeholder text */
    ::placeholder {
        color: var(--text-muted) !important;
        opacity: 0.8;
    }
    
    /* Dropdown/select text */
    .stSelectbox [data-baseweb="select"] span {
        color: var(--text-primary) !important;
    }
    
    /* Expander text */
    .streamlit-expanderHeader p {
        color: var(--text-primary) !important;
    }
    
    /* Radio button text */
    .stRadio label p {
        color: var(--text-primary) !important;
    }
    
    /* Number input */
    .stNumberInput > div > div > input {
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        border-radius: 6px;
    }
    
    /* Chat messages */
    .stChatMessage {
        background: var(--bg-card);
        border: 1px solid var(--border-color);
        border-radius: 8px;
    }
    
    /* Scrollbar styling */
    ::-webkit-scrollbar {
        width: 8px;
        height: 8px;
    }
    
    ::-webkit-scrollbar-track {
        background: var(--bg-dark);
    }
    
    ::-webkit-scrollbar-thumb {
        background: var(--border-color);
        border-radius: 4px;
    }
    
    ::-webkit-scrollbar-thumb:hover {
        background: var(--text-secondary);
    }
    </style>
    """,
        unsafe_allow_html=True,
    )

    init_session_state()
    render_sidebar()
    render_main_area()


if __name__ == "__main__":
    main()
