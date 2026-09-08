import React from "react";
import ReactMarkdown from "react-markdown";
import { User, Sparkles, CheckCircle2, AlertCircle } from "lucide-react";
import clsx from "clsx";
import { Message } from "#/api/types";
import { ThoughtBox } from "./thought-box";
import { Badge } from "#/components/ui/badge";

interface MessageItemProps {
  message: Message;
}

export const MessageItem: React.FC<MessageItemProps> = ({ message }) => {
  const isUser = message.role === "user";

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
      <div className="flex-1 min-w-0 space-y-1.5">
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
            <div className="text-xs text-content-muted italic">正在生成预览指令与分镜脚本...</div>
          ) : null}
        </div>
      </div>
    </div>
  );
};
