import { useState, useEffect, useCallback } from 'react';
import { FetchModels, CheckEndpointHealth } from '../../wailsjs/go/main/App';
import Dropdown from './Dropdown';

interface ModelInfo {
  id: string;
  object: string;
  owned_by: string;
}

interface Config {
  api_key: string;
  endpoint: string;
  model: string;
  execution_timeout: number;
  mode: string;
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
  const [availableModels, setAvailableModels] = useState<ModelInfo[]>([]);
  const [isFetchingModels, setIsFetchingModels] = useState(false);
  const [modelFetchError, setModelFetchError] = useState<string | null>(null);
  const [endpointStatus, setEndpointStatus] = useState<'unknown' | 'checking' | 'online' | 'offline'>('unknown');
  const [useManualModel, setUseManualModel] = useState(false);
  const [formData, setFormData] = useState<Config>({
    api_key: '',
    endpoint: 'https://api.openai.com/v1',
    model: '',
    execution_timeout: 60,
    mode: 'CLOUD_OPS',
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

  const handlePresetChange = (preset: string) => {
    const presets: Record<string, string> = {
      'openai': 'https://api.openai.com/v1',
      'lmstudio': 'http://localhost:1234/v1',
      'openrouter': 'https://openrouter.ai/api/v1',
      'custom': formData.endpoint,
    };
    const newEndpoint = presets[preset] || formData.endpoint;
    setFormData(prev => ({
      ...prev,
      endpoint: newEndpoint,
    }));
    // Auto-fetch models for local endpoints
    if (isLocalEndpoint(newEndpoint)) {
      fetchModelsFromEndpoint(newEndpoint, formData.api_key);
    } else {
      setAvailableModels([]);
      setEndpointStatus('unknown');
      setUseManualModel(false);
    }
  };

  const isLocalEndpoint = (endpoint: string): boolean => {
    return endpoint.includes('localhost') || endpoint.includes('127.0.0.1');
  };

  const getModelPlaceholder = (endpoint: string): string => {
    if (endpoint.includes('api.openai.com')) return 'gpt-4o';
    if (isLocalEndpoint(endpoint)) return 'google/gemma-3-4b';
    if (endpoint.includes('openrouter.ai')) return 'deepseek/deepseek-chat';
    return 'model-id';
  };

  const fetchModelsFromEndpoint = useCallback(async (endpoint: string, apiKey: string) => {
    setIsFetchingModels(true);
    setModelFetchError(null);
    setAvailableModels([]);
    setEndpointStatus('checking');
    try {
      const models = await FetchModels(endpoint, apiKey);
      setAvailableModels(models || []);
      setEndpointStatus('online');
      // Auto-select first model if none currently selected
      if (models && models.length > 0 && !formData.model) {
        setFormData(prev => ({ ...prev, model: models[0].id }));
      }
    } catch (err) {
      setModelFetchError(String(err));
      setEndpointStatus('offline');
    }
    setIsFetchingModels(false);
  }, [formData.model]);

  const getPresetFromEndpoint = (endpoint: string): string => {
    if (endpoint.includes('api.openai.com')) return 'openai';
    if (isLocalEndpoint(endpoint)) return 'lmstudio';
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
      if (!formData.model && isLocalEndpoint(formData.endpoint)) {
        // No model selected yet — just check endpoint health
        const result = await CheckEndpointHealth(formData.endpoint, formData.api_key);
        if (Array.isArray(result)) {
          setTestResult({ success: result[0] as boolean, message: result[1] as string });
          setEndpointStatus((result[0] as boolean) ? 'online' : 'offline');
        }
      } else {
        const result = await onTestConnection();
        setTestResult(result);
      }
    } catch (err) {
      setTestResult({ success: false, message: String(err) });
    }
    setIsTesting(false);
  };

  // Auto-fetch models when editing opens with a local endpoint
  useEffect(() => {
    if (isEditing && isLocalEndpoint(formData.endpoint)) {
      fetchModelsFromEndpoint(formData.endpoint, formData.api_key);
    }
  }, [isEditing]);

  // Check endpoint health on mount for configured local endpoints
  useEffect(() => {
    if (config && isLocalEndpoint(config.endpoint)) {
      setEndpointStatus('checking');
      CheckEndpointHealth(config.endpoint, config.api_key).then(result => {
        if (Array.isArray(result)) {
          setEndpointStatus((result[0] as boolean) ? 'online' : 'offline');
        }
      }).catch(() => setEndpointStatus('offline'));
    }
  }, [config?.endpoint]);

  const isConfigured = config &&
    (config.api_key || isLocalEndpoint(config.endpoint)) &&
    config.endpoint &&
    config.model;

  useEffect(() => {
    if (!isConfigured) {
      setIsCollapsed(false);
    }
  }, [isConfigured]);

  const getProviderName = (endpoint: string): string => {
    if (endpoint.includes('api.openai.com')) return 'OpenAI';
    if (isLocalEndpoint(endpoint)) return 'LM_Studio';
    if (endpoint.includes('openrouter.ai')) return 'OpenRouter';
    try {
      const url = new URL(endpoint);
      return url.hostname;
    } catch {
      return 'Custom';
    }
  };

  return (
    <aside className="w-60 bg-brand-darker border-r border-brand-border h-full overflow-y-auto flex flex-col">
      {/* Header */}
      <div className="border-b border-brand-border">
        {/* IG brand strip */}
        <div className="h-[3px] w-full" style={{ background: 'linear-gradient(90deg, #00D6F2, #FDCD01, #FF0068)' }} />
        <div className="p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <h1 className="text-sm font-semibold">
                Nimbus
              </h1>
              <span className="flex items-center gap-1 ml-1" title="Insight Global">
                <span className="w-1.5 h-1.5 rounded-full bg-brand-cyan" />
                <span className="w-1.5 h-1.5 rounded-full bg-brand-yellow" />
                <span className="w-1.5 h-1.5 rounded-full bg-brand-magenta" />
              </span>
            </div>
            {onCollapse && (
              <button
                onClick={onCollapse}
                className="p-1.5 hover:bg-white/10 rounded text-white/50 hover:text-white transition-all"
                title="Collapse sidebar"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
                </svg>
              </button>
            )}
          </div>
          <div className="text-[11px] text-white/35 mt-2">
            v1.0.0
          </div>
        </div>
      </div>

      {/* Configuration Section */}
      <div className="border-b border-brand-border">
        <button
          onClick={() => !isEditing && setIsCollapsed(!isCollapsed)}
          className={`w-full p-3 flex items-center justify-between hover:bg-white/5 transition-colors ${isEditing ? 'cursor-default' : 'cursor-pointer'}`}
        >
          <div className="flex items-center gap-2">
            <span className={`text-white/40 text-xs transition-transform ${isCollapsed ? '' : 'rotate-90'}`}>
              {isCollapsed ? '▶' : '▼'}
            </span>
            <span className="text-xs font-medium text-white/70">
              Connection
            </span>
          </div>
          {isConfigured && !isEditing && (
            <span className="flex items-center gap-1.5">
              <span className="status-online"></span>
              <span className="text-[11px] text-brand-cyan">Connected</span>
            </span>
          )}
        </button>

        {!isCollapsed && (
          <div className="px-3 pb-4">
            {isConfigured && !isEditing ? (
              <div className="space-y-2 text-xs">
                <div className="flex items-center justify-between text-white/50">
                  <span>Provider</span>
                  <span className="text-white">{getProviderName(config.endpoint)}</span>
                </div>
                <div className="flex items-center justify-between text-white/50">
                  <span>Model</span>
                  <span className="text-white truncate max-w-[140px]">{config.model}</span>
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setIsEditing(true);
                  }}
                  className="text-xs text-brand-cyan hover:text-white mt-2 transition-colors"
                >
                  Edit configuration
                </button>
              </div>
            ) : (
              <div className="space-y-3">
                <div>
                  <label className="block text-xs font-medium text-white/60 mb-1">
                    Provider
                  </label>
                  <Dropdown
                    value={getPresetFromEndpoint(formData.endpoint)}
                    onChange={handlePresetChange}
                    options={[
                      { value: 'openai', label: 'OpenAI' },
                      { value: 'lmstudio', label: 'LM Studio (local)' },
                      { value: 'openrouter', label: 'OpenRouter' },
                      { value: 'custom', label: 'Custom endpoint' },
                    ]}
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-white/60 mb-1">
                    Endpoint URL
                  </label>
                  <input
                    type="text"
                    name="endpoint"
                    value={formData.endpoint}
                    onChange={handleChange}
                    placeholder="https://api.openai.com/v1"
                    className="input-field text-xs"
                  />
                  {isLocalEndpoint(formData.endpoint) && endpointStatus !== 'unknown' && (
                    <div className="flex items-center gap-1.5 mt-1.5">
                      <span className={`w-1.5 h-1.5 rounded-full ${
                        endpointStatus === 'online' ? 'bg-brand-cyan' :
                        endpointStatus === 'checking' ? 'bg-brand-yellow animate-pulse' :
                        'bg-brand-magenta'
                      }`} />
                      <span className={`text-[11px] ${
                        endpointStatus === 'online' ? 'text-brand-cyan' :
                        endpointStatus === 'checking' ? 'text-brand-yellow' :
                        'text-brand-magenta'
                      }`}>
                        {endpointStatus === 'online' ? 'LM Studio online' :
                         endpointStatus === 'checking' ? 'Checking…' :
                         'LM Studio offline'}
                      </span>
                    </div>
                  )}
                </div>

                <div>
                  <label className="block text-xs font-medium text-white/60 mb-1">
                    API key
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
                    <p className="text-[11px] text-white/35 mt-1">Optional for local models</p>
                  )}
                </div>

                <div>
                  <label className="block text-xs font-medium text-white/60 mb-1">
                    Model
                  </label>
                  {availableModels.length > 0 && !useManualModel ? (
                    <>
                      <Dropdown
                        value={formData.model}
                        onChange={(id) => setFormData(prev => ({ ...prev, model: id }))}
                        placeholder="Select a model"
                        options={availableModels.map(m => ({ value: m.id, label: m.id }))}
                      />
                      <button
                        type="button"
                        onClick={() => setUseManualModel(true)}
                        className="text-[11px] text-white/40 hover:text-brand-cyan mt-1 block transition-colors"
                      >
                        Enter manually
                      </button>
                    </>
                  ) : (
                    <>
                      <input
                        type="text"
                        name="model"
                        value={formData.model}
                        onChange={handleChange}
                        placeholder={getModelPlaceholder(formData.endpoint)}
                        className="input-field text-xs"
                      />
                      {availableModels.length > 0 && (
                        <button
                          type="button"
                          onClick={() => setUseManualModel(false)}
                          className="text-[11px] text-white/40 hover:text-brand-cyan mt-1 block transition-colors"
                        >
                          Select from list
                        </button>
                      )}
                    </>
                  )}
                  {isFetchingModels && (
                    <p className="text-[11px] text-brand-yellow mt-1 flex items-center gap-1.5">
                      <span className="w-1.5 h-1.5 bg-brand-yellow rounded-full animate-pulse" />
                      Loading models…
                    </p>
                  )}
                  {modelFetchError && (
                    <p className="text-[11px] text-brand-magenta mt-1">
                      Could not reach endpoint
                    </p>
                  )}
                  {isLocalEndpoint(formData.endpoint) && !isFetchingModels && (
                    <button
                      type="button"
                      onClick={() => fetchModelsFromEndpoint(formData.endpoint, formData.api_key)}
                      className="text-[11px] text-white/40 hover:text-brand-cyan mt-1 transition-colors"
                    >
                      Refresh models
                    </button>
                  )}
                </div>

                <div>
                  <label className="block text-xs font-medium text-white/60 mb-1">
                    Timeout (seconds)
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

                <div>
                  <label className="block text-xs font-medium text-white/60 mb-1">
                    Agent mode
                  </label>
                  <div className="flex rounded-lg border border-brand-border overflow-hidden text-xs">
                    <button
                      type="button"
                      onClick={() => setFormData(prev => ({ ...prev, mode: 'CLOUD_OPS' }))}
                      className={`flex-1 px-2 py-1.5 transition-colors ${
                        formData.mode === 'CLOUD_OPS'
                          ? 'bg-brand-cyan/15 text-brand-cyan'
                          : 'text-white/50 hover:text-white'
                      }`}
                    >
                      Cloud Ops
                    </button>
                    <button
                      type="button"
                      onClick={() => setFormData(prev => ({ ...prev, mode: 'GENERAL' }))}
                      className={`flex-1 px-2 py-1.5 transition-colors border-l border-brand-border ${
                        formData.mode === 'GENERAL'
                          ? 'bg-brand-cyan/15 text-brand-cyan'
                          : 'text-white/50 hover:text-white'
                      }`}
                    >
                      General
                    </button>
                  </div>
                  <p className="text-[10px] text-white/35 mt-1">
                    Cloud Ops exposes az_query/az_propose for Azure work. General exposes local shell/file tools.
                  </p>
                </div>

                {testResult && (
                  <div className={`p-2.5 rounded-lg text-xs message-animate ${
                    testResult.success
                      ? 'bg-brand-cyan/10 border border-brand-cyan/30 text-brand-cyan'
                      : 'bg-brand-magenta/10 border border-brand-magenta/30 text-brand-magenta'
                  }`}>
                    <span className="flex items-center gap-2">
                      {testResult.success ? (
                        <span>✓</span>
                      ) : (
                        <span>✗</span>
                      )}
                      {testResult.message}
                    </span>
                  </div>
                )}

                <div className="flex gap-2 pt-3">
                  <button
                    onClick={handleTest}
                    disabled={isTesting || !formData.endpoint || (!formData.model && !isLocalEndpoint(formData.endpoint))}
                    className="btn-secondary text-xs flex-1 py-2"
                  >
                    {isTesting ? (
                      <span className="flex items-center justify-center gap-2">
                        <span className="w-1.5 h-1.5 bg-current rounded-full animate-pulse" />
                        Testing
                      </span>
                    ) : 'Test'}
                  </button>
                  <button
                    onClick={handleSave}
                    className="btn-primary text-xs flex-1 py-2"
                  >
                    Save
                  </button>
                </div>

                {isEditing && isConfigured && (
                  <button
                    onClick={() => {
                      setIsEditing(false);
                      setFormData(config!);
                      setTestResult(null);
                    }}
                    className="text-xs text-white/40 hover:text-white w-full text-center mt-2 transition-colors"
                  >
                    Cancel
                  </button>
                )}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Token Usage Section */}
      {tokenUsage.total_tokens > 0 && (
        <div className="p-4 border-b border-brand-border">
          <h2 className="text-xs font-medium text-white/60 mb-3">
            Token usage
          </h2>
          <div className="space-y-2 text-xs">
            <div className="flex justify-between items-center">
              <span className="text-white/50">Input</span>
              <span>{tokenUsage.prompt_tokens.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-white/50">Output</span>
              <span>{tokenUsage.completion_tokens.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center pt-2 border-t border-brand-border">
              <span className="text-white/50">Total</span>
              <span className="font-semibold">{tokenUsage.total_tokens.toLocaleString()}</span>
            </div>
          </div>
        </div>
      )}

      {/* System Status */}
      <div className="p-4 mt-auto border-t border-brand-border">
        <div className="text-[11px] text-white/35 space-y-1.5">
          <div className="flex items-center justify-between">
            <span>Status</span>
            <span className="flex items-center gap-1.5">
              <span className={`w-1.5 h-1.5 rounded-full ${
                isConfigured
                  ? (endpointStatus === 'online' || !isLocalEndpoint(config?.endpoint || ''))
                    ? 'bg-brand-cyan'
                    : 'bg-brand-magenta'
                  : 'bg-brand-yellow'
              }`} />
              <span className={
                isConfigured
                  ? (endpointStatus === 'online' || !isLocalEndpoint(config?.endpoint || ''))
                    ? 'text-brand-cyan'
                    : 'text-brand-magenta'
                  : 'text-brand-yellow'
              }>
                {isConfigured
                  ? (endpointStatus === 'online' || !isLocalEndpoint(config?.endpoint || ''))
                    ? 'Operational'
                    : 'Endpoint down'
                  : 'Not configured'}
              </span>
            </span>
          </div>
        </div>
      </div>
    </aside>
  );
}
