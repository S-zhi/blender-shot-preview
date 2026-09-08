import React, { useState, useRef, useEffect } from "react";
import { PanelLeft, Eye, MessageSquareCode, FileText, CheckCircle } from "lucide-react";
import { useConversationStore } from "#/stores/conversation-store";
import { useSidebarStore } from "#/stores/sidebar-store";
import { MessageItem } from "./message-item";
import { InputBox } from "./input-box";
import { EmptyState } from "./empty-state";
import { Badge } from "#/components/ui/badge";

export const ChatArea: React.FC = () => {
  const { conversations, activeConversationId, isGenerating } = useConversationStore();
  const { isOpen, toggleSidebar } = useSidebarStore();
  const [activeTab, setActiveTab] = useState<"chat" | "preview" | "logs">("chat");
  const [selectedPrompt, setSelectedPrompt] = useState<string>("");
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const currentConversation = conversations.find(
    (c) => c.id === activeConversationId
  );

  const messages = currentConversation?.messages || [];

  // Scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isGenerating]);

  return (
    <div className="flex-1 flex flex-col h-full bg-background overflow-hidden">
      {/* Top Header Bar */}
      <header className="h-13 border-b border-border-subtle bg-surface-400/60 backdrop-blur flex items-center justify-between px-4 shrink-0 select-none">
        <div className="flex items-center gap-3">
          {!isOpen && (
            <button
              onClick={toggleSidebar}
              title="打开侧边栏"
              className="p-1.5 text-gray-400 hover:text-white hover:bg-surface-200 rounded-lg transition-colors md:flex"
            >
              <PanelLeft size={17} />
            </button>
          )}

          <div className="flex items-center gap-2">
            <h2 className="text-sm font-semibold text-gray-200 truncate max-w-xs md:max-w-md">
              {currentConversation?.title || "分镜预览控制台"}
            </h2>
            <Badge variant="info" className="hidden sm:inline-flex text-[11px]">
              <CheckCircle size={11} />
              {isGenerating ? "生成中" : "就绪 (Ready)"}
            </Badge>
          </div>
        </div>

        {/* View Tabs */}
        <div className="flex items-center gap-1 bg-surface-300 p-0.5 rounded-lg border border-border-subtle text-xs">
          <button
            onClick={() => setActiveTab("chat")}
            className={`flex items-center gap-1.5 px-3 py-1 rounded-md font-medium transition-colors ${
              activeTab === "chat"
                ? "bg-surface-100 text-white shadow-sm"
                : "text-gray-400 hover:text-gray-200"
            }`}
          >
            <MessageSquareCode size={13} />
            <span>对话</span>
          </button>
          <button
            onClick={() => setActiveTab("preview")}
            className={`flex items-center gap-1.5 px-3 py-1 rounded-md font-medium transition-colors ${
              activeTab === "preview"
                ? "bg-surface-100 text-white shadow-sm"
                : "text-gray-400 hover:text-gray-200"
            }`}
          >
            <Eye size={13} />
            <span>分镜视图</span>
          </button>
          <button
            onClick={() => setActiveTab("logs")}
            className={`flex items-center gap-1.5 px-3 py-1 rounded-md font-medium transition-colors ${
              activeTab === "logs"
                ? "bg-surface-100 text-white shadow-sm"
                : "text-gray-400 hover:text-gray-200"
            }`}
          >
            <FileText size={13} />
            <span>渲染日志</span>
          </button>
        </div>
      </header>

      {/* Main Tab Content */}
      <div className="flex-1 overflow-y-auto flex flex-col">
        {activeTab === "chat" && (
          <>
            {messages.length === 0 ? (
              <EmptyState onSelectPrompt={(prompt) => setSelectedPrompt(prompt)} />
            ) : (
              <div className="flex-1 py-4 space-y-1 max-w-4xl w-full mx-auto">
                {messages.map((msg) => (
                  <MessageItem key={msg.id} message={msg} />
                ))}
                <div ref={messagesEndRef} />
              </div>
            )}
          </>
        )}

        {activeTab === "preview" && (
          <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-gray-400 select-none">
            <div className="w-96 h-56 rounded-xl border border-dashed border-border-strong bg-surface-300/40 flex flex-col items-center justify-center gap-3 p-4">
              <Eye size={32} className="text-brand-primary/60" />
              <div className="text-xs text-gray-300 font-medium">
                Blender 实时分镜视口 (Viewport)
              </div>
              <p className="text-[11px] text-gray-500 max-w-xs">
                当后台 ShotPreview 渲染任务完成时，生成的镜头序列帧或 WebRTC 实时流将嵌入在此处渲染。
              </p>
            </div>
          </div>
        )}

        {activeTab === "logs" && (
          <div className="flex-1 p-4 font-mono text-xs text-gray-400 bg-surface-500 overflow-y-auto">
            <div className="text-emerald-400">[INFO] Kitex RPC client initialized.</div>
            <div className="text-gray-500">[DEBUG] Registered ShotPreviewServiceV0_1.CreateShotPreviewTask handler.</div>
            <div className="text-gray-500">[DEBUG] Registered LLMKeyServiceV0_1.ManageLLMKey handler.</div>
            <div className="text-cyan-400 mt-2">[INFO] Ready to receive shot preview instructions.</div>
          </div>
        )}
      </div>

      {/* Bottom Floating Input Box */}
      <div className="shrink-0">
        <InputBox
          initialPrompt={selectedPrompt}
          onClearInitialPrompt={() => setSelectedPrompt("")}
        />
      </div>
    </div>
  );
};
