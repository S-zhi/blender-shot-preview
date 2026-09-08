import React, { useState } from "react";
import {
  CheckCircle2,
  Clock,
  Loader2,
  AlertTriangle,
  ChevronDown,
  ChevronRight,
  Eye,
  Check,
  Zap,
} from "lucide-react";
import { TaskNodeView } from "#/api/types";
import { Button } from "#/components/ui/button";

interface WorkflowProgressPanelProps {
  nodes?: TaskNodeView[];
  waitingNode?: TaskNodeView | null;
  autoConfirm?: boolean;
  onInspectNode: (node: TaskNodeView) => void;
  onConfirmStep: (nodeId: string) => Promise<void>;
  onToggleAutoConfirm: () => void;
}

interface WorkflowStepMeta {
  id: string;
  name: string;
  kind: "agent" | "tool" | "step";
  target: string;
}

const WORKFLOW_STEPS: WorkflowStepMeta[] = [
  { id: "Initialize", name: "工作流初始化", kind: "step", target: "Initialize" },
  { id: "Intent", name: "提示词意图解析", kind: "agent", target: "intent-agent" },
  { id: "ValidateSpec", name: "场景规格契约校验", kind: "step", target: "ValidateSpec" },
  { id: "ScenePlan", name: "分镜与场景规划", kind: "agent", target: "scene-planner-agent" },
  { id: "CreateAssets", name: "资产调度与生产", kind: "step", target: "asset-creator-agent" },
  { id: "DesignShots", name: "机位与分镜设计", kind: "agent", target: "shot-designer-agent" },
  { id: "AssembleScene", name: "Blender 场景工程组装", kind: "agent", target: "scene-assembly-agent" },
  { id: "PreviewRender", name: "单帧低精度预览渲染", kind: "tool", target: "blender.render.submit" },
  { id: "Inspect", name: "分镜质检与审核", kind: "step", target: "Inspect" },
  { id: "FinalRender", name: "完整序列帧渲染", kind: "tool", target: "blender.render.submit" },
  { id: "Encode", name: "视频编码 (H.264)", kind: "tool", target: "ffmpeg.encode" },
  { id: "Verify", name: "视频规格校验", kind: "tool", target: "ffprobe.inspect" },
  { id: "Publish", name: "分镜视频制品发布", kind: "step", target: "Publish" },
];

