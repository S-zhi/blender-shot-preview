import React from "react";
import { Plus, Settings, Search, PanelLeftClose, PanelLeftOpen, Film } from "lucide-react";
import clsx from "clsx";
import { useSidebarStore } from "#/stores/sidebar-store";
import { useConversationStore } from "#/stores/conversation-store";
import { useSettingsStore } from "#/stores/settings-store";
import { ConversationList } from "./conversation-list";
import { Button } from "#/components/ui/button";

export const Sidebar: React.FC = () => {
  const { isOpen, toggleSidebar, searchQuery, setSearchQuery } = useSidebarStore();
  const { createNewConversation } = useConversationStore();
  const { setModalOpen } = useSettingsStore();

  if (!isOpen) {
    return (
      <div className="hidden md:flex flex-col items-center justify-between py-3 px-2 border-r border-border-subtle bg-surface-400 w-14 shrink-0 select-none">
        <div className="flex flex-col items-center gap-3">
          <button
            onClick={toggleSidebar}
            title="展开侧边栏"
            className="p-2 text-gray-400 hover:text-white hover:bg-surface-200 rounded-lg transition-colors"
          >
            <PanelLeftOpen size={18} />
          </button>
          <button
            onClick={() => createNewConversation()}
            title="新建会话"
            className="p-2 text-brand-primary hover:bg-surface-200 rounded-lg transition-colors"
          >
            <Plus size={18} />
          </button>
        </div>

        <button
          onClick={() => setModalOpen(true)}
          title="模型与凭证设置"
          className="p-2 text-gray-400 hover:text-white hover:bg-surface-200 rounded-lg transition-colors"
        >
          <Settings size={18} />
        </button>
      </div>
    );
  }

  return (
    <aside
      className={clsx(
        "flex flex-col h-full bg-surface-400 border-r border-border-subtle select-none transition-all duration-200 shrink-0",
        "w-64 lg:w-72"
      )}
    >
      {/* Header & Logo */}
      <div className="flex items-center justify-between p-3.5 border-b border-border-subtle">
        <div className="flex items-center gap-2.5">
          <div className="h-7 w-7 rounded-lg bg-brand-primary/20 text-brand-primary flex items-center justify-center font-bold">
            <Film size={16} />
          </div>
          <div>
            <div className="text-sm font-semibold text-gray-100 flex items-center gap-1.5">
              Shot Preview
              <span className="text-[10px] uppercase tracking-wider px-1.5 py-0.2 rounded bg-surface-200 text-gray-400 border border-border-subtle">
                Canvas
              </span>
            </div>
          </div>
        </div>

        <button
          onClick={toggleSidebar}
          title="收起侧边栏"
          className="p-1.5 text-gray-400 hover:text-white hover:bg-surface-200 rounded-md transition-colors"
        >
          <PanelLeftClose size={16} />
        </button>
      </div>

      {/* Action Buttons: New Chat */}
      <div className="p-3 space-y-2">
        <Button
          onClick={() => createNewConversation()}
          variant="primary"
          className="w-full justify-start text-xs font-semibold py-2 px-3 gap-2"
        >
          <Plus size={15} />
          新建分镜会话
        </Button>

        {/* Search Box */}
        <div className="relative">
          <Search
            size={14}
            className="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-500"
          />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="搜索分镜历史..."
            className="w-full h-8 pl-8 pr-3 text-xs rounded-lg bg-surface-300 border border-border-subtle text-gray-200 placeholder-gray-500 focus:outline-none focus:border-brand-primary/60 transition-colors"
          />
        </div>
      </div>

      {/* History Conversation List */}
      <ConversationList />

      {/* Footer Settings */}
      <div className="p-3 border-t border-border-subtle mt-auto bg-surface-400/90">
        <button
          onClick={() => setModalOpen(true)}
          className="w-full flex items-center justify-between px-3 py-2 text-xs font-medium text-gray-300 hover:text-white hover:bg-surface-200 rounded-lg transition-colors"
        >
          <div className="flex items-center gap-2">
            <Settings size={15} className="text-gray-400" />
            <span>LLM Gateway 凭证设置</span>
          </div>
          <span className="text-[10px] bg-surface-100 px-1.5 py-0.5 rounded text-gray-400 border border-border-subtle">
            RPC v0.1
          </span>
        </button>
      </div>
    </aside>
  );
};
