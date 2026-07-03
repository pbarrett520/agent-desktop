import { useState, useEffect, useCallback } from 'react';
import { GetAzureContext, ListSubscriptions, SetSubscription } from '../../wailsjs/go/main/App';
import { azure } from '../../wailsjs/go/models';
import Dropdown from './Dropdown';

export default function AzureContextStrip() {
  const [ctx, setCtx] = useState<azure.Context | null>(null);
  const [subscriptions, setSubscriptions] = useState<azure.Subscription[]>([]);
  const [switching, setSwitching] = useState(false);
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const c = await GetAzureContext();
      setCtx(c);
      if (c.logged_in) {
        try {
          const subs = await ListSubscriptions();
          setSubscriptions(subs || []);
        } catch {
          setSubscriptions([]);
        }
      } else {
        setSubscriptions([]);
      }
    } catch {
      setCtx({ installed: false, logged_in: false, error: 'Failed to check az CLI status.' } as azure.Context);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const handleSwitch = async (id: string) => {
    if (!id || id === ctx?.subscription_id) return;
    setSwitching(true);
    try {
      await SetSubscription(id);
      await refresh();
    } finally {
      setSwitching(false);
    }
  };

  if (loading && !ctx) {
    return (
      <div className="px-3 py-1.5 text-xs text-white/40">Checking az CLI…</div>
    );
  }

  if (!ctx || !ctx.installed || !ctx.logged_in) {
    return (
      <div className="flex items-center gap-2 px-3 py-1.5 rounded-lg border border-brand-yellow/30 bg-brand-yellow-muted text-xs">
        <span className="w-1.5 h-1.5 rounded-full bg-brand-yellow" />
        <span className="text-brand-yellow-dim">{ctx?.error || 'az CLI unavailable'}</span>
        <button onClick={refresh} className="text-white/50 hover:text-white ml-1" title="Recheck">
          ↻
        </button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-3 px-3 py-1.5 rounded-lg border border-brand-border bg-brand-panel text-xs">
      <span className="status-online" />
      <div className="flex items-center gap-1.5 text-white/70">
        <span className="text-white/40">Tenant</span>
        <span className="font-mono text-[11px] truncate max-w-[110px]" title={ctx.tenant_id}>
          {ctx.tenant_id?.slice(0, 8) || '—'}
        </span>
      </div>
      <div className="w-px h-3 bg-brand-border" />
      <div className="flex items-center gap-1.5 text-white/70">
        <span className="text-white/40">Sub</span>
        {subscriptions.length > 1 ? (
          <Dropdown
            value={ctx.subscription_id || ''}
            onChange={handleSwitch}
            disabled={switching}
            className="w-40"
            options={subscriptions.map(s => ({ value: s.id, label: s.name }))}
          />
        ) : (
          <span className="text-brand-cyan truncate max-w-[160px]">{ctx.subscription_name}</span>
        )}
      </div>
      <div className="w-px h-3 bg-brand-border" />
      <div className="flex items-center gap-1.5 text-white/70">
        <span className="text-white/40">User</span>
        <span className="truncate max-w-[140px]" title={ctx.user}>{ctx.user}</span>
      </div>
      <button onClick={refresh} className="text-white/40 hover:text-white" title="Refresh">
        ↻
      </button>
    </div>
  );
}
