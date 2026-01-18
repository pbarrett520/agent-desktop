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

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [chatMessages, currentSteps]);

  // Focus input when conversation changes
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

  // Count user messages for display
  const userMessageCount = chatMessages.filter(m => m.role === 'user').length;

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
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <h2 className="text-h4 font-bold text-secondary-navy truncate">
              {activeConversation?.title || 'New Chat'}
            </h2>
            <p className="text-sm text-neutral-gray">
              {activeConversation 
                ? `${userMessageCount} ${userMessageCount === 1 ? 'message' : 'messages'}`
                : 'Start a conversation with your AI assistant'}
            </p>
          </div>
          {sessionInfo && (
            <div className="flex-shrink-0 text-right">
              <div className="text-xs text-neutral-gray font-mono bg-gray-100 px-2 py-1 rounded truncate max-w-[200px]" title={sessionInfo.cwd}>
                📁 {sessionInfo.cwd.split('\\').pop() || sessionInfo.cwd.split('/').pop()}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Messages Area */}
      <div className="flex-1 overflow-y-auto p-4 bg-gray-50">
        {chatMessages.length === 0 && !isRunning ? (
          <div className="text-center text-neutral-gray py-12">
            <div className="text-4xl mb-3">💬</div>
            <p className="font-medium">How can I help you today?</p>
            <p className="text-sm mt-2">
              Ask me to run commands, read/write files, or complete tasks.
            </p>
          </div>
        ) : (
          <div className="space-y-4 max-w-4xl mx-auto">
            {chatMessages.map((msg) => (
              <div key={msg.id}>
                {/* Message bubble */}
                <div
                  className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-[80%] rounded-2xl px-4 py-3 shadow-sm ${
                      msg.role === 'user'
                        ? 'bg-primary-blue text-white rounded-br-md'
                        : msg.role === 'system'
                        ? 'bg-yellow-50 border border-yellow-200 text-yellow-800 rounded-bl-md'
                        : 'bg-white border border-neutral-light text-secondary-navy rounded-bl-md'
                    }`}
                  >
                    <div className="whitespace-pre-wrap break-words">{msg.content}</div>
                  </div>
                </div>

                {/* Show steps if this assistant message has them */}
                {msg.role === 'assistant' && msg.steps && msg.steps.length > 0 && (
                  <div className="mt-2 ml-4">
                    <details className="group">
                      <summary className="cursor-pointer text-xs text-neutral-gray hover:text-secondary-navy flex items-center gap-1">
                        <svg className="w-3 h-3 transition-transform group-open:rotate-90" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                        </svg>
                        <span>{msg.steps.filter(s => s.type !== 'usage').length} tool operations</span>
                      </summary>
                      <div className="mt-2 space-y-2 pl-4 border-l-2 border-neutral-light">
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
              <div className="space-y-2 p-3 bg-blue-50 rounded-lg border border-blue-100">
                <div className="text-xs font-medium text-primary-blue mb-2 flex items-center gap-2">
                  <div className="animate-spin h-3 w-3 border-2 border-primary-blue border-t-transparent rounded-full"></div>
                  Working...
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
                <div className="bg-white border border-neutral-light rounded-2xl rounded-bl-md px-4 py-3 shadow-sm">
                  <div className="flex items-center gap-2 text-neutral-gray">
                    <div className="flex gap-1">
                      <div className="w-2 h-2 bg-primary-blue rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
                      <div className="w-2 h-2 bg-primary-blue rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
                      <div className="w-2 h-2 bg-primary-blue rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>
        )}
      </div>

      {/* Input Area */}
      <div className="p-4 bg-white border-t border-neutral-light">
        <form onSubmit={handleSubmit} className="max-w-4xl mx-auto">
          <div className="space-y-2">
            {showContext && (
              <div>
                <label className="block text-sm font-medium text-neutral-gray mb-1">
                  Additional Context
                </label>
                <textarea
                  value={context}
                  onChange={(e) => setContext(e.target.value)}
                  placeholder="Any specific requirements or constraints..."
                  className="input-field resize-none text-sm"
                  rows={2}
                  disabled={isRunning}
                />
              </div>
            )}

            <div className="flex gap-2">
              <div className="flex-1 relative">
                <textarea
                  ref={inputRef}
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Send a message..."
                  className="input-field resize-none pr-12"
                  rows={1}
                  disabled={isRunning}
                  style={{ minHeight: '44px', maxHeight: '200px' }}
                />
                <button
                  type="button"
                  onClick={() => setShowContext(!showContext)}
                  className={`absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded hover:bg-gray-100 ${
                    showContext ? 'text-primary-blue' : 'text-neutral-gray'
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
                  className="btn-danger whitespace-nowrap"
                >
                  Stop
                </button>
              ) : (
                <button
                  type="submit"
                  disabled={!message.trim()}
                  className="btn-primary whitespace-nowrap"
                >
                  Send
                </button>
              )}
            </div>

            <div className="flex items-center justify-between text-xs text-neutral-gray">
              <span>Press Enter to send, Shift+Enter for new line</span>
              {chatMessages.length > 0 && (
                <button
                  type="button"
                  onClick={onNewConversation}
                  className="hover:text-primary-blue hover:underline"
                >
                  Start new chat
                </button>
              )}
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
