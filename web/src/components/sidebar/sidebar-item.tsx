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
          ? "bg-surface-200 text-white font-medium border border-border-subtle/80 shadow-sm"
          : "text-gray-400 hover:bg-surface-300/80 hover:text-gray-200"
      )}
    >
      <div className="flex items-center gap-2.5 truncate pr-6">
        <MessageSquare
          size={16}
          className={clsx(
            "shrink-0 transition-colors",
            isActive ? "text-brand-primary" : "text-gray-500 group-hover:text-gray-400"
          )}
        />
        <span className="truncate">{conversation.title || "未命名分镜"}</span>
      </div>

      {isHovered && (
        <button
          onClick={handleDelete}
          title="删除会话"
          className="absolute right-2 p-1 text-gray-400 hover:text-red-400 hover:bg-surface-100 rounded transition-colors"
        >
          <Trash2 size={14} />
        </button>
      )}
    </div>
  );
};
