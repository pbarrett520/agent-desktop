export interface Proposal {
  id: string;
  command: string;
  explanation: string;
  rollback_hint: string;
  tier: 'READ' | 'MUTATE' | 'DESTRUCTIVE' | string;
  sensitive: boolean;
}

interface ApprovalCardProps {
  proposal: Proposal;
  onApprove: () => void;
  onDeny: () => void;
  resolving: boolean;
}

const tierStyles: Record<string, { border: string; bg: string; text: string; label: string }> = {
  DESTRUCTIVE: {
    border: 'border-brand-magenta/50',
    bg: 'bg-brand-magenta-muted',
    text: 'text-brand-magenta',
    label: 'DESTRUCTIVE',
  },
  MUTATE: {
    border: 'border-brand-yellow/40',
    bg: 'bg-brand-yellow-muted',
    text: 'text-brand-yellow',
    label: 'MUTATE',
  },
};

export default function ApprovalCard({ proposal, onApprove, onDeny, resolving }: ApprovalCardProps) {
  const style = tierStyles[proposal.tier] || tierStyles.MUTATE;

  return (
    <div className={`rounded-lg border ${style.border} ${style.bg} p-4 max-w-2xl mx-auto`}>
      <div className="flex items-center gap-2 mb-3">
        <span className={`text-[10px] font-bold tracking-wider px-2 py-0.5 rounded ${style.text} border ${style.border}`}>
          {style.label}
        </span>
        <span className="text-xs text-white/50">Approval required</span>
      </div>

      <div className="font-mono text-xs bg-black/40 rounded px-3 py-2 mb-3 overflow-x-auto whitespace-pre">
        {proposal.command}
      </div>

      <div className="text-sm text-white/85 mb-2">{proposal.explanation}</div>

      <div className="text-xs text-white/50 mb-4">
        <span className="text-white/40">Rollback: </span>
        {proposal.rollback_hint}
      </div>

      <div className="flex gap-3">
        <button
          onClick={onApprove}
          disabled={resolving}
          className="btn-primary text-sm px-4 py-2 disabled:opacity-50"
        >
          Approve
        </button>
        <button
          onClick={onDeny}
          disabled={resolving}
          className={`text-sm px-4 py-2 rounded border ${style.border} ${style.text} hover:bg-white/5 disabled:opacity-50`}
        >
          Deny
        </button>
      </div>
    </div>
  );
}
