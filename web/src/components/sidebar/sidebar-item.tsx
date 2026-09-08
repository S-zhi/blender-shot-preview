import React, { useState } from "react";
import { MessageSquare, Trash2 } from "lucide-react";
import clsx from "clsx";
import { Conversation } from "#/api/types";
import { useConversationStore } from "#/stores/conversation-store";

interface SidebarItemProps {
  conversation: Conversation;
  isActive: boolean;
}

export const SidebarItem: React.FC<SidebarItemProps> = ({
  conversation,
  isActive,
}) => {
  const { selectConversation, deleteConversation } = useConversationStore();
  const [isHovered, setIsHovered] = useState(false);

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    deleteConversation(conversation.id);
  };

  return (
    <div
      onClick={() => selectConversation(conversation.id)}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      className={clsx(
        "group relative flex items-center justify-between px-3 py-2 text-sm rounded-lg cursor-pointer transition-all duration-150 select-none",
        isActive
          ? "bg-surface-elevated text-content font-medium border-l-2 border-l-brand-primary border border-border shadow-sm"
          : "text-content-muted hover:bg-surface-outline hover:text-content"
      )}
    >
      <div className="flex items-center gap-2.5 truncate pr-6">
        <MessageSquare
          size={16}
          className={clsx(
            "shrink-0 transition-colors",
            isActive ? "text-brand-primary" : "text-content-icon group-hover:text-content-muted"
          )}
        />
        <span className="truncate">{conversation.title || "未命名分镜"}</span>
      </div>

      {isHovered && (
        <button
          onClick={handleDelete}
          title="删除会话"
          className="absolute right-2 p-1 text-content-muted hover:text-status-fail-text hover:bg-surface-elevated rounded transition-colors"
        >
          <Trash2 size={14} />
        </button>
      )}
    </div>
  );
};
