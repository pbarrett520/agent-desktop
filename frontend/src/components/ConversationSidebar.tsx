import { useState } from 'react';
import { conversation } from '../../wailsjs/go/models';

interface ConversationSidebarProps {
  conversations: conversation.Summary[];
  activeConversationId: string | null;
  onNewConversation: () => void;
  onLoadConversation: (id: string) => void;
  onDeleteConversation: (id: string) => void;
  onRenameConversation: (id: string, title: string) => void;
}

export default function ConversationSidebar({
  conversations,
  activeConversationId,
  onNewConversation,
  onLoadConversation,
  onDeleteConversation,
  onRenameConversation,
}: ConversationSidebarProps) {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState('');
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  const handleStartRename = (conv: conversation.Summary) => {
    setEditingId(conv.id);
    setEditTitle(conv.title);
  };

  const handleSaveRename = () => {
    if (editingId && editTitle.trim()) {
      onRenameConversation(editingId, editTitle.trim());
    }
    setEditingId(null);
    setEditTitle('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveRename();
    } else if (e.key === 'Escape') {
      setEditingId(null);
      setEditTitle('');
    }
  };

  const formatDate = (dateStr: string | Date) => {
    try {
      const date = new Date(dateStr);
      const now = new Date();
      const diffMs = now.getTime() - date.getTime();
      const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

      if (diffDays === 0) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      } else if (diffDays === 1) {
        return 'Yesterday';
      } else if (diffDays < 7) {
        return date.toLocaleDateString([], { weekday: 'short' });
      } else {
        return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
      }
    } catch {
      return '';
    }
  };

  return (
    <div className="w-56 bg-matrix-dark border-r border-matrix-green/10 flex flex-col h-full panel-depth">
      {/* Header */}
      <div className="p-3 border-b border-matrix-green/10">
        <button
          onClick={onNewConversation}
          className="w-full btn-primary text-[11px] py-2.5 flex items-center justify-center gap-2 uppercase tracking-wider glitch-hover"
        >
          <span className="text-base leading-none">+</span>
          NEW_SESSION
        </button>
      </div>

      {/* Session List Header */}
      <div className="px-4 py-2.5 border-b border-matrix-border bg-matrix-darker/50">
        <div className="text-[10px] text-matrix-green-dim uppercase tracking-wider flex items-center gap-2 font-medium">
          <span className="text-matrix-green text-glow">▸</span>
          <span>Sessions</span>
          <span className="text-matrix-cyan ml-auto">[{conversations.length}]</span>
        </div>
      </div>

      {/* Conversation list */}
      <div className="flex-1 overflow-y-auto">
        {conversations.length === 0 ? (
          <div className="p-4 text-center">
            <div className="text-matrix-green-dark text-[10px] font-mono mb-2">
              NO_SESSIONS_FOUND
            </div>
            <div className="text-[10px] text-matrix-green-dim">
              Initialize new session to begin
            </div>
          </div>
        ) : (
          <div className="py-1">
            {conversations.map((conv, index) => (
              <div
                key={conv.id}
                className={`group relative px-4 py-2.5 cursor-pointer transition-all duration-200 ${
                  activeConversationId === conv.id 
                    ? 'bg-matrix-green/10 border-l-2 border-matrix-green shadow-inner-glow' 
                    : 'hover:bg-matrix-green/5 border-l-2 border-transparent hover:border-matrix-green-dark'
                }`}
                onClick={() => {
                  if (editingId !== conv.id) {
                    onLoadConversation(conv.id);
                  }
                }}
              >
                {editingId === conv.id ? (
                  <input
                    type="text"
                    value={editTitle}
                    onChange={(e) => setEditTitle(e.target.value)}
                    onBlur={handleSaveRename}
                    onKeyDown={handleKeyDown}
                    className="w-full px-2 py-1.5 text-[11px] bg-matrix-black border border-matrix-green rounded text-matrix-green focus:outline-none focus:shadow-glow-sm"
                    autoFocus
                    onClick={(e) => e.stopPropagation()}
                  />
                ) : (
                  <>
                    {/* Session number indicator */}
                    <div className="text-[9px] text-matrix-green-dark font-mono mb-1 flex items-center gap-2">
                      <span className="text-matrix-cyan-dim">◆</span>
                      SESSION_{String(conversations.length - index).padStart(3, '0')}
                    </div>
                    
                    {/* Title */}
                    <div className={`text-[11px] truncate pr-8 transition-all ${
                      activeConversationId === conv.id 
                        ? 'text-matrix-green text-glow font-medium' 
                        : 'text-matrix-green-dim group-hover:text-matrix-green'
                    }`} title={conv.title || 'New_Session'}>
                      {activeConversationId === conv.id && <span className="text-matrix-green-bright mr-1.5">▸</span>}
                      {conv.title || 'New_Session'}
                    </div>
                    
                    {/* Metadata */}
                    <div className="flex items-center justify-between text-[9px] text-matrix-green-dark mt-1.5 font-mono">
                      <span className="flex items-center gap-1">
                        <span className="opacity-60">⬡</span>
                        {conv.turn_count} {conv.turn_count === 1 ? 'msg' : 'msgs'}
                      </span>
                      <span>{formatDate(conv.updated_at)}</span>
                    </div>

                    {/* Action buttons (show on hover) */}
                    <div className="absolute right-2 top-1/2 -translate-y-1/2 hidden group-hover:flex items-center gap-1 bg-matrix-dark/90 rounded px-1">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleStartRename(conv);
                        }}
                        className="p-1.5 hover:bg-matrix-green/20 rounded text-matrix-green-dim hover:text-matrix-green transition-all"
                        title="Rename"
                      >
                        <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          if (confirmDeleteId === conv.id) {
                            onDeleteConversation(conv.id);
                            setConfirmDeleteId(null);
                          } else {
                            setConfirmDeleteId(conv.id);
                            setTimeout(() => setConfirmDeleteId(null), 3000);
                          }
                        }}
                        className={`p-1.5 rounded transition-all ${
                          confirmDeleteId === conv.id 
                            ? 'bg-matrix-red/30 text-matrix-red animate-pulse' 
                            : 'hover:bg-matrix-red/20 text-matrix-green-dim hover:text-matrix-red'
                        }`}
                        title={confirmDeleteId === conv.id ? 'Click to confirm' : 'Delete'}
                      >
                        <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="p-3 border-t border-matrix-border bg-matrix-darker/50">
        <div className="text-[9px] text-matrix-green-dark font-mono text-center flex items-center justify-center gap-2">
          <span className="w-1.5 h-1.5 bg-matrix-green rounded-full animate-pulse" />
          MEMORY_BANK_ACTIVE
        </div>
      </div>
    </div>
  );
}
