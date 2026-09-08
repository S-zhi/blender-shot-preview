import React, { useState } from "react";
import { X, Key, Server, Check, Loader2 } from "lucide-react";
import { useSettingsStore } from "#/stores/settings-store";
import { LLMProvider } from "#/api/types";
import { Button } from "#/components/ui/button";

export const LLMKeyModal: React.FC = () => {
  const {
    isModalOpen,
    setModalOpen,
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

  if (!isModalOpen) return null;

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    const ok = await saveKey();
    if (ok) {
      setSaveSuccess(true);
      setTimeout(() => {
        setSaveSuccess(false);
      }, 1500);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-muted-overlay backdrop-blur-sm p-4 select-none">
      <div className="w-full max-w-md rounded-2xl bg-surface-outline border border-border shadow-2xl overflow-hidden animate-modal-enter">
        {/* Modal Header */}
        <div className="flex items-center justify-between px-5 py-4 border-b border-border">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-lg bg-brand-primary/10 text-brand-primary">
              <Key size={18} />
            </div>
            <div>
              <h3 className="text-sm font-semibold text-content">LLM Gateway 凭证管理</h3>
              <p className="text-[11px] text-content-muted">
                配置 Kitex LLMKeyServiceV0_1 访问凭证
              </p>
            </div>
          </div>

          <button
            onClick={() => setModalOpen(false)}
            className="p-1.5 text-content-muted hover:text-content hover:bg-surface-elevated rounded-lg transition-colors"
          >
            <X size={16} />
          </button>
        </div>

        {/* Modal Body */}
        <form onSubmit={handleSave} className="p-5 space-y-4 text-xs">
          {/* User ID */}
          <div>
            <label className="block text-content-muted font-medium mb-1">
              用户标识 (user_id)
            </label>
            <input
              type="text"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              placeholder="e.g. user_default_001"
              required
              className="w-full h-8 px-3 rounded-lg bg-surface-background border border-border text-content focus:outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary/30"
            />
          </div>

          {/* Provider Selection */}
          <div>
            <label className="block text-content-muted font-medium mb-1">
              模型服务提供方 (LLMProvider)
            </label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => setProvider(LLMProvider.OPENAI)}
                className={`py-2 px-3 rounded-lg border font-medium transition-all ${
                  provider === LLMProvider.OPENAI
                    ? "bg-brand-primary/10 border-brand-primary text-brand-primary"
                    : "bg-surface-elevated border-border text-content-muted hover:text-content"
                }`}
              >
                OpenAI (1)
              </button>
              <button
                type="button"
                onClick={() => setProvider(LLMProvider.ANTHROPIC)}
                className={`py-2 px-3 rounded-lg border font-medium transition-all ${
                  provider === LLMProvider.ANTHROPIC
                    ? "bg-brand-primary/10 border-brand-primary text-brand-primary"
                    : "bg-surface-elevated border-border text-content-muted hover:text-content"
                }`}
              >
                Anthropic (2)
              </button>
            </div>
          </div>

          {/* Model Name */}
          <div>
            <label className="block text-content-muted font-medium mb-1">
              模型名称 (name)
            </label>
            <input
              type="text"
              value={modelName}
              onChange={(e) => setModelName(e.target.value)}
              placeholder="gpt-4o, claude-3-5-sonnet-latest"
              required
              className="w-full h-8 px-3 rounded-lg bg-surface-background border border-border text-content focus:outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary/30"
            />
          </div>

          {/* API Key */}
          <div>
            <label className="block text-content-muted font-medium mb-1">
              API 密钥 (api_key)
            </label>
            <input
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="sk-..."
              className="w-full h-8 px-3 rounded-lg bg-surface-background border border-border text-content focus:outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary/30 font-mono"
            />
          </div>

          {/* Base URL */}
          <div>
            <label className="block text-content-muted font-medium mb-1">
              API 基础路径 (base_url)
            </label>
            <div className="relative">
              <Server size={14} className="absolute left-2.5 top-2 text-content-icon" />
              <input
                type="text"
                value={baseUrl}
                onChange={(e) => setBaseUrl(e.target.value)}
                placeholder="https://api.openai.com/v1"
                className="w-full h-8 pl-8 pr-3 rounded-lg bg-surface-background border border-border text-content focus:outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary/30 font-mono text-[11px]"
              />
            </div>
          </div>

          {savedKeyId && (
            <div className="p-2.5 rounded-lg bg-surface-elevated border border-border text-[11px] text-content-muted">
              当前绑定凭证 ID: <span className="text-brand-primary font-mono">{savedKeyId}</span>
            </div>
          )}

          {/* Footer Buttons */}
          <div className="flex items-center justify-end gap-2 pt-2 border-t border-border">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setModalOpen(false)}
            >
              取消
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="sm"
              disabled={isSaving}
              className="min-w-20"
            >
              {isSaving ? (
                <Loader2 size={13} className="animate-spin" />
              ) : saveSuccess ? (
                <>
                  <Check size={13} /> 已保存
                </>
              ) : (
                "保存凭据"
              )}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
};
