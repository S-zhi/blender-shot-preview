import React, { useState, useRef, useEffect } from "react";
import {
  ArrowUp,
  Plus,
  ChevronDown,
  Folder,
  Box,
  Github,
  MessageSquare,
  Bot,
  Sparkles,
  Cpu,
} from "lucide-react";
import clsx from "clsx";
import { useConversationStore } from "#/stores/conversation-store";
import { useSettingsStore } from "#/stores/settings-store";
import { useNavigationStore } from "#/stores/navigation-store";

export const HomeCanvas: React.FC = () => {
  const [prompt, setPrompt] = useState("");
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const { sendMessage } = useConversationStore();
  const { modelName, setModalOpen } = useSettingsStore();
  const { setActiveView } = useNavigationStore();

  useEffect(() => {
    textareaRef.current?.focus();
  }, []);

  const handleSend = () => {
    const trimmed = prompt.trim();
    if (!trimmed) return;
    setActiveView("chat");
    sendMessage(trimmed);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const recommendations = [
    {
      icon: Github,
      iconColor: "text-white",
      iconBg: "bg-white/10",
      title: "GitHub Code Review Agent",
      desc: "Watch for a configurable label on GitHub pull requests, inspect full code diffs...",
      prompt: "创建代码审查代理，检查 pull request 变更并给出详尽的优化与安全建议",
    },
    {
      icon: Github,
      iconColor: "text-white",
      iconBg: "bg-white/10",
      title: "GitHub Issue to PR Agent",
      desc: "Watch for a configurable label on GitHub issues, implement the feature and open PR...",
      prompt: "针对 GitHub Issue 自动实现代码并创建对应的 Pull Request",
    },
    {
      icon: MessageSquare,
      iconColor: "text-[#e01e5a]",
      iconBg: "bg-[#e01e5a]/10",
      title: "Slack channel monitor",
      desc: "Watch Slack channels for @openhands mentions, open a dedicated workspace...",
      prompt: "监控 Slack 频道中的 @openhands 提及并自动响应用户提问与任务分派",
    },
    {
      icon: Bot,
      iconColor: "text-[#cfb755]",
      iconBg: "bg-[#cfb755]/10",
      title: "AGENTS Repo Maintainer",
      desc: "Keep AGENTS.md up to date with repository changes, verify agent workflow rules...",
      prompt: "保持工作区文档与 AGENTS.md 准则同步，自动检查项目架构规则与分支规范",
    },
  ];

  return (
    <div className="flex-1 h-full overflow-y-auto bg-[#090b0e] text-content flex flex-col items-center px-4 py-16 select-none font-sans relative">
      {/* Subtle Ambient Radial Lighting in the background */}
      <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[700px] h-[340px] bg-gradient-to-b from-brand-primary/5 via-blue-500/5 to-transparent blur-3xl pointer-events-none -z-0" />

      <div className="w-full max-w-[680px] flex flex-col items-center z-10">
        {/* 1. Big Hero Title with clean letter-spacing */}
        <h1 className="text-3xl md:text-[34px] font-semibold tracking-[-0.03em] text-white/95 mb-8">
          让我们开始开发！
        </h1>

        {/* 2. Signature OpenHands Input Box (High-fidelity Hero Card) */}
        <div className="w-full rounded-2xl bg-[#14171d]/90 backdrop-blur-md border border-[#232834] shadow-[0_16px_48px_-12px_rgba(0,0,0,0.7)] p-4 focus-within:border-[#384154] focus-within:shadow-[0_20px_60px_-15px_rgba(0,0,0,0.9)] focus-within:ring-1 focus-within:ring-white/10 transition-all duration-200">
          <textarea
            ref={textareaRef}
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="你想要构建什么？"
            rows={3}
            className="w-full bg-transparent text-[15px] text-white/90 placeholder-[#636c7a] resize-none focus:outline-none leading-relaxed selection:bg-brand-primary/30"
          />

          {/* Card Bottom Controls */}
          <div className="flex items-center justify-between pt-3 border-t border-[#1f242f]/80 mt-1">
            {/* Left: Plus & Model Selector */}
            <div className="flex items-center gap-2">
              <button
                type="button"
                title="添加附件或工作区上下文"
                className="w-7 h-7 rounded-lg bg-[#1a1f29] hover:bg-[#242b38] text-content-muted hover:text-white flex items-center justify-center transition-all duration-150 border border-white/5 active:scale-95"
              >
                <Plus size={15} />
              </button>

              <button
                type="button"
                onClick={() => setModalOpen(true)}
                className="flex items-center gap-1.5 px-3 py-1 rounded-full bg-[#1a1f29] hover:bg-[#242b38] text-xs font-medium text-white/80 hover:text-white transition-all duration-150 border border-white/10 active:scale-95 shadow-xs"
              >
                <Cpu size={12} className="text-[#cfb755]" />
                <span>{modelName || "Grok"}</span>
                <ChevronDown size={13} className="text-content-icon ml-0.5" />
              </button>
            </div>

            {/* Right: Signature Circular Arrow Button */}
            <button
              type="button"
              onClick={handleSend}
              disabled={!prompt.trim()}
              className={clsx(
                "w-8 h-8 rounded-full flex items-center justify-center transition-all duration-200",
                prompt.trim()
                  ? "bg-white text-black hover:bg-white/90 hover:scale-105 shadow-md active:scale-95 cursor-pointer"
                  : "bg-[#212631] text-[#555d6b] cursor-not-allowed"
              )}
            >
              <ArrowUp size={16} strokeWidth={2.6} />
            </button>
          </div>
        </div>

        {/* 3. Action Pills directly below Input Box */}
        <div className="w-full flex items-center gap-2 mt-4">
          <button
            onClick={() => setActiveView("chat")}
            className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-[#14171d]/80 hover:bg-[#1a1f29] border border-[#232834] text-xs font-medium text-content-muted hover:text-white transition-all duration-150 active:scale-98 shadow-xs"
          >
            <Folder size={13.5} className="text-content-icon" />
            <span>打开工作区</span>
          </button>

          <button
            onClick={() => setActiveView("custom")}
            className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-[#14171d]/80 hover:bg-[#1a1f29] border border-[#232834] text-xs font-medium text-content-muted hover:text-white transition-all duration-150 active:scale-98 shadow-xs"
          >
            <Box size={13.5} className="text-content-icon" />
            <span>插件</span>
          </button>

          <button
            onClick={() => {
              setPrompt("为赛博朋克风格雨夜街道创建由远及近的特写推镜头，重点展现路面霓虹反光与积水波纹");
              textareaRef.current?.focus();
            }}
            className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-[#14171d]/80 hover:bg-[#1a1f29] border border-[#232834] text-xs font-medium text-content-muted hover:text-white transition-all duration-150 ml-auto active:scale-98 shadow-xs"
          >
            <Sparkles size={13} className="text-[#cfb755]" />
            <span>常用分镜预设</span>
          </button>
        </div>

        {/* 4. Recommendations Header & 4-Column Grid */}
        <div className="w-full mt-12">
          <div className="text-[12px] font-medium text-[#8490a5] mb-3.5 tracking-wide">
            推荐的自动化
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
            {recommendations.map((item, idx) => {
              const Icon = item.icon;
              return (
                <div
                  key={idx}
                  onClick={() => {
                    setPrompt(item.prompt);
                    textareaRef.current?.focus();
                  }}
                  className="rounded-xl bg-[#14171d]/70 hover:bg-[#191e26] border border-[#202530] hover:border-[#353d4f] hover:-translate-y-1 p-3.5 flex flex-col justify-between transition-all duration-200 cursor-pointer group shadow-sm hover:shadow-lg"
                >
                  <div>
                    <div className={clsx("w-8 h-8 rounded-lg flex items-center justify-center mb-3 group-hover:scale-105 transition-transform", item.iconBg)}>
                      <Icon size={16} className={item.iconColor} />
                    </div>
                    <h3 className="text-xs font-semibold text-white/90 mb-1.5 line-clamp-1 group-hover:text-white tracking-tight">
                      {item.title}
                    </h3>
                    <p className="text-[11px] text-[#717b8c] leading-relaxed line-clamp-3">
                      {item.desc}
                    </p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* 5. Automations Section Footer */}
        <div className="w-full mt-12 pb-16">
          <div className="flex items-center justify-between mb-3 text-xs">
            <span className="font-medium text-[#8490a5] tracking-wide">自动化</span>
            <button
              onClick={() => setActiveView("automate")}
              className="text-[11px] text-[#8490a5] hover:text-white transition-colors flex items-center gap-1 font-medium"
            >
              <Plus size={13} />
              <span>添加</span>
            </button>
          </div>

          <div className="rounded-xl bg-[#14171d]/70 border border-[#202530] divide-y divide-[#1e232d] shadow-sm overflow-hidden">
            <div
              onClick={() => setActiveView("automate")}
              className="p-3.5 flex items-center justify-between hover:bg-[#191e26] transition-colors cursor-pointer text-xs"
            >
              <div className="flex items-center gap-3">
                <span className="relative flex h-2 w-2">
                  <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                  <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
                </span>
                <span className="font-medium text-white/90">【OK】代码Review</span>
                <span className="text-[11px] text-[#717b8c] font-mono">cron: */15 * * * *</span>
              </div>
              <span className="text-[11px] text-[#717b8c]">4小时前</span>
            </div>

            <div
              onClick={() => setActiveView("automate")}
              className="p-3.5 flex items-center justify-between hover:bg-[#191e26] transition-colors cursor-pointer text-xs"
            >
              <div className="flex items-center gap-3">
                <span className="w-3.5 h-3.5 rounded-full border border-amber-500/40 bg-amber-500/10 flex items-center justify-center text-[9px] text-amber-400 font-bold">
                  !
                </span>
                <span className="font-medium text-white/90">问题发现 — 2026-09-08</span>
                <span className="text-[11px] text-[#717b8c] font-mono">cron: 0 9 * * *</span>
              </div>
              <span className="text-[11px] text-[#717b8c]">31m 前</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