export const WorkflowProgressPanel: React.FC<WorkflowProgressPanelProps> = ({
  nodes = [],
  waitingNode,
  autoConfirm = false,
  onInspectNode,
  onConfirmStep,
  onToggleAutoConfirm,
}) => {
  const [isExpanded, setIsExpanded] = useState(true);
  const [confirmingNodeId, setConfirmingNodeId] = useState<string | null>(null);

  const nodeMap = new Map<string, TaskNodeView>();
  nodes.forEach((n) => nodeMap.set(n.node_id, n));

  const completedCount = nodes.filter((n) => n.status === "succeeded").length;
  const isAllSucceeded = nodes.length > 0 && completedCount === WORKFLOW_STEPS.length;
  const hasWaiting = !!waitingNode;

  const handleConfirm = async (nodeId: string) => {
    setConfirmingNodeId(nodeId);
    try {
      await onConfirmStep(nodeId);
    } finally {
      setConfirmingNodeId(null);
    }
  };

  if (nodes.length === 0) return null;

  return (
    <div className="my-3 border border-border bg-surface-card rounded-2xl overflow-hidden shadow-sm transition-all">
      {/* Header bar */}
      <div className="flex items-center justify-between px-4 py-3 bg-surface-elevated/80 border-b border-border">
        <button
          type="button"
          onClick={() => setIsExpanded(!isExpanded)}
          className="flex items-center gap-2 text-xs font-semibold text-content hover:text-brand-primary transition-colors text-left"
        >
          {hasWaiting ? (
            <span className="relative flex h-2.5 w-2.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-amber-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-amber-500"></span>
            </span>
          ) : isAllSucceeded ? (
            <CheckCircle2 size={14} className="text-emerald-400" />
          ) : (
            <Loader2 size={14} className="animate-spin text-brand-primary" />
          )}

          <span>已编排 Agent 工作流执行状态</span>
          <span className="text-[11px] font-normal text-content-muted">
            ({completedCount}/{WORKFLOW_STEPS.length} 步已完成)
          </span>
          {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        </button>

        {/* Auto confirm switch */}
        <div className="flex items-center gap-2 text-[11px]">
          <label className="flex items-center gap-1.5 cursor-pointer select-none text-content-muted hover:text-content">
            <input
              type="checkbox"
              checked={autoConfirm}
              onChange={onToggleAutoConfirm}
              className="rounded border-border text-brand-primary focus:ring-0 focus:ring-offset-0 bg-surface-background h-3.5 w-3.5"
            />
            <span className="flex items-center gap-1">
              <Zap size={11} className={autoConfirm ? "text-amber-400" : "text-content-icon"} />
              自动确认后续步骤
            </span>
          </label>
        </div>
      </div>

      {/* Pending Confirmation Callout Banner if any */}
      {waitingNode && (
        <div className="p-3.5 bg-amber-500/10 border-b border-amber-500/30 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 animate-fade-in">
          <div className="flex items-start gap-2.5">
            <AlertTriangle size={16} className="text-amber-400 shrink-0 mt-0.5" />
            <div>
              <div className="text-xs font-semibold text-amber-300">
                等待确认：[{waitingNode.node_id}] {WORKFLOW_STEPS.find((s) => s.id === waitingNode.node_id)?.name}
              </div>
              <div className="text-[11px] text-content-muted mt-0.5">
                当前节点输出已生成，流程已暂停。请检查数据契约是否符合预期，若不符合可在线调整后确认。
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 self-end sm:self-center shrink-0">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => onInspectNode(waitingNode)}
              className="text-xs h-7 px-2.5 bg-surface-card hover:bg-surface-elevated text-content"
            >
              <Eye size={12} className="mr-1" />
              检查 / 调整数据
            </Button>
            <Button
              variant="primary"
              size="sm"
              onClick={() => handleConfirm(waitingNode.node_id)}
              disabled={confirmingNodeId === waitingNode.node_id}
              className="text-xs h-7 px-3 bg-amber-500 hover:bg-amber-600 text-black font-semibold"
            >
              {confirmingNodeId === waitingNode.node_id ? (
                <Loader2 size={12} className="animate-spin mr-1" />
              ) : (
                <Check size={12} className="mr-1" />
              )}
              确认执行下一步
            </Button>
          </div>
        </div>
      )}

      {/* Steps List */}
      {isExpanded && (
        <div className="p-3 space-y-1 bg-surface-card/60 divide-y divide-border/40">
          {WORKFLOW_STEPS.map((step, idx) => {
            const node = nodeMap.get(step.id);
            const status = node ? node.status : "pending";
            const isWaitingThis = waitingNode?.node_id === step.id;

            return (
              <div
                key={step.id}
                className={`pt-1.5 pb-1.5 px-2 rounded-lg flex items-center justify-between text-xs transition-colors ${
                  isWaitingThis
                    ? "bg-amber-500/15 border border-amber-500/40"
                    : status === "running"
                    ? "bg-brand-primary/10 border border-brand-primary/20"
                    : "hover:bg-surface-elevated/40"
                }`}
              >
                {/* Left: Step index, icon, title & agent badge */}
                <div className="flex items-center gap-2.5 min-w-0">
                  <span className="text-[10px] font-mono text-content-icon w-4 shrink-0 text-right">
                    {idx + 1}
                  </span>

                  <div className="shrink-0">
                    {status === "succeeded" ? (
                      <CheckCircle2 size={14} className="text-emerald-400" />
                    ) : status === "waiting_confirmation" ? (
                      <AlertTriangle size={14} className="text-amber-400 animate-pulse" />
                    ) : status === "running" ? (
                      <Loader2 size={14} className="animate-spin text-brand-primary" />
                    ) : (
                      <Clock size={14} className="text-content-icon" />
                    )}
                  </div>

                  <span
                    className={`truncate font-medium ${
                      status === "succeeded"
                        ? "text-content"
                        : status === "waiting_confirmation"
                        ? "text-amber-300 font-semibold"
                        : status === "running"
                        ? "text-brand-primary font-semibold"
                        : "text-content-muted"
                    }`}
                  >
                    {step.name}
                  </span>

                  {step.kind === "agent" && (
                    <span className="text-[9px] px-1.5 py-0.2 rounded bg-brand-primary/20 text-brand-primary font-mono hidden sm:inline">
                      Agent
                    </span>
                  )}
                </div>

                {/* Right: Actions & State */}
                <div className="flex items-center gap-2 shrink-0">
                  {/* Inspect Button if node has data */}
                  {(node?.input || node?.output) && (
                    <button
                      type="button"
                      onClick={() => node && onInspectNode(node)}
                      className="text-[11px] text-content-muted hover:text-brand-primary flex items-center gap-1 px-1.5 py-0.5 rounded hover:bg-surface-background transition-colors"
                      title="查看该步骤的 Input/Output 数据"
                    >
                      <Eye size={11} />
                      <span>查看数据</span>
                    </button>
                  )}

                  {/* Status Tag */}
                  <span
                    className={`text-[10px] px-1.5 py-0.5 rounded font-mono ${
                      status === "succeeded"
                        ? "bg-emerald-500/10 text-emerald-400"
                        : status === "waiting_confirmation"
                        ? "bg-amber-500/20 text-amber-300 font-semibold"
                        : status === "running"
                        ? "bg-brand-primary/20 text-brand-primary animate-pulse"
                        : "text-content-icon"
                    }`}
                  >
                    {status === "succeeded"
                      ? "完成"
                      : status === "waiting_confirmation"
                      ? "待确认"
                      : status === "running"
                      ? "运行中"
                      : "等待"}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
