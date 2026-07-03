import { useState, useEffect, useCallback } from 'react';
import { GetAzureContext, ListResourceGroups, ListVMPowerStates, GetMonthlyCost } from '../../wailsjs/go/main/App';
import { azure } from '../../wailsjs/go/models';

interface DashboardPanelProps {
  onClose: () => void;
}

const powerStateColor = (state: string) => {
  if (state.toLowerCase().includes('running')) return 'text-brand-cyan';
  if (state.toLowerCase().includes('deallocat') || state.toLowerCase().includes('stopped')) return 'text-white/40';
  return 'text-brand-yellow';
};

export default function DashboardPanel({ onClose }: DashboardPanelProps) {
  const [groups, setGroups] = useState<azure.ResourceGroupSummary[]>([]);
  const [vms, setVms] = useState<azure.VMPowerState[]>([]);
  const [cost, setCost] = useState<azure.CostSummary | null>(null);
  const [costError, setCostError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    setCostError(null);
    try {
      const ctx = await GetAzureContext();
      if (!ctx.logged_in || !ctx.subscription_id) {
        setError('Not logged in to az CLI, or no active subscription.');
        setGroups([]);
        setVms([]);
        setCost(null);
        return;
      }

      const [groupResults, vmResults] = await Promise.all([
        ListResourceGroups(),
        ListVMPowerStates(),
      ]);
      setGroups(groupResults || []);
      setVms(vmResults || []);

      try {
        const c = await GetMonthlyCost(ctx.subscription_id);
        setCost(c);
      } catch (err) {
        setCost(null);
        setCostError('Cost data unavailable (costmanagement extension not installed or no access).');
      }
    } catch (err) {
      setError(String(err));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-6">
      <div className="bg-brand-darker border border-brand-border rounded-lg w-full max-w-4xl max-h-[85vh] flex flex-col">
        <div className="flex items-center justify-between px-5 py-4 border-b border-brand-border">
          <h2 className="text-base font-semibold">Subscription overview</h2>
          <div className="flex items-center gap-3">
            <button onClick={load} className="text-xs text-white/50 hover:text-white">↻ Refresh</button>
            <button onClick={onClose} className="text-white/50 hover:text-white text-lg leading-none">×</button>
          </div>
        </div>

        <div className="overflow-y-auto flex-1 px-5 py-4 space-y-6">
          {loading && <div className="text-xs text-white/40">Loading…</div>}
          {error && <div className="text-xs text-brand-magenta">{error}</div>}

          {!loading && !error && (
            <>
              <div>
                <div className="flex items-center justify-between mb-2">
                  <h3 className="text-sm font-medium text-white/80">This month's cost</h3>
                </div>
                {cost ? (
                  <div className="text-2xl font-semibold">
                    {cost.amount_to_date.toFixed(2)} <span className="text-sm text-white/50">{cost.currency}</span>
                  </div>
                ) : (
                  <div className="text-xs text-white/40">{costError}</div>
                )}
              </div>

              <div>
                <h3 className="text-sm font-medium text-white/80 mb-2">Resource groups ({groups.length})</h3>
                {groups.length === 0 ? (
                  <div className="text-xs text-white/40">No resource groups found.</div>
                ) : (
                  <table className="w-full text-xs">
                    <thead className="text-white/40">
                      <tr className="text-left">
                        <th className="px-2 py-1 font-medium">Name</th>
                        <th className="px-2 py-1 font-medium">Location</th>
                        <th className="px-2 py-1 font-medium">Resources</th>
                      </tr>
                    </thead>
                    <tbody>
                      {groups.map(g => (
                        <tr key={g.name} className="border-t border-brand-border/60">
                          <td className="px-2 py-1.5 text-white/85">{g.name}</td>
                          <td className="px-2 py-1.5 text-white/60">{g.location}</td>
                          <td className="px-2 py-1.5 text-white/60">{g.resource_count}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>

              <div>
                <h3 className="text-sm font-medium text-white/80 mb-2">VM power states ({vms.length})</h3>
                {vms.length === 0 ? (
                  <div className="text-xs text-white/40">No VMs found.</div>
                ) : (
                  <table className="w-full text-xs">
                    <thead className="text-white/40">
                      <tr className="text-left">
                        <th className="px-2 py-1 font-medium">Name</th>
                        <th className="px-2 py-1 font-medium">Resource group</th>
                        <th className="px-2 py-1 font-medium">State</th>
                      </tr>
                    </thead>
                    <tbody>
                      {vms.map(vm => (
                        <tr key={`${vm.resource_group}/${vm.name}`} className="border-t border-brand-border/60">
                          <td className="px-2 py-1.5 text-white/85">{vm.name}</td>
                          <td className="px-2 py-1.5 text-white/60">{vm.resource_group}</td>
                          <td className={`px-2 py-1.5 font-medium ${powerStateColor(vm.power_state)}`}>{vm.power_state}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                )}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
