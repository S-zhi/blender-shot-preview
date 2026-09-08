import React, { useState } from "react";
import {
  Server,
  Check,
  Loader2,
  Cpu,
  ShieldCheck,
  Sliders,
  Plus,
  Trash2,
  Activity,
  AlertCircle,
  CheckCircle2,
} from "lucide-react";
import clsx from "clsx";
import { useSettingsStore, ProviderItem } from "#/stores/settings-store";
import { Button } from "#/components/ui/button";

export const SettingsView: React.FC = () => {
  const {
    userId,
    setUserId,
    activeTab,
    setActiveTab,
    providers,
    selectedProviderId,
    setSelectedProviderId,
    updateProvider,
    addProvider,
    removeProvider,
    testConnection,
    saveCurrentProviderKey,
    isSaving,
    isTesting,
    renderEngine,
    setRenderEngine,
    renderResolution,
    setRenderResolution,
  } = useSettingsStore();

  const [saveSuccess, setSaveSuccess] = useState(false);
  const [isCreatingModalOpen, setIsCreatingModalOpen] = useState(false);
  const [newTabId, setNewTabId] = useState("");
  const [newLabel, setNewLabel] = useState("");
  const [newType, setNewType] = useState<"openai" | "azure_openai">("openai");

  const currentProvider = providers.find((p) => p.id === selectedProviderId) || providers[0];

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    const ok = await saveCurrentProviderKey();
    if (ok) {
      setSaveSuccess(true);
      setTimeout(() => {
        setSaveSuccess(false);
      }, 2500);
    }
  };

  const handleTest = async () => {
    if (!currentProvider) return;
    await testConnection(currentProvider.id);
  };

  const handleCreateSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const cleanId = newTabId.trim().toLowerCase().replace(/[^a-z0-9_-]/g, "_");
    if (!cleanId) return;

    const newProvider: ProviderItem = {
      id: cleanId,
      label: newLabel.trim() || cleanId,
      providerType: newType,
      modelName: "gpt-4o",
      apiKey: "",
      baseUrl: newType === "azure_openai" ? "https://your-resource.openai.azure.com" : "https://api.openai.com/v1",
      apiVersion: newType === "azure_openai" ? "2024-02-15-preview" : undefined,
    };

    addProvider(newProvider);
    setIsCreatingModalOpen(false);
    setNewTabId("");
    setNewLabel("");
  };

  return (
    <div className="flex-1 h-full flex bg-[#090b0e] text-content font-sans select-none overflow-hidden">
      {/* 1. Internal Settings Sidebar */}
      <div className="w-64 border-r border-[#1a1d24] bg-[#0c0e12] flex flex-col shrink-0 p-4 justify-between">
        <div className="space-y-6">
          {/* Header & Title */}
          <div>
            <div className="text-[11px] text-[#8490a5] font-mono uppercase tracking-wider mb-1">
              Configuration
            </div>
            <h2 className="text-base font-bold text-white tracking-tight">
              系统与模型设置
            </h2>
          </div>

          {/* Model Providers Section */}
          <div className="space-y-2">
            <div className="flex items-center justify-between px-1">
              <span className="text-xs font-semibold text-[#8b95a5] uppercase tracking-wider">
                模型提供商 (Providers)
              </span>
              <button
                onClick={() => setIsCreatingModalOpen(true)}
                title="创建新的提供商配置"
                className="flex items-center gap-1 text-[11px] text-brand-primary hover:text-white px-2 py-0.5 rounded bg-brand-primary/10 hover:bg-brand-primary/20 transition-all font-medium cursor-pointer"
              >
                <Plus size={12} />
                <span>创建</span>
              </button>
            </div>

            <div className="space-y-1">
              {providers.map((p) => {
                const isSelected = activeTab === p.id;
                return (
                  <div
                    key={p.id}
                    onClick={() => {
                      setSelectedProviderId(p.id);
                      setActiveTab(p.id);
                    }}
                    className={clsx(
                      "w-full px-3 py-2 rounded-xl flex items-center justify-between text-xs font-medium cursor-pointer transition-all border group",
                      isSelected
                        ? "bg-white text-black font-semibold border-white shadow-xs"
                        : "bg-[#14171d]/60 border-[#202532] text-[#8b95a5] hover:text-white hover:bg-[#1a1e27]"
                    )}
                  >
                    <div className="flex items-center gap-2 truncate">
                      <Cpu size={14} className={isSelected ? "text-black" : "text-brand-primary"} />
                      <span className="font-mono truncate">{p.id}</span>
                    </div>

                    <div className="flex items-center gap-1.5 shrink-0">
                      {p.lastTestResult && (
                        <span
                          title={p.lastTestResult.message}
                          className={clsx(
                            "w-2 h-2 rounded-full",
                            p.lastTestResult.success ? "bg-emerald-500" : "bg-red-500"
                          )}
                        />
                      )}
                      {providers.length > 1 && (
                        <button
                          type="button"
                          onClick={(e) => {
                            e.stopPropagation();
                            removeProvider(p.id);
                          }}
                          title="删除此配置"
                          className={clsx(
                            "opacity-0 group-hover:opacity-100 p-0.5 hover:text-red-400 transition-opacity",
                            isSelected ? "text-gray-600 hover:text-red-600" : "text-gray-400"
                          )}
                        >
                          <Trash2 size={12} />
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Preferences Section */}
          <div className="space-y-2">
            <span className="text-xs font-semibold text-[#8b95a5] uppercase tracking-wider px-1">
              渲染与环境偏好
            </span>
            <div
              onClick={() => setActiveTab("render_preferences")}
              className={clsx(
                "w-full px-3 py-2 rounded-xl flex items-center gap-2 text-xs font-medium cursor-pointer transition-all border",
                activeTab === "render_preferences"
                  ? "bg-white text-black font-semibold border-white shadow-xs"
                  : "bg-[#14171d]/60 border-[#202532] text-[#8b95a5] hover:text-white hover:bg-[#1a1e27]"
              )}
            >
              <Sliders size={14} className={activeTab === "render_preferences" ? "text-black" : "text-[#cfb755]"} />
              <span className="font-mono">render_preferences</span>
            </div>
          </div>
        </div>

        {/* Sidebar Footer Info */}
        <div className="pt-4 border-t border-[#1a1d24] text-[11px] text-[#636e82]">
          <div className="flex items-center gap-1.5 text-emerald-400 font-mono mb-1">
            <CheckCircle2 size={12} />
            <span>Kitex RPC 连接正常</span>
          </div>
          <div className="truncate">用户: {userId}</div>
        </div>
      </div>

      {/* 2. Main Content Area */}
      <div className="flex-1 h-full overflow-y-auto p-6 lg:p-10 relative">
        <div className="max-w-3xl mx-auto space-y-6">
          {activeTab === "render_preferences" ? (
            /* Tab: Blender Render Preferences */
            <div className="space-y-6 animate-fade-in">
              <div className="border-b border-[#171a22] pb-4">
                <h1 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
                  <Sliders size={20} className="text-[#cfb755]" />
                  <span>Blender 分镜渲染偏好 (render_preferences)</span>
                </h1>
                <p className="text-xs text-[#717b8c] mt-1">
                  设置后台无头渲染器进程与视口预览生成的默认规格与渲染引擎
                </p>
              </div>

              <div className="rounded-2xl bg-[#14171d]/90 border border-[#202532] p-6 space-y-6 text-xs">
                <div>
                  <label className="block text-[#8490a5] font-medium mb-2">
                    默认分镜渲染器 (Engine)
                  </label>
                  <div className="grid grid-cols-2 gap-3">
                    <button
                      type="button"
                      onClick={() => setRenderEngine("eevee")}
                      className={clsx(
                        "py-3 px-4 rounded-xl border font-medium flex items-center justify-between transition-all cursor-pointer",
                        renderEngine === "eevee"
                          ? "bg-white text-black font-semibold border-white shadow-xs"
                          : "bg-[#0d0f13] border-[#212632] text-[#8490a5] hover:text-white"
                      )}
                    >
                      <span className="text-xs">EEVEE-Next (实时高速渲染)</span>
                      <span className="text-[10px] font-mono opacity-70">默认首选</span>
                    </button>

                    <button
                      type="button"
                      onClick={() => setRenderEngine("cycles")}
                      className={clsx(
                        "py-3 px-4 rounded-xl border font-medium flex items-center justify-between transition-all cursor-pointer",
                        renderEngine === "cycles"
                          ? "bg-white text-black font-semibold border-white shadow-xs"
                          : "bg-[#0d0f13] border-[#212632] text-[#8490a5] hover:text-white"
                      )}
                    >
                      <span className="text-xs">Cycles (物理光线追踪)</span>
                      <span className="text-[10px] font-mono opacity-70">高精画质</span>
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-[#8490a5] font-medium mb-2">
                    默认预览分辨率 (Resolution)
                  </label>
                  <div className="grid grid-cols-3 gap-3">
                    {(["720p", "1080p", "4K"] as const).map((res) => (
                      <button
                        key={res}
                        type="button"
                        onClick={() => setRenderResolution(res)}
                        className={clsx(
                          "py-2.5 px-3 rounded-xl border font-medium transition-all text-center cursor-pointer",
                          renderResolution === res
                            ? "bg-white text-black font-semibold border-white shadow-xs"
                            : "bg-[#0d0f13] border-[#212632] text-[#8490a5] hover:text-white"
                        )}
                      >
                        {res}
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          ) : (
            /* Tab: Individual LLM Provider Configuration */
            <div className="space-y-6 animate-fade-in">
              <div className="border-b border-[#171a22] pb-4 flex items-center justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h1 className="text-xl font-bold text-white tracking-tight font-mono">
                      {currentProvider.id}
                    </h1>
                    <span className="px-2 py-0.5 rounded-md bg-[#1d222b] text-[11px] font-mono text-brand-primary border border-white/5">
                      {currentProvider.providerType}
                    </span>
                  </div>
                  <p className="text-xs text-[#717b8c] mt-1">
                    配置该提供商的独立访问网关地址与 API 凭证密钥
                  </p>
                </div>

                <div className="flex items-center gap-2">
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-500/10 text-[11px] font-mono text-emerald-400 border border-emerald-500/20">
                    <ShieldCheck size={13} />
                    <span>独立协议配置</span>
                  </span>
                </div>
              </div>

              {/* Provider Config Form */}
              <div className="rounded-2xl bg-[#14171d]/90 border border-[#202532] p-6 space-y-5 text-xs">
                {/* User ID */}
                <div>
                  <label className="block text-[#8490a5] font-medium mb-1.5">
                    绑定用户识别标识 (user_id)
                  </label>
                  <input
                    type="text"
                    value={userId}
                    onChange={(e) => setUserId(e.target.value)}
                    placeholder="default_user_001"
                    className="w-full h-9 px-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs"
                  />
                </div>

                {/* Model Name (Freeform, NO HARDCODED RESTRICTIONS!) */}
                <div>
                  <label className="block text-[#8490a5] font-medium mb-1.5">
                    指定大模型名称 (model_name / deployment)
                  </label>
                  <input
                    type="text"
                    value={currentProvider.modelName}
                    onChange={(e) =>
                      updateProvider(currentProvider.id, { modelName: e.target.value })
                    }
                    placeholder="可自由指定模型名称，如 gpt-4o, deepseek-chat, qwen-plus..."
                    required
                    className="w-full h-9 px-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs"
                  />
                  <p className="text-[11px] text-[#636e82] mt-1">
                    支持输入任何 OpenAI 兼容协议支持的模型标识或 Azure 部署名称，无版本写死限制。
                  </p>
                </div>

                {/* API Key */}
                <div>
                  <label className="block text-[#8490a5] font-medium mb-1.5">
                    API 访问密钥 (api_key)
                  </label>
                  <input
                    type="password"
                    value={currentProvider.apiKey}
                    onChange={(e) =>
                      updateProvider(currentProvider.id, { apiKey: e.target.value })
                    }
                    placeholder="sk-..."
                    className="w-full h-9 px-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs tracking-wider"
                  />
                </div>

                {/* Base URL */}
                <div>
                  <label className="block text-[#8490a5] font-medium mb-1.5">
                    网关服务基础地址 (base_url)
                  </label>
                  <div className="relative">
                    <Server size={14} className="absolute left-3 top-2.5 text-[#646e80]" />
                    <input
                      type="text"
                      value={currentProvider.baseUrl}
                      onChange={(e) =>
                        updateProvider(currentProvider.id, { baseUrl: e.target.value })
                      }
                      placeholder="https://api.openai.com/v1 或兼容网关地址"
                      className="w-full h-9 pl-9 pr-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs"
                    />
                  </div>
                </div>

                {/* Azure API Version if applicable */}
                {currentProvider.providerType === "azure_openai" && (
                  <div>
                    <label className="block text-[#8490a5] font-medium mb-1.5">
                      Azure API Version (api-version)
                    </label>
                    <input
                      type="text"
                      value={currentProvider.apiVersion || "2024-02-15-preview"}
                      onChange={(e) =>
                        updateProvider(currentProvider.id, { apiVersion: e.target.value })
                      }
                      placeholder="2024-02-15-preview"
                      className="w-full h-9 px-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs"
                    />
                  </div>
                )}

                {/* Test Connection Feedback Display */}
                {currentProvider.lastTestResult && (
                  <div
                    className={clsx(
                      "p-3 rounded-xl border flex items-start gap-2.5 text-xs transition-all",
                      currentProvider.lastTestResult.success
                        ? "bg-emerald-500/10 border-emerald-500/30 text-emerald-400"
                        : "bg-red-500/10 border-red-500/30 text-red-400"
                    )}
                  >
                    {currentProvider.lastTestResult.success ? (
                      <CheckCircle2 size={16} className="shrink-0 mt-0.5" />
                    ) : (
                      <AlertCircle size={16} className="shrink-0 mt-0.5" />
                    )}
                    <div className="flex-1">
                      <div className="font-semibold">
                        {currentProvider.lastTestResult.success ? "连通性测试通过" : "连通性测试未通过"}
                      </div>
                      <div className="text-[11px] opacity-90 break-all mt-0.5">
                        {currentProvider.lastTestResult.message}
                      </div>
                    </div>
                  </div>
                )}

                {/* Buttons: Test Connection + Save */}
                <div className="pt-2 flex items-center justify-between border-t border-[#1e232d]">
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={handleTest}
                      disabled={isTesting}
                      className="px-4 py-2 h-9 rounded-lg border border-[#2c3240] hover:bg-[#1a1e27] text-white font-medium flex items-center gap-1.5 transition-colors cursor-pointer text-xs disabled:opacity-50"
                    >
                      {isTesting ? (
                        <>
                          <Loader2 size={13} className="animate-spin text-brand-primary" />
                          <span>正在发起连通测试...</span>
                        </>
                      ) : (
                        <>
                          <Activity size={13} className="text-brand-primary" />
                          <span>测试连通性 (Test)</span>
                        </>
                      )}
                    </button>
                    <span className="text-[11px] text-[#636e82]">
                      测试不通过亦可正常保存配置
                    </span>
                  </div>

                  <Button
                    type="button"
                    onClick={handleSave}
                    variant="primary"
                    disabled={isSaving}
                    className="px-6 py-2 h-9 text-xs font-semibold"
                  >
                    {isSaving ? (
                      <>
                        <Loader2 size={14} className="animate-spin mr-1" />
                        <span>正在同步凭据...</span>
                      </>
                    ) : saveSuccess ? (
                      <>
                        <Check size={14} className="mr-1" />
                        <span>凭证已保存</span>
                      </>
                    ) : (
                      "保存此提供商配置"
                    )}
                  </Button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* 3. Create Provider Modal */}
      {isCreatingModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4 select-none">
          <div className="w-full max-w-md rounded-2xl bg-[#14171d] border border-[#202532] shadow-2xl p-6 space-y-4">
            <h3 className="text-base font-bold text-white tracking-tight flex items-center gap-2">
              <Plus size={16} className="text-brand-primary" />
              <span>添加大模型提供商配置</span>
            </h3>

            <form onSubmit={handleCreateSubmit} className="space-y-4 text-xs">
              <div>
                <label className="block text-[#8490a5] font-medium mb-1">
                  英文 Tab 标识 ID (如: deepseek, qwen, local_vllm)
                </label>
                <input
                  type="text"
                  value={newTabId}
                  onChange={(e) => setNewTabId(e.target.value)}
                  placeholder="deepseek"
                  required
                  className="w-full h-8 px-3 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs"
                />
              </div>

              <div>
                <label className="block text-[#8490a5] font-medium mb-1">
                  协议格式类型
                </label>
                <div className="grid grid-cols-2 gap-2">
                  <button
                    type="button"
                    onClick={() => setNewType("openai")}
                    className={clsx(
                      "py-2 px-3 rounded-lg border font-medium transition-all text-center",
                      newType === "openai"
                        ? "bg-white text-black font-semibold border-white"
                        : "bg-[#0d0f13] border-[#212632] text-[#8490a5]"
                    )}
                  >
                    OpenAI 兼容协议
                  </button>
                  <button
                    type="button"
                    onClick={() => setNewType("azure_openai")}
                    className={clsx(
                      "py-2 px-3 rounded-lg border font-medium transition-all text-center",
                      newType === "azure_openai"
                        ? "bg-white text-black font-semibold border-white"
                        : "bg-[#0d0f13] border-[#212632] text-[#8490a5]"
                    )}
                  >
                    Azure OpenAI 协议
                  </button>
                </div>
              </div>

              <div className="flex items-center justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsCreatingModalOpen(false)}
                  className="px-4 py-2 rounded-lg text-xs text-[#8490a5] hover:text-white"
                >
                  取消
                </button>
                <Button type="submit" variant="primary" className="px-5 py-2 text-xs">
                  确认添加
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
