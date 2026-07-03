import { useState, useEffect, useCallback } from 'react';
import { GetAuditLog } from '../../wailsjs/go/main/App';
import { audit } from '../../wailsjs/go/models';
import Dropdown from './Dropdown';

interface AuditPanelProps {
  onClose: () => void;
}

const tierColor: Record<string, string> = {
  READ: 'text-brand-cyan',
  MUTATE: 'text-brand-yellow',
  DESTRUCTIVE: 'text-brand-magenta',
};

export default function AuditPanel({ onClose }: AuditPanelProps) {
  const [events, setEvents] = useState<audit.Event[]>([]);
  const [tierFilter, setTierFilter] = useState<string>('ALL');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const log = await GetAuditLog();
      setEvents((log || []).slice().reverse()); // newest first
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const filtered = tierFilter === 'ALL' ? events : events.filter(e => e.tier === tierFilter);

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-6">
      <div className="bg-brand-darker border border-brand-border rounded-lg w-full max-w-4xl max-h-[85vh] flex flex-col">
        <div className="flex items-center justify-between px-5 py-4 border-b border-brand-border">
          <h2 className="text-base font-semibold">Audit trail</h2>
          <div className="flex items-center gap-3">
            <Dropdown
              value={tierFilter}
              onChange={setTierFilter}
              className="w-32"
              options={[
                { value: 'ALL', label: 'All tiers' },
                { value: 'READ', label: 'READ' },
                { value: 'MUTATE', label: 'MUTATE' },
                { value: 'DESTRUCTIVE', label: 'DESTRUCTIVE' },
              ]}
            />
            <button onClick={load} className="text-xs text-white/50 hover:text-white">↻ Refresh</button>
            <button onClick={onClose} className="text-white/50 hover:text-white text-lg leading-none">×</button>
          </div>
        </div>

        <div className="overflow-y-auto flex-1 px-2 py-2">
          {loading && <div className="text-xs text-white/40 px-3 py-4">Loading…</div>}
          {error && <div className="text-xs text-brand-magenta px-3 py-4">{error}</div>}
          {!loading && !error && filtered.length === 0 && (
            <div className="text-xs text-white/40 px-3 py-4">No audit events recorded yet.</div>
          )}
          {!loading && filtered.length > 0 && (
            <table className="w-full text-xs">
              <thead className="text-white/40 sticky top-0 bg-brand-darker">
                <tr className="text-left">
                  <th className="px-3 py-2 font-medium">Time</th>
                  <th className="px-3 py-2 font-medium">Tier</th>
                  <th className="px-3 py-2 font-medium">Command</th>
                  <th className="px-3 py-2 font-medium">Decision</th>
                  <th className="px-3 py-2 font-medium">Exit</th>
                  <th className="px-3 py-2 font-medium">Duration</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((e, i) => (
                  <tr key={i} className="border-t border-brand-border/60 hover:bg-white/5">
                    <td className="px-3 py-2 whitespace-nowrap text-white/60">{new Date(e.timestamp).toLocaleString()}</td>
                    <td className={`px-3 py-2 font-medium ${tierColor[e.tier] || ''}`}>{e.tier}</td>
                    <td className="px-3 py-2 font-mono text-white/80 max-w-[320px] truncate" title={e.command}>{e.command}</td>
                    <td className="px-3 py-2 text-white/60">{e.decision}</td>
                    <td className="px-3 py-2 text-white/60">{e.exit_code}</td>
                    <td className="px-3 py-2 text-white/60">{e.duration_ms}ms</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
