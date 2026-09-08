import React, { useState } from "react";
import { Download, Film, Check, Copy } from "lucide-react";
import { ShotPreviewService } from "#/api/shot-preview-service";
import { TaskArtifactView } from "#/api/types";
import { Button } from "#/components/ui/button";

interface VideoDownloadCardProps {
  taskId: string;
  artifacts?: TaskArtifactView[];
}

export const VideoDownloadCard: React.FC<VideoDownloadCardProps> = ({
  taskId,
  artifacts,
}) => {
  const [copied, setCopied] = useState(false);

  const videoArtifact = artifacts?.find(
    (a) => a.type === "video" || a.name.endsWith(".mp4")
  );
  const downloadUrl =
    videoArtifact?.uri || ShotPreviewService.getArtifactDownloadUrl(taskId);

  const handleCopyLink = () => {
    const fullUrl = `${window.location.origin}${downloadUrl}`;
    navigator.clipboard.writeText(fullUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="my-3 p-4 rounded-xl bg-surface-card border border-brand-primary/40 bg-gradient-to-r from-brand-primary/5 to-surface-card shadow-lg animate-fade-in">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-brand-primary/20 text-brand-primary border border-brand-primary/30 flex items-center justify-center shrink-0">
            <Film size={20} />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <h4 className="text-sm font-semibold text-content">
                分镜视频生成完毕
              </h4>
              <span className="text-[10px] bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 px-1.5 py-0.5 rounded-md font-mono">
                MP4
              </span>
            </div>
            <p className="text-xs text-content-muted mt-0.5">
              已完成所有 Agent 编排规划、Eevee 渲染与 H.264 编码
            </p>
            <p className="text-[11px] text-content-icon mt-1 italic">
              提示：已按要求直接提供下载链接，暂不提供在线预览播放。
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2 w-full sm:w-auto">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCopyLink}
            className="text-xs"
          >
            {copied ? (
              <>
                <Check size={13} className="mr-1 text-emerald-400" />
                已复制
              </>
            ) : (
              <>
                <Copy size={13} className="mr-1" />
                复制链接
              </>
            )}
          </Button>

          <a
            href={downloadUrl}
            download="shot-preview.mp4"
            className="flex-1 sm:flex-initial inline-flex items-center justify-center gap-1.5 px-4 py-2 rounded-lg bg-brand-primary hover:bg-brand-primary/90 text-white text-xs font-semibold shadow-md shadow-brand-primary/25 transition-all hover:scale-[1.02] active:scale-[0.98]"
          >
            <Download size={14} />
            <span>下载分镜视频</span>
          </a>
        </div>
      </div>
    </div>
  );
};
