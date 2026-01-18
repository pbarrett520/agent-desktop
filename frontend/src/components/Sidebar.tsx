import { useState, useEffect } from 'react';

interface Config {
  openai_subscription_key: string;
  openai_endpoint: string;
  openai_deployment: string;
  openai_model_name: string;
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
  const [isCollapsed, setIsCollapsed] = useState(true); // Start collapsed
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null);
  const [isTesting, setIsTesting] = useState(false);
  const [formData, setFormData] = useState<Config>({
    openai_subscription_key: '',
    openai_endpoint: '',
    openai_deployment: '',
    openai_model_name: '',
    execution_timeout: 60,
  });

  useEffect(() => {
    if (config) {
      setFormData(config);
    }
  }, [config]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? parseInt(value) || 0 : value,
    }));
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
    config.openai_subscription_key && 
    config.openai_endpoint && 
    config.openai_deployment && 
    config.openai_model_name;

  // Auto-expand if not configured
  useEffect(() => {
    if (!isConfigured) {
      setIsCollapsed(false);
    }
  }, [isConfigured]);

  return (
    <aside className="w-52 bg-white border-r border-neutral-light h-full overflow-y-auto flex flex-col">
      {/* Header */}
      <div className="p-3 border-b border-neutral-light">
        <div className="flex items-center justify-between">
          <h1 className="text-sm font-bold text-secondary-navy">Agent Desktop</h1>
          {onCollapse && (
            <button
              onClick={onCollapse}
              className="p-1 hover:bg-gray-100 rounded text-neutral-gray hover:text-secondary-navy"
              title="Collapse sidebar"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
              </svg>
            </button>
          )}
        </div>
      </div>

      {/* Configuration Section - Collapsible */}
      <div className="border-b border-neutral-light">
        <button
          onClick={() => !isEditing && setIsCollapsed(!isCollapsed)}
          className={`w-full p-3 flex items-center justify-between hover:bg-gray-50 transition-colors ${isEditing ? 'cursor-default' : 'cursor-pointer'}`}
        >
          <div className="flex items-center gap-2">
            <svg 
              className={`w-4 h-4 text-neutral-gray transition-transform ${isCollapsed ? '' : 'rotate-90'}`} 
              fill="none" 
              stroke="currentColor" 
              viewBox="0 0 24 24"
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
            <span className="text-sm font-semibold text-neutral-gray">LLM Provider</span>
          </div>
          {isConfigured && !isEditing && (
            <span className="flex items-center gap-1 text-xs text-secondary-lime">
              <span className="w-1.5 h-1.5 rounded-full bg-secondary-lime"></span>
              Connected
            </span>
          )}
        </button>

        {/* Collapsible content */}
        {!isCollapsed && (
          <div className="px-4 pb-4">
            {isConfigured && !isEditing ? (
              <div className="space-y-2 text-sm">
                <div className="text-neutral-gray truncate">
                  <span className="font-medium">Endpoint:</span>{' '}
                  <span className="text-xs">{config.openai_endpoint.replace('https://', '').split('.')[0]}</span>
                </div>
                <div className="text-neutral-gray">
                  <span className="font-medium">Model:</span> {config.openai_model_name}
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setIsEditing(true);
                  }}
                  className="text-xs text-primary-blue hover:underline mt-2"
                >
                  Edit configuration
                </button>
              </div>
            ) : (
              <div className="space-y-3">
                <div>
                  <label className="block text-xs font-medium text-neutral-gray mb-1">
                    Endpoint
                  </label>
                  <input
                    type="text"
                    name="openai_endpoint"
                    value={formData.openai_endpoint}
                    onChange={handleChange}
                    placeholder="https://your-resource.openai.azure.com"
                    className="input-field text-sm"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-neutral-gray mb-1">
                    API Key
                  </label>
                  <input
                    type="password"
                    name="openai_subscription_key"
                    value={formData.openai_subscription_key}
                    onChange={handleChange}
                    placeholder="Enter your API key"
                    className="input-field text-sm"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-neutral-gray mb-1">
                    Deployment Name
                  </label>
                  <input
                    type="text"
                    name="openai_deployment"
                    value={formData.openai_deployment}
                    onChange={handleChange}
                    placeholder="e.g., gpt-4o"
                    className="input-field text-sm"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-neutral-gray mb-1">
                    Model Name
                  </label>
                  <input
                    type="text"
                    name="openai_model_name"
                    value={formData.openai_model_name}
                    onChange={handleChange}
                    placeholder="e.g., gpt-4o"
                    className="input-field text-sm"
                  />
                </div>

                <div>
                  <label className="block text-xs font-medium text-neutral-gray mb-1">
                    Timeout (seconds)
                  </label>
                  <input
                    type="number"
                    name="execution_timeout"
                    value={formData.execution_timeout}
                    onChange={handleChange}
                    min={10}
                    max={300}
                    className="input-field text-sm"
                  />
                </div>

                {testResult && (
                  <div className={`p-2 rounded-md text-xs ${
                    testResult.success 
                      ? 'bg-secondary-lime/20 text-secondary-navy' 
                      : 'bg-secondary-coral/20 text-secondary-coral'
                  }`}>
                    {testResult.message}
                  </div>
                )}

                <div className="flex gap-2 pt-2">
                  <button
                    onClick={handleTest}
                    disabled={isTesting || !formData.openai_endpoint || !formData.openai_subscription_key}
                    className="btn-secondary text-xs flex-1 py-1.5"
                  >
                    {isTesting ? 'Testing...' : 'Test'}
                  </button>
                  <button
                    onClick={handleSave}
                    className="btn-primary text-xs flex-1 py-1.5"
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
                    className="text-xs text-neutral-gray hover:underline w-full text-center"
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
        <div className="p-4 border-b border-neutral-light">
          <h2 className="text-xs font-semibold text-neutral-gray mb-2">Token Usage</h2>
          <div className="space-y-1 text-xs font-mono">
            <div className="flex justify-between">
              <span className="text-neutral-gray">Prompt:</span>
              <span className="text-primary-blue">{tokenUsage.prompt_tokens.toLocaleString()}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-neutral-gray">Completion:</span>
              <span className="text-secondary-lime">{tokenUsage.completion_tokens.toLocaleString()}</span>
            </div>
            <div className="flex justify-between pt-1 border-t border-neutral-light">
              <span className="font-bold text-neutral-gray">Total:</span>
              <span className="font-bold text-secondary-navy">{tokenUsage.total_tokens.toLocaleString()}</span>
            </div>
          </div>
        </div>
      )}

      {/* Spacer */}
      <div className="flex-1"></div>
    </aside>
  );
}
