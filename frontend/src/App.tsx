import { useState, useEffect, useCallback, useRef } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { 
  GetConfig, 
  SaveConfig, 
  TestConnection, 
  IsConfigured,
  GetSessionInfo,
  NewConversation,
  LoadConversation,
  ListConversations,
  DeleteConversation,
  RenameConversation,
  GetActiveConversation,
  SendMessage,
  StopAgent
} from '../wailsjs/go/main/App';
import { conversation } from '../wailsjs/go/models';
import Sidebar from './components/Sidebar';
import ChatInterface from './components/ChatInterface';
import ConversationSidebar from './components/ConversationSidebar';
import './style.css';

interface Config {
  api_key: string;
  endpoint: string;
  model: string;
  execution_timeout: number;
}

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

// Message displayed in chat (derived from conversation + steps)
interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  steps?: Step[]; // Tool calls/results associated with this response
}

function App() {
  const [config, setConfig] = useState<Config | null>(null);
  const [isConfigured, setIsConfigured] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [sessionInfo, setSessionInfo] = useState<SessionInfo | null>(null);
  const [tokenUsage, setTokenUsage] = useState({
    prompt_tokens: 0,
    completion_tokens: 0,
    total_tokens: 0,
  });

  // Conversation state
  const [conversations, setConversations] = useState<conversation.Summary[]>([]);
  const [activeConversation, setActiveConversation] = useState<conversation.Conversation | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [currentSteps, setCurrentSteps] = useState<Step[]>([]); // Steps for current turn
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  
  // Use ref to track steps for event handlers (avoids stale closure issues)
  const currentStepsRef = useRef<Step[]>([]);

  // Load initial state
  useEffect(() => {
    const loadInitialState = async () => {
      try {
        const cfg = await GetConfig();
        setConfig(cfg);
        const configured = await IsConfigured();
        setIsConfigured(configured);
        const info = await GetSessionInfo();
        setSessionInfo(info as SessionInfo);
        
        // Load conversations list
        await refreshConversations();
        
        // Check if there's an active conversation
        const active = await GetActiveConversation();
        if (active && active.id) {
          setActiveConversation(active);
          updateChatMessagesFromConversation(active);
        }
      } catch (err) {
        console.error('Failed to load initial state:', err);
      }
    };
    loadInitialState();
  }, []);

  // Refresh conversations list
  const refreshConversations = async () => {
    try {
      const list = await ListConversations();
      setConversations(list || []);
    } catch (err) {
      console.error('Failed to load conversations:', err);
    }
  };

  // Convert conversation messages to chat messages
  const updateChatMessagesFromConversation = (conv: conversation.Conversation) => {
    if (!conv.messages) {
      setChatMessages([]);
      return;
    }

    const messages: ChatMessage[] = [];
    const rawMessages = conv.messages;
    
    // Track steps for current turn (between user messages)
    let currentSteps: Step[] = [];
    let stepNumber = 0;
    
    for (let i = 0; i < rawMessages.length; i++) {
      const msg = rawMessages[i];
      
      // Skip system messages
      if (msg.role === 'system') continue;
      
      // Add user messages - this starts a new turn, so reset steps
      if (msg.role === 'user' && msg.content && msg.content.trim()) {
        currentSteps = [];
        stepNumber = 0;
        messages.push({
          id: `msg-${messages.length}`,
          role: 'user',
          content: msg.content,
          timestamp: new Date(),
        });
        continue;
      }
      
      // For assistant messages with tool_calls, create tool_call steps
      if (msg.role === 'assistant' && msg.tool_calls && msg.tool_calls.length > 0) {
        for (const toolCall of msg.tool_calls) {
          stepNumber++;
          let args: Record<string, unknown> = {};
          try {
            args = JSON.parse(toolCall.arguments || '{}');
          } catch {
            args = { raw: toolCall.arguments };
          }
          currentSteps.push({
            step_number: stepNumber,
            type: 'tool_call',
            content: '',
            tool_name: toolCall.name,
            tool_args: args,
          });
        }
        continue;
      }
      
      // For tool messages, create tool_result steps
      if (msg.role === 'tool' && msg.content) {
        stepNumber++;
        const isTaskComplete = msg.content.includes('Task completed!');
        
        // Add as a tool result step
        currentSteps.push({
          step_number: stepNumber,
          type: isTaskComplete ? 'complete' : 'tool_result',
          content: msg.content,
          tool_name: isTaskComplete ? 'task_complete' : undefined,
          tool_result: {
            success: !msg.content.toLowerCase().includes('error'),
            output: msg.content,
          },
        });
        
        // If this is a task_complete result, create the assistant message with all steps
        if (isTaskComplete) {
          // Clean up the content - remove the emoji prefix if present
          let content = msg.content;
          const taskCompleteIndex = content.indexOf('Task completed!');
          if (taskCompleteIndex > 0) {
            content = content.substring(taskCompleteIndex);
          }
          content = content.replace(/^[✅\s]+/, '').trim();
          
          messages.push({
            id: `msg-${messages.length}`,
            role: 'assistant',
            content: content,
            timestamp: new Date(),
            steps: currentSteps.length > 0 ? [...currentSteps] : undefined,
          });
        }
        continue;
      }
      
      // For assistant messages with actual content (not just tool calls)
      if (msg.role === 'assistant' && msg.content && msg.content.trim()) {
        messages.push({
          id: `msg-${messages.length}`,
          role: 'assistant',
          content: msg.content,
          timestamp: new Date(),
          steps: currentSteps.length > 0 ? [...currentSteps] : undefined,
        });
      }
    }
    setChatMessages(messages);
  };

  // Keep ref in sync with state
  useEffect(() => {
    currentStepsRef.current = currentSteps;
  }, [currentSteps]);

  // Subscribe to agent events (only once on mount)
  useEffect(() => {
    const unsubscribeStep = EventsOn('agent:step', (step: Step) => {
      if (step.type === 'usage' && step.usage) {
        setTokenUsage(prev => ({
          prompt_tokens: prev.prompt_tokens + step.usage!.prompt_tokens,
          completion_tokens: prev.completion_tokens + step.usage!.completion_tokens,
          total_tokens: prev.total_tokens + step.usage!.total_tokens,
        }));
      } else {
        setCurrentSteps(prev => [...prev, step]);
      }
    });

    const unsubscribeComplete = EventsOn('agent:complete', (content: string) => {
      setIsRunning(false);
      // Add assistant message to chat with current steps from ref
      const steps = currentStepsRef.current;
      setChatMessages(prev => [...prev, {
        id: `msg-${Date.now()}`,
        role: 'assistant',
        content: content,
        timestamp: new Date(),
        steps: steps.length > 0 ? [...steps] : undefined,
      }]);
      setCurrentSteps([]);
      // Refresh conversations (title may have been generated)
      refreshConversations();
      // Refresh session info
      GetSessionInfo().then(info => setSessionInfo(info as SessionInfo));
    });

    const unsubscribeMessage = EventsOn('agent:message', (content: string) => {
      setIsRunning(false);
      // Conversational response (not task completion)
      const steps = currentStepsRef.current;
      setChatMessages(prev => [...prev, {
        id: `msg-${Date.now()}`,
        role: 'assistant',
        content: content,
        timestamp: new Date(),
        steps: steps.length > 0 ? [...steps] : undefined,
      }]);
      setCurrentSteps([]);
      // Refresh conversations (title may have been generated)
      refreshConversations();
    });

    const unsubscribeError = EventsOn('agent:error', (errorMsg: string) => {
      setIsRunning(false);
      // Show error as a system message
      setChatMessages(prev => [...prev, {
        id: `msg-${Date.now()}`,
        role: 'system',
        content: `Error: ${errorMsg}`,
        timestamp: new Date(),
      }]);
      setCurrentSteps([]);
    });

    return () => {
      unsubscribeStep();
      unsubscribeComplete();
      unsubscribeMessage();
      unsubscribeError();
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleConfigChange = useCallback(async (newConfig: Config) => {
    try {
      await SaveConfig(newConfig);
      setConfig(newConfig);
      const configured = await IsConfigured();
      setIsConfigured(configured);
    } catch (err) {
      console.error('Failed to save config:', err);
    }
  }, []);

  const handleTestConnection = useCallback(async () => {
    try {
      const result = await TestConnection();
      if (Array.isArray(result)) {
        return { success: result[0] as boolean, message: result[1] as string };
      }
      return { success: Boolean(result), message: result ? 'Connected!' : 'Failed' };
    } catch (err) {
      return { success: false, message: String(err) };
    }
  }, []);

  // Conversation management
  const handleNewConversation = useCallback(async () => {
    try {
      const conv = await NewConversation();
      setActiveConversation(conv);
      setChatMessages([]);
      setCurrentSteps([]);
      setTokenUsage({ prompt_tokens: 0, completion_tokens: 0, total_tokens: 0 });
      await refreshConversations();
    } catch (err) {
      console.error('Failed to create conversation:', err);
    }
  }, []);

  const handleLoadConversation = useCallback(async (id: string) => {
    try {
      const conv = await LoadConversation(id);
      setActiveConversation(conv);
      updateChatMessagesFromConversation(conv);
      setCurrentSteps([]);
    } catch (err) {
      console.error('Failed to load conversation:', err);
    }
  }, []);

  const handleDeleteConversation = useCallback(async (id: string) => {
    try {
      await DeleteConversation(id);
      await refreshConversations();
      
      // If we deleted the active conversation, clear it
      if (activeConversation?.id === id) {
        setActiveConversation(null);
        setChatMessages([]);
      }
    } catch (err) {
      console.error('Failed to delete conversation:', err);
    }
  }, [activeConversation]);

  const handleRenameConversation = useCallback(async (id: string, title: string) => {
    try {
      await RenameConversation(id, title);
      await refreshConversations();
      
      // Update active conversation if it's the one we renamed
      if (activeConversation?.id === id) {
        // Reload to get the updated conversation object
        const updated = await LoadConversation(id);
        setActiveConversation(updated);
      }
    } catch (err) {
      console.error('Failed to rename conversation:', err);
    }
  }, [activeConversation]);

  // Send message in chat
  const handleSendMessage = useCallback(async (message: string, context: string) => {
    // Add user message to chat immediately for responsiveness
    setChatMessages(prev => [...prev, {
      id: `msg-${Date.now()}`,
      role: 'user',
      content: message,
      timestamp: new Date(),
    }]);

    setCurrentSteps([]);
    setIsRunning(true);

    try {
      await SendMessage(message, context);
    } catch (err) {
      console.error('Failed to send message:', err);
      setIsRunning(false);
    }
  }, []);

  const handleStopAgent = useCallback(async () => {
    try {
      await StopAgent();
      setIsRunning(false);
      setChatMessages(prev => [...prev, {
        id: `msg-${Date.now()}`,
        role: 'system',
        content: 'Cancelled by user',
        timestamp: new Date(),
      }]);
      setCurrentSteps([]);
    } catch (err) {
      console.error('Failed to stop agent:', err);
    }
  }, []);

  return (
    <div className="h-screen flex bg-gray-50 overflow-hidden">
      {/* Collapsible sidebar container */}
      <div className={`flex transition-all duration-300 ${sidebarCollapsed ? 'w-12' : ''}`}>
        {sidebarCollapsed ? (
          /* Collapsed state - just show expand button */
          <div className="w-12 bg-white border-r border-neutral-light flex flex-col items-center py-3 gap-3">
            <button
              onClick={() => setSidebarCollapsed(false)}
              className="p-2 hover:bg-gray-100 rounded-lg text-neutral-gray hover:text-secondary-navy"
              title="Expand sidebar"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
              </svg>
            </button>
            <button
              onClick={handleNewConversation}
              className="p-2 hover:bg-gray-100 rounded-lg text-primary-blue"
              title="New Chat"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </button>
            {isConfigured && (
              <div className="w-2 h-2 rounded-full bg-secondary-lime" title="Connected" />
            )}
          </div>
        ) : (
          /* Expanded state - show both sidebars */
          <>
            {/* Config sidebar */}
            <Sidebar
              config={config}
              onConfigChange={handleConfigChange}
              tokenUsage={tokenUsage}
              onTestConnection={handleTestConnection}
              onCollapse={() => setSidebarCollapsed(true)}
            />
            
            {/* Conversation list */}
            <ConversationSidebar
              conversations={conversations}
              activeConversationId={activeConversation?.id || null}
              onNewConversation={handleNewConversation}
              onLoadConversation={handleLoadConversation}
              onDeleteConversation={handleDeleteConversation}
              onRenameConversation={handleRenameConversation}
            />
          </>
        )}
      </div>
      
      {/* Main chat interface */}
      <ChatInterface
        isConfigured={isConfigured}
        chatMessages={chatMessages}
        currentSteps={currentSteps}
        isRunning={isRunning}
        sessionInfo={sessionInfo}
        activeConversation={activeConversation}
        onSendMessage={handleSendMessage}
        onStopAgent={handleStopAgent}
        onNewConversation={handleNewConversation}
      />
    </div>
  );
}

export default App;
