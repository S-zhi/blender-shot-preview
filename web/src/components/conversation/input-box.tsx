import React, { useState, useRef, useEffect } from "react";
import { ArrowUp, Square, Paperclip, Cpu, Sparkles } from "lucide-react";
import clsx from "clsx";
import { useConversationStore } from "#/stores/conversation-store";
import { useSettingsStore } from "#/stores/settings-store";
import { LLMProvider } from "#/api/types";

interface InputBoxProps {
  initialPrompt?: string;
  onClearInitialPrompt?: () => void;
}

export const InputBox: React.FC<InputBoxProps> = ({
  initialPrompt,
  onClearInitialPrompt,
}) => {
  const [content, setContent] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { sendMessage, isGenerating, stopGenerating } = useConversationStore();
  const { provider, modelName, setModalOpen } = useSettingsStore();

  useEffect(() => {
    if (initialPrompt) {
      setContent(initialPrompt);
      if (textareaRef.current) {
        textareaRef.current.focus();
      }
      onClearInitialPrompt?.();
    }
  }, [initialPrompt, onClearInitialPrompt]);

  // Auto-resize textarea height
  useEffect(() => {
    const el = textareaRef.current;
    if (el) {
      el.style.height = "auto";
      el.style.height = `${Math.min(el.scrollHeight, 180)}px`;
    }
  }, [content]);

  const handleSend = () => {
    const trimmed = content.trim();
    if (!trimmed || isGenerating) return;
    sendMessage(trimmed);
    setContent("");
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="w-full max-w-3xl mx-auto px-4 pb-4">
      {/* Container styled like OpenHands floating card input */}
      <div className="relative rounded-2xl bg-surface-300 border border-border-default/80 shadow-xl focus-within:border-border-strong focus-within:ring-1 focus-within:ring-brand-primary/40 transition-all p-3">
        {/* Main Textarea */}
        <textarea
          ref={textareaRef}
          value={content}
          onChange={(e) => setContent(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="描述分镜需求（例如：设计一个俯角 45° 赛博朋克街道推镜头）... Enter 发送，Shift + Enter 换行"
          rows={1}
          disabled={isGenerating}
          className="w-full bg-transparent text-sm text-gray-100 placeholder-gray-500 resize-none focus:outline-none max-h-44 min-h-[44px] py-1 px-1 leading-relaxed"
        />

        {/* Bottom Toolbar */}
        <div className="flex items-center justify-between pt-2 border-t border-border-subtle/50 mt-1 select-none">
          {/* Left Action Buttons */}
          <div className="flex items-center gap-1.5 text-xs text-gray-400">
            {/* Model Badge */}
            <button
              type="button"
              onClick={() => setModalOpen(true)}
              title="切换模型 / 配置 LLM Key"
              className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-surface-200/80 hover:bg-surface-100 text-gray-300 hover:text-white border border-border-subtle transition-colors cursor-pointer"
            >
              <Cpu size={13} className="text-brand-primary" />
              <span className="font-mono text-[11px] font-medium">
                {provider === LLMProvider.OPENAI ? "OpenAI" : "Anthropic"}: {modelName}
              </span>
            </button>

            {/* Quick action: Attachment / Scene Preset */}
            <button
              type="button"
              title="添加参考图片或 Blender 场景资产"
              className="p-1.5 rounded-lg hover:bg-surface-200 text-gray-400 hover:text-gray-200 transition-colors"
            >
              <Paperclip size={15} />
            </button>

            <button
              type="button"
              title="使用预置分镜模版"
              onClick={() =>
                setContent("创建 180 帧平滑环绕运镜，摄影机定焦中心主体，光圈 f/2.8")
              }
              className="hidden sm:flex items-center gap-1 px-2 py-1 rounded-lg hover:bg-surface-200 text-[11px] text-gray-400 hover:text-gray-200 transition-colors"
            >
              <Sparkles size={12} className="text-amber-400" />
              <span>常用镜头</span>
            </button>
          </div>

          {/* Right Action: Send / Stop button */}
          <div>
            {isGenerating ? (
              <button
                type="button"
                onClick={stopGenerating}
                title="停止生成"
                className="h-8 w-8 rounded-xl bg-red-500/20 text-red-400 hover:bg-red-500/30 flex items-center justify-center transition-colors cursor-pointer"
              >
                <Square size={13} fill="currentColor" />
              </button>
            ) : (
              <button
                type="button"
                onClick={handleSend}
                disabled={!content.trim()}
                title="发送分镜指令"
                className={clsx(
                  "h-8 w-8 rounded-xl flex items-center justify-center transition-all cursor-pointer",
                  content.trim()
                    ? "bg-brand-primary text-gray-950 hover:bg-brand-accent shadow-sm"
                    : "bg-surface-200 text-gray-500 cursor-not-allowed"
                )}
              >
                <ArrowUp size={16} strokeWidth={2.5} />
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
