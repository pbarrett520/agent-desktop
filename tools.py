"""
MCP-style tools for Agent Desktop.

This module provides a set of tools that the AI agent can use to interact
with the system. These are designed to be compatible with OpenAI's function
calling and can also be exposed as an MCP server.
"""

import os
import subprocess
import sys
import shutil
from pathlib import Path
from typing import Any
from dataclasses import dataclass, field
import json

# Windows known folder detection (handles OneDrive redirection)
if sys.platform == "win32":
    try:
        import winreg

        def _get_windows_known_folder(folder_name: str) -> str | None:
            """Get the actual path of a Windows known folder (handles OneDrive)."""
            folder_map = {
                "desktop": r"Software\Microsoft\Windows\CurrentVersion\Explorer\User Shell Folders",
            }
            value_map = {
                "desktop": "Desktop",
                "documents": "Personal",
                "downloads": "{374DE290-123F-4565-9164-39C4925E467B}",
            }
            try:
                key = winreg.OpenKey(
                    winreg.HKEY_CURRENT_USER, folder_map.get("desktop")
                )
                value, _ = winreg.QueryValueEx(
                    key, value_map.get(folder_name.lower(), folder_name)
                )
                winreg.CloseKey(key)
                # Expand environment variables like %USERPROFILE%
                return os.path.expandvars(value)
            except Exception:
                return None

    except ImportError:

        def _get_windows_known_folder(folder_name: str) -> str | None:
            return None

else:

    def _get_windows_known_folder(folder_name: str) -> str | None:
        return None


def _expand_path(path: str, cwd: str) -> str:
    """
    Expand a path, handling:
    - ~ (home directory)
    - Relative paths (relative to cwd)
    - Windows known folders like Desktop, Documents (handles OneDrive redirection)
    - Fixes incorrect absolute paths (e.g., C:\\Users\\X\\Desktop -> actual Desktop path)
    """
    if not path:
        return cwd

    # Handle home directory
    if path.startswith("~"):
        return os.path.expanduser(path)

    # Normalize path separators
    normalized = path.replace("\\", "/")

    # Check for Windows known folder patterns in absolute paths
    # e.g., "C:/Users/pbarr/Desktop/test" should use actual Desktop location
    if sys.platform == "win32":
        home = str(Path.home()).replace("\\", "/").lower()
        path_lower = normalized.lower()

        for folder_name in ("desktop", "documents", "downloads"):
            # Check if path contains the default (possibly wrong) known folder path
            wrong_path = f"{home}/{folder_name}"
            if path_lower.startswith(wrong_path):
                actual_path = _get_windows_known_folder(folder_name)
                if actual_path and actual_path.replace("\\", "/").lower() != wrong_path:
                    # Replace the wrong path prefix with the correct one
                    remainder = normalized[len(wrong_path) :]
                    return actual_path + remainder.replace("/", os.sep)

    # Handle absolute paths (that don't need fixing)
    if os.path.isabs(path):
        return path

    # Check for Windows known folder at start of relative path (e.g., "Desktop/test")
    parts = normalized.split("/")
    first_part = parts[0].lower()

    if first_part in ("desktop", "documents", "downloads"):
        known_path = _get_windows_known_folder(first_part)
        if known_path:
            # Replace first part with actual known folder path
            if len(parts) > 1:
                return os.path.join(known_path, *parts[1:])
            return known_path

    # Otherwise, treat as relative to cwd
    return os.path.join(cwd, path)


@dataclass
class ToolResult:
    """Result of a tool execution."""

    success: bool
    output: str
    error: str = ""


@dataclass
class ShellSession:
    """Maintains state for shell command execution."""

    cwd: str = field(default_factory=lambda: str(Path.home()))
    env: dict = field(default_factory=lambda: dict(os.environ))
    history: list = field(default_factory=list)


# Global shell session
_session = ShellSession()


# =============================================================================
# Tool Definitions (OpenAI Function Calling Format)
# =============================================================================

