import React, { useState, useEffect } from "react";
import { X, Check, Save, AlertCircle, Eye, Edit3 } from "lucide-react";
import { TaskNodeView } from "#/api/types";
import { Button } from "#/components/ui/button";
import { Badge } from "#/components/ui/badge";

interface NodeInspectorModalProps {
  isOpen: boolean;
  node: TaskNodeView | null;
  onClose: () => void;
  onSaveAdjust: (outputJson: string) => Promise<void>;
  onConfirm: (adjustedOutput?: string) => Promise<void>;
}

const nodeNameMap: Record<string, string> = {
  Initialize: "初始化准备 (Initialize)",
  Intent: "意图解析 Agent (Intent Agent)",
  ValidateSpec: "规格契约校验 (ValidateSpec)",
  ScenePlan: "场景分镜规划 Agent (Scene Planner)",
  CreateAssets: "资产生产调度 (CreateAssets FanOut)",
  DesignShots: "机位分镜设计 Agent (Shot Designer)",
  AssembleScene: "场景工程组装 Agent (Scene Assembly)",
  PreviewRender: "低精度预览渲染 (PreviewRender)",
  Inspect: "分镜质检与审核 (Inspect)",
  FinalRender: "高质量序列帧渲染 (FinalRender)",
  Encode: "视频编码生成 (Encode MP4)",
  Verify: "成品视频校验 (Verify)",
  Publish: "分镜成品发布 (Publish)",
};

