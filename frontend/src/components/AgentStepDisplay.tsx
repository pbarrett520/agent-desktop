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
        return { icon: '◇', color: 'text-matrix-cyan', label: 'ANALYZE' };
      case 'tool_call':
        return { icon: '▶', color: 'text-matrix-amber', label: 'EXEC' };
      case 'tool_result':
        return step.tool_result?.success 
          ? { icon: '✓', color: 'text-matrix-green', label: 'OK' }
          : { icon: '✗', color: 'text-matrix-red', label: 'ERR' };
      case 'complete':
        return { icon: '■', color: 'text-matrix-green', label: 'DONE' };
      case 'assistant_message':
        return { icon: '◆', color: 'text-matrix-cyan', label: 'MSG' };
      case 'error':
        return { icon: '!', color: 'text-matrix-red', label: 'FAIL' };
      default:
        return { icon: '•', color: 'text-matrix-green-dim', label: 'INFO' };
    }
  };

  const getBorderColor = () => {
    switch (step.type) {
      case 'thinking':
        return 'border-matrix-cyan/50';
      case 'tool_call':
        return 'border-matrix-amber/50';
      case 'tool_result':
        return step.tool_result?.success ? 'border-matrix-green/50' : 'border-matrix-red/50';
      case 'complete':
        return 'border-matrix-green/50';
      case 'error':
        return 'border-matrix-red/50';
      default:
        return 'border-matrix-border';
    }
  };

  const indicator = getStepIndicator();

  const renderContent = () => {
    switch (step.type) {
      case 'thinking':
        return (
          <div className="text-matrix-cyan-dim text-xs whitespace-pre-wrap font-mono">
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
              <span className="text-[10px] text-matrix-green-dim group-hover:text-matrix-green transition-all duration-200">
                {isExpanded ? '▼' : '▶'}
              </span>
              <span className="font-mono text-xs text-matrix-amber flex items-center gap-2">
                <span className="text-matrix-amber-dim">⚡</span>
                {step.tool_name}
              </span>
            </div>
            {isExpanded && step.tool_args && (
              <pre className="mt-2 p-3 bg-matrix-black border border-matrix-amber/20 rounded text-[10px] overflow-x-auto text-matrix-green-dim font-mono max-w-full break-words whitespace-pre-wrap panel-depth">
                {JSON.stringify(step.tool_args, null, 2)}
              </pre>
            )}
          </div>
        );

      case 'tool_result':
        const success = step.tool_result?.success ?? true;
        return (
          <div className="message-animate">
            <div 
              className="flex items-center gap-2 cursor-pointer select-none group py-0.5"
              onClick={() => setIsExpanded(!isExpanded)}
            >
              <span className="text-[10px] text-matrix-green-dim group-hover:text-matrix-green transition-all duration-200">
                {isExpanded ? '▼' : '▶'}
              </span>
              <span className={`font-mono text-xs flex items-center gap-2 ${success ? 'text-matrix-green' : 'text-matrix-red'}`}>
                {success ? <span className="text-matrix-green">✓</span> : <span className="text-matrix-red">✗</span>}
                {step.tool_name || 'OUTPUT'}
              </span>
            </div>
            {isExpanded && (
              <pre className={`mt-2 p-3 rounded text-[10px] overflow-x-auto whitespace-pre-wrap break-words max-w-full font-mono panel-depth ${
                success 
                  ? 'bg-matrix-green/5 border border-matrix-green/20 text-matrix-green-dim' 
                  : 'bg-matrix-red/5 border border-matrix-red/20 text-matrix-red'
              }`}>
                {step.content || 'No output'}
              </pre>
            )}
          </div>
        );

      case 'complete':
        return (
          <div className="bg-matrix-green/10 border border-matrix-green/30 p-3 rounded shadow-glow-sm message-animate">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-matrix-green text-glow animate-pulse">■</span>
              <span className="font-mono text-xs text-matrix-green uppercase tracking-wider brand-text">TASK_COMPLETE</span>
            </div>
            <div className="text-matrix-green-dim text-xs whitespace-pre-wrap font-mono">
              {step.content}
            </div>
          </div>
        );

      case 'error':
        return (
          <div className="bg-matrix-red/10 border border-matrix-red/30 p-3 rounded overflow-hidden message-animate">
            <div className="flex items-center gap-2 mb-2">
              <span className="text-matrix-red animate-pulse">!</span>
              <span className="font-mono text-xs text-matrix-red uppercase tracking-wider">SYSTEM_ERROR</span>
            </div>
            <div className="text-matrix-red text-xs whitespace-pre-wrap break-words font-mono">
              {step.content}
            </div>
          </div>
        );

      default:
        return <div className="text-matrix-green-dim text-xs font-mono">{step.content}</div>;
    }
  };

  return (
    <div className={`border-l-2 ${getBorderColor()} ${compact ? 'pl-2 py-1' : 'pl-3 py-2'}`}>
      <div className="flex items-start gap-2">
        {/* Step indicator */}
        <div className={`flex-shrink-0 ${compact ? 'text-xs' : 'text-sm'}`}>
          <span className={`${indicator.color} font-mono`}>{indicator.icon}</span>
        </div>
        
        {/* Label (only in non-compact mode) */}
        {!compact && (
          <div className="flex-shrink-0 w-12">
            <span className={`text-[9px] font-mono uppercase tracking-wide ${indicator.color}`}>
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
