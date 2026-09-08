import React, { useMemo } from "react";
import { useConversationStore } from "#/stores/conversation-store";
import { useSidebarStore } from "#/stores/sidebar-store";
import { SidebarItem } from "./sidebar-item";
import { Conversation } from "#/api/types";

export const ConversationList: React.FC = () => {
  const { conversations, activeConversationId } = useConversationStore();
  const { searchQuery } = useSidebarStore();

  const filteredConversations = useMemo(() => {
    if (!searchQuery.trim()) return conversations;
    const q = searchQuery.toLowerCase();
    return conversations.filter((c) => c.title.toLowerCase().includes(q));
  }, [conversations, searchQuery]);

  const { today, previous } = useMemo(() => {
    const now = Date.now();
    const oneDay = 24 * 60 * 60 * 1000;
    const startOfToday = new Date().setHours(0, 0, 0, 0);

    const todayList: Conversation[] = [];
    const prevList: Conversation[] = [];

    filteredConversations.forEach((conv) => {
      if (conv.createdAt >= startOfToday || now - conv.createdAt < oneDay) {
        todayList.push(conv);
      } else {
        prevList.push(conv);
      }
    });

    return { today: todayList, previous: prevList };
  }, [filteredConversations]);

  if (filteredConversations.length === 0) {
    return (
      <div className="px-4 py-8 text-center text-xs text-content-muted">
        暂无匹配的历史分镜会话
      </div>
    );
  }

  return (
    <div className="flex-1 overflow-y-auto px-2.5 space-y-4">
      {today.length > 0 && (
        <div className="space-y-1">
          <div className="px-3 py-1 text-[11px] font-semibold tracking-wider text-content-muted uppercase">
            今天
          </div>
          {today.map((conv) => (
            <SidebarItem
              key={conv.id}
              conversation={conv}
              isActive={conv.id === activeConversationId}
            />
          ))}
        </div>
      )}

      {previous.length > 0 && (
        <div className="space-y-1">
          <div className="px-3 py-1 text-[11px] font-semibold tracking-wider text-content-muted uppercase">
            更早之前
          </div>
          {previous.map((conv) => (
            <SidebarItem
              key={conv.id}
              conversation={conv}
              isActive={conv.id === activeConversationId}
            />
          ))}
        </div>
      )}
    </div>
  );
};
