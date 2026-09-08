import React, { useState, useEffect, useCallback, useRef } from "react";
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
  FileUp,
  Eye,
  Bone,
  CheckCircle2,
  ExternalLink,
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

  // 3D Asset Preview modal state
  const [previewAsset, setPreviewAsset] = useState<AssetView | null>(null);
  const [isPreviewModalOpen, setIsPreviewModalOpen] = useState(false);

  // Drag & Drop Upload modal state
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const [uploadName, setUploadName] = useState("");
  const [uploadType, setUploadType] = useState<AssetType>(AssetType.MODEL_3D);
  const [uploadDescription, setUploadDescription] = useState("");
  const [uploadTags, setUploadTags] = useState("");
  const [isUploading, setIsUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

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

  const inferTypeFromExt = (fileName: string): AssetType => {
    const ext = fileName.split(".").pop()?.toLowerCase() || "";
    if (["blend", "fbx", "obj", "gltf", "glb", "usd", "usda", "usdc", "usdz"].includes(ext)) {
      return AssetType.MODEL_3D;
    }
    if (["json", "py"].includes(ext)) {
      return AssetType.SHOT_PRESET;
    }
    if (["png", "jpg", "jpeg", "hdr", "exr", "tif", "tiff", "tga"].includes(ext)) {
      return AssetType.MATERIAL;
    }
    if (["abc", "bvh"].includes(ext)) {
      return AssetType.ANIMATION;
    }
    return AssetType.MODEL_3D;
  };

  const handleFileChosen = (file: File) => {
    setSelectedFile(file);
    setUploadName(file.name);
    const inferred = inferTypeFromExt(file.name);
    setUploadType(inferred);
    setUploadDescription(
      `本地拖拽导入: ${file.name} (${formatBytes(file.size)})`
    );
    const ext = file.name.split(".").pop()?.toLowerCase() || "";
    setUploadTags(`本地导入, ${ext}`);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFileChosen(e.dataTransfer.files[0]);
    }
  };

  const resetUploadModal = () => {
    setSelectedFile(null);
    setIsDragging(false);
    setUploadName("");
    setUploadType(AssetType.MODEL_3D);
    setUploadDescription("");
    setUploadTags("");
    setIsUploadModalOpen(false);
  };

  const handleUploadSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!uploadName.trim()) return;
    try {
      setIsUploading(true);
      const tagsArray = uploadTags
        .split(/[,，]/)
        .map((t) => t.trim())
        .filter(Boolean);

      if (selectedFile) {
        await AssetService.uploadAsset(selectedFile, {
          user_id: "default_user_001",
          name: uploadName,
          asset_type: uploadType,
          description: uploadDescription,
          tags: tagsArray,
        });
      } else {
        await AssetService.registerAsset({
          user_id: "default_user_001",
          name: uploadName,
          asset_type: uploadType,
          file_format: uploadName.split(".").pop() || "blend",
          file_size_bytes: 1024 * 1024,
          storage_uri: `blender://assets/${uploadName}`,
          description: uploadDescription || "分镜工程注册资产",
          tags: tagsArray,
        });
      }
      resetUploadModal();
      fetchAssetsAndStats();
    } catch (err) {
      alert("导入失败: " + (err as Error).message);
    } finally {
      setIsUploading(false);
    }
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
              直接对接后端 Kitex AssetService 磁盘持久化存储，提供场景模型、相机预设与贴图真实存取。
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
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 pb-16">
            {assets.map((asset) => {
              const hasRig = asset.metadata?.has_face_rig || asset.metadata?.rig_type;
              const is3DModel = asset.asset_type === AssetType.MODEL_3D;

              return (
                <div
                  key={asset.asset_id}
                  className="rounded-2xl bg-[#14171d]/85 border border-[#202532] hover:border-[#333d4e] p-5 flex flex-col justify-between transition-all duration-200 shadow-sm hover:shadow-md group"
                >
                  <div>
                    {/* Card Top: Format Badge + Status */}
                    <div className="flex items-center justify-between mb-3">
                      <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-md bg-[#1a202c] text-[11px] font-mono font-semibold text-brand-primary border border-brand-primary/20">
                        {asset.file_format.toUpperCase()} · {getTypeName(asset.asset_type)}
                      </span>
                      <div className="flex items-center gap-1.5 text-xs text-[#8490a5]">
                        <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400">
                          <CheckCircle2 size={12} />
                          就绪
                        </span>
                        <button
                          onClick={() => handleDelete(asset.asset_id)}
                          title="删除资产 (真实调用 Go 后端)"
                          className="p-1 text-[#646e80] hover:text-red-400 hover:bg-white/5 rounded transition-colors cursor-pointer"
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
                      {asset.description || "暂无描述"}
                    </p>

                    {/* Blender Inspector Metadata Badges */}
                    {asset.metadata && (
                      <div className="p-2 rounded-lg bg-[#0e1117] border border-[#1b212c] mb-3.5 flex items-center justify-between text-[11px] font-mono text-[#8490a5]">
                        {asset.metadata.poly_count && (
                          <div>
                            面数: <span className="text-white/90">{(asset.metadata.poly_count / 1000).toFixed(0)}k</span>
                          </div>
                        )}
                        {hasRig && (
                          <div className="flex items-center gap-1 text-emerald-400">
                            <Bone size={12} />
                            <span>{asset.metadata.rig_type ? "带骨骼绑定" : "基础骨骼"}</span>
                          </div>
                        )}
                        {asset.metadata.duration_frames && (
                          <div className="text-[#cfb755]">
                            {asset.metadata.duration_frames}帧
                          </div>
                        )}
                      </div>
                    )}

                    {/* Tags row */}
                    {asset.tags && asset.tags.length > 0 && (
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
                    )}
                  </div>

                  {/* Card Bottom Meta Info & Actions */}
                  <div className="pt-3 border-t border-[#1d222b] flex items-center justify-between text-xs text-[#646e80]">
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-white/80">
                        {formatBytes(asset.file_size_bytes)}
                      </span>
                      <span>·</span>
                      <span className="truncate max-w-28">{asset.updated_at}</span>
                    </div>

                    <div className="flex items-center gap-2">
                      {/* 3D Preview Button */}
                      {is3DModel && (
                        <button
                          onClick={() => {
                            setPreviewAsset(asset);
                            setIsPreviewModalOpen(true);
                          }}
                          className="flex items-center gap-1 text-[11px] px-2 py-1 rounded bg-[#1b2230] text-emerald-400 hover:bg-[#232c3d] border border-emerald-500/20 transition-all cursor-pointer font-medium"
                          title="在 Three.js 视口中 3D 交互预览"
                        >
                          <Eye size={12} />
                          <span>3D 预览</span>
                        </button>
                      )}

                      <button
                        onClick={() => alert(`已在当前分镜工程中装载资产: ${asset.name}`)}
                        className="flex items-center gap-1 text-[11px] text-brand-primary hover:underline cursor-pointer"
                      >
                        <span>装入分镜</span>
                        <ExternalLink size={11} />
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

      {/* 6. Drag & Drop Upload/Import Asset Modal */}
      {isUploadModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-xs p-4 animate-in fade-in duration-150">
          <div className="w-full max-w-lg rounded-2xl bg-[#14171d] border border-[#282f3d] p-6 shadow-2xl space-y-5 text-content">
            {/* Header */}
            <div className="flex items-center justify-between border-b border-[#212733] pb-3.5">
              <div>
                <h3 className="text-base font-semibold text-white flex items-center gap-2">
                  <UploadCloud size={18} className="text-brand-primary" />
                  <span>拖拽导入分镜素材资产</span>
                </h3>
                <p className="text-[11px] text-[#788497] mt-0.5">
                  支持拖入 Blender 模型、相机轨迹脚本或贴图，流式落盘至工作区
                </p>
              </div>
              <button
                onClick={resetUploadModal}
                className="p-1.5 text-[#8490a5] hover:text-white hover:bg-white/5 rounded-lg transition-colors cursor-pointer"
              >
                <X size={16} />
              </button>
            </div>

            <form onSubmit={handleUploadSubmit} className="space-y-4 text-xs">
              {/* Hidden file input */}
              <input
                ref={fileInputRef}
                type="file"
                className="hidden"
                onChange={(e) => {
                  if (e.target.files && e.target.files.length > 0) {
                    handleFileChosen(e.target.files[0]);
                  }
                }}
              />

              {/* Drag & Drop Zone */}
              {!selectedFile ? (
                <div
                  onDragOver={handleDragOver}
                  onDragLeave={handleDragLeave}
                  onDrop={handleDrop}
                  onClick={() => fileInputRef.current?.click()}
                  className={clsx(
                    "border-2 border-dashed rounded-xl p-7 text-center transition-all cursor-pointer flex flex-col items-center justify-center gap-2.5 select-none",
                    isDragging
                      ? "border-brand-primary bg-brand-primary/10 shadow-lg shadow-brand-primary/5 scale-[1.01]"
                      : "border-[#2b3342] bg-[#0e1014]/60 hover:border-[#3d485c] hover:bg-[#12161f]"
                  )}
                >
                  <div className="w-12 h-12 rounded-full bg-[#1c222c] flex items-center justify-center text-brand-primary shadow-xs">
                    <FileUp size={22} className={clsx(isDragging && "animate-bounce")} />
                  </div>
                  <div>
                    <div className="text-xs font-semibold text-white/90">
                      {isDragging ? "松开鼠标完成投放" : "拖拽文件到此处，或点击浏览本地文件"}
                    </div>
                    <div className="text-[11px] text-[#6b778c] mt-1">
                      支持 .blend, .fbx, .obj, .json, .py, .hdr, .abc 等
                    </div>
                  </div>
                  <div className="flex flex-wrap items-center justify-center gap-1.5 pt-1">
                    {["3D模型 .blend", "镜头预设 .json/.py", "环境贴图 .hdr", "动画缓存 .abc"].map((t) => (
                      <span
                        key={t}
                        className="px-2 py-0.5 rounded text-[10px] font-mono bg-[#161a22] text-[#8490a5] border border-[#232936]"
                      >
                        {t}
                      </span>
                    ))}
                  </div>
                </div>
              ) : (
                /* Selected File Card */
                <div className="rounded-xl bg-[#0e1014] border border-[#2b3342] p-4 flex items-center justify-between gap-3">
                  <div className="flex items-center gap-3 overflow-hidden">
                    <div className="w-10 h-10 rounded-lg bg-brand-primary/10 border border-brand-primary/30 flex items-center justify-center shrink-0 text-brand-primary font-mono text-xs font-bold">
                      .{selectedFile.name.split(".").pop()?.toUpperCase() || "FILE"}
                    </div>
                    <div className="overflow-hidden">
                      <div className="text-xs font-medium text-white truncate">
                        {selectedFile.name}
                      </div>
                      <div className="text-[11px] text-[#717b8c] font-mono flex items-center gap-2 mt-0.5">
                        <span>{formatBytes(selectedFile.size)}</span>
                        <span>·</span>
                        <span className="text-emerald-400 flex items-center gap-1">
                          <CheckCircle2 size={11} />
                          文件就绪
                        </span>
                      </div>
                    </div>
                  </div>

                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    className="px-2.5 py-1 rounded-md text-[11px] text-[#8490a5] hover:text-white hover:bg-white/5 border border-[#252b37] shrink-0 transition-colors cursor-pointer"
                  >
                    更换文件
                  </button>
                </div>
              )}

              {/* Metadata Form Fields */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
                <div>
                  <label className="block text-[#8490a5] mb-1 font-medium">资产展示名称</label>
                  <input
                    type="text"
                    value={uploadName}
                    onChange={(e) => setUploadName(e.target.value)}
                    placeholder="例如: cyberpunk_camera_rig.blend"
                    required
                    className="w-full h-8 px-3 rounded-lg bg-[#0e1014] border border-[#212733] text-white focus:outline-none focus:border-brand-primary text-xs"
                  />
                </div>

                <div>
                  <label className="block text-[#8490a5] mb-1 font-medium">资产分类 (AssetType)</label>
                  <select
                    value={uploadType}
                    onChange={(e) => setUploadType(Number(e.target.value) as AssetType)}
                    className="w-full h-8 px-3 rounded-lg bg-[#0e1014] border border-[#212733] text-white focus:outline-none focus:border-brand-primary text-xs cursor-pointer"
                  >
                    <option value={AssetType.MODEL_3D}>3D 模型 (.blend, .fbx, .obj)</option>
                    <option value={AssetType.SHOT_PRESET}>镜头预设 (.json, .py)</option>
                    <option value={AssetType.MATERIAL}>材质贴图 (.hdr, .exr, .png)</option>
                    <option value={AssetType.ANIMATION}>动画序列 (.abc, .bvh)</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-[#8490a5] mb-1 font-medium">标签 Tags (以逗号分隔)</label>
                <input
                  type="text"
                  value={uploadTags}
                  onChange={(e) => setUploadTags(e.target.value)}
                  placeholder="例如: 主角, 特写, 4K"
                  className="w-full h-8 px-3 rounded-lg bg-[#0e1014] border border-[#212733] text-white focus:outline-none focus:border-brand-primary text-xs"
                />
              </div>

              <div>
                <label className="block text-[#8490a5] mb-1 font-medium">资产描述 Description</label>
                <textarea
                  value={uploadDescription}
                  onChange={(e) => setUploadDescription(e.target.value)}
                  rows={2}
                  placeholder="补充关于该模型或镜头的分镜用途..."
                  className="w-full p-2.5 rounded-lg bg-[#0e1014] border border-[#212733] text-white focus:outline-none focus:border-brand-primary text-xs resize-none"
                />
              </div>

              {/* Status Hint */}
              <div className="p-3 rounded-lg bg-[#0e1014] border border-[#212733] text-[11px] text-[#717b8c] leading-relaxed">
                {selectedFile ? (
                  <span className="text-emerald-400/90 font-mono">
                    ✓ 导入后物理文件将流式写入后端工作区 assets 目录，并同步写入 metadata.json 持久化保存。
                  </span>
                ) : (
                  <span>
                    提示：可将本地文件直接拖拽投放至上方区域，自动完成物理落盘与分类索引。
                  </span>
                )}
              </div>

              {/* Action Buttons */}
              <div className="flex items-center justify-end gap-2.5 pt-2 border-t border-[#1c212c]">
                <button
                  type="button"
                  onClick={resetUploadModal}
                  className="px-3.5 py-1.5 rounded-lg text-xs text-[#8490a5] hover:text-white transition-colors cursor-pointer"
                >
                  取消
                </button>
                <button
                  type="submit"
                  disabled={isUploading || !uploadName.trim()}
                  className="px-5 py-1.5 rounded-lg bg-white text-black text-xs font-semibold hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1.5 transition-all cursor-pointer shadow-md"
                >
                  {isUploading && <Loader2 size={13} className="animate-spin" />}
                  <span>{isUploading ? "正在导入落盘..." : "确认导入资产"}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
