import React, { useState } from "react";
import ReactMarkdown from "react-markdown";
import { User, Sparkles, CheckCircle2, AlertCircle } from "lucide-react";
import clsx from "clsx";
import { Message, TaskNodeView } from "#/api/types";
import { ThoughtBox } from "./thought-box";
import { Badge } from "#/components/ui/badge";
import { WorkflowProgressPanel } from "./workflow-progress-panel";
import { VideoDownloadCard } from "./video-download-card";
import { NodeInspectorModal } from "./node-inspector-modal";
import { useConversationStore } from "#/stores/conversation-store";

interface MessageItemProps {
  message: Message;
}

export const MessageItem: React.FC<MessageItemProps> = ({ message }) => {
  const isUser = message.role === "user";
  const [inspectingNode, setInspectingNode] = useState<TaskNodeView | null>(null);
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);

  const confirmNodeStep = useConversationStore((state) => state.confirmNodeStep);
  const adjustNodeStep = useConversationStore((state) => state.adjustNodeStep);
  const toggleAutoConfirm = useConversationStore((state) => state.toggleAutoConfirm);

  const handleInspectNode = (node: TaskNodeView) => {
    setInspectingNode(node);
    setIsInspectorOpen(true);
  };

  const handleSaveAdjust = async (outputJson: string) => {
    if (!inspectingNode) return;
    await adjustNodeStep(message.id, inspectingNode.node_id, outputJson);
  };

  const handleConfirmStep = async (adjustedOutput?: string) => {
    if (!inspectingNode) return;
    await confirmNodeStep(message.id, inspectingNode.node_id, adjustedOutput);
  };

  const handleDirectConfirm = async (nodeId: string) => {
    await confirmNodeStep(message.id, nodeId);
  };

  const hasVideoArtifact =
    message.artifacts?.some(
      (a) => a.type === "video" || a.name.endsWith(".mp4")
    ) || message.status === "done";

  return (
    <div
      className={clsx(
        "flex gap-3.5 py-4 px-2 md:px-6 transition-colors rounded-xl animate-fade-in",
        isUser ? "bg-surface-card/50" : "bg-transparent"
      )}
    >
      {/* Avatar */}
      <div
        className={clsx(
          "h-7 w-7 rounded-lg shrink-0 flex items-center justify-center text-xs font-semibold select-none",
          isUser
            ? "bg-surface-background text-content border border-border"
            : "bg-brand-primary/20 text-brand-primary border border-brand-primary/30"
        )}
      >
        {isUser ? <User size={15} /> : <Sparkles size={15} />}
      </div>

      {/* Main Content Area */}
      <div className="flex-1 min-w-0 space-y-2">
        <div className="flex items-center gap-2 select-none">
          <span className="text-xs font-semibold text-content-muted">
            {isUser ? "You" : "Shot Preview Assistant"}
          </span>
          <span className="text-[10px] text-content-icon">
            {new Date(message.timestamp).toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
            })}
          </span>

          {message.taskId && (
            <Badge variant="success" className="text-[10px] py-0 px-1.5 ml-1">
              <CheckCircle2 size={10} />
              RPC Task: {message.taskId}
            </Badge>
          )}

          {message.status === "error" && (
            <Badge variant="warning" className="text-[10px] py-0 px-1.5 ml-1">
              <AlertCircle size={10} />
              调用异常
            </Badge>
          )}
        </div>

        {/* Workflow Progress Panel for Agent DAG (Step Confirmation & Progress) */}
        {!isUser && message.nodes && message.nodes.length > 0 && (
          <WorkflowProgressPanel
            nodes={message.nodes}
            waitingNode={message.waitingNode}
            autoConfirm={message.autoConfirm}
            onInspectNode={handleInspectNode}
            onConfirmStep={handleDirectConfirm}
            onToggleAutoConfirm={() => toggleAutoConfirm(message.id)}
          />
        )}

        {/* Thought Chain Accordion if present */}
        {(message.thoughts || message.isThinking) && (
          <ThoughtBox
            thoughts={message.thoughts || []}
            isThinking={message.isThinking}
          />
        )}

        {/* Message Body with Markdown */}
        <div className="markdown-body prose prose-invert prose-sm max-w-none text-content text-sm leading-relaxed break-words">
          {message.content ? (
            <ReactMarkdown>{message.content}</ReactMarkdown>
          ) : message.isThinking ? (
            <div className="text-xs text-content-muted italic">
              实时执行流已连接，正在处理各节点计算任务...
            </div>
          ) : null}
        </div>

        {/* Final Video Download Card (strictly download link, no preview player) */}
        {!isUser && message.taskId && hasVideoArtifact && (
          <VideoDownloadCard
            taskId={message.taskId}
            artifacts={message.artifacts}
          />
        )}

        {/* Node Inspector & Adjustment Modal */}
        <NodeInspectorModal
          isOpen={isInspectorOpen}
          node={inspectingNode}
          onClose={() => setIsInspectorOpen(false)}
          onSaveAdjust={handleSaveAdjust}
          onConfirm={handleConfirmStep}
        />
      </div>
    </div>
  );
};
