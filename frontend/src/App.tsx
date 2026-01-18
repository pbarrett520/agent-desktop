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

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  steps?: Step[];
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

  const [conversations, setConversations] = useState<conversation.Summary[]>([]);
  const [activeConversation, setActiveConversation] = useState<conversation.Conversation | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [currentSteps, setCurrentSteps] = useState<Step[]>([]);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  
  const currentStepsRef = useRef<Step[]>([]);

  useEffect(() => {
    const loadInitialState = async () => {
      try {
        const cfg = await GetConfig();
        setConfig(cfg);
        const configured = await IsConfigured();
        setIsConfigured(configured);
        const info = await GetSessionInfo();
        setSessionInfo(info as SessionInfo);
        
        await refreshConversations();
        
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

  const refreshConversations = async () => {
    try {
      const list = await ListConversations();
      setConversations(list || []);
    } catch (err) {
      console.error('Failed to load conversations:', err);
    }
  };

  const updateChatMessagesFromConversation = (conv: conversation.Conversation) => {
    if (!conv.messages) {
      setChatMessages([]);
      return;
    }

    const messages: ChatMessage[] = [];
    const rawMessages = conv.messages;
    
    let currentSteps: Step[] = [];
    let stepNumber = 0;
    
    for (let i = 0; i < rawMessages.length; i++) {
      const msg = rawMessages[i];
      
      if (msg.role === 'system') continue;
      
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
      
      if (msg.role === 'tool' && msg.content) {
        stepNumber++;
        const isTaskComplete = msg.content.includes('Task completed!');
        
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
        
        if (isTaskComplete) {
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

  useEffect(() => {
    currentStepsRef.current = currentSteps;
  }, [currentSteps]);

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
      const steps = currentStepsRef.current;
      setChatMessages(prev => [...prev, {
        id: `msg-${Date.now()}`,
        role: 'assistant',
        content: content,
        timestamp: new Date(),
        steps: steps.length > 0 ? [...steps] : undefined,
      }]);
      setCurrentSteps([]);
      refreshConversations();
      GetSessionInfo().then(info => setSessionInfo(info as SessionInfo));
    });

    const unsubscribeMessage = EventsOn('agent:message', (content: string) => {
      setIsRunning(false);
      const steps = currentStepsRef.current;
      setChatMessages(prev => [...prev, {
        id: `msg-${Date.now()}`,
        role: 'assistant',
        content: content,
        timestamp: new Date(),
        steps: steps.length > 0 ? [...steps] : undefined,
      }]);
      setCurrentSteps([]);
      refreshConversations();
    });

    const unsubscribeError = EventsOn('agent:error', (errorMsg: string) => {
      setIsRunning(false);
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
      
      if (activeConversation?.id === id) {
        const updated = await LoadConversation(id);
        setActiveConversation(updated);
      }
    } catch (err) {
      console.error('Failed to rename conversation:', err);
    }
  }, [activeConversation]);

  const handleSendMessage = useCallback(async (message: string, context: string) => {
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
    <div className="h-screen flex bg-matrix-black overflow-hidden relative">
      {/* CRT Effects Overlays */}
      <div className="crt-overlay" />
      <div className="noise-overlay" />
      <div className="vignette" />
      
      {/* Collapsible sidebar container */}
      <div className={`flex transition-all duration-300 ${sidebarCollapsed ? 'w-12' : ''}`}>
        {sidebarCollapsed ? (
          /* Collapsed state */
          <div className="w-12 bg-matrix-darker border-r border-matrix-border flex flex-col items-center py-3 gap-3">
            <button
              onClick={() => setSidebarCollapsed(false)}
              className="p-2 hover:bg-matrix-green/10 rounded text-matrix-green-dim hover:text-matrix-green transition-colors"
              title="Expand sidebar"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 5l7 7-7 7M5 5l7 7-7 7" />
              </svg>
            </button>
            <button
              onClick={handleNewConversation}
              className="p-2 hover:bg-matrix-green/10 rounded text-matrix-green transition-colors"
              title="New Session"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
            </button>
            {isConfigured && (
              <div className="status-online" title="Connected" />
            )}
          </div>
        ) : (
          /* Expanded state */
          <>
            <Sidebar
              config={config}
              onConfigChange={handleConfigChange}
              tokenUsage={tokenUsage}
              onTestConnection={handleTestConnection}
              onCollapse={() => setSidebarCollapsed(true)}
            />
            
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
