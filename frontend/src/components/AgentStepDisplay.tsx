import { useState } from 'react';

interface ToolResult {
  success: boolean;
  output: string;
  error?: string;
}

interface Step {
  step_number: number;
  type: 'thinking' | 'tool_call' | 'tool_result' | 'complete' | 'error' | 'usage' | 'assistant_message';
  content: string;
  tool_name?: string;
  tool_args?: Record<string, unknown>;
  tool_result?: ToolResult;
  usage?: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

interface AgentStepDisplayProps {
  step: Step;
  compact?: boolean; // Compact mode for chat history
}

export default function AgentStepDisplay({ step, compact = false }: AgentStepDisplayProps) {
  const [isExpanded, setIsExpanded] = useState(!compact); // Collapsed by default in compact mode

  // Skip usage steps in display (they're tracked separately)
  if (step.type === 'usage') {
    return null;
  }

  const getStepIcon = () => {
    switch (step.type) {
      case 'thinking':
        return '💭';
      case 'tool_call':
        return '🔧';
      case 'tool_result':
        return step.tool_result?.success ? '✅' : '❌';
      case 'complete':
        return '✅';
      case 'assistant_message':
        return '💬';
      case 'error':
        return '❌';
      default:
        return '•';
    }
  };

  const getStepColor = () => {
    switch (step.type) {
      case 'thinking':
        return 'border-l-primary-blue';
      case 'tool_call':
        return 'border-l-secondary-purple';
      case 'tool_result':
        return step.tool_result?.success ? 'border-l-secondary-lime' : 'border-l-secondary-coral';
      case 'complete':
        return 'border-l-secondary-lime';
      case 'error':
        return 'border-l-secondary-coral';
      default:
        return 'border-l-neutral-light';
    }
  };

  const renderContent = () => {
    switch (step.type) {
      case 'thinking':
        return (
          <div className="text-neutral-gray whitespace-pre-wrap">
            {step.content}
          </div>
        );

      case 'tool_call':
        return (
          <div>
            <div 
              className="flex items-center gap-2 cursor-pointer select-none"
              onClick={() => setIsExpanded(!isExpanded)}
            >
              <span className="text-sm text-neutral-gray">{isExpanded ? '▼' : '▶'}</span>
              <span className="font-medium text-secondary-navy">{step.tool_name}</span>
            </div>
            {isExpanded && step.tool_args && (
              <pre className="mt-2 p-3 bg-neutral-light/30 rounded-md text-sm overflow-x-auto text-neutral-gray max-w-full break-words whitespace-pre-wrap">
                {JSON.stringify(step.tool_args, null, 2)}
              </pre>
            )}
          </div>
        );

      case 'tool_result':
        const success = step.tool_result?.success ?? true;
        return (
          <div>
            <div 
              className="flex items-center gap-2 cursor-pointer select-none"
              onClick={() => setIsExpanded(!isExpanded)}
            >
              <span className="text-sm text-neutral-gray">{isExpanded ? '▼' : '▶'}</span>
              <span className={`font-medium ${success ? 'text-secondary-navy' : 'text-secondary-coral'}`}>
                Result: {step.tool_name}
              </span>
            </div>
            {isExpanded && (
              <pre className={`mt-2 p-3 rounded-md text-sm overflow-x-auto whitespace-pre-wrap break-words max-w-full ${
                success ? 'bg-secondary-lime/10 text-neutral-gray' : 'bg-secondary-coral/10 text-secondary-coral'
              }`}>
                {step.content || 'No output'}
              </pre>
            )}
          </div>
        );

      case 'complete':
        return (
          <div className="bg-secondary-lime/10 p-3 rounded-md">
            <span className="font-bold text-secondary-navy">Task Complete</span>
            <div className="mt-2 text-neutral-gray whitespace-pre-wrap">
              {step.content}
            </div>
          </div>
        );

      case 'error':
        return (
          <div className="bg-secondary-coral/10 p-3 rounded-md overflow-hidden">
            <span className="font-bold text-secondary-coral">Error</span>
            <div className="mt-2 text-secondary-coral whitespace-pre-wrap break-words">
              {step.content}
            </div>
          </div>
        );

      default:
        return <div className="text-neutral-gray">{step.content}</div>;
    }
  };

  return (
    <div className={`border-l-4 ${getStepColor()} ${compact ? 'pl-2 py-1' : 'pl-4 py-2'}`}>
      <div className="flex items-start gap-2">
        <span className={`${compact ? 'text-sm' : 'text-lg'} flex-shrink-0`}>{getStepIcon()}</span>
        <div className={`flex-1 min-w-0 ${compact ? 'text-sm' : ''}`}>
          {renderContent()}
        </div>
      </div>
    </div>
  );
}