TOOL_DEFINITIONS = [
    {
        "type": "function",
        "function": {
            "name": "run_command",
            "description": "Execute a shell command and return the output. Use this to run any command-line operation.",
            "parameters": {
                "type": "object",
                "properties": {
                    "command": {
                        "type": "string",
                        "description": "The shell command to execute",
                    },
                    "working_dir": {
                        "type": "string",
                        "description": "Directory to run the command in. If not specified, uses the current working directory.",
                    },
                    "timeout": {
                        "type": "integer",
                        "description": "Maximum time in seconds to wait for the command. Default is 60.",
                        "default": 60,
                    },
                },
                "required": ["command"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "read_file",
            "description": "Read the contents of a file.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {
                        "type": "string",
                        "description": "Path to the file to read",
                    },
                    "max_lines": {
                        "type": "integer",
                        "description": "Maximum number of lines to read. If not specified, reads entire file.",
                    },
                },
                "required": ["path"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "write_file",
            "description": "Write content to a file. Creates the file if it doesn't exist.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {
                        "type": "string",
                        "description": "Path to the file to write",
                    },
                    "content": {
                        "type": "string",
                        "description": "Content to write to the file",
                    },
                    "append": {
                        "type": "boolean",
                        "description": "If true, append to the file instead of overwriting. Default is false.",
                        "default": False,
                    },
                },
                "required": ["path", "content"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "list_directory",
            "description": "List files and directories in a path.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {
                        "type": "string",
                        "description": "Path to the directory to list. Defaults to current working directory.",
                    },
                    "show_hidden": {
                        "type": "boolean",
                        "description": "Whether to show hidden files (starting with .). Default is false.",
                        "default": False,
                    },
                },
                "required": [],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "get_current_directory",
            "description": "Get the current working directory.",
            "parameters": {"type": "object", "properties": {}, "required": []},
        },
    },
    {
        "type": "function",
        "function": {
            "name": "change_directory",
            "description": "Change the current working directory.",
            "parameters": {
                "type": "object",
                "properties": {
                    "path": {"type": "string", "description": "Path to change to"}
                },
                "required": ["path"],
            },
        },
    },
    {
        "type": "function",
        "function": {
            "name": "task_complete",
            "description": "Call this when you have completed the user's task. Provide a summary of what was done.",
            "parameters": {
                "type": "object",
                "properties": {
                    "summary": {
                        "type": "string",
                        "description": "A summary of what was accomplished",
                    },
                    "files_modified": {
                        "type": "array",
                        "items": {"type": "string"},
                        "description": "List of files that were created or modified",
                    },
                },
                "required": ["summary"],
            },
        },
    },
]


# =============================================================================
# Tool Implementations
# =============================================================================


def run_command(command: str, working_dir: str = None, timeout: int = 60) -> ToolResult:
    """Execute a shell command."""
    global _session

    # Expand working directory (handles ~, relative paths, and Windows known folders like Desktop)
    if working_dir:
        cwd = _expand_path(working_dir, _session.cwd)
    else:
        cwd = _session.cwd

    try:
        # Use shell=True for proper command parsing
        # On Windows, use cmd.exe; on Unix, use bash
        if sys.platform == "win32":
            result = subprocess.run(
                command,
                shell=True,
                cwd=cwd,
                capture_output=True,
                text=True,
                timeout=timeout,
                env=_session.env,
            )
        else:
            result = subprocess.run(
                command,
                shell=True,
                cwd=cwd,
                capture_output=True,
                text=True,
                timeout=timeout,
                executable="/bin/bash",
                env=_session.env,
            )

        # Record in history
        _session.history.append(
            {"command": command, "cwd": cwd, "exit_code": result.returncode}
        )

        output = result.stdout
        error = result.stderr

        if result.returncode != 0:
            return ToolResult(
                success=False,
                output=output,
                error=f"Command failed with exit code {result.returncode}\n{error}",
            )

        return ToolResult(success=True, output=output, error=error)

    except subprocess.TimeoutExpired:
        return ToolResult(
            success=False, output="", error=f"Command timed out after {timeout} seconds"
        )
    except Exception as e:
        return ToolResult(success=False, output="", error=str(e))


def read_file(path: str, max_lines: int = None) -> ToolResult:
    """Read contents of a file."""
    try:
        # Expand path (handles ~, relative paths, and Windows known folders like Desktop)
        path = Path(_expand_path(path, _session.cwd))

        if not path.exists():
            return ToolResult(success=False, output="", error=f"File not found: {path}")

        if not path.is_file():
            return ToolResult(success=False, output="", error=f"Not a file: {path}")

        content = path.read_text(encoding="utf-8", errors="replace")

        if max_lines:
            lines = content.split("\n")[:max_lines]
            content = "\n".join(lines)
            if len(lines) == max_lines:
                content += f"\n... (truncated, showing first {max_lines} lines)"

        return ToolResult(success=True, output=content)

    except Exception as e:
        return ToolResult(success=False, output="", error=str(e))


def write_file(path: str, content: str, append: bool = False) -> ToolResult:
    """Write content to a file."""
    try:
        # Expand path (handles ~, relative paths, and Windows known folders like Desktop)
        path = Path(_expand_path(path, _session.cwd))

        # Create parent directories if needed
        path.parent.mkdir(parents=True, exist_ok=True)

        mode = "a" if append else "w"
        with open(path, mode, encoding="utf-8") as f:
            f.write(content)

        action = "Appended to" if append else "Wrote"
        return ToolResult(
            success=True, output=f"{action} {path} ({len(content)} bytes)"
        )

    except Exception as e:
        return ToolResult(success=False, output="", error=str(e))


