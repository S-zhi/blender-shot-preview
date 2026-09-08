import React, { useState } from "react";
import { ChevronDown, ChevronRight, Loader2, Sparkles } from "lucide-react";

interface ThoughtBoxProps {
  thoughts: string[];
  isThinking?: boolean;
}

export const ThoughtBox: React.FC<ThoughtBoxProps> = ({
  thoughts,
  isThinking = false,
}) => {
  const [isExpanded, setIsExpanded] = useState(true);

  if (thoughts.length === 0 && !isThinking) return null;

  return (
    <div className="my-2 border border-border bg-surface-card rounded-xl overflow-hidden text-xs transition-all">
      {/* Header Bar */}
      <button
        type="button"
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between px-3.5 py-2 hover:bg-surface-elevated transition-colors text-left"
      >
        <div className="flex items-center gap-2 text-content-muted font-medium">
          {isThinking ? (
            <Loader2 size={13} className="animate-spin text-brand-primary" />
          ) : (
            <Sparkles size={13} className="text-amber-400" />
          )}
          <span>
            {isThinking
              ? "Agent 正在规划镜头轨迹与分析场景..."
              : `思考与场景分析过程 (${thoughts.length} 步)`}
          </span>
        </div>

        <div className="text-content-icon">
          {isExpanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        </div>
      </button>

      {/* Expanded Step List */}
      {isExpanded && (
        <div className="px-3.5 pb-3 pt-1 space-y-1.5 border-t border-border font-mono text-[11px] text-content-muted">
          {thoughts.map((step, idx) => (
            <div key={idx} className="flex items-start gap-2">
              <span className="text-brand-primary/80 shrink-0">[{idx + 1}]</span>
              <span className="leading-relaxed text-content-muted">{step}</span>
            </div>
          ))}
          {isThinking && (
            <div className="flex items-center gap-2 text-brand-primary italic animate-pulse">
              <span>...</span>
              <span>执行下一步计算中</span>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
