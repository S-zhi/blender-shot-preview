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
  const activeMessage = [...messages].reverse().find((message) => message.role === "assistant");
  const connectionFailed = activeMessage?.status === "error" || activeMessage?.status === "timeout" || activeMessage?.status === "disconnected";

  // Scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isGenerating]);

  return (
    <div className="flex-1 flex flex-col h-full bg-surface overflow-hidden">
      {/* Top Header Bar */}
      <header className="h-13 border-b border-border bg-surface-card/80 backdrop-blur flex items-center justify-between px-4 shrink-0 select-none">
        <div className="flex items-center gap-3">
          {!isOpen && (
            <button
              onClick={toggleSidebar}
              title="打开侧边栏"
              className="p-1.5 text-content-muted hover:text-content hover:bg-surface-elevated rounded-lg transition-colors md:flex"
            >
              <PanelLeft size={17} />
            </button>
          )}

          <div className="flex items-center gap-2">
            <h2 className="text-sm font-semibold text-content truncate max-w-xs md:max-w-md">
              {currentConversation?.title || "分镜预览控制台"}
            </h2>
            <Badge variant={connectionFailed ? "danger" : isGenerating ? "info" : "success"} className="hidden sm:inline-flex text-[11px]">
              <CheckCircle size={11} />
              {connectionFailed ? "LLM / 后端连接失败" : isGenerating ? "生成中" : "就绪 (Ready)"}
            </Badge>
          </div>
        </div>

        {/* View Tabs */}
        <div className="flex items-center gap-1 bg-pill-bg p-0.5 rounded-lg border border-border text-xs">
          <button
            onClick={() => setActiveTab("chat")}
            className={`flex items-center gap-1.5 px-3 py-1 rounded-md font-medium transition-colors ${
              activeTab === "chat"
                ? "bg-surface-elevated text-content shadow-sm"
                : "text-content-muted hover:text-content"
            }`}
          >
            <MessageSquareCode size={13} />
            <span>对话</span>
          </button>
          <button
            onClick={() => setActiveTab("preview")}
            className={`flex items-center gap-1.5 px-3 py-1 rounded-md font-medium transition-colors ${
              activeTab === "preview"
                ? "bg-surface-elevated text-content shadow-sm"
                : "text-content-muted hover:text-content"
            }`}
          >
            <Eye size={13} />
            <span>分镜视图</span>
          </button>
          <button
            onClick={() => setActiveTab("logs")}
            className={`flex items-center gap-1.5 px-3 py-1 rounded-md font-medium transition-colors ${
              activeTab === "logs"
                ? "bg-surface-elevated text-content shadow-sm"
                : "text-content-muted hover:text-content"
            }`}
          >
            <FileText size={13} />
            <span>渲染日志</span>
          </button>
        </div>
      </header>

      {connectionFailed && activeMessage?.content && (
        <div className="shrink-0 border-b border-status-fail-border bg-status-fail-bg px-4 py-2 text-xs text-status-fail-text">
          {activeMessage.content.replace(/^### .*\n\n/, "").replace(/\*\*/g, "")}
        </div>
      )}

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
          <div className="flex-1 flex flex-col items-center justify-center p-6 select-none bg-surface">
            <div className="w-full max-w-3xl aspect-video rounded-2xl border border-border bg-surface-card relative overflow-hidden flex flex-col justify-between p-4 shadow-2xl group">
              {/* Camera HUD Header */}
              <div className="flex items-center justify-between z-10">
                <div className="flex items-center gap-2">
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-surface-elevated/80 backdrop-blur border border-border text-[11px] font-mono text-brand-primary">
                    <span className="w-1.5 h-1.5 rounded-full bg-brand-primary animate-pulse" />
                    CAM_01: Active
                  </span>
                  <span className="px-2 py-1 rounded-md bg-surface-elevated/60 text-[10px] font-mono text-content-muted border border-border">
                    35mm · f/2.8 · 24 FPS
                  </span>
                </div>
                <div className="text-[11px] font-mono text-content-muted bg-surface-elevated/60 px-2 py-0.5 rounded border border-border">
                  1920 × 1080
                </div>
              </div>

              {/* Viewport Center Grid & Crosshair */}
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none opacity-20">
                <div className="w-full h-full border border-dashed border-content-muted/30 grid grid-cols-3 grid-rows-3" />
                <div className="absolute w-8 h-8 border-t border-b border-l border-r border-brand-primary/40" />
              </div>

              {/* Viewport Center Indicator */}
              <div className="flex flex-col items-center justify-center gap-2.5 z-10 my-auto">
                <div className="w-12 h-12 rounded-2xl bg-brand-primary/10 border border-brand-primary/30 flex items-center justify-center text-brand-primary">
                  <Eye size={22} />
                </div>
                <div className="text-xs font-semibold text-content">
                  Blender 实时分镜视口 (Agent Viewport)
                </div>
                <p className="text-[11px] text-content-muted max-w-sm text-center leading-relaxed">
                  通过 RPC 调用触发 Blender 后台任务后，生成的镜头序列帧或实时预览流将无缝嵌入此画板。
                </p>
              </div>

              {/* Viewport HUD Footer */}
              <div className="flex items-center justify-between z-10 text-[10px] font-mono text-content-icon border-t border-border/40 pt-2">
                <div>Frame: 001 / 180 (0.0s)</div>
                <div className="flex items-center gap-3">
                  <span>Engine: EEVEE-Next</span>
                  <span>SSR: ON</span>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === "logs" && (
          <div className="flex-1 p-4 bg-surface flex flex-col">
            <div className="flex-1 rounded-2xl border border-border bg-surface-card overflow-hidden flex flex-col font-mono text-xs shadow-xl">
              {/* Terminal Window Header */}
              <div className="h-9 px-3.5 bg-surface-elevated/80 border-b border-border flex items-center justify-between select-none">
                <div className="flex items-center gap-2">
                  <div className="flex items-center gap-1.5">
                    <div className="w-2.5 h-2.5 rounded-full bg-status-fail-solid/80" />
                    <div className="w-2.5 h-2.5 rounded-full bg-amber-500/80" />
                    <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/80" />
                  </div>
                  <span className="text-[11px] text-content-muted ml-2 font-medium">
                    shot-preview-daemon · stdout
                  </span>
                </div>
                <span className="text-[10px] text-content-icon">Kitex RPC v0.1</span>
              </div>

              {/* Terminal Log Content */}
              <div className="flex-1 p-4 space-y-2 overflow-y-auto leading-relaxed">
                <div className="flex items-start gap-2 text-status-success-text">
                  <span className="text-content-icon">[13:00:01]</span>
                  <span>[INFO] Kitex RPC client initialized on 127.0.0.1:8888.</span>
                </div>
                <div className="flex items-start gap-2 text-content-muted">
                  <span className="text-content-icon">[13:00:02]</span>
                  <span>[DEBUG] Registered ShotPreviewServiceV0_1.CreateShotPreviewTask handler.</span>
                </div>
                <div className="flex items-start gap-2 text-content-muted">
                  <span className="text-content-icon">[13:00:02]</span>
                  <span>[DEBUG] Registered LLMKeyServiceV0_1.ManageLLMKey handler.</span>
                </div>
                <div className="flex items-start gap-2 text-cyan-400">
                  <span className="text-content-icon">[13:00:03]</span>
                  <span>[INFO] Connection verified. Agent Canvas ready to receive shot instructions.</span>
                </div>
              </div>
            </div>
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
