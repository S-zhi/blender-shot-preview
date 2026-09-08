import React, { useState } from "react";
import {
  Key,
  Server,
  Check,
  Loader2,
  Cpu,
  ShieldCheck,
  Sliders,
  Radio,
  CheckCircle2,
} from "lucide-react";
import { useSettingsStore } from "#/stores/settings-store";
import { LLMProvider } from "#/api/types";
import { Button } from "#/components/ui/button";

export const SettingsView: React.FC = () => {
  const {
    userId,
    setUserId,
    provider,
    setProvider,
    modelName,
    setModelName,
    apiKey,
    setApiKey,
    baseUrl,
    setBaseUrl,
    saveKey,
    isSaving,
    savedKeyId,
  } = useSettingsStore();

  const [saveSuccess, setSaveSuccess] = useState(false);
  const [engine, setEngine] = useState("eevee");
  const [resolution, setResolution] = useState("1080p");

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    const ok = await saveKey();
    if (ok) {
      setSaveSuccess(true);
      setTimeout(() => {
        setSaveSuccess(false);
      }, 2500);
    }
  };

  return (
    <div className="flex-1 h-full overflow-y-auto bg-[#090b0e] text-content p-6 lg:p-10 font-sans select-none relative">
      {/* Background Ambient Glow */}
      <div className="absolute top-0 right-1/4 w-[500px] h-[260px] bg-brand-primary/5 blur-[120px] pointer-events-none -z-0" />

      <div className="max-w-4xl mx-auto space-y-8 z-10 relative">
        {/* 1. Page Header */}
        <div className="border-b border-[#171a22] pb-6">
          <div className="text-[12px] text-[#8490a5] mb-1 font-medium tracking-wide">
            Preferences
          </div>
          <h1 className="text-2xl lg:text-[28px] font-bold tracking-tight text-white/95 mb-1.5">
            系统设置 (Settings)
          </h1>
          <p className="text-xs text-[#717b8c]">
            配置大语言模型访问凭据与 Blender 分镜渲染引擎运行参数。
          </p>
        </div>

        {/* 2. Core Section: LLM Key Configuration (Required!) */}
        <div className="rounded-2xl bg-[#14171d]/90 border border-[#202532] p-6 lg:p-7 shadow-sm space-y-6">
          <div className="flex items-center justify-between border-b border-[#1e232d] pb-4">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-xl bg-[#1c222c] border border-white/5 flex items-center justify-center text-brand-primary">
                <Key size={18} />
              </div>
              <div>
                <h2 className="text-base font-semibold text-white/95">
                  LLM 密钥与模型网关配置
                </h2>
                <p className="text-xs text-[#717b8c]">
                  直接对接 Kitex RPC `LLMKeyServiceV0_1.ManageLLMKey` 服务
                </p>
              </div>
            </div>

            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-500/10 text-[11px] font-mono font-medium text-emerald-400 border border-emerald-500/20">
              <ShieldCheck size={13} />
              RPC 鉴权就绪
            </span>
          </div>

          <form onSubmit={handleSave} className="space-y-5 text-xs">
            {/* User ID */}
            <div>
              <label className="block text-[#8490a5] font-medium mb-1.5">
                用户识别标识 (user_id)
              </label>
              <input
                type="text"
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                placeholder="例如: default_user_001"
                required
                className="w-full h-9 px-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs"
              />
            </div>

            {/* Provider Selection */}
            <div>
              <label className="block text-[#8490a5] font-medium mb-1.5">
                模型服务提供方 (LLMProvider)
              </label>
              <div className="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  onClick={() => setProvider(LLMProvider.OPENAI)}
                  className={`py-2.5 px-4 rounded-xl border font-medium flex items-center justify-between transition-all cursor-pointer ${
                    provider === LLMProvider.OPENAI
                      ? "bg-white text-black font-semibold border-white shadow-xs"
                      : "bg-[#0d0f13] border-[#212632] text-[#8490a5] hover:text-white"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Cpu size={15} />
                    <span>OpenAI (GPT-4o, o1, o3)</span>
                  </div>
                  <span className="text-[10px] font-mono opacity-80">enum: 1</span>
                </button>

                <button
                  type="button"
                  onClick={() => setProvider(LLMProvider.ANTHROPIC)}
                  className={`py-2.5 px-4 rounded-xl border font-medium flex items-center justify-between transition-all cursor-pointer ${
                    provider === LLMProvider.ANTHROPIC
                      ? "bg-white text-black font-semibold border-white shadow-xs"
                      : "bg-[#0d0f13] border-[#212632] text-[#8490a5] hover:text-white"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Cpu size={15} />
                    <span>Anthropic (Claude-3.5-Sonnet)</span>
                  </div>
                  <span className="text-[10px] font-mono opacity-80">enum: 2</span>
                </button>
              </div>
            </div>

            {/* Model Name */}
            <div>
              <label className="block text-[#8490a5] font-medium mb-1.5">
                指定模型名称 (model_name)
              </label>
              <input
                type="text"
                value={modelName}
                onChange={(e) => setModelName(e.target.value)}
                placeholder="gpt-4o, claude-3-5-sonnet-latest"
                required
                className="w-full h-9 px-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs"
              />
            </div>

            {/* API Key */}
            <div>
              <label className="block text-[#8490a5] font-medium mb-1.5">
                API 访问密钥 (api_key)
              </label>
              <input
                type="password"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="sk-..."
                className="w-full h-9 px-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs tracking-wider"
              />
            </div>

            {/* Base URL */}
            <div>
              <label className="block text-[#8490a5] font-medium mb-1.5">
                API 网关基础地址 (base_url)
              </label>
              <div className="relative">
                <Server size={14} className="absolute left-3 top-2.5 text-[#646e80]" />
                <input
                  type="text"
                  value={baseUrl}
                  onChange={(e) => setBaseUrl(e.target.value)}
                  placeholder="https://api.openai.com/v1"
                  className="w-full h-9 pl-9 pr-3.5 rounded-lg bg-[#0d0f13] border border-[#212632] text-white focus:outline-none focus:border-brand-primary font-mono text-xs"
                />
              </div>
            </div>

            {/* Saved Key Status */}
            {savedKeyId && (
              <div className="p-3 rounded-xl bg-[#0d0f13] border border-[#212632] flex items-center justify-between text-xs">
                <span className="text-[#8490a5]">当前已挂载凭证 ID:</span>
                <span className="font-mono text-brand-primary font-medium">{savedKeyId}</span>
              </div>
            )}

            {/* Submit Button */}
            <div className="pt-2 flex items-center justify-end">
              <Button
                type="submit"
                variant="primary"
                disabled={isSaving}
                className="px-6 py-2 h-9 text-xs font-semibold"
              >
                {isSaving ? (
                  <>
                    <Loader2 size={14} className="animate-spin mr-1" />
                    <span>正在同步 RPC...</span>
                  </>
                ) : saveSuccess ? (
                  <>
                    <Check size={14} className="mr-1" />
                    <span>凭据已成功保存</span>
                  </>
                ) : (
                  "保存并同步 LLM 凭证"
                )}
              </Button>
            </div>
          </form>
        </div>

        {/* 3. Auxiliary Section: Blender Rendering Settings */}
        <div className="rounded-2xl bg-[#14171d]/90 border border-[#202532] p-6 lg:p-7 shadow-sm space-y-6">
          <div className="flex items-center gap-3 border-b border-[#1e232d] pb-4">
            <div className="w-9 h-9 rounded-xl bg-[#1c222c] border border-white/5 flex items-center justify-center text-[#cfb755]">
              <Sliders size={18} />
            </div>
            <div>
              <h2 className="text-base font-semibold text-white/95">
                Blender 分镜渲染偏好
              </h2>
              <p className="text-xs text-[#717b8c]">
                配置后台渲染守护进程使用的默认渲染引擎与分辨率
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-5 text-xs">
            <div>
              <label className="block text-[#8490a5] font-medium mb-1.5">
                默认分镜渲染器
              </label>
              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => setEngine("eevee")}
                  className={`py-2 px-3 rounded-lg border font-medium transition-colors ${
                    engine === "eevee"
                      ? "bg-white text-black font-semibold border-white"
                      : "bg-[#0d0f13] border-[#212632] text-[#8490a5]"
                  }`}
                >
                  EEVEE-Next (实时)
                </button>
                <button
                  type="button"
                  onClick={() => setEngine("cycles")}
                  className={`py-2 px-3 rounded-lg border font-medium transition-colors ${
                    engine === "cycles"
                      ? "bg-white text-black font-semibold border-white"
                      : "bg-[#0d0f13] border-[#212632] text-[#8490a5]"
                  }`}
                >
                  Cycles (离线光追)
                </button>
              </div>
            </div>

            <div>
              <label className="block text-[#8490a5] font-medium mb-1.5">
                默认预览分辨率
              </label>
              <div className="grid grid-cols-3 gap-2">
                {["720p", "1080p", "4K"].map((r) => (
                  <button
                    key={r}
                    type="button"
                    onClick={() => setResolution(r)}
                    className={`py-2 px-3 rounded-lg border font-medium transition-colors ${
                      resolution === r
                        ? "bg-white text-black font-semibold border-white"
                        : "bg-[#0d0f13] border-[#212632] text-[#8490a5]"
                    }`}
                  >
                    {r}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* 4. RPC Microservice Health Status */}
        <div className="rounded-2xl bg-[#14171d]/60 border border-[#202532] p-5 flex items-center justify-between text-xs">
          <div className="flex items-center gap-3">
            <Radio size={16} className="text-emerald-400 animate-pulse" />
            <div>
              <div className="font-medium text-white">Kitex RPC 守护服务状态</div>
              <div className="text-[11px] text-[#717b8c] font-mono">
                127.0.0.1:8888 · ShotPreviewServiceV0_1 / LLMKeyServiceV0_1
              </div>
            </div>
          </div>

          <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400 font-medium">
            <CheckCircle2 size={13} />
            全部服务运行中
          </span>
        </div>
      </div>
    </div>
  );
};
