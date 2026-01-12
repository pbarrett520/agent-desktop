"""
Agentic loop for Agent Desktop.

This module implements the core agent loop that allows the LLM to 
iteratively call tools until the task is complete.
"""

import json
import sys
from typing import Generator, Callable, Optional
from dataclasses import dataclass
from openai import OpenAI, AzureOpenAI

from tools import TOOL_DEFINITIONS, execute_tool, ToolResult, reset_session


def get_os_instructions() -> str:
    """Get OS-specific instructions for the system prompt."""
    if sys.platform == "darwin":
        return "The user is on macOS, so use Unix-compatible commands (mv, cp, rm, ls, etc.) or Python scripts."
    elif sys.platform == "win32":
        return "The user is on Windows, so use Windows-compatible commands (dir, copy, del, etc.) or Python scripts."
    else:
        return "The user is on Linux, so use Unix-compatible commands (mv, cp, rm, ls, etc.) or Python scripts."


@dataclass
class AgentMessage:
    """A message in the agent conversation."""
    role: str  # "user", "assistant", "tool"
    content: str
    tool_calls: list = None
    tool_call_id: str = None
    name: str = None  # tool name for tool messages


@dataclass
class AgentStep:
    """A single step in the agent's execution."""
    step_number: int
    type: str  # "thinking", "tool_call", "tool_result", "complete", "error"
    content: str
    tool_name: str = None
    tool_args: dict = None
    tool_result: ToolResult = None


SYSTEM_PROMPT_TEMPLATE = """You are an AI assistant that helps users accomplish tasks by executing commands and managing files.

You have access to the following tools:
- run_command: Execute shell commands
- read_file: Read file contents
- write_file: Write to files
- list_directory: List directory contents
- get_current_directory: Get current working directory
- change_directory: Change working directory
- task_complete: Signal that the task is finished

CRITICAL RULES:
1. You MUST call task_complete when you have finished the user's task
2. Do NOT output multiple text responses - always make a tool call
3. After getting a tool result that completes the task, immediately call task_complete
4. Break complex tasks into smaller steps
5. If a command fails, try to understand why and fix it
6. Be careful with destructive operations - list files before deleting

{os_instructions}

WORKFLOW:
1. Analyze the task
2. Call appropriate tools to complete it
3. Once done, ALWAYS call task_complete with a summary
"""


def get_system_prompt() -> str:
    """Get the system prompt with OS-specific instructions."""
    return SYSTEM_PROMPT_TEMPLATE.format(os_instructions=get_os_instructions())


