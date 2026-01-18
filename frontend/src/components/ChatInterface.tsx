import { useState, useRef, useEffect } from 'react';
import { conversation } from '../../wailsjs/go/models';
import AgentStepDisplay from './AgentStepDisplay';

interface Step {
  step_number: number;
  type: 'thinking' | 'tool_call' | 'tool_result' | 'complete' | 'error' | 'usage' | 'assistant_message';
  content: string;
  tool_name?: string;
  tool_args?: Record<string, unknown>;
  tool_result?: {
    success: boolean;
    output: string;
    error?: string;
  };
}

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  steps?: Step[];
}

interface SessionInfo {
  cwd: string;
  history_count: number;
}

interface ChatInterfaceProps {
  isConfigured: boolean;
  chatMessages: ChatMessage[];
  currentSteps: Step[];
  isRunning: boolean;
  sessionInfo: SessionInfo | null;
  activeConversation: conversation.Conversation | null;
  onSendMessage: (message: string, context: string) => void;
  onStopAgent: () => void;
  onNewConversation: () => void;
}

export default function ChatInterface({
  isConfigured,
  chatMessages,
  currentSteps,
  isRunning,
  sessionInfo,
  activeConversation,
  onSendMessage,
  onStopAgent,
  onNewConversation,
}: ChatInterfaceProps) {
  const [message, setMessage] = useState('');
  const [showContext, setShowContext] = useState(false);
  const [context, setContext] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [chatMessages, currentSteps]);

  useEffect(() => {
    if (inputRef.current && !isRunning) {
      inputRef.current.focus();
    }
  }, [activeConversation, isRunning]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (message.trim() && !isRunning) {
      onSendMessage(message.trim(), context.trim());
      setMessage('');
      setContext('');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e as unknown as React.FormEvent);
    }
  };

  const userMessageCount = chatMessages.filter(m => m.role === 'user').length;

  if (!isConfigured) {
    return (
      <div className="flex-1 flex items-center justify-center p-8 bg-matrix-black relative overflow-hidden">
        {/* Animated grid background */}
        <div className="absolute inset-0 opacity-20" style={{
          backgroundImage: 'linear-gradient(rgba(0, 255, 65, 0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 255, 65, 0.03) 1px, transparent 1px)',
          backgroundSize: '50px 50px'
        }} />
        
        <div className="text-center max-w-md relative z-10">
          {/* ASCII Art Logo */}
          <pre className="text-matrix-green text-glow text-[8px] leading-tight mb-6 font-mono">
{`
    ╔═══════════════════════════════════════╗
    ║     █████╗  ██████╗ ███████╗███╗   ██╗║
    ║    ██╔══██╗██╔════╝ ██╔════╝████╗  ██║║
    ║    ███████║██║  ███╗█████╗  ██╔██╗ ██║║
    ║    ██╔══██║██║   ██║██╔══╝  ██║╚██╗██║║
    ║    ██║  ██║╚██████╔╝███████╗██║ ╚████║║
    ║    ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝║
    ╚═══════════════════════════════════════╝
`}
          </pre>
          
          <div className="text-matrix-green text-glow font-mono">
            <div className="text-xl mb-2 brand-text tracking-widest">NEURAL_LINK_REQUIRED</div>
            <div className="text-sm text-matrix-green-dim">
              Configure LLM connection in the sidebar to initialize the system.
            </div>
          </div>
          <div className="mt-6 text-[10px] text-matrix-green-dark font-mono cursor-blink">
            AWAITING_CONFIGURATION
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full min-w-0 overflow-hidden bg-matrix-black">
      {/* Header */}
      <div className="px-4 py-3 border-b border-matrix-green/20 bg-matrix-darker panel-depth relative overflow-hidden">
        {/* Subtle scan line effect */}
        <div className="absolute inset-0 opacity-30 pointer-events-none" style={{
          background: 'repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0, 255, 65, 0.02) 2px, rgba(0, 255, 65, 0.02) 4px)'
        }} />
        
        <div className="flex items-center justify-between gap-4 relative z-10">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-3">
              {/* Status indicator */}
              <div className={`w-2 h-2 rounded-full ${isRunning ? 'bg-matrix-amber animate-pulse' : 'status-online'}`} />
              
              <h2 className="text-base font-medium text-matrix-green text-glow truncate brand-text tracking-wide">
                {activeConversation?.title || 'NEW_SESSION'}
              </h2>
            </div>
            <p className="text-[10px] text-matrix-green-dim font-mono mt-1 ml-5 uppercase tracking-wider">
              {activeConversation 
                ? `${userMessageCount} ${userMessageCount === 1 ? 'transmission' : 'transmissions'} logged`
                : 'Awaiting user input...'}
            </p>
          </div>
          
          {sessionInfo && (
            <div className="flex-shrink-0">
              <div className="text-[10px] font-mono bg-matrix-panel/80 px-3 py-1.5 rounded border border-matrix-border flex items-center gap-2 max-w-[220px]" title={sessionInfo.cwd}>
                <span className="text-matrix-cyan-dim">⌂</span>
                <span className="text-matrix-green truncate">
                  {sessionInfo.cwd.split('\\').pop() || sessionInfo.cwd.split('/').pop()}
                </span>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto p-4 bg-matrix-black">
        {chatMessages.length === 0 && !isRunning ? (
          <div className="h-full flex items-center justify-center relative">
            {/* Subtle grid pattern */}
            <div className="absolute inset-0 opacity-10" style={{
              backgroundImage: 'linear-gradient(rgba(0, 255, 65, 0.05) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 255, 65, 0.05) 1px, transparent 1px)',
              backgroundSize: '30px 30px'
            }} />
            
            <div className="text-center py-12 max-w-lg relative z-10">
              {/* Animated terminal prompt */}
              <div className="text-matrix-green text-glow-intense text-5xl mb-6 font-mono text-flicker">
                <span className="text-matrix-green-bright">❯</span>
                <span className="cursor-blink"></span>
              </div>
              
              <div className="text-matrix-green font-mono text-xl mb-3 brand-text tracking-wider">
                SYSTEM_READY
              </div>
              
              <p className="text-sm text-matrix-green-dim font-mono mb-8">
                Execute commands • Manipulate files • Automate tasks
              </p>
              
              <div className="inline-block border border-matrix-border bg-matrix-panel/50 rounded px-6 py-4">
                <div className="text-[11px] text-matrix-green-dim font-mono space-y-2 text-left">
                  <div className="flex items-center gap-3">
                    <span className="text-matrix-green">▸</span>
                    <span>Type your command below</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-matrix-amber">⏎</span>
                    <span>Press ENTER to transmit</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-matrix-cyan">⇧</span>
                    <span>SHIFT+ENTER for new line</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="space-y-4 max-w-4xl mx-auto">
            {chatMessages.map((msg, index) => (
              <div 
                key={msg.id} 
                className="message-animate"
                style={{ animationDelay: `${Math.min(index * 0.05, 0.3)}s` }}
              >
                {/* Message bubble */}
                <div className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                  <div className={`max-w-[85%] rounded px-4 py-3 transition-all duration-200 hover:shadow-glow-sm ${
                    msg.role === 'user'
                      ? 'message-user'
                      : msg.role === 'system'
                      ? 'message-system'
                      : 'message-assistant'
                  }`}>
                    {/* Role indicator */}
                    <div className={`text-[9px] font-mono mb-1 uppercase tracking-wide ${
                      msg.role === 'user' 
                        ? 'text-matrix-green-dim' 
                        : msg.role === 'system'
                        ? 'text-matrix-red-dim'
                        : 'text-matrix-cyan-dim'
                    }`}>
                      {msg.role === 'user' ? '> USER_INPUT' : msg.role === 'system' ? '! SYSTEM_MSG' : '< AI_RESPONSE'}
                    </div>
                    <div className="whitespace-pre-wrap break-words text-sm">{msg.content}</div>
                  </div>
                </div>

                {/* Show steps if this assistant message has them */}
                {msg.role === 'assistant' && msg.steps && msg.steps.length > 0 && (
                  <div className="mt-2 ml-4">
                    <details className="group">
                      <summary className="cursor-pointer text-[10px] text-matrix-green-dim hover:text-matrix-green flex items-center gap-2 font-mono uppercase">
                        <span className="group-open:rotate-90 transition-transform">▶</span>
                        <span>EXECUTION_LOG [{msg.steps.filter(s => s.type !== 'usage').length} operations]</span>
                      </summary>
                      <div className="mt-2 space-y-1 pl-4 border-l border-matrix-border">
                        {msg.steps
                          .filter(s => s.type !== 'usage')
                          .map((step, idx) => (
                            <AgentStepDisplay key={idx} step={step} compact />
                          ))}
                      </div>
                    </details>
                  </div>
                )}
              </div>
            ))}

            {/* Current steps (while agent is working) */}
            {currentSteps.length > 0 && (
              <div className="space-y-2 p-3 bg-matrix-panel rounded border border-matrix-green/30">
                <div className="text-[10px] font-mono text-matrix-green mb-2 flex items-center gap-2 uppercase tracking-wide">
                  <div className="typing-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                  EXECUTING...
                </div>
                {currentSteps
                  .filter(s => s.type !== 'usage')
                  .map((step, idx) => (
                    <AgentStepDisplay key={idx} step={step} />
                  ))}
              </div>
            )}

            {/* Loading indicator */}
            {isRunning && currentSteps.length === 0 && (
              <div className="flex justify-start">
                <div className="bg-matrix-panel border border-matrix-border rounded px-4 py-3">
                  <div className="flex items-center gap-3 text-matrix-green font-mono text-sm">
                    <div className="typing-dots">
                      <span></span>
                      <span></span>
                      <span></span>
                    </div>
                    <span className="text-matrix-green-dim">PROCESSING...</span>
                  </div>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* Input Area */}
      <div className="p-4 bg-matrix-darker border-t border-matrix-green/20 panel-depth">
        <form onSubmit={handleSubmit} className="max-w-4xl mx-auto">
          <div className="space-y-3">
            {showContext && (
              <div className="message-animate">
                <label className="block text-[10px] font-mono text-matrix-cyan mb-1.5 uppercase tracking-wider flex items-center gap-2">
                  <span className="text-matrix-cyan-dim">◇</span>
                  Additional_Context
                </label>
                <textarea
                  value={context}
                  onChange={(e) => setContext(e.target.value)}
                  placeholder="// Additional parameters or constraints..."
                  className="input-field resize-none text-xs border-matrix-cyan/30 focus:border-matrix-cyan"
                  rows={2}
                  disabled={isRunning}
                />
              </div>
            )}

            <div className="flex gap-3">
              <div className="flex-1 relative group">
                {/* Animated prompt indicator */}
                <div className={`absolute left-3 top-1/2 -translate-y-1/2 font-mono text-lg pointer-events-none transition-all duration-300 ${
                  isRunning ? 'text-matrix-amber animate-pulse' : 'text-matrix-green-bright text-glow'
                }`}>
                  ❯
                </div>
                <textarea
                  ref={inputRef}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder={isRunning ? "Processing..." : "Enter command..."}
                  className={`input-field resize-none pl-9 pr-12 font-mono text-sm ${
                    isRunning ? 'opacity-60' : 'group-focus-within:border-matrix-green group-focus-within:shadow-glow-sm'
                  }`}
                  rows={1}
                  disabled={isRunning}
                  style={{ minHeight: '48px', maxHeight: '200px' }}
                />
                <button
                  type="button"
                  onClick={() => setShowContext(!showContext)}
                  className={`absolute right-3 top-1/2 -translate-y-1/2 p-1.5 rounded-sm transition-all duration-200 ${
                    showContext 
                      ? 'text-matrix-cyan bg-matrix-cyan/10 shadow-glow-sm' 
                      : 'text-matrix-green-dim hover:text-matrix-green hover:bg-matrix-green/10'
                  }`}
                  title={showContext ? 'Hide context' : 'Add context'}
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                  </svg>
                </button>
              </div>

              {isRunning ? (
                <button
                  type="button"
                  onClick={onStopAgent}
                  className="btn-danger whitespace-nowrap text-xs uppercase tracking-wider px-5 py-3 glitch-hover"
                >
                  <span className="flex items-center gap-2">
                    <span className="w-2 h-2 bg-matrix-red rounded-full animate-pulse" />
                    ABORT
                  </span>
                </button>
              ) : (
                <button
                  type="submit"
                  disabled={!message.trim()}
                  className="btn-primary whitespace-nowrap text-xs uppercase tracking-wider px-5 py-3 glitch-hover"
                >
                  <span className="flex items-center gap-2">
                    <span>▶</span>
                    EXEC
                  </span>
                </button>
              )}
            </div>

            <div className="flex items-center justify-between text-[10px] text-matrix-green-dark font-mono pt-1">
              <div className="flex items-center gap-4">
                <span className="flex items-center gap-1">
                  <span className="text-matrix-green-dim">⏎</span> transmit
                </span>
                <span className="flex items-center gap-1">
                  <span className="text-matrix-green-dim">⇧⏎</span> newline
                </span>
              </div>
              {chatMessages.length > 0 && (
                <button
                  type="button"
                  onClick={onNewConversation}
                  className="text-matrix-green-dim hover:text-matrix-green uppercase tracking-wider transition-colors hover:text-glow"
                >
                  + NEW_SESSION
                </button>
              )}
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
