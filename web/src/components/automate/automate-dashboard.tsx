import React, { useState } from "react";
import {
  FolderGit2,
  ChevronDown,
  Box,
  AlertCircle,
  Activity,
  Clock,
  Search,
  SlidersHorizontal,
  LayoutGrid,
  Play,
  MoreVertical,
  XCircle,
} from "lucide-react";
import { useNavigationStore } from "#/stores/navigation-store";

export const AutomateDashboard: React.FC = () => {
  const [search, setSearch] = useState("");
  const { setActiveView } = useNavigationStore();

  const cards = [
    {
      id: "auto-1",
      title: "【OK】代码Review",
      desc: "# GitHub 代码审查者（OpenHands）你是 S-zhi/LynxSense 仓库的代码审查代理。审查带有 openhands-review 标签的 open PR 并提交评审...",
      cron: "cron",
      error: null,
      runs: "---",
      successRate: "---",
      avgDuration: "---",
    },
    {
      id: "auto-2",
      title: "问题发现",
      desc: "# Mavis 每日 Bug 发现与建单提示词 你是 Mavis，负责对 Agent Canvas 自动克隆到工作区的 GitHub 仓库 'S-zhi/LynxSense' 进行代码静态检测...",
      cron: "cron",
      error: null,
      runs: "---",
      successRate: "---",
      avgDuration: "---",
    },
    {
      id: "auto-3",
      title: "【S-zhi/ThirdBrain】自动合并（GH1-步...",
      desc: "# GitHub PR 自动合并提示词 ## 头部：任务描述 请完成自动化任务：'【S-zhi/ThirdBrain】自动合并（GH1-步骤4）'，你负责将 feature 分支安全合并...",
      cron: "cron",
      error: "失败 LLMAuthenticationError: litellm.AuthenticationError: Provider API Key invalid · 4小时前",
      runs: "2",
      successRate: "50%",
      avgDuration: "2m",
    },
    {
      id: "auto-4",
      title: "Blender 实时分镜自动生成流水线",
      desc: "# Blender Shot Preview Daemon 自动接收场景提示词，调用 EEVEE-Next 引擎批量渲染相机轨迹序列帧并生成实时 WebRTC 预览流...",
      cron: "cron",
      error: null,
      runs: "12",
      successRate: "100%",
      avgDuration: "45s",
    },
  ];

  const filteredCards = cards.filter((c) =>
    c.title.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="flex-1 h-full overflow-y-auto bg-[#090b0e] text-content p-6 lg:p-10 font-sans select-none relative">
      {/* Background Ambient Glow */}
      <div className="absolute top-0 right-1/4 w-[600px] h-[300px] bg-brand-primary/5 blur-[120px] pointer-events-none -z-0" />

      <div className="max-w-6xl mx-auto space-y-7 z-10 relative">
        {/* 1. Header & Breadcrumb */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#171a22] pb-6">
          <div>
            <div className="text-[12px] text-[#8490a5] mb-1 font-medium tracking-wide">
              Automate
            </div>
            <h1 className="text-2xl lg:text-[28px] font-bold tracking-tight text-white/95 mb-1.5">
              Dashboard
            </h1>
            <p className="text-xs text-[#717b8c]">
              Health, activity, and run performance across your automations.
            </p>
          </div>

          <div className="flex items-center gap-2.5 shrink-0">
            <button
              onClick={() => alert("Git 仓库已与本地配置同步完成")}
              className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-[#14171d] hover:bg-[#1a1f29] border border-[#232834] text-xs font-medium text-white/90 hover:text-white transition-all shadow-xs active:scale-98"
            >
              <FolderGit2 size={14} className="text-[#8490a5]" />
              <span>Git 同步</span>
            </button>

            <button
              onClick={() => setActiveView("home")}
              className="flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-[#1c212c] hover:bg-[#252c3b] border border-[#2e3646] text-xs font-medium text-white transition-all shadow-sm active:scale-98"
            >
              <span>添加自动化</span>
              <ChevronDown size={14} className="text-[#8490a5]" />
            </button>
          </div>
        </div>

        {/* 2. 4 Metric KPI Cards */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {/* Card 1: Automations */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>Automations</span>
              <Box size={14} className="text-[#646e80]" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-white/95">15</div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">3 active</div>
            </div>
          </div>

          {/* Card 2: Needs attention */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>Needs attention</span>
              <AlertCircle size={14} className="text-amber-400" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-amber-400">1</div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">Latest run failed</div>
            </div>
          </div>

          {/* Card 3: Total runs */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>Total runs</span>
              <Activity size={14} className="text-[#646e80]" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-white/95">100</div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">Across loaded automations</div>
            </div>
          </div>

          {/* Card 4: Average duration */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>Average duration</span>
              <Clock size={14} className="text-[#646e80]" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-white/95">4m</div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">Recent completed runs</div>
            </div>
          </div>
        </div>

        {/* 3. Search & Filter Toolbar */}
        <div className="flex items-center justify-between gap-3 pt-2">
          <div className="relative flex-1 max-w-md">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#646e80]" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索自动化..."
              className="w-full h-8 pl-9 pr-3 rounded-lg bg-[#14171d] border border-[#202532] text-xs text-white placeholder-[#5a6372] focus:outline-none focus:border-[#384254]"
            />
          </div>

          <div className="flex items-center gap-2">
            <button className="flex items-center gap-1.5 h-8 px-3 rounded-lg bg-[#14171d] hover:bg-[#1a1f29] border border-[#202532] text-xs text-[#8490a5] hover:text-white transition-colors">
              <SlidersHorizontal size={13} className="text-[#646e80]" />
              <span>筛选</span>
              <ChevronDown size={13} className="text-[#646e80]" />
            </button>

            <button
              title="网格视图"
              className="h-8 w-8 rounded-lg bg-[#14171d] hover:bg-[#1a1f29] border border-[#202532] text-[#8490a5] hover:text-white flex items-center justify-center transition-colors"
            >
              <LayoutGrid size={14} />
            </button>
          </div>
        </div>

        {/* 4. Activity Section Header */}
        <div className="flex items-center gap-2 pt-2">
          <span className="text-sm font-semibold text-white/90">活动</span>
          <span className="px-2 py-0.5 rounded-full bg-[#1c212c] text-[10px] font-mono text-[#8490a5] border border-[#28303f]">
            {filteredCards.length}
          </span>
        </div>

        {/* 5. 2-Column Activity Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pb-16">
          {filteredCards.map((item) => (
            <div
              key={item.id}
              className="rounded-2xl bg-[#14171d]/85 border border-[#202532] p-5 flex flex-col justify-between hover:border-[#333d4e] transition-all duration-200 shadow-sm hover:shadow-md"
            >
              <div>
                <div className="flex items-center justify-between mb-2.5">
                  <h3 className="text-sm font-semibold text-white truncate max-w-[80%] tracking-tight">
                    {item.title}
                  </h3>
                  <div className="flex items-center gap-1">
                    <button
                      title="立即执行"
                      className="p-1.5 text-[#8490a5] hover:text-white hover:bg-white/5 rounded-md transition-colors"
                    >
                      <Play size={13} fill="currentColor" />
                    </button>
                    <button
                      title="更多选项"
                      className="p-1.5 text-[#8490a5] hover:text-white hover:bg-white/5 rounded-md transition-colors"
                    >
                      <MoreVertical size={13} />
                    </button>
                  </div>
                </div>

                {/* Prompt Description */}
                <p className="text-[11.5px] text-[#788497] font-mono leading-relaxed line-clamp-2 mb-3 bg-[#0d0f13] p-2.5 rounded-lg border border-[#191d25]">
                  {item.desc}
                </p>

                {/* Tag pill */}
                <div className="flex items-center gap-2 mb-4">
                  <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded bg-[#1a1f29] text-[10px] font-mono text-[#8490a5] border border-[#272e3d]">
                    <Clock size={11} className="text-[#646e80]" />
                    <span>{item.cron}</span>
                  </span>
                </div>

                {/* Error Banner if present */}
                {item.error && (
                  <div className="flex items-center gap-2 p-2.5 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-xs mb-4">
                    <XCircle size={14} className="shrink-0 text-red-400" />
                    <span className="truncate">{item.error}</span>
                  </div>
                )}
              </div>

              {/* Card Footer: Metrics row */}
              <div className="pt-3 border-t border-[#1d222b] grid grid-cols-3 text-center text-xs">
                <div>
                  <div className="text-[10px] text-[#646e80] mb-0.5 font-medium">Runs</div>
                  <div className="font-mono text-white/90 font-medium">{item.runs}</div>
                </div>
                <div>
                  <div className="text-[10px] text-[#646e80] mb-0.5 font-medium">Recent success</div>
                  <div className="font-mono text-white/90 font-medium">{item.successRate}</div>
                </div>
                <div>
                  <div className="text-[10px] text-[#646e80] mb-0.5 font-medium">Avg. duration</div>
                  <div className="font-mono text-white/90 font-medium">{item.avgDuration}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
