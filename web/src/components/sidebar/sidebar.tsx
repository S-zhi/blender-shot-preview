import React, { useState } from "react";
import {
  Search,
  Plus,
  Atom,
  Settings,
  LayoutTemplate,
  FolderPlus,
  SlidersHorizontal,
  ChevronDown,
  ChevronRight,
  Folder,
  ChevronLeft,
  Menu,
} from "lucide-react";
import clsx from "clsx";
import { useSidebarStore } from "#/stores/sidebar-store";
import { useConversationStore } from "#/stores/conversation-store";
import { useSettingsStore } from "#/stores/settings-store";
import { useNavigationStore } from "#/stores/navigation-store";

export const Sidebar: React.FC = () => {
  const { isOpen, toggleSidebar, searchQuery, setSearchQuery } = useSidebarStore();
  const { conversations, activeConversationId, selectConversation, createNewConversation } =
    useConversationStore();
  const { setModalOpen } = useSettingsStore();
  const { activeView, setActiveView } = useNavigationStore();

  const [workspaceOpen, setWorkspaceOpen] = useState(true);

  const filteredConversations = conversations.filter((c) =>
    c.title.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleNewChat = () => {
    const id = createNewConversation();
    selectConversation(id);
    setActiveView("chat");
  };

  const handleSelectConv = (id: string) => {
    selectConversation(id);
    setActiveView("chat");
  };

  if (!isOpen) {
    return (
      <div className="hidden md:flex flex-col items-center justify-between py-3 px-2 border-r border-[#1e2229] bg-[#0c0e12] w-14 shrink-0 select-none">
        <div className="flex flex-col items-center gap-3">
          <button
            onClick={toggleSidebar}
            title="展开侧边栏"
            className="p-2 text-[#8b95a5] hover:text-white hover:bg-white/5 rounded-lg transition-colors"
          >
            <Menu size={18} />
          </button>
          <button
            onClick={handleNewChat}
            title="新建对话"
            className="p-2 text-[#cfb755] hover:bg-white/5 rounded-lg transition-colors"
          >
            <Plus size={18} />
          </button>
        </div>

        <button
          onClick={() => setModalOpen(true)}
          title="模型与凭证设置"
          className="p-2 text-[#8b95a5] hover:text-white hover:bg-white/5 rounded-lg transition-colors"
        >
          <Settings size={18} />
        </button>
      </div>
    );
  }

  return (
    <aside className="w-[244px] h-full flex flex-col bg-[#0c0e12] border-r border-[#1a1d24] select-none shrink-0 text-[13px] font-sans">
      {/* 1. Header: Logo and Close Button */}
      <div className="h-12 px-3 flex items-center justify-between border-b border-[#161920]">
        <div
          onClick={() => setActiveView("home")}
          className="flex items-center gap-2 cursor-pointer group"
        >
          {/* Hands Logo - OpenHands Signature Warm Gold Hands */}
          <div className="w-6 h-6 flex items-center justify-center text-[#cfb755] group-hover:brightness-110 transition-all">
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="w-5 h-5 drop-shadow-[0_0_8px_rgba(207,183,85,0.25)]"
            >
              <path d="M18 11V6a2 2 0 0 0-2-2v0a2 2 0 0 0-2 2v0" />
              <path d="M14 10V4a2 2 0 0 0-2-2v0a2 2 0 0 0-2 2v2" />
              <path d="M10 10.5V6a2 2 0 0 0-2-2v0a2 2 0 0 0-2 2v8" />
              <path d="M18 8a2 2 0 1 1 4 0v6a8 8 0 0 1-8 8h-2c-2.8 0-4.5-.86-5.99-2.34l-3.6-3.6a2 2 0 0 1 2.83-2.82L7 15" />
            </svg>
          </div>
          <span className="font-semibold text-white/90 tracking-tight text-sm group-hover:text-white transition-colors">
            OpenHands
          </span>
        </div>

        <button
          onClick={toggleSidebar}
          title="收起侧边栏"
          className="p-1 text-[#6b7688] hover:text-white hover:bg-white/5 rounded transition-colors"
        >
          <ChevronLeft size={16} />
        </button>
      </div>

      {/* 2. Search commands with shortcut ⌘K */}
      <div className="p-2.5 pb-1">
        <div className="relative flex items-center">
          <Search size={13} className="absolute left-2.5 text-[#6b7688]" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder="Search commands"
            className="w-full h-7 pl-8 pr-8 rounded-lg bg-[#14171d] border border-[#212630] text-xs text-white placeholder-[#5a6372] focus:outline-none focus:border-[#384254] transition-colors"
          />
          <span className="absolute right-2 px-1 py-0.2 rounded text-[10px] font-mono text-[#6b7688] bg-[#1a1f28] border border-[#262c38]">
            ⌘K
          </span>
        </div>
      </div>

      {/* 3. Main Navigation Actions */}
      <div className="px-2 py-1.5 space-y-0.5 border-b border-[#161920]">
        <button
          onClick={handleNewChat}
          className="w-full flex items-center gap-2.5 px-2 py-1.5 rounded-lg text-white/90 hover:bg-[#151921] hover:text-white transition-colors text-left"
        >
          <Plus size={15} className="text-[#8490a5]" />
          <span className="font-medium">新建对话</span>
        </button>

        <button
          onClick={() => setActiveView("custom")}
          className={clsx(
            "w-full flex items-center gap-2.5 px-2 py-1.5 rounded-lg transition-colors text-left font-medium",
            activeView === "custom"
              ? "bg-[#181d26] text-white"
              : "text-[#8490a5] hover:bg-[#151921] hover:text-white"
          )}
        >
          <Atom size={15} className="text-[#8490a5]" />
          <span>自定义</span>
        </button>

        <button
          onClick={() => setActiveView("automate")}
          className={clsx(
            "w-full flex items-center gap-2.5 px-2 py-1.5 rounded-lg transition-colors text-left font-medium",
            activeView === "automate"
              ? "bg-[#181d26] text-white"
              : "text-[#8490a5] hover:bg-[#151921] hover:text-white"
          )}
        >
          <Settings size={15} className="text-[#8490a5]" />
          <span>Automate</span>
        </button>

        <button
          onClick={() => setActiveView("templates")}
          className={clsx(
            "w-full flex items-center gap-2.5 px-2 py-1.5 rounded-lg transition-colors text-left font-medium",
            activeView === "templates"
              ? "bg-[#181d26] text-white"
              : "text-[#8490a5] hover:bg-[#151921] hover:text-white"
          )}
        >
          <LayoutTemplate size={15} className="text-[#8490a5]" />
          <span>自动化模板</span>
        </button>
      </div>

      {/* 4. Conversations Category Header */}
      <div className="px-3 pt-3 pb-1 flex items-center justify-between text-xs">
        <span className="font-medium text-[#8490a5] text-[11px] tracking-wider uppercase">
          对话
        </span>
        <div className="flex items-center gap-1">
          <button
            title="新建工作区文件夹"
            className="p-1 text-[#6b7688] hover:text-white hover:bg-white/5 rounded transition-colors"
          >
            <FolderPlus size={13} />
          </button>
          <button
            title="筛选"
            className="p-1 text-[#6b7688] hover:text-white hover:bg-white/5 rounded transition-colors"
          >
            <SlidersHorizontal size={13} />
          </button>
        </div>
      </div>

      {/* 5. Workspace Tree & Conversation List */}
      <div className="flex-1 overflow-y-auto px-2 space-y-1">
        <div>
          <button
            onClick={() => setWorkspaceOpen(!workspaceOpen)}
            className="w-full flex items-center justify-between px-2 py-1 rounded text-[#8490a5] hover:text-white hover:bg-[#151921] text-xs transition-colors"
          >
            <div className="flex items-center gap-1.5">
              {workspaceOpen ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
              <Folder size={13} className="text-[#6b7688]" />
              <span className="font-medium">无工作区</span>
            </div>
            <Plus size={12} className="text-[#6b7688] hover:text-white" />
          </button>

          {workspaceOpen && (
            <div className="mt-0.5 space-y-0.5 pl-1.5">
              {filteredConversations.map((conv, idx) => {
                const isActive = conv.id === activeConversationId && activeView === "chat";
                const isError = idx === 0 || idx === 3;
                const timeLabel = idx === 0 ? "31m" : idx === 1 ? "4h" : idx === 2 ? "9h" : "11h";

                return (
                  <div
                    key={conv.id}
                    onClick={() => handleSelectConv(conv.id)}
                    className={clsx(
                      "group flex items-center justify-between px-2 py-1.5 rounded-md cursor-pointer text-xs transition-colors",
                      isActive
                        ? "bg-[#181d26] text-white font-medium border-l-2 border-[#cfb755]"
                        : "text-[#8490a5] hover:bg-[#13161c] hover:text-white"
                    )}
                  >
                    <div className="flex items-center gap-2 truncate pr-2">
                      <span
                        className={clsx(
                          "w-1.5 h-1.5 rounded-full shrink-0",
                          isError
                            ? "bg-[#f87171] shadow-[0_0_6px_rgba(248,113,113,0.4)]"
                            : "bg-[#34d399] shadow-[0_0_6px_rgba(52,211,153,0.4)]"
                        )}
                      />
                      <span className="truncate">{conv.title}</span>
                    </div>
                    <span className="text-[10px] text-[#5a6372] shrink-0 font-mono">
                      {timeLabel}
                    </span>
                  </div>
                );
              })}
            </div>
          )}
        </div>

        <div className="pt-2 px-2">
          <button className="text-[11px] text-[#5a6372] hover:text-[#8490a5] transition-colors">
            更多 · 加载更多
          </button>
        </div>
      </div>

      {/* 6. Sidebar Footer: Tenant + Settings */}
      <div className="h-12 px-3 border-t border-[#161920] flex items-center justify-between bg-[#0a0c0f]">
        <button
          onClick={() => setModalOpen(true)}
          className="flex items-center gap-2 hover:bg-[#151921] px-2 py-1 rounded transition-colors text-xs text-white/90"
        >
          <span className="relative flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
            <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
          </span>
          <span className="font-semibold">Bytedance</span>
          <ChevronDown size={13} className="text-[#6b7688]" />
        </button>

        <button
          onClick={() => setModalOpen(true)}
          title="模型设置与凭证"
          className="p-1.5 text-[#8490a5] hover:text-white hover:bg-[#151921] rounded transition-colors"
        >
          <Settings size={16} />
        </button>
      </div>
    </aside>
  );
};
