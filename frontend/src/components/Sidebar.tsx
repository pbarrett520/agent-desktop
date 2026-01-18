import { useState, useEffect } from 'react';

interface Config {
  api_key: string;
  endpoint: string;
  model: string;
  execution_timeout: number;
}

interface TokenUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

interface SidebarProps {
  config: Config | null;
  onConfigChange: (config: Config) => void;
  tokenUsage: TokenUsage;
  onTestConnection: () => Promise<{ success: boolean; message: string }>;
  onCollapse?: () => void;
}

export default function Sidebar({ config, onConfigChange, tokenUsage, onTestConnection, onCollapse }: SidebarProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [isCollapsed, setIsCollapsed] = useState(true);
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [isTesting, setIsTesting] = useState(false);
  const [formData, setFormData] = useState<Config>({
    api_key: '',
    endpoint: 'https://api.openai.com/v1',
    model: '',
    execution_timeout: 60,
  });

  useEffect(() => {
    if (config) {
      setFormData(config);
    }
  }, [config]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value,
    }));
  };

  const handlePresetChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const preset = e.target.value;
    const presets: Record<string, string> = {
      'openai': 'https://api.openai.com/v1',
      'lmstudio': 'http://localhost:1234/v1',
      'openrouter': 'https://openrouter.ai/api/v1',
      'custom': formData.endpoint,
    };
    setFormData(prev => ({
      ...prev,
      endpoint: presets[preset] || prev.endpoint,
    }));
  };

  const getPresetFromEndpoint = (endpoint: string): string => {
    if (endpoint.includes('api.openai.com')) return 'openai';
    if (endpoint.includes('localhost:1234')) return 'lmstudio';
    if (endpoint.includes('openrouter.ai')) return 'openrouter';
    return 'custom';
  };

  const handleSave = () => {
    onConfigChange(formData);
    setIsEditing(false);
    setTestResult(null);
  };

  const handleTest = async () => {
    setIsTesting(true);
    setTestResult(null);
    try {
      const result = await onTestConnection();
      setTestResult(result);
    } catch (err) {
      setTestResult({ success: false, message: String(err) });
    }
    setIsTesting(false);
  };

  const isConfigured = config && 
    config.api_key && 
    config.endpoint && 
    config.model;

  useEffect(() => {
    if (!isConfigured) {
      setIsCollapsed(false);
    }
  }, [isConfigured]);

  const getProviderName = (endpoint: string): string => {
    if (endpoint.includes('api.openai.com')) return 'OpenAI';
    if (endpoint.includes('localhost:1234')) return 'LM_Studio';
    if (endpoint.includes('openrouter.ai')) return 'OpenRouter';
    try {
      const url = new URL(endpoint);
      return url.hostname;
    } catch {
      return 'Custom';
    }
  };

  return (
    <aside className="w-60 bg-matrix-darker border-r border-matrix-green/10 h-full overflow-y-auto flex flex-col panel-depth relative">
      {/* Subtle grid pattern overlay */}
      <div className="absolute inset-0 opacity-5 pointer-events-none" style={{
        backgroundImage: 'linear-gradient(rgba(0, 255, 65, 0.1) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 255, 65, 0.1) 1px, transparent 1px)',
        backgroundSize: '20px 20px'
      }} />
      
      {/* Header with ASCII-style branding */}
      <div className="p-4 border-b border-matrix-green/20 relative z-10">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-matrix-green rounded-full animate-pulse shadow-glow" />
            <h1 className="text-sm font-bold text-matrix-green text-glow-intense tracking-widest brand-text">
              AGENT
            </h1>
          </div>
          {onCollapse && (
            <button
              onClick={onCollapse}
              className="p-1.5 hover:bg-matrix-green/10 rounded text-matrix-green-dim hover:text-matrix-green transition-all hover:shadow-glow-sm"
              title="Collapse sidebar"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
              </svg>
            </button>
          )}
        </div>
        <div className="text-[10px] text-matrix-green-dark mt-2 font-mono flex items-center gap-2">
          <span className="text-matrix-green-dim">v1.0.0</span>
          <span className="text-matrix-border">│</span>
          <span className="text-matrix-cyan-dim text-flicker">NEURAL_INTERFACE</span>
        </div>
      </div>

      {/* Configuration Section */}
      <div className="border-b border-matrix-border">
        <button
          onClick={() => !isEditing && setIsCollapsed(!isCollapsed)}
          className={`w-full p-3 flex items-center justify-between hover:bg-matrix-green/5 transition-colors ${isEditing ? 'cursor-default' : 'cursor-pointer'}`}
        >
          <div className="flex items-center gap-2">
            <span className={`text-matrix-green-dim text-xs transition-transform ${isCollapsed ? '' : 'rotate-90'}`}>
              {isCollapsed ? '▶' : '▼'}
            </span>
            <span className="text-xs font-medium text-matrix-green-dim uppercase tracking-wide">
              LLM_CONFIG
            </span>
          </div>
          {isConfigured && !isEditing && (
            <span className="flex items-center gap-1.5">
              <span className="status-online"></span>
              <span className="text-[10px] text-matrix-green uppercase">LINKED</span>
            </span>
          )}
        </button>

        {!isCollapsed && (
          <div className="px-3 pb-4">
            {isConfigured && !isEditing ? (
              <div className="space-y-2 text-xs">
                <div className="flex items-center gap-2 text-matrix-green-dim">
                  <span className="text-matrix-green">▸</span>
                  <span className="uppercase">Provider:</span>
                  <span className="text-matrix-green">{getProviderName(config.endpoint)}</span>
                </div>
                <div className="flex items-center gap-2 text-matrix-green-dim">
                  <span className="text-matrix-green">▸</span>
                  <span className="uppercase">Model:</span>
                  <span className="text-matrix-green truncate">{config.model}</span>
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setIsEditing(true);
                  }}
                  className="text-[10px] text-matrix-green-dim hover:text-matrix-green mt-2 uppercase tracking-wide"
                >
                  [MODIFY_CONFIG]
                </button>
              </div>
            ) : (
              <div className="space-y-3">
                <div>
                  <label className="block text-[10px] font-medium text-matrix-green-dim mb-1 uppercase tracking-wide">
                    Provider_Preset
                  </label>
                  <select
                    value={getPresetFromEndpoint(formData.endpoint)}
                    onChange={handlePresetChange}
                    className="input-field text-xs"
                  >
                    <option value="openai">OpenAI</option>
                    <option value="lmstudio">LM_Studio [LOCAL]</option>
                    <option value="openrouter">OpenRouter</option>
                    <option value="custom">Custom_Endpoint</option>
                  </select>
                </div>

                <div>
                  <label className="block text-[10px] font-medium text-matrix-green-dim mb-1 uppercase tracking-wide">
                    Endpoint_URL
                  </label>
                  <input
                    type="text"
                    name="endpoint"
                    value={formData.endpoint}
                    onChange={handleChange}
                    placeholder="https://api.openai.com/v1"
                    className="input-field text-xs"
                  />
                </div>

                <div>
                  <label className="block text-[10px] font-medium text-matrix-green-dim mb-1 uppercase tracking-wide">
                    API_Key
                  </label>
                  <input
                    type="password"
                    name="api_key"
                    value={formData.api_key}
                    onChange={handleChange}
                    placeholder="••••••••••••••••"
                    className="input-field text-xs"
                  />
                  {getPresetFromEndpoint(formData.endpoint) === 'lmstudio' && (
                    <p className="text-[10px] text-matrix-green-dark mt-1">[OPTIONAL FOR LOCAL]</p>
                  )}
                </div>

                <div>
                  <label className="block text-[10px] font-medium text-matrix-green-dim mb-1 uppercase tracking-wide">
                    Model_ID
                  </label>
                  <input
                    type="text"
                    name="model"
                    value={formData.model}
                    onChange={handleChange}
                    placeholder="gpt-4o / deepseek-chat"
                    className="input-field text-xs"
                  />
                </div>

                <div>
                  <label className="block text-[10px] font-medium text-matrix-green-dim mb-1 uppercase tracking-wide">
                    Timeout_Sec
                  </label>
                  <input
                    type="number"
                    name="execution_timeout"
                    value={formData.execution_timeout}
                    onChange={handleChange}
                    min={10}
                    max={300}
                    className="input-field text-xs"
                  />
                </div>

                {testResult && (
                  <div className={`p-2.5 rounded text-[10px] font-mono message-animate ${
                    testResult.success 
                      ? 'bg-matrix-green/10 border border-matrix-green/30 text-matrix-green shadow-glow-sm' 
                      : 'bg-matrix-red/10 border border-matrix-red/30 text-matrix-red'
                  }`}>
                    <span className="flex items-center gap-2">
                      {testResult.success ? (
                        <span className="text-matrix-green text-glow">✓</span>
                      ) : (
                        <span className="text-matrix-red">✗</span>
                      )}
                      {testResult.message}
                    </span>
                  </div>
                )}

                <div className="flex gap-2 pt-3">
                  <button
                    onClick={handleTest}
                    disabled={isTesting || !formData.endpoint || !formData.model}
                    className="btn-secondary text-[10px] flex-1 py-2 uppercase tracking-wider glitch-hover"
                  >
                    {isTesting ? (
                      <span className="flex items-center justify-center gap-2">
                        <span className="w-1.5 h-1.5 bg-current rounded-full animate-pulse" />
                        TESTING
                      </span>
                    ) : 'TEST'}
                  </button>
                  <button
                    onClick={handleSave}
                    className="btn-primary text-[10px] flex-1 py-2 uppercase tracking-wider glitch-hover"
                  >
                    SAVE
                  </button>
                </div>

                {isEditing && isConfigured && (
                  <button
                    onClick={() => {
                      setIsEditing(false);
                      setFormData(config!);
                      setTestResult(null);
                    }}
                    className="text-[10px] text-matrix-green-dim hover:text-matrix-green w-full text-center uppercase tracking-wider mt-2 transition-colors"
                  >
                    [CANCEL]
                  </button>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Token Usage Section */}
      {tokenUsage.total_tokens > 0 && (
        <div className="p-4 border-b border-matrix-border relative z-10">
          <h2 className="text-[10px] font-medium text-matrix-green-dim mb-3 uppercase tracking-wider flex items-center gap-2">
            <span className="text-matrix-amber">◇</span>
            Token_Usage
          </h2>
          <div className="space-y-2 text-[10px] font-mono">
            <div className="flex justify-between items-center">
              <span className="text-matrix-green-dim flex items-center gap-1.5">
                <span className="text-matrix-green">↑</span> INPUT
              </span>
              <span className="text-matrix-green">{tokenUsage.prompt_tokens.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-matrix-green-dim flex items-center gap-1.5">
                <span className="text-matrix-cyan">↓</span> OUTPUT
              </span>
              <span className="text-matrix-cyan">{tokenUsage.completion_tokens.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center pt-2 border-t border-matrix-border">
              <span className="text-matrix-green-dim">TOTAL</span>
              <span className="text-matrix-green text-glow font-bold text-xs">{tokenUsage.total_tokens.toLocaleString()}</span>
            </div>
          </div>
        </div>
      )}

      {/* System Status */}
      <div className="p-4 mt-auto border-t border-matrix-border bg-matrix-darker/50 relative z-10">
        <div className="text-[9px] text-matrix-green-dark font-mono space-y-1">
          <div className="flex items-center justify-between">
            <span>SYS_STATUS</span>
            <span className="flex items-center gap-1.5">
              <span className="w-1.5 h-1.5 bg-matrix-green rounded-full animate-pulse" />
              <span className="text-matrix-green">OPERATIONAL</span>
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span>NEURAL_NET</span>
            <span className="text-matrix-cyan text-flicker">ACTIVE</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