def run_agent_loop(
    client: OpenAI | AzureOpenAI,
    model: str,
    task: str,
    context: str = "",
    max_steps: int = 20,
    on_step: Optional[Callable[[AgentStep], None]] = None
) -> Generator[AgentStep, None, None]:
    """
    Run the agent loop to complete a task.
    
    Args:
        client: OpenAI-compatible client
        model: Model name to use
        task: The user's task description
        context: Additional context (working directory, etc.)
        max_steps: Maximum number of tool calls allowed
        on_step: Optional callback for each step
        
    Yields:
        AgentStep objects for each step of execution
    """
    # Reset session for fresh start
    reset_session()
    
    # Build initial messages
    messages = [
        {"role": "system", "content": get_system_prompt()},
        {"role": "user", "content": f"{task}\n\n{context}" if context else task}
    ]
    
    step_number = 0
    consecutive_text_responses = 0  # Track responses without tool calls
    max_text_responses = 2  # Force completion after this many text-only responses
    
    while step_number < max_steps:
        step_number += 1
        
        try:
            # Call the LLM with tools
            response = client.chat.completions.create(
                model=model,
                messages=messages,
                tools=TOOL_DEFINITIONS,
                tool_choice="auto"
            )
            
            message = response.choices[0].message
            
            # Check if the model wants to use tools
            if message.tool_calls:
                consecutive_text_responses = 0  # Reset counter when tools are called
                # Add assistant message with tool calls
                messages.append({
                    "role": "assistant",
                    "content": message.content or "",
                    "tool_calls": [
                        {
                            "id": tc.id,
                            "type": "function",
                            "function": {
                                "name": tc.function.name,
                                "arguments": tc.function.arguments
                            }
                        }
                        for tc in message.tool_calls
                    ]
                })
                
                # If there's thinking content, yield it
                if message.content:
                    step = AgentStep(
                        step_number=step_number,
                        type="thinking",
                        content=message.content
                    )
                    if on_step:
                        on_step(step)
                    yield step
                
                # Process each tool call
                for tool_call in message.tool_calls:
                    tool_name = tool_call.function.name
                    
                    try:
                        tool_args = json.loads(tool_call.function.arguments)
                    except json.JSONDecodeError:
                        tool_args = {}
                    
                    # Yield the tool call step
                    call_step = AgentStep(
                        step_number=step_number,
                        type="tool_call",
                        content=f"Calling {tool_name}",
                        tool_name=tool_name,
                        tool_args=tool_args
                    )
                    if on_step:
                        on_step(call_step)
                    yield call_step
                    
                    # Execute the tool
                    result = execute_tool(tool_name, tool_args)
                    
                    # Build result content
                    result_content = result.output
                    if result.error:
                        result_content += f"\n\nError: {result.error}"
                    
                    # Add tool result to messages
                    messages.append({
                        "role": "tool",
                        "tool_call_id": tool_call.id,
                        "content": result_content
                    })
                    
                    # Yield the tool result step
                    result_step = AgentStep(
                        step_number=step_number,
                        type="tool_result",
                        content=result_content,
                        tool_name=tool_name,
                        tool_result=result
                    )
                    if on_step:
                        on_step(result_step)
                    yield result_step
                    
                    # Check if task_complete was called
                    if tool_name == "task_complete":
                        complete_step = AgentStep(
                            step_number=step_number,
                            type="complete",
                            content=result.output
                        )
                        if on_step:
                            on_step(complete_step)
                        yield complete_step
                        return
            
            else:
                # No tool calls - model is done or wants to respond
                consecutive_text_responses += 1
                
                if message.content:
                    # Check if this looks like a completion
                    content = message.content.lower()
                    is_complete = any(phrase in content for phrase in [
                        "completed", "done", "finished", "task complete",
                        "let me know", "anything else", "help you with"
                    ])
                    
                    if is_complete or consecutive_text_responses >= max_text_responses:
                        complete_step = AgentStep(
                            step_number=step_number,
                            type="complete",
                            content=message.content
                        )
                        if on_step:
                            on_step(complete_step)
                        yield complete_step
                        return
                    else:
                        # Model wants to say something without tools
                        thinking_step = AgentStep(
                            step_number=step_number,
                            type="thinking",
                            content=message.content
                        )
                        if on_step:
                            on_step(thinking_step)
                        yield thinking_step
                        
                        # Add to messages and continue
                        messages.append({
                            "role": "assistant",
                            "content": message.content
                        })
                else:
                    # Empty response - something went wrong
                    error_step = AgentStep(
                        step_number=step_number,
                        type="error",
                        content="Received empty response from model"
                    )
                    if on_step:
                        on_step(error_step)
                    yield error_step
                    return
                    
        except Exception as e:
            error_step = AgentStep(
                step_number=step_number,
                type="error",
                content=f"Error: {str(e)}"
            )
            if on_step:
                on_step(error_step)
            yield error_step
            return
    
    # Max steps reached
    max_step = AgentStep(
        step_number=step_number,
        type="error",
        content=f"Maximum steps ({max_steps}) reached without completing the task"
    )
    if on_step:
        on_step(max_step)
    yield max_step


def format_step_for_display(step: AgentStep) -> dict:
    """Format a step for UI display."""
    icons = {
        "thinking": "💭",
        "tool_call": "🔧",
        "tool_result": "📤",
        "complete": "✅",
        "error": "❌"
    }
    
    return {
        "icon": icons.get(step.type, "•"),
        "type": step.type,
        "step": step.step_number,
        "content": step.content,
        "tool_name": step.tool_name,
        "tool_args": step.tool_args,
        "success": step.tool_result.success if step.tool_result else None
    }

