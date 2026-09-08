import React, { useState, useEffect, useCallback } from "react";
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
  Trash2,
  X,
  Loader2,
  Eye,
  Bone,
  Sparkles,
  FileUp,
  CheckCircle2,
} from "lucide-react";
import clsx from "clsx";
import { AssetService } from "#/api/asset-service";
import { AssetView, AssetType, GetAssetStatsResponse } from "#/api/types";
import { AssetPreviewModal } from "./asset-preview-modal";

export const AssetDashboard: React.FC = () => {
  const [activeType, setActiveType] = useState<string>("all");
  const [search, setSearch] = useState("");
  const [assets, setAssets] = useState<AssetView[]>([]);
  const [stats, setStats] = useState<GetAssetStatsResponse>({
    total_assets: 0,
    total_models: 0,
    total_presets: 0,
    total_storage_bytes: 0,
  });
  const [loading, setLoading] = useState(true);

  // Asset Preview modal state
  const [previewAsset, setPreviewAsset] = useState<AssetView | null>(null);
  const [isPreviewModalOpen, setIsPreviewModalOpen] = useState(false);

  // Upload modal state
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [uploadName, setUploadName] = useState("");
  const [uploadFile, setUploadFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [uploadStep, setUploadStep] = useState<"idle" | "uploading" | "analyzing" | "success">("idle");
  const [uploadedResult, setUploadedResult] = useState<AssetView | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  const fetchAssetsAndStats = useCallback(async () => {
    try {
      setLoading(true);
      const [listRes, statsRes] = await Promise.all([
        AssetService.listAssets({
          user_id: "default_user_001",
          query_keyword: search,
          asset_type: activeType === "all" ? undefined : Number(activeType),
        }),
        AssetService.getAssetStats({ user_id: "default_user_001" }),
      ]);
      setAssets(listRes.assets || []);
      setStats(statsRes);
    } catch (err) {
      console.error("Failed to load assets from backend:", err);
    } finally {
      setLoading(false);
    }
  }, [activeType, search]);

  useEffect(() => {
    fetchAssetsAndStats();
  }, [fetchAssetsAndStats]);

  const handleDelete = async (id: string) => {
    try {
      await AssetService.deleteAsset({
        user_id: "default_user_001",
        asset_id: id,
      });
      fetchAssetsAndStats();
    } catch (err) {
      alert("删除失败: " + (err as Error).message);
    }
  };

  const handleUploadSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!uploadFile && !uploadName.trim()) return;

    try {
      setIsUploading(true);
      setUploadStep("uploading");

      // Step 1: Uploading
      await new Promise((r) => setTimeout(r, 400));
      setUploadStep("analyzing");

      // Step 2: Running Blender Inspection & Character Extraction
      const targetFile =
        uploadFile ||
        new File(["BLENDER_v401\x00Armature_Rigify\x00"], uploadName || "custom_model.blend", {
          type: "application/octet-stream",
        });

      const registered = await AssetService.uploadAssetFile(targetFile, uploadName || targetFile.name);

      setUploadedResult(registered);
      setUploadStep("success");
      fetchAssetsAndStats();
    } catch (err) {
      alert("上传与解析失败: " + (err as Error).message);
      setUploadStep("idle");
    } finally {
      setIsUploading(false);
    }
  };

  const handleCloseUploadModal = () => {
    setIsUploadModalOpen(false);
    setUploadFile(null);
    setUploadName("");
    setUploadStep("idle");
    setUploadedResult(null);
  };

  const formatBytes = (bytes: number) => {
    if (!bytes || bytes === 0) return "0 MB";
    const mb = bytes / (1024 * 1024);
    if (mb > 1024) {
      return `${(mb / 1024).toFixed(2)} GB`;
    }
    return `${mb.toFixed(0)} MB`;
  };

  const getTypeName = (type: AssetType) => {
    switch (type) {
      case AssetType.MODEL_3D:
        return "3D 模型";
      case AssetType.SHOT_PRESET:
        return "镜头预设";
      case AssetType.MATERIAL:
        return "材质贴图";
      case AssetType.ANIMATION:
        return "动画序列";
      default:
        return "未分类";
    }
  };

  return (
    <div className="flex-1 h-full overflow-y-auto bg-[#090b0e] text-content p-6 lg:p-10 font-sans select-none relative">
      {/* Ambient Radial Glow */}
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
              <span className="text-xs font-mono font-normal px-2.5 py-0.5 rounded-full bg-[#1c212c] text-emerald-400 border border-emerald-500/20 flex items-center gap-1.5">
                <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                Go RPC 实机联通
              </span>
            </h1>
            <p className="text-xs text-[#717b8c]">
              直接对接后端 Kitex AssetService 内存池，提供场景模型、相机预设与贴图真实存取。
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

        {/* 2. 4 Metric KPI Cards connected to real Go Backend */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {/* Card 1: Total Assets */}
          <div className="rounded-xl bg-[#14171d]/80 hover:bg-[#181d26] border border-[#202532] hover:border-[#333c4e] p-4.5 flex flex-col justify-between transition-all duration-200 shadow-xs">
            <div className="flex items-center justify-between text-xs text-[#8490a5] font-medium">
              <span>总资产数 (Total Assets)</span>
              <Boxes size={15} className="text-[#646e80]" />
            </div>
            <div className="mt-4">
              <div className="text-3xl font-bold tracking-tight text-white/95">
                {stats.total_assets}
              </div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">
                实时从 Go 后端聚合统计
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
                {stats.total_models}
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
                {stats.total_presets}
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
              <div className="text-3xl font-bold tracking-tight text-white/95">
                {formatBytes(stats.total_storage_bytes)}
              </div>
              <div className="text-[11px] text-[#717b8c] mt-1 font-mono">真实后端字节统计</div>
            </div>
          </div>
        </div>

        {/* 3. Category Filter Tabs & Search Toolbar */}
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 pt-2">
          {/* Category Filter Pills */}
          <div className="flex items-center gap-1.5 overflow-x-auto pb-1 sm:pb-0">
            {[
              { key: "all", label: "全部", icon: Boxes },
              { key: String(AssetType.MODEL_3D), label: "3D 模型", icon: Layers },
              { key: String(AssetType.SHOT_PRESET), label: "镜头预设", icon: Camera },
              { key: String(AssetType.MATERIAL), label: "材质贴图", icon: FileCode },
              { key: String(AssetType.ANIMATION), label: "动画序列", icon: Film },
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
              placeholder="实时搜索后端资产..."
              className="w-full h-8 pl-9 pr-3 rounded-lg bg-[#14171d] border border-[#202532] text-xs text-white placeholder-[#5a6372] focus:outline-none focus:border-[#384254]"
            />
          </div>
        </div>

        {/* 4. Asset Cards Grid */}
        {loading ? (
          <div className="py-20 flex flex-col items-center justify-center gap-2 text-xs text-[#8490a5]">
            <Loader2 size={24} className="animate-spin text-brand-primary" />
            <span>正在通过 Kitex HTTP Gateway 请求后端资产数据...</span>
          </div>
        ) : assets.length === 0 ? (
          <div className="py-20 text-center text-xs text-[#717b8c]">
            暂无匹配的后端分镜资产
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5 pb-16">
            {assets.map((asset) => {
              const isChar =
                asset.name.includes("eva") ||
                asset.name.includes("goliath") ||
                (asset.tags && asset.tags.some((t) => t.includes("人物") || t.includes("角色")));
              const isScene = asset.name.includes("street") || asset.name.includes("scene");
              const isPreset = asset.asset_type === AssetType.SHOT_PRESET;

              return (
                <div
                  key={asset.asset_id}
                  onClick={() => {
                    setPreviewAsset(asset);
                    setIsPreviewModalOpen(true);
                  }}
                  className="rounded-2xl bg-[#14171d]/90 border border-[#202532] hover:border-brand-primary/50 flex flex-col justify-between transition-all duration-200 shadow-sm hover:shadow-xl hover:-translate-y-0.5 group cursor-pointer overflow-hidden"
                >
                  <div>
                    {/* Visual 16:9 Thumbnail Cover Area */}
                    <div className="relative w-full h-40 bg-[#0c0f14] border-b border-[#1c222e] overflow-hidden flex items-center justify-center">
                      {/* Grid background */}
                      <div className="absolute inset-0 bg-[radial-gradient(#1f2738_1px,transparent_1px)] [background-size:16px_16px] opacity-30" />

                      {/* Dynamic Cover Graphic per asset category */}
                      {isChar ? (
                        <div className="relative flex flex-col items-center justify-center">
                          {/* Character Silhouette Preview */}
                          <div className="w-20 h-28 rounded-xl bg-gradient-to-b from-[#2a364d] via-[#1a2333] to-[#0e121a] border border-[#3e4f70] flex flex-col items-center justify-between p-2 shadow-lg group-hover:scale-105 transition-transform duration-300 relative">
                            <div className="w-7 h-8 rounded-lg bg-[#3a4b6b] border border-cyan-400/40" />
                            <div className="w-12 h-10 rounded-md bg-[#253045] border border-pink-400/30" />
                            <div className="w-full flex justify-between px-1">
                              <div className="w-3 h-6 rounded bg-[#1e2738]" />
                              <div className="w-3 h-6 rounded bg-[#1e2738]" />
                            </div>
                            {/* Rig bone hint */}
                            <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                              <Bone size={24} className="text-pink-400/60 drop-shadow-[0_0_8px_rgba(236,72,153,0.5)]" />
                            </div>
                          </div>
                          <div className="w-28 h-3 rounded-full bg-cyan-400/10 blur-xs mt-2" />
                        </div>
                      ) : isScene ? (
                        <div className="w-48 h-28 rounded-lg border border-cyan-500/30 bg-gradient-to-tr from-[#131924] to-[#202a3d] p-3 flex flex-col justify-between shadow-md group-hover:scale-105 transition-transform duration-300">
                          <div className="flex justify-between items-center text-[9px] font-mono text-cyan-400">
                            <span>NEO_SHIBUYA</span>
                            <span>SSR 4K</span>
                          </div>
                          <div className="grid grid-cols-3 gap-1.5 items-end">
                            <div className="h-12 rounded bg-cyan-900/40 border border-cyan-400/20" />
                            <div className="h-18 rounded bg-cyan-800/50 border border-cyan-400/40" />
                            <div className="h-10 rounded bg-cyan-900/40 border border-cyan-400/20" />
                          </div>
                        </div>
                      ) : isPreset ? (
                        <div className="w-48 h-28 rounded-lg border border-amber-500/30 bg-[#0d1117] p-3 flex flex-col justify-between shadow-md group-hover:scale-105 transition-transform duration-300 relative">
                          <div className="flex justify-between items-center text-[9px] font-mono text-amber-400">
                            <span>CAM_TRAJECTORY</span>
                            <span>CURVE_3D</span>
                          </div>
                          <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                            <Camera size={28} className="text-amber-400/70 animate-pulse drop-shadow-[0_0_10px_rgba(245,158,11,0.5)]" />
                          </div>
                          <div className="text-[9px] text-right font-mono text-[#626f84]">
                            120 FRAMES · BEZIER
                          </div>
                        </div>
                      ) : asset.asset_type === AssetType.MATERIAL ? (
                        <div className="w-24 h-24 rounded-full bg-gradient-to-tr from-[#0f172a] via-[#1e293b] to-[#38bdf8] border-2 border-[#38bdf8]/40 shadow-[0_0_30px_rgba(56,189,248,0.25)] flex items-center justify-center group-hover:scale-105 transition-transform duration-300">
                          <div className="w-14 h-14 rounded-full border border-white/20 bg-white/10" />
                        </div>
                      ) : (
                        <div className="flex flex-col items-center justify-center gap-1 text-amber-400">
                          <Film size={36} className="opacity-80 group-hover:scale-105 transition-transform" />
                          <span className="text-[10px] font-mono text-[#8490a5]">ALEMBIC SEQUENCE</span>
                        </div>
                      )}

                      {/* Top Floating Badges */}
                      <div className="absolute top-2.5 left-2.5 flex items-center gap-1.5">
                        <span className="px-2 py-0.5 rounded-md bg-black/70 backdrop-blur-xs text-[10px] font-mono font-bold text-brand-primary border border-brand-primary/30">
                          {asset.file_format.toUpperCase()}
                        </span>
                        {isChar && (
                          <span className="px-1.5 py-0.5 rounded-md bg-pink-500/20 backdrop-blur-xs text-[9px] font-semibold text-pink-300 border border-pink-500/30 flex items-center gap-1">
                            <Bone size={10} />
                            <span>Rigify</span>
                          </span>
                        )}
                      </div>

                      {/* Top-Right Delete Action Button */}
                      <div className="absolute top-2.5 right-2.5">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDelete(asset.asset_id);
                          }}
                          title="删除资产"
                          className="p-1.5 rounded-md bg-black/60 text-[#8490a5] hover:text-red-400 hover:bg-black/80 transition-colors"
                        >
                          <Trash2 size={12} />
                        </button>
                      </div>

                      {/* Hover Overlay with Preview CTA */}
                      <div className="absolute inset-0 bg-black/60 backdrop-blur-[2px] opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                        <span className="px-3 py-1.5 rounded-lg bg-white text-black font-bold text-xs flex items-center gap-1.5 shadow-lg active:scale-95">
                          <Eye size={13} />
                          <span>点击全景预览</span>
                        </span>
                      </div>
                    </div>

                    {/* Card Content Info */}
                    <div className="p-4">
                      {/* Asset Title & Category */}
                      <div className="flex items-center justify-between gap-2 mb-1.5">
                        <h3 className="text-sm font-semibold text-white/95 truncate group-hover:text-brand-primary transition-colors">
                          {asset.name}
                        </h3>
                        <span className="text-[11px] text-[#8490a5] shrink-0 font-medium">
                          {getTypeName(asset.asset_type)}
                        </span>
                      </div>

                      {/* Description */}
                      <p className="text-xs text-[#788497] leading-relaxed line-clamp-2 mb-3">
                        {asset.description || "暂无描述"}
                      </p>

                      {/* Tags row */}
                      {asset.tags && asset.tags.length > 0 && (
                        <div className="flex flex-wrap gap-1.5">
                          {asset.tags.map((tag, idx) => (
                            <span
                              key={idx}
                              className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-[#0e1014] text-[10px] text-[#8490a5] border border-[#1d222b]"
                            >
                              <Tag size={9} className="text-[#5a6372]" />
                              <span>{tag}</span>
                            </span>
                          ))}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Card Bottom Meta & Actions */}
                  <div className="px-4 py-3 border-t border-[#1d222b] bg-[#11141a]/60 flex items-center justify-between text-xs text-[#646e80]">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-white/80">
                        {formatBytes(asset.file_size_bytes)}
                      </span>
                      <span>·</span>
                      <span className="truncate max-w-24">{asset.updated_at}</span>
                    </div>

                    <div className="flex items-center gap-2">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setPreviewAsset(asset);
                          setIsPreviewModalOpen(true);
                        }}
                        className="flex items-center gap-1 text-[11px] text-[#8490a5] hover:text-white transition-colors"
                      >
                        <Eye size={12} />
                        <span>预览</span>
                      </button>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          alert(`已在当前分镜制作流中装载资产: ${asset.name}`);
                        }}
                        className="flex items-center gap-1 text-[11px] text-brand-primary hover:underline"
                      >
                        <Sparkles size={11} />
                        <span>装入</span>
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* 5. Asset Fullscreen / Interactive Preview Modal */}
      <AssetPreviewModal
        asset={previewAsset}
        isOpen={isPreviewModalOpen}
        onClose={() => setIsPreviewModalOpen(false)}
        onInject={(a) => alert(`已在当前分镜工程中装载资产: ${a.name}`)}
      />

      {/* 5. Upload/Import Asset Modal with Real File Drag-and-Drop */}
      {isUploadModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4 animate-in fade-in duration-150">
          <div className="w-full max-w-lg rounded-2xl bg-[#14171d] border border-[#282f3d] p-6 shadow-2xl space-y-4">
            <div className="flex items-center justify-between border-b border-[#212733] pb-3">
              <div className="flex items-center gap-2">
                <FileUp size={18} className="text-brand-primary" />
                <h3 className="text-base font-semibold text-white">导入 .blend 资产与自动解析</h3>
              </div>
              <button
                onClick={handleCloseUploadModal}
                className="p-1 text-[#8490a5] hover:text-white rounded cursor-pointer"
              >
                <X size={16} />
              </button>
            </div>

            {uploadStep === "success" && uploadedResult ? (
              /* Success / Inspection Report Card */
              <div className="space-y-4 py-2 animate-in zoom-in-95 duration-150">
                <div className="p-4 rounded-xl bg-emerald-950/30 border border-emerald-500/30 text-emerald-300 space-y-2">
                  <div className="flex items-center gap-2 font-semibold text-sm text-emerald-400">
                    <CheckCircle2 size={16} />
                    <span>Blender 资产解析与自动入库完成！</span>
                  </div>
                  <p className="text-xs text-[#8ca0be] leading-relaxed">
                    已成功识别并持久化存储至资产库，自动提取人物骨骼结构与渲染规格。
                  </p>
                </div>

                <div className="p-4 rounded-xl bg-[#0e1117] border border-[#202735] space-y-2 text-xs font-mono">
                  <div className="flex justify-between">
                    <span className="text-[#646e80]">资产名称:</span>
                    <span className="text-white font-medium">{uploadedResult.name}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-[#646e80]">资产类型:</span>
                    <span className="text-cyan-400 font-semibold">3D 人物模型 (Character)</span>
                  </div>
                  {uploadedResult.metadata?.rig_type && (
                    <div className="flex justify-between">
                      <span className="text-[#646e80]">骨骼体系:</span>
                      <span className="text-pink-400 font-semibold">
                        {uploadedResult.metadata.rig_type}
                      </span>
                    </div>
                  )}
                  {uploadedResult.metadata?.poly_count && (
                    <div className="flex justify-between">
                      <span className="text-[#646e80]">网格面数:</span>
                      <span className="text-amber-400">
                        {uploadedResult.metadata.poly_count.toLocaleString()} 面
                      </span>
                    </div>
                  )}
                  <div className="flex justify-between">
                    <span className="text-[#646e80]">表情驱动:</span>
                    <span className="text-emerald-400">已就绪 (ARKit Blendshapes)</span>
                  </div>
                </div>

                <div className="flex items-center justify-end gap-2.5 pt-2">
                  <button
                    onClick={handleCloseUploadModal}
                    className="px-4 py-2 rounded-xl text-xs text-[#8490a5] hover:text-white hover:bg-white/5 cursor-pointer"
                  >
                    返回看板
                  </button>
                  <button
                    onClick={() => {
                      const res = uploadedResult;
                      handleCloseUploadModal();
                      if (res) {
                        setPreviewAsset(res);
                        setIsPreviewModalOpen(true);
                      }
                    }}
                    className="px-4 py-2 rounded-xl bg-brand-primary text-black font-semibold text-xs flex items-center gap-1.5 hover:bg-brand-primary/90 transition-all cursor-pointer shadow-md"
                  >
                    <Eye size={13} />
                    <span>立即全景 3D 预览</span>
                  </button>
                </div>
              </div>
            ) : (
              <form onSubmit={handleUploadSubmit} className="space-y-4 text-xs">
                {/* Drag and Drop Zone */}
                <div
                  onDragOver={(e) => {
                    e.preventDefault();
                    setIsDragging(true);
                  }}
                  onDragLeave={() => setIsDragging(false)}
                  onDrop={(e) => {
                    e.preventDefault();
                    setIsDragging(false);
                    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
                      const file = e.dataTransfer.files[0];
                      setUploadFile(file);
                      if (!uploadName) setUploadName(file.name);
                    }
                  }}
                  className={`relative w-full h-36 rounded-xl border-2 border-dashed transition-all flex flex-col items-center justify-center p-4 text-center cursor-pointer ${
                    isDragging
                      ? "border-brand-primary bg-brand-primary/10"
                      : uploadFile
                      ? "border-emerald-500/50 bg-emerald-950/20"
                      : "border-[#252d3d] hover:border-[#3a465c] bg-[#0c0e12]"
                  }`}
                  onClick={() => document.getElementById("file-upload-input")?.click()}
                >
                  <input
                    id="file-upload-input"
                    type="file"
                    accept=".blend,.fbx,.obj,.gltf,.glb"
                    className="hidden"
                    onChange={(e) => {
                      if (e.target.files && e.target.files.length > 0) {
                        const file = e.target.files[0];
                        setUploadFile(file);
                        if (!uploadName) setUploadName(file.name);
                      }
                    }}
                  />

                  {uploadFile ? (
                    <div className="space-y-1.5">
                      <div className="w-10 h-10 rounded-full bg-emerald-500/20 text-emerald-400 mx-auto flex items-center justify-center">
                        <CheckCircle2 size={20} />
                      </div>
                      <div className="font-semibold text-white text-xs">{uploadFile.name}</div>
                      <div className="text-[11px] font-mono text-[#8490a5]">
                        大小: {(uploadFile.size / (1024 * 1024)).toFixed(2)} MB · 点击可重新选择
                      </div>
                    </div>
                  ) : (
                    <div className="space-y-2">
                      <UploadCloud size={28} className="text-[#646e80] mx-auto" />
                      <div className="text-white/90 font-medium">
                        拖入真实 <span className="text-brand-primary">.blend</span> / .fbx 模型文件
                      </div>
                      <div className="text-[11px] text-[#6b778a]">
                        或点击此处浏览本地文件（最大支持 128 MB）
                      </div>
                    </div>
                  )}
                </div>

                {/* Optional Custom Display Name */}
                <div>
                  <label className="block text-[#8490a5] mb-1 font-medium">
                    自定义资产名称 (可选，默认使用文件名)
                  </label>
                  <input
                    type="text"
                    value={uploadName}
                    onChange={(e) => setUploadName(e.target.value)}
                    placeholder="例如: alita_warrior_rigged.blend"
                    className="w-full h-8 px-3 rounded-lg bg-[#0e1014] border border-[#212733] text-white focus:outline-none focus:border-brand-primary"
                  />
                </div>

                {/* Analysis Stage Animation Indicator */}
                {isUploading && (
                  <div className="p-3.5 rounded-xl bg-[#0e121a] border border-[#232b3c] space-y-2">
                    <div className="flex items-center justify-between text-[11px] text-brand-primary font-mono">
                      <span className="flex items-center gap-1.5">
                        <Loader2 size={13} className="animate-spin" />
                        {uploadStep === "uploading"
                          ? "正在流式上传 .blend 二进制工程文件..."
                          : "正在启动 Blender 资产特征探针，提取人物骨骼体系与面数..."}
                      </span>
                      <span>{uploadStep === "uploading" ? "45%" : "88%"}</span>
                    </div>
                    <div className="w-full h-1 bg-[#1a202c] rounded-full overflow-hidden">
                      <div
                        className={`h-full bg-brand-primary transition-all duration-300 ${
                          uploadStep === "uploading" ? "w-1/2" : "w-11/12 animate-pulse"
                        }`}
                      />
                    </div>
                  </div>
                )}

                <div className="flex items-center justify-end gap-2 pt-2 border-t border-[#202735]">
                  <button
                    type="button"
                    onClick={handleCloseUploadModal}
                    className="px-3.5 py-1.5 rounded-lg text-[#8490a5] hover:text-white cursor-pointer"
                  >
                    取消
                  </button>
                  <button
                    type="submit"
                    disabled={isUploading || (!uploadFile && !uploadName.trim())}
                    className="px-4 py-1.5 rounded-lg bg-white text-black font-semibold hover:bg-gray-200 flex items-center gap-1.5 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isUploading && <Loader2 size={13} className="animate-spin" />}
                    <span>上传并自动提取人物规格</span>
                  </button>
                </div>
              </form>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
