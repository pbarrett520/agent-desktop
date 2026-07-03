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
  compact?: boolean;
}

export default function AgentStepDisplay({ step, compact = false }: AgentStepDisplayProps) {
  const [isExpanded, setIsExpanded] = useState(!compact);

  if (step.type === 'usage') {
    return null;
  }

  const getStepIndicator = () => {
    switch (step.type) {
      case 'thinking':
        return { icon: '◇', color: 'text-brand-yellow', label: 'Thinking' };
      case 'tool_call':
        return { icon: '▶', color: 'text-brand-yellow', label: 'Running' };
      case 'tool_result':
        return step.tool_result?.success
          ? { icon: '✓', color: 'text-brand-cyan', label: 'Done' }
          : { icon: '✗', color: 'text-brand-magenta', label: 'Failed' };
      case 'complete':
        return { icon: '■', color: 'text-brand-cyan', label: 'Complete' };
      case 'assistant_message':
        return { icon: '◆', color: 'text-brand-yellow', label: 'Message' };
      case 'error':
        return { icon: '!', color: 'text-brand-magenta', label: 'Error' };
      default:
        return { icon: '•', color: 'text-white/40', label: 'Info' };
    }
  };

  const getBorderColor = () => {
    switch (step.type) {
      case 'thinking':
        return 'border-brand-yellow/50';
      case 'tool_call':
        return 'border-brand-yellow/50';
      case 'tool_result':
        return step.tool_result?.success ? 'border-brand-cyan/50' : 'border-brand-magenta/50';
      case 'complete':
        return 'border-brand-cyan/50';
      case 'error':
        return 'border-brand-magenta/50';
      default:
        return 'border-brand-border';
    }
  };

  const indicator = getStepIndicator();

  const renderContent = () => {
    switch (step.type) {
      case 'thinking':
        return (
          <div className="text-white/60 text-xs whitespace-pre-wrap">
            {step.content}
          </div>
        );

      case 'tool_call':
        return (
          <div className="message-animate">
            <div
              className="flex items-center gap-2 cursor-pointer select-none group py-0.5"
              onClick={() => setIsExpanded(!isExpanded)}
            >
              <span className="text-[10px] text-white/35 group-hover:text-white/70 transition-all duration-200">
                {isExpanded ? '▼' : '▶'}
              </span>
              <span className="text-xs text-brand-yellow flex items-center gap-2">
                {step.tool_name}
              </span>
            </div>
            {isExpanded && step.tool_args && (
              <pre className="mt-2 p-3 bg-brand-black border border-brand-border rounded-lg text-[11px] overflow-x-auto text-white/60 font-mono max-w-full break-words whitespace-pre-wrap">
                {JSON.stringify(step.tool_args, null, 2)}
              </pre>
            )}
          </div>
        );

      case 'tool_result': {
        const success = step.tool_result?.success ?? true;
        return (
          <div className="message-animate">
            <div
              className="flex items-center gap-2 cursor-pointer select-none group py-0.5"
              onClick={() => setIsExpanded(!isExpanded)}
            >
              <span className="text-[10px] text-white/35 group-hover:text-white/70 transition-all duration-200">
                {isExpanded ? '▼' : '▶'}
              </span>
              <span className={`text-xs flex items-center gap-2 ${success ? 'text-brand-cyan' : 'text-brand-magenta'}`}>
                {success ? <span>✓</span> : <span>✗</span>}
                {step.tool_name || 'Output'}
              </span>
            </div>
            {isExpanded && (
              <pre className={`mt-2 p-3 rounded-lg text-[11px] overflow-x-auto whitespace-pre-wrap break-words max-w-full font-mono ${
                success
                  ? 'bg-brand-cyan/5 border border-brand-cyan/20 text-white/60'
                  : 'bg-brand-magenta/5 border border-brand-magenta/20 text-brand-magenta'
              }`}>
                {step.content || 'No output'}
              </pre>
            )}
          </div>
        );
      }

      case 'complete':
        return (
          <div className="bg-brand-cyan/10 border border-brand-cyan/30 p-3 rounded-lg message-animate">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-brand-cyan">■</span>
              <span className="text-xs font-medium text-brand-cyan">Task complete</span>
            </div>
            <div className="text-white/70 text-xs whitespace-pre-wrap">
              {step.content}
            </div>
          </div>
        );

      case 'error':
        return (
          <div className="bg-brand-magenta/10 border border-brand-magenta/30 p-3 rounded-lg overflow-hidden message-animate">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-brand-magenta">!</span>
              <span className="text-xs font-medium text-brand-magenta">Error</span>
            </div>
            <div className="text-brand-magenta text-xs whitespace-pre-wrap break-words">
              {step.content}
            </div>
          </div>
        );

      default:
        return <div className="text-white/60 text-xs">{step.content}</div>;
    }
  };

  return (
    <div className={`border-l-2 ${getBorderColor()} ${compact ? 'pl-2 py-1' : 'pl-3 py-2'}`}>
      <div className="flex items-start gap-2">
        {/* Step indicator */}
        <div className={`flex-shrink-0 ${compact ? 'text-xs' : 'text-sm'}`}>
          <span className={indicator.color}>{indicator.icon}</span>
        </div>

        {/* Label (only in non-compact mode) */}
        {!compact && (
          <div className="flex-shrink-0 w-16">
            <span className={`text-[10px] font-medium ${indicator.color}`}>
              {indicator.label}
            </span>
          </div>
        )}

        {/* Content */}
        <div className={`flex-1 min-w-0 ${compact ? 'text-xs' : ''}`}>
          {renderContent()}
        </div>
      </div>
    </div>
  );
}
