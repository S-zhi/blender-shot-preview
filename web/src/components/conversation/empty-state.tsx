import React from "react";
import { Film, Video, Camera, Sparkles, Wand2 } from "lucide-react";

interface EmptyStateProps {
  onSelectPrompt: (prompt: string) => void;
}

export const EmptyState: React.FC<EmptyStateProps> = ({ onSelectPrompt }) => {
  const suggestions = [
    {
      title: "赛博朋克雨夜街道推镜头",
      desc: "俯视 45° 广角推至平视特写，展现霓虹灯反光与雨滴折射",
      icon: Camera,
      prompt: "为赛博朋克风格雨夜街道创建由远及近的特写推镜头，重点展现路面霓虹反光与积水波纹",
    },
    {
      title: "工业产品 360° 环绕展示",
      desc: "围绕中心机械结构做圆周匀速运转，维持精准对焦",
      icon: Video,
      prompt: "针对中心机械结构生成 360 度圆周匀速环绕镜头，摄影机保持定焦并带有微妙俯仰角度",
    },
    {
      title: "自然风光无人机大景深飞越",
      desc: "穿过峡谷树冠的大范围快速航拍，模拟无人机运动轨迹",
      icon: Film,
      prompt: "模拟 FPV 无人机穿过山谷与森林树冠的快速穿梭镜头，设定动态运动模糊与加速曲线",
    },
    {
      title: "人物特写焦点平移 (Rack Focus)",
      desc: "景深由前景前景物快速平滑过渡到背景人物面部",
      icon: Wand2,
      prompt: "设计一个平滑变焦焦点转移镜头：先对焦于前景杯沿水珠，0.8秒后平滑对焦至后景主角眼神",
    },
  ];

  return (
    <div className="flex-1 flex flex-col items-center justify-center p-6 text-center select-none max-w-2xl mx-auto">
      <div className="h-14 w-14 rounded-2xl bg-brand-primary/10 border border-brand-primary/20 text-brand-primary flex items-center justify-center mb-5 shadow-lg shadow-brand-primary/5">
        <Sparkles size={28} />
      </div>

      <h1 className="text-2xl font-bold tracking-tight text-white mb-2">
        Blender 分镜智能预览工作台
      </h1>
      <p className="text-sm text-gray-400 mb-8 max-w-md leading-relaxed">
        基于 OpenHands 架构设计。描述您期望的镜头轨迹、光影构图与机位运镜，系统将调度后台任务生成分镜预览。
      </p>

      {/* Suggestion Prompts Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 w-full text-left">
        {suggestions.map((item, idx) => {
          const Icon = item.icon;
          return (
            <button
              key={idx}
              onClick={() => onSelectPrompt(item.prompt)}
              className="p-3.5 rounded-xl bg-surface-300/70 hover:bg-surface-200/90 border border-border-subtle hover:border-border-default transition-all duration-150 group cursor-pointer text-left"
            >
              <div className="flex items-center gap-2.5 mb-1.5 text-gray-200 group-hover:text-brand-primary font-medium text-xs">
                <Icon size={15} className="text-brand-primary/80" />
                <span>{item.title}</span>
              </div>
              <p className="text-[11px] text-gray-400 line-clamp-2 leading-normal">
                {item.desc}
              </p>
            </button>
          );
        })}
      </div>
    </div>
  );
};
