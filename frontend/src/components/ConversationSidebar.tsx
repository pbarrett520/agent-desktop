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
    <div className="w-48 bg-white border-r border-neutral-light flex flex-col h-full">
      {/* Header */}
      <div className="p-2 border-b border-neutral-light">
        <button
          onClick={onNewConversation}
          className="w-full btn-primary text-sm py-1.5 flex items-center justify-center gap-1.5"
        >
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          New Chat
        </button>
      </div>

      {/* Conversation list */}
      <div className="flex-1 overflow-y-auto">
        {conversations.length === 0 ? (
          <div className="p-3 text-center text-neutral-gray text-xs">
            No conversations yet
          </div>
        ) : (
          <div className="py-1">
            {conversations.map((conv) => (
              <div
                key={conv.id}
                className={`group relative px-2 py-1.5 cursor-pointer hover:bg-gray-50 ${
                  activeConversationId === conv.id ? 'bg-blue-50 border-l-2 border-primary-blue' : ''
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
                    className="w-full px-1.5 py-0.5 text-xs border rounded focus:outline-none focus:ring-1 focus:ring-primary-blue"
                    autoFocus
                    onClick={(e) => e.stopPropagation()}
                  />
                ) : (
                  <>
                    <div className="text-xs font-medium text-secondary-navy truncate pr-1" title={conv.title || 'New Conversation'}>
                      {conv.title || 'New Conversation'}
                    </div>
                    <div className="flex items-center justify-between text-[10px] text-neutral-gray mt-0.5">
                      <span>{conv.turn_count} {conv.turn_count === 1 ? 'turn' : 'turns'}</span>
                      <span>{formatDate(conv.updated_at)}</span>
                    </div>

                    {/* Action buttons (show on hover) */}
                    <div className="absolute right-1 top-1/2 -translate-y-1/2 hidden group-hover:flex items-center gap-0.5 bg-white shadow-sm rounded px-0.5">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleStartRename(conv);
                        }}
                        className="p-0.5 hover:bg-gray-100 rounded text-neutral-gray hover:text-secondary-navy"
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
                        className={`p-0.5 hover:bg-gray-100 rounded ${
                          confirmDeleteId === conv.id 
                            ? 'text-red-600 bg-red-50' 
                            : 'text-neutral-gray hover:text-red-600'
                        }`}
                        title={confirmDeleteId === conv.id ? 'Click again to confirm' : 'Delete'}
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