def list_directory(path: str = None, show_hidden: bool = False) -> ToolResult:
    """List contents of a directory."""
    try:
        # Expand path (handles ~, relative paths, and Windows known folders like Desktop)
        if path is None:
            path = Path(_session.cwd)
        else:
            path = Path(_expand_path(path, _session.cwd))

        if not path.exists():
            return ToolResult(
                success=False, output="", error=f"Directory not found: {path}"
            )

        if not path.is_dir():
            return ToolResult(
                success=False, output="", error=f"Not a directory: {path}"
            )

        entries = []
        for entry in sorted(path.iterdir()):
            name = entry.name

            # Skip hidden files unless requested
            if not show_hidden and name.startswith("."):
                continue

            if entry.is_dir():
                entries.append(f"📁 {name}/")
            else:
                size = entry.stat().st_size
                entries.append(f"📄 {name} ({_format_size(size)})")

        output = f"Directory: {path}\n\n" + "\n".join(entries)
        return ToolResult(success=True, output=output)

    except Exception as e:
        return ToolResult(success=False, output="", error=str(e))


def get_current_directory() -> ToolResult:
    """Get the current working directory."""
    return ToolResult(success=True, output=_session.cwd)


def change_directory(path: str) -> ToolResult:
    """Change the current working directory."""
    global _session

    try:
        # Expand path (handles ~, relative paths, and Windows known folders like Desktop)
        path = os.path.abspath(_expand_path(path, _session.cwd))

        if not os.path.exists(path):
            return ToolResult(
                success=False, output="", error=f"Directory not found: {path}"
            )

        if not os.path.isdir(path):
            return ToolResult(
                success=False, output="", error=f"Not a directory: {path}"
            )

        _session.cwd = path
        return ToolResult(success=True, output=f"Changed directory to: {path}")

    except Exception as e:
        return ToolResult(success=False, output="", error=str(e))


def task_complete(summary: str, files_modified: list = None) -> ToolResult:
    """Signal that the task is complete."""
    output = f"✅ Task completed!\n\n{summary}"
    if files_modified:
        output += f"\n\nFiles modified:\n" + "\n".join(
            f"  • {f}" for f in files_modified
        )
    return ToolResult(success=True, output=output)


# =============================================================================
# Tool Dispatcher
# =============================================================================

TOOL_FUNCTIONS = {
    "run_command": run_command,
    "read_file": read_file,
    "write_file": write_file,
    "list_directory": list_directory,
    "get_current_directory": get_current_directory,
    "change_directory": change_directory,
    "task_complete": task_complete,
}


def execute_tool(name: str, arguments: dict) -> ToolResult:
    """Execute a tool by name with the given arguments."""
    if name not in TOOL_FUNCTIONS:
        return ToolResult(success=False, output="", error=f"Unknown tool: {name}")

    try:
        func = TOOL_FUNCTIONS[name]
        return func(**arguments)
    except TypeError as e:
        return ToolResult(
            success=False, output="", error=f"Invalid arguments for {name}: {e}"
        )
    except Exception as e:
        return ToolResult(
            success=False, output="", error=f"Error executing {name}: {e}"
        )


def reset_session():
    """Reset the shell session to initial state."""
    global _session
    _session = ShellSession()


def get_session_info() -> dict:
    """Get current session information."""
    return {
        "cwd": _session.cwd,
        "history_count": len(_session.history),
        "last_commands": _session.history[-5:] if _session.history else [],
    }


# =============================================================================
# Helpers
# =============================================================================


def _format_size(size: int) -> str:
    """Format file size in human-readable form."""
    for unit in ["B", "KB", "MB", "GB"]:
        if size < 1024:
            return f"{size:.1f} {unit}"
        size /= 1024
    return f"{size:.1f} TB"


# =============================================================================
# MCP Server (Optional - for external MCP client connections)
# =============================================================================


def get_mcp_tool_schemas() -> list:
    """Get tool definitions in MCP format."""
    mcp_tools = []
    for tool in TOOL_DEFINITIONS:
        func = tool["function"]
        mcp_tools.append(
            {
                "name": func["name"],
                "description": func["description"],
                "inputSchema": func["parameters"],
            }
        )
    return mcp_tools


def handle_mcp_tool_call(name: str, arguments: dict) -> dict:
    """Handle an MCP tool call and return MCP-formatted result."""
    result = execute_tool(name, arguments)

    content = result.output
    if result.error:
        content += f"\n\nError: {result.error}"

    return {
        "content": [{"type": "text", "text": content}],
        "isError": not result.success,
    }
