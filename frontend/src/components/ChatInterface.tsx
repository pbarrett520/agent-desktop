import { useState, useRef, useEffect } from 'react';
import { conversation } from '../../wailsjs/go/models';
import AgentStepDisplay from './AgentStepDisplay';
import AzureContextStrip from './AzureContextStrip';
import ApprovalCard, { Proposal } from './ApprovalCard';

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
  pendingProposal: Proposal | null;
  onApproveProposal: () => void;
  onDenyProposal: () => void;
  resolvingProposal: boolean;
  onOpenAudit: () => void;
  onOpenDashboard: () => void;
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
  pendingProposal,
  onApproveProposal,
  onDenyProposal,
  resolvingProposal,
  onOpenAudit,
  onOpenDashboard,
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

  const inputDisabled = isRunning || !!pendingProposal;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (message.trim() && !inputDisabled) {
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
      <div className="flex-1 flex items-center justify-center p-8 bg-brand-black">
        <div className="text-center max-w-md">
          <div className="w-14 h-14 mx-auto mb-6 rounded-2xl bg-brand-cyan/10 border border-brand-cyan/20 flex items-center justify-center">
            <svg className="w-7 h-7 text-brand-cyan" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
          </div>

          <div className="text-xl font-semibold mb-2">Connect a model provider</div>
          <p className="text-sm text-white/60">
            Configure an LLM connection in the sidebar to get started.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full min-w-0 overflow-hidden bg-brand-black">
      {/* Header */}
      <div className="border-b border-brand-border bg-brand-darker relative">
        {/* IG brand strip */}
        <div className="h-[3px] w-full" style={{ background: 'linear-gradient(90deg, #00D6F2, #FDCD01, #FF0068)' }} />

        <div className="px-4 py-3 flex items-center justify-between gap-4">
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-3">
              {/* Status indicator */}
              <div className={`w-2 h-2 rounded-full ${isRunning ? 'bg-brand-yellow animate-pulse' : 'status-online'}`} />

              <h2 className="text-base font-medium truncate">
                {activeConversation?.title || 'New chat'}
              </h2>
            </div>
            <p className="text-xs text-white/45 mt-1 ml-5">
              {activeConversation
                ? `${userMessageCount} ${userMessageCount === 1 ? 'message' : 'messages'}`
                : 'Waiting for your first message'}
            </p>
          </div>

          <div className="flex-shrink-0 flex items-center gap-2">
            <button
              onClick={onOpenDashboard}
              className="text-xs px-2.5 py-1.5 rounded-lg border border-brand-border bg-brand-panel text-white/60 hover:text-white transition-colors"
            >
              Overview
            </button>
            <button
              onClick={onOpenAudit}
              className="text-xs px-2.5 py-1.5 rounded-lg border border-brand-border bg-brand-panel text-white/60 hover:text-white transition-colors"
            >
              Audit
            </button>
            <AzureContextStrip />
            {sessionInfo && (
              <div className="text-xs bg-brand-panel px-3 py-1.5 rounded-lg border border-brand-border flex items-center gap-2 max-w-[220px] text-white/60" title={sessionInfo.cwd}>
                <span className="text-brand-yellow-dim">⌂</span>
                <span className="truncate">
                  {sessionInfo.cwd.split('\\').pop() || sessionInfo.cwd.split('/').pop()}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto p-4 bg-brand-black">
        {chatMessages.length === 0 && !isRunning ? (
          <div className="h-full flex items-center justify-center">
            <div className="text-center py-12 max-w-lg">
              <div className="w-14 h-14 mx-auto mb-6 rounded-2xl bg-brand-cyan/10 border border-brand-cyan/20 flex items-center justify-center">
                <svg className="w-7 h-7 text-brand-cyan" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>

              <div className="text-xl font-semibold mb-3">Ready when you are</div>

              <p className="text-sm text-white/60 mb-8">
                Execute commands, manipulate files, and automate tasks.
              </p>

              <div className="inline-block border border-brand-border bg-brand-panel rounded-lg px-6 py-4">
                <div className="text-xs text-white/60 space-y-2.5 text-left">
                  <div className="flex items-center gap-3">
                    <kbd className="px-1.5 py-0.5 rounded border border-brand-border-light text-[10px] text-white/80">Enter</kbd>
                    <span>Send message</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <kbd className="px-1.5 py-0.5 rounded border border-brand-border-light text-[10px] text-white/80">Shift+Enter</kbd>
                    <span>New line</span>
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
                  <div className={`max-w-[85%] rounded px-4 py-3 transition-all duration-200 ${
                    msg.role === 'user'
                      ? 'message-user'
                      : msg.role === 'system'
                      ? 'message-system'
                      : 'message-assistant'
                  }`}>
                    {/* Role indicator */}
                    <div className={`text-[10px] font-medium mb-1 ${
                      msg.role === 'user'
                        ? 'text-brand-cyan-dim'
                        : msg.role === 'system'
                        ? 'text-brand-magenta-dim'
                        : 'text-brand-yellow-dim'
                    }`}>
                      {msg.role === 'user' ? 'You' : msg.role === 'system' ? 'System' : 'Agent'}
                    </div>
                    <div className="whitespace-pre-wrap break-words text-sm">{msg.content}</div>
                  </div>
                </div>

                {/* Show steps if this assistant message has them */}
                {msg.role === 'assistant' && msg.steps && msg.steps.length > 0 && (
                  <div className="mt-2 ml-4">
                    <details className="group">
                      <summary className="cursor-pointer text-xs text-white/45 hover:text-white/70 flex items-center gap-2">
                        <span className="group-open:rotate-90 transition-transform">▶</span>
                        <span>{msg.steps.filter(s => s.type !== 'usage').length} steps</span>
                      </summary>
                      <div className="mt-2 space-y-1 pl-4 border-l border-brand-border">
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
              <div className="space-y-2 p-3 bg-brand-panel rounded-lg border border-brand-border">
                <div className="text-xs font-medium mb-2 flex items-center gap-2 text-white/70">
                  <div className="typing-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                  Working…
                </div>
                {currentSteps
                  .filter(s => s.type !== 'usage')
                  .map((step, idx) => (
                    <AgentStepDisplay key={idx} step={step} />
                  ))}
              </div>
            )}

            {/* Approval card for a pending az_propose */}
            {pendingProposal && (
              <div className="message-animate">
                <ApprovalCard
                  proposal={pendingProposal}
                  onApprove={onApproveProposal}
                  onDeny={onDenyProposal}
                  resolving={resolvingProposal}
                />
              </div>
            )}

            {/* Loading indicator */}
            {isRunning && currentSteps.length === 0 && !pendingProposal && (
              <div className="flex justify-start">
                <div className="bg-brand-panel border border-brand-border rounded-lg px-4 py-3">
                  <div className="flex items-center gap-3 text-sm text-white/70">
                    <div className="typing-dots">
                      <span></span>
                      <span></span>
                      <span></span>
                    </div>
                    <span>Thinking…</span>
                  </div>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* Input Area */}
      <div className="p-4 bg-brand-darker border-t border-brand-border">
        <form onSubmit={handleSubmit} className="max-w-4xl mx-auto">
          <div className="space-y-3">
            {showContext && (
              <div className="message-animate">
                <label className="block text-xs font-medium text-brand-yellow mb-1.5 flex items-center gap-2">
                  Additional context
                </label>
                <textarea
                  value={context}
                  onChange={(e) => setContext(e.target.value)}
                  placeholder="// Additional parameters or constraints..."
                  className="input-field resize-none text-xs border-brand-yellow/30 focus:border-brand-yellow"
                  rows={2}
                  disabled={inputDisabled}
                />
              </div>
            )}

            <div className="flex gap-3">
              <div className="flex-1 relative group">
                <textarea
                  ref={inputRef}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder={pendingProposal ? "Waiting for approval decision…" : isRunning ? "Processing…" : "Message the agent…"}
                  className={`input-field resize-none pr-12 text-sm ${
                    inputDisabled ? 'opacity-60' : 'group-focus-within:border-brand-cyan'
                  }`}
                  rows={1}
                  disabled={inputDisabled}
                  style={{ minHeight: '48px', maxHeight: '200px' }}
                />
                <button
                  type="button"
                  onClick={() => setShowContext(!showContext)}
                  className={`absolute right-3 top-1/2 -translate-y-1/2 p-1.5 rounded-sm transition-all duration-200 ${
                    showContext
                      ? 'text-brand-yellow bg-brand-yellow/10'
                      : 'text-white/40 hover:text-white hover:bg-white/10'
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
                  className="btn-danger whitespace-nowrap text-sm px-5 py-3"
                >
                  <span className="flex items-center gap-2">
                    <span className="w-2 h-2 bg-brand-magenta rounded-full animate-pulse" />
                    Stop
                  </span>
                </button>
              ) : (
                <button
                  type="submit"
                  disabled={!message.trim()}
                  className="btn-primary whitespace-nowrap text-sm px-5 py-3"
                >
                  Send
                </button>
              )}
            </div>

            <div className="flex items-center justify-between text-xs text-white/35 pt-1">
              <div className="flex items-center gap-4">
                <span className="flex items-center gap-1.5">
                  <kbd className="px-1 py-0.5 rounded border border-brand-border text-[10px]">Enter</kbd> to send
                </span>
                <span className="flex items-center gap-1.5">
                  <kbd className="px-1 py-0.5 rounded border border-brand-border text-[10px]">Shift+Enter</kbd> for new line
                </span>
              </div>
              {chatMessages.length > 0 && (
                <button
                  type="button"
                  onClick={onNewConversation}
                  className="text-white/50 hover:text-white transition-colors"
                >
                  + New chat
                </button>
              )}
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