export const NodeInspectorModal: React.FC<NodeInspectorModalProps> = ({
  isOpen,
  node,
  onClose,
  onSaveAdjust,
  onConfirm,
}) => {
  const [activeTab, setActiveTab] = useState<"input" | "output">("output");
  const [outputDraft, setOutputDraft] = useState<string>("");
  const [jsonError, setJsonError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  useEffect(() => {
    if (node) {
      const formattedOutput = formatJSON(node.output || "{}");
      setOutputDraft(formattedOutput);
      setJsonError(null);
      if (node.status === "waiting_confirmation" || node.output) {
        setActiveTab("output");
      } else {
        setActiveTab("input");
      }
    }
  }, [node]);

  if (!isOpen || !node) return null;

  const nodeDisplayName = nodeNameMap[node.node_id] || node.node_id;

  const handleOutputChange = (val: string) => {
    setOutputDraft(val);
    try {
      if (val.trim()) {
        JSON.parse(val);
      }
      setJsonError(null);
    } catch (err: unknown) {
      setJsonError((err as Error).message);
    }
  };

  const handleFormatJson = () => {
    try {
      const parsed = JSON.parse(outputDraft);
      setOutputDraft(JSON.stringify(parsed, null, 2));
      setJsonError(null);
    } catch (err: unknown) {
      setJsonError("无法格式化：JSON 语法错误 (" + (err as Error).message + ")");
    }
  };

  const handleSaveOnly = async () => {
    if (jsonError) return;
    setIsSaving(true);
    try {
      await onSaveAdjust(outputDraft);
      onClose();
    } finally {
      setIsSaving(false);
    }
  };

  const handleSaveAndConfirm = async () => {
    if (jsonError) return;
    setIsSaving(true);
    try {
      await onConfirm(outputDraft);
      onClose();
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-fade-in">
      <div className="bg-surface-elevated border border-border rounded-2xl w-full max-w-3xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-border bg-surface-card">
          <div className="flex items-center gap-3">
            <h3 className="text-base font-semibold text-content">
              {nodeDisplayName}
            </h3>
            <Badge
              variant={
                node.status === "waiting_confirmation"
                  ? "warning"
                  : node.status === "succeeded"
                  ? "success"
                  : node.status === "running"
                  ? "info"
                  : "neutral"
              }
              className="text-xs"
            >
              {node.status === "waiting_confirmation"
                ? "⏸ 等待确认中"
                : node.status === "succeeded"
                ? "✔ 已完成"
                : node.status === "running"
                ? "▶ 运行中"
                : node.status}
            </Badge>
          </div>
          <button
            onClick={onClose}
            className="text-content-muted hover:text-content p-1 rounded-lg hover:bg-surface-background transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* Tab switcher */}
        <div className="flex items-center justify-between px-6 pt-3 border-b border-border bg-surface-background text-xs">
          <div className="flex gap-2">
            <button
              onClick={() => setActiveTab("input")}
              className={`flex items-center gap-1.5 px-3 py-2 border-b-2 font-medium transition-colors ${
                activeTab === "input"
                  ? "border-brand-primary text-brand-primary"
                  : "border-transparent text-content-muted hover:text-content"
              }`}
            >
              <Eye size={13} />
              <span>输入数据 (Input JSON)</span>
            </button>
            <button
              onClick={() => setActiveTab("output")}
              className={`flex items-center gap-1.5 px-3 py-2 border-b-2 font-medium transition-colors ${
                activeTab === "output"
                  ? "border-brand-primary text-brand-primary"
                  : "border-transparent text-content-muted hover:text-content"
              }`}
            >
              <Edit3 size={13} />
              <span>输出数据 (Output JSON)</span>
              {node.status === "waiting_confirmation" && (
                <span className="w-2 h-2 rounded-full bg-amber-400 animate-pulse ml-0.5" />
              )}
            </button>
          </div>

          {activeTab === "output" && (
            <button
              onClick={handleFormatJson}
              className="text-[11px] text-brand-primary hover:underline"
            >
              格式化对齐
            </button>
          )}
        </div>

        {/* Content body */}
        <div className="p-6 overflow-y-auto flex-1 font-mono text-xs">
          {activeTab === "input" ? (
            <div className="space-y-2">
              <div className="text-[11px] text-content-muted">
                当前节点接收的上游依赖数据与执行参数（只读）：
              </div>
              <pre className="p-4 rounded-xl bg-surface-background border border-border text-content-muted overflow-x-auto whitespace-pre-wrap leading-relaxed">
                {formatJSON(node.input || "{}")}
              </pre>
            </div>
          ) : (
            <div className="space-y-2">
              <div className="flex items-center justify-between text-[11px] text-content-muted">
                <span>
                  当前节点生成的数据契约（可按需编辑调整，调整后将作为下游基准）：
                </span>
              </div>
              <textarea
                value={outputDraft}
                onChange={(e) => handleOutputChange(e.target.value)}
                disabled={node.status !== "waiting_confirmation" && node.status !== "succeeded"}
                rows={14}
                className="w-full p-4 rounded-xl bg-surface-background border border-border focus:border-brand-primary focus:outline-none font-mono text-xs text-content leading-relaxed resize-none shadow-inner"
                placeholder="{ ... }"
              />
              {jsonError && (
                <div className="flex items-center gap-1.5 text-rose-400 text-[11px]">
                  <AlertCircle size={13} />
                  <span>{jsonError}</span>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="px-6 py-4 border-t border-border bg-surface-card flex items-center justify-between">
          <div className="text-[11px] text-content-icon">
            {node.status === "waiting_confirmation"
              ? "若数据符合预期可直接确认，若有偏差请修改输出 JSON 后保存继续。"
              : "该节点已执行完毕。"}
          </div>

          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={onClose}>
              取消
            </Button>

            {node.status === "waiting_confirmation" && (
              <>
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={handleSaveOnly}
                  disabled={isSaving || !!jsonError}
                >
                  <Save size={13} className="mr-1" />
                  仅保存调整
                </Button>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleSaveAndConfirm}
                  disabled={isSaving || !!jsonError}
                >
                  <Check size={13} className="mr-1" />
                  保存并确认下一步
                </Button>
              </>
            )}

            {node.status !== "waiting_confirmation" && (
              <Button variant="primary" size="sm" onClick={onClose}>
                确定
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

function formatJSON(raw: string): string {
  try {
    const parsed = typeof raw === "string" ? JSON.parse(raw) : raw;
    return JSON.stringify(parsed, null, 2);
  } catch {
    return raw || "{}";
  }
}
