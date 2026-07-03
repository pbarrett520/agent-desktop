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
    <div className="w-56 bg-brand-dark border-r border-brand-border flex flex-col h-full">
      {/* Header */}
      <div className="p-3 border-b border-brand-border">
        <button
          onClick={onNewConversation}
          className="w-full btn-primary text-xs py-2.5 flex items-center justify-center gap-2"
        >
          <span className="text-base leading-none">+</span>
          New chat
        </button>
      </div>

      {/* Session List Header */}
      <div className="px-4 py-2.5 border-b border-brand-border">
        <div className="text-xs text-white/45 flex items-center gap-2 font-medium">
          <span>Chats</span>
          <span className="text-brand-yellow ml-auto">{conversations.length}</span>
        </div>
      </div>

      {/* Conversation list */}
      <div className="flex-1 overflow-y-auto">
        {conversations.length === 0 ? (
          <div className="p-4 text-center">
            <div className="text-sm text-white/50 mb-1">
              No chats yet
            </div>
            <div className="text-xs text-white/35">
              Start a new chat to begin
            </div>
          </div>
        ) : (
          <div className="py-1">
            {conversations.map((conv) => (
              <div
                key={conv.id}
                className={`group relative px-4 py-2.5 cursor-pointer transition-all duration-200 ${
                  activeConversationId === conv.id
                    ? 'bg-brand-cyan/10 border-l-2 border-brand-cyan'
                    : 'hover:bg-white/5 border-l-2 border-transparent'
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
                    className="w-full px-2 py-1.5 text-xs bg-brand-black border border-brand-cyan rounded text-white focus:outline-none"
                    autoFocus
                    onClick={(e) => e.stopPropagation()}
                  />
                ) : (
                  <>
                    {/* Title */}
                    <div className={`text-xs truncate pr-8 transition-all ${
                      activeConversationId === conv.id
                        ? 'text-white font-medium'
                        : 'text-white/60 group-hover:text-white'
                    }`} title={conv.title || 'New chat'}>
                      {conv.title || 'New chat'}
                    </div>

                    {/* Metadata */}
                    <div className="flex items-center justify-between text-[11px] text-white/35 mt-1.5">
                      <span>{conv.turn_count} {conv.turn_count === 1 ? 'msg' : 'msgs'}</span>
                      <span>{formatDate(conv.updated_at)}</span>
                    </div>

                    {/* Action buttons (show on hover) */}
                    <div className="absolute right-2 top-1/2 -translate-y-1/2 hidden group-hover:flex items-center gap-1 bg-brand-dark rounded px-1">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleStartRename(conv);
                        }}
                        className="p-1.5 hover:bg-white/10 rounded text-white/45 hover:text-white transition-all"
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
                            ? 'bg-brand-magenta/30 text-brand-magenta animate-pulse'
                            : 'hover:bg-brand-magenta/20 text-white/45 hover:text-brand-magenta'
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
    </div>
  );
}
