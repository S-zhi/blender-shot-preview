import React, { useState } from "react";
import {
  Boxes,
  Camera,
  Layers,
  Film,
  UploadCloud,
  Search,
  HardDrive,
  FileCode,
  Tag,
  ExternalLink,
  Trash2,
  CheckCircle2,
  X,
} from "lucide-react";
import clsx from "clsx";

interface Asset {
  id: string;
  name: string;
  type: "model" | "preset" | "material" | "animation";
  typeName: string;
  format: string;
  size: string;
  tags: string[];
  updatedAt: string;
  status: "ready" | "processing";
  desc: string;
}

export const AssetDashboard: React.FC = () => {
  const [activeType, setActiveType] = useState<string>("all");
  const [search, setSearch] = useState("");
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [uploadName, setUploadName] = useState("");
  const [uploadType, setUploadType] = useState("model");

  const [assets, setAssets] = useState<Asset[]>([
    {
      id: "asset_001",
      name: "cyberpunk_street_night.blend",
      type: "model",
      typeName: "3D 模型",
      format: "BLEND",
      size: "348 MB",
      tags: ["赛博朋克", "雨夜", "SSR光影"],
      updatedAt: "10分钟前",
      status: "ready",
      desc: "包含高细节湿漉路面贴图、霓虹招牌顶点着色与体积雾灯光设置",
    },
    {
      id: "asset_002",
      name: "dolly_zoom_rack_focus_35to85.json",
      type: "preset",
      typeName: "镜头预设",
      format: "JSON",
      size: "12 KB",
      tags: ["推拉变焦", "平移对焦", "180帧"],
      updatedAt: "1小时前",
      status: "ready",
      desc: "Blender 摄像机平滑变焦焦点转移曲线数据，定焦前景水珠到背景人脸",
    },
    {
      id: "asset_003",
      name: "mechanical_watch_center.blend",
      type: "model",
      typeName: "3D 模型",
      format: "BLEND",
      size: "142 MB",
      tags: ["工业产品", "精密零件", "金属材质"],
      updatedAt: "昨天",
      status: "ready",
      desc: "高精度机械机芯齿轮结构模型，带独立中心对焦旋转轴",
    },
    {
      id: "asset_004",
      name: "orbit_360_smooth_yaw.py",
      type: "preset",
      typeName: "镜头预设",
      format: "PY",
      size: "8.4 KB",
      tags: ["圆周环绕", "DampedTrack", "恒速"],
      updatedAt: "2天前",
      status: "ready",
      desc: "Python 脚本驱动的 360 度圆周运镜轨迹生成算法",
    },
    {
      id: "asset_005",
      name: "tokyo_night_rain_4k.hdr",
      type: "material",
      typeName: "材质贴图",
      format: "HDR",
      size: "68 MB",
      tags: ["HDRI", "夜景环境光", "4K"],
      updatedAt: "3天前",
      status: "ready",
      desc: "32-bit 浮点高动态范围环境贴图，提供高真实度反射与环境照明",
    },
    {
      id: "asset_006",
      name: "fpv_drone_canyon_run.abc",
      type: "animation",
      typeName: "动画序列",
      format: "ABC",
      size: "215 MB",
      tags: ["FPV无人机", "大景深", "动态模糊"],
      updatedAt: "5天前",
      status: "ready",
      desc: "Alembic 格式无人机快速俯冲与穿越山谷相机运动缓存序列",
    },
  ]);

  const filteredAssets = assets.filter((item) => {
    const matchesType = activeType === "all" || item.type === activeType;
    const matchesSearch =
      item.name.toLowerCase().includes(search.toLowerCase()) ||
      item.tags.some((t) => t.toLowerCase().includes(search.toLowerCase()));
    return matchesType && matchesSearch;
  });

  const handleDelete = (id: string) => {
    setAssets((prev) => prev.filter((a) => a.id !== id));
  };

  const handleUploadSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!uploadName.trim()) return;
    const newAsset: Asset = {
      id: `asset_${Date.now()}`,
      name: uploadName.endsWith(".blend") ? uploadName : `${uploadName}.blend`,
      type: uploadType as Asset["type"],
      typeName:
        uploadType === "model"
          ? "3D 模型"
          : uploadType === "preset"
          ? "镜头预设"
          : uploadType === "material"
          ? "材质贴图"
          : "动画序列",
      format: uploadType === "model" ? "BLEND" : uploadType === "preset" ? "JSON" : "PNG",
      size: "54 MB",
      tags: ["新导入", "分镜素材"],
      updatedAt: "刚刚",
      status: "ready",
      desc: "由前端看板注册的新建分镜工程资产",
    };
    setAssets([newAsset, ...assets]);
    setIsUploadModalOpen(false);
    setUploadName("");
  };

  return (
    <div className="flex-1 h-full overflow-y-auto bg-[#090b0e] text-content p-6 lg:p-10 font-sans select-none relative">
      {/* Subtle Ambient Radial Lighting */}
      <div className="absolute top-0 right-1/3 w-[600px] h-[280px] bg-brand-primary/5 blur-[120px] pointer-events-none -z-0" />

      <div className="max-w-6xl mx-auto space-y-7 z-10 relative">
        {/* 1. Header & Actions */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#171a22] pb-6">
          <div>
            <div className="text-[12px] text-[#8490a5] mb-1 font-medium tracking-wide">
              Blender Shot Preview
            </div>
            <h1 className="text-2xl lg:text-[28px] font-bold tracking-tight text-white/95 mb-1.5 flex items-center gap-2.5">
              <span>资产管理看板</span>
              <span className="text-xs font-mono font-normal px-2 py-0.5 rounded-full bg-[#1c212c] text-[#8490a5] border border-[#28303f]">
                AssetService RPC v0.1
              </span>
            </h1>
            <p className="text-xs text-[#717b8c]">
              管理分镜三维模型、相机轨迹预设、HDRI 环境贴图与动画序列缓存。
            </p>
          </div>

          <div className="flex items-center gap-2.5 shrink-0">
            <button
              onClick={() => setIsUploadModalOpen(true)}
              className="flex items-center gap-1.5 px-4 py-1.5 rounded-lg bg-white text-black hover:bg-white/90 text-xs font-semibold transition-all shadow-md active:scale-98 cursor-pointer"
            >
              <UploadCloud size={15} />
              <span>导入/上传资产</span>
            </button>
          </div>
        </div>

        {/* 2. 4 Metric KPI Cards for Assets */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {/* Card 1: Total Assets */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>总资产数 (Total Assets)</span>
              <Boxes size={15} className="text-[#646e80]" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-white/95">
                {assets.length}
              </div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">
                {assets.filter((a) => a.status === "ready").length} 处于可用就绪态
              </div>
            </div>
          </div>

          {/* Card 2: 3D Models */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>3D 场景与模型</span>
              <Layers size={15} className="text-brand-primary" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-white/95">
                {assets.filter((a) => a.type === "model").length}
              </div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">.blend / .fbx 格式</div>
            </div>
          </div>

          {/* Card 3: Shot Presets */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>运镜与轨迹预设</span>
              <Camera size={15} className="text-[#cfb755]" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-[#cfb755]">
                {assets.filter((a) => a.type === "preset").length}
              </div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">
                支持一键注入 Blender 相机
              </div>
            </div>
          </div>

          {/* Card 4: Storage */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>存储池占用 (Storage)</span>
              <HardDrive size={15} className="text-[#646e80]" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-white/95">841 MB</div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">本地 RPC 缓存驱动</div>
            </div>
          </div>
        </div>

        {/* 3. Category Filter Tabs & Search Toolbar */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 pt-2">
          {/* Category Filter Pills */}
          <div className="flex items-center gap-1.5 overflow-x-auto pb-1 sm:pb-0">
            {[
              { key: "all", label: "全部", icon: Boxes },
              { key: "model", label: "3D 模型", icon: Layers },
              { key: "preset", label: "镜头预设", icon: Camera },
              { key: "material", label: "材质贴图", icon: FileCode },
              { key: "animation", label: "动画序列", icon: Film },
            ].map((tab) => {
              const Icon = tab.icon;
              const isSelected = activeType === tab.key;
              return (
                <button
                  key={tab.key}
                  onClick={() => setActiveType(tab.key)}
                  className={clsx(
                    "flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all cursor-pointer shrink-0",
                    isSelected
                      ? "bg-white text-black font-semibold shadow-xs"
                      : "bg-[#14171d] text-[#8490a5] hover:text-white hover:bg-[#1c212c] border border-[#202532]"
                  )}
                >
                  <Icon size={13} />
                  <span>{tab.label}</span>
                </button>
              );
            })}
          </div>

          {/* Search Box */}
          <div className="relative w-full sm:w-64">
            <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#646e80]" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索资产名称或标签..."
              className="w-full h-8 pl-9 pr-3 rounded-lg bg-[#14171d] border border-[#202532] text-xs text-white placeholder-[#5a6372] focus:outline-none focus:border-[#384254]"
            />
          </div>
        </div>

        {/* 4. Asset Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pb-16">
          {filteredAssets.map((asset) => (
            <div
              key={asset.id}
              className="rounded-2xl bg-[#14171d]/85 border border-[#202532] hover:border-[#333d4e] p-5 flex flex-col justify-between transition-all duration-200 shadow-sm hover:shadow-md group"
            >
              <div>
                {/* Card Top: Format Badge + Status */}
                <div className="flex items-center justify-between mb-3">
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md bg-[#1a202c] text-[11px] font-mono font-semibold text-brand-primary border border-brand-primary/20">
                    {asset.format} · {asset.typeName}
                  </span>
                  <div className="flex items-center gap-1.5 text-xs text-[#8490a5]">
                    <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400">
                      <CheckCircle2 size={12} />
                      就绪
                    </span>
                    <button
                      onClick={() => handleDelete(asset.id)}
                      title="删除资产"
                      className="p-1 text-[#646e80] hover:text-red-400 hover:bg-white/5 rounded transition-colors"
                    >
                      <Trash2 size={13} />
                    </button>
                  </div>
                </div>

                {/* Asset Title */}
                <h3 className="text-sm font-semibold text-white/95 mb-1.5 truncate group-hover:text-white transition-colors">
                  {asset.name}
                </h3>

                {/* Description */}
                <p className="text-xs text-[#788497] leading-relaxed line-clamp-2 mb-3.5">
                  {asset.desc}
                </p>

                {/* Tags row */}
                <div className="flex flex-wrap gap-1.5 mb-4">
                  {asset.tags.map((tag, idx) => (
                    <span
                      key={idx}
                      className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-[#0e1014] text-[10px] text-[#8490a5] border border-[#1d222b]"
                    >
                      <Tag size={10} className="text-[#5a6372]" />
                      <span>{tag}</span>
                    </span>
                  ))}
                </div>
              </div>

              {/* Card Bottom Meta Info */}
              <div className="pt-3 border-t border-[#1d222b] flex items-center justify-between text-xs text-[#646e80]">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-white/80">{asset.size}</span>
                  <span>·</span>
                  <span>{asset.updatedAt}</span>
                </div>

                <button
                  onClick={() => alert(`已在当前分镜工程中装载资产: ${asset.name}`)}
                  className="flex items-center gap-1 text-[11px] text-brand-primary hover:underline"
                >
                  <span>装入分镜</span>
                  <ExternalLink size={11} />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 5. Upload/Import Asset Modal */}
      {isUploadModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xs p-4">
          <div className="w-full max-w-md rounded-2xl bg-[#14171d] border border-[#282f3d] p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-[#212733] pb-3">
              <h3 className="text-base font-semibold text-white">导入分镜素材资产</h3>
              <button
                onClick={() => setIsUploadModalOpen(false)}
                className="p-1 text-[#8490a5] hover:text-white rounded"
              >
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleUploadSubmit} className="space-y-4 text-xs">
              <div>
                <label className="block text-[#8490a5] mb-1 font-medium">资产名称</label>
                <input
                  type="text"
                  value={uploadName}
                  onChange={(e) => setUploadName(e.target.value)}
                  placeholder="例如: camera_dolly_cyberpunk.blend"
                  required
                  className="w-full h-8 px-3 rounded-lg bg-[#0e1014] border border-[#212733] text-white focus:outline-none focus:border-brand-primary"
                />
              </div>

              <div>
                <label className="block text-[#8490a5] mb-1 font-medium">资产类别 (AssetType)</label>
                <select
                  value={uploadType}
                  onChange={(e) => setUploadType(e.target.value)}
                  className="w-full h-8 px-3 rounded-lg bg-[#0e1014] border border-[#212733] text-white focus:outline-none focus:border-brand-primary"
                >
                  <option value="model">3D 模型 (.blend, .fbx, .obj)</option>
                  <option value="preset">镜头预设 (.json, .py)</option>
                  <option value="material">材质与贴图 (.png, .exr, .hdr)</option>
                  <option value="animation">动画序列 (.abc, .bvh)</option>
                </select>
              </div>

              <div className="p-3 rounded-lg bg-[#0e1014] border border-[#212733] text-[#717b8c] text-[11px] leading-relaxed">
                提示：本地选中的三维文件将通过 Kitex RPC 接口同步注册至后端 Blender Asset Pipeline。
              </div>

              <div className="flex items-center justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsUploadModalOpen(false)}
                  className="px-3 py-1.5 rounded-lg text-[#8490a5] hover:text-white"
                >
                  取消
                </button>
                <button
                  type="submit"
                  className="px-4 py-1.5 rounded-lg bg-white text-black font-semibold hover:bg-gray-200"
                >
                  确认导入
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
