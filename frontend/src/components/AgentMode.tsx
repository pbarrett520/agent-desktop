import { useState, useRef, useEffect } from 'react';
import AgentStepDisplay from './AgentStepDisplay';

interface Step {
  step_number: number;
  type: 'thinking' | 'tool_call' | 'tool_result' | 'complete' | 'error' | 'usage';
  content: string;
  tool_name?: string;
  tool_args?: Record<string, unknown>;
  tool_result?: {
    success: boolean;
    output: string;
    error?: string;
  };
  usage?: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

interface SessionInfo {
  cwd: string;
  history_count: number;
}

interface AgentModeProps {
  isConfigured: boolean;
  steps: Step[];
  isRunning: boolean;
  sessionInfo: SessionInfo | null;
  onRunTask: (task: string, context: string) => void;
  onStopAgent: () => void;
  onResetSession: () => void;
}

export default function AgentMode({
  isConfigured,
  steps,
  isRunning,
  sessionInfo,
  onRunTask,
  onStopAgent,
  onResetSession,
}: AgentModeProps) {
  const [task, setTask] = useState('');
  const [context, setContext] = useState('');
  const [showContext, setShowContext] = useState(false);
  const stepsEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to bottom when new steps arrive
  useEffect(() => {
    if (stepsEndRef.current) {
      stepsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [steps]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (task.trim() && !isRunning) {
      onRunTask(task.trim(), context.trim());
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      handleSubmit(e as unknown as React.FormEvent);
    }
  };

  if (!isConfigured) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="text-center max-w-md">
          <div className="text-6xl mb-4">⚙️</div>
          <h2 className="text-h4 font-bold text-secondary-navy mb-2">
            Configure Azure OpenAI
          </h2>
          <p className="text-neutral-gray">
            Please configure your Azure OpenAI credentials in the sidebar to get started.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full min-w-0 overflow-hidden">
      {/* Header */}
      <div className="p-4 border-b border-neutral-light bg-white">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-h4 font-bold text-secondary-navy">Agent Mode</h2>
            <p className="text-sm text-neutral-gray">
              The AI assistant can execute commands and manage files to complete tasks.
            </p>
          </div>
          {sessionInfo && (
            <div className="text-right text-sm text-neutral-gray">
              <div className="font-mono truncate max-w-xs" title={sessionInfo.cwd}>
                {sessionInfo.cwd}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Task Input Form */}
      <form onSubmit={handleSubmit} className="p-4 bg-white border-b border-neutral-light">
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-neutral-gray mb-1">
              What would you like me to do?
            </label>
            <textarea
              value={task}
              onChange={(e) => setTask(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Example: Find all Python files larger than 1MB in my Documents folder"
              className="input-field resize-none"
              rows={3}
              disabled={isRunning}
            />
            <p className="text-xs text-neutral-light mt-1">
              Press Ctrl+Enter to run
            </p>
          </div>

          {showContext && (
            <div>
              <label className="block text-sm font-medium text-neutral-gray mb-1">
                Additional Context
              </label>
              <textarea
                value={context}
                onChange={(e) => setContext(e.target.value)}
                placeholder="Any specific requirements or constraints..."
                className="input-field resize-none"
                rows={2}
                disabled={isRunning}
              />
            </div>
          )}

          <div className="flex items-center gap-3">
            {isRunning ? (
              <button
                type="button"
                onClick={onStopAgent}
                className="btn-danger"
              >
                Stop Agent
              </button>
            ) : (
              <button
                type="submit"
                disabled={!task.trim()}
                className="btn-primary"
              >
                Run Task
              </button>
            )}

            <button
              type="button"
              onClick={() => setShowContext(!showContext)}
              className="text-sm text-primary-blue hover:underline"
            >
              {showContext ? 'Hide Context' : 'Add Context'}
            </button>

            {steps.length > 0 && !isRunning && (
              <button
                type="button"
                onClick={() => {
                  onResetSession();
                  setTask('');
                  setContext('');
                }}
                className="text-sm text-neutral-gray hover:underline ml-auto"
              >
                Clear & Reset
              </button>
            )}
          </div>
        </div>
      </form>

      {/* Steps Display */}
      <div className="flex-1 overflow-y-auto p-4 bg-gray-50">
        {steps.length === 0 ? (
          <div className="text-center text-neutral-gray py-12">
            <div className="text-4xl mb-3">🤖</div>
            <p>Enter a task above to get started.</p>
            <p className="text-sm mt-2">
              The agent will execute commands and manage files to complete your task.
            </p>
          </div>
        ) : (
          <div className="space-y-3">
            {steps.map((step, index) => (
              <AgentStepDisplay key={index} step={step} />
            ))}
            <div ref={stepsEndRef} />
          </div>
        )}

        {isRunning && (
          <div className="flex items-center gap-2 text-primary-blue mt-4">
            <div className="animate-spin h-4 w-4 border-2 border-primary-blue border-t-transparent rounded-full"></div>
            <span>Agent is working...</span>
          </div>
        )}
      </div>
    </div>
  );
}
