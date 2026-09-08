import React, { useState } from "react";
import { KeyRound, ShieldAlert, Loader2, ArrowRight, Eye, EyeOff, Lock } from "lucide-react";
import { loginWithToken } from "#/api/client";

interface LoginPageProps {
  onSuccess: () => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ onSuccess }) => {
  const [token, setToken] = useState("");
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");
  const [showPassword, setShowPassword] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const cleanToken = token.trim();
    if (!cleanToken) {
      setErrorMsg("请输入服务器 Access Token");
      return;
    }

    setLoading(true);
    setErrorMsg("");

    const result = await loginWithToken(cleanToken);
    setLoading(false);

    if (result.success) {
      onSuccess();
    } else {
      setErrorMsg(result.error || "访问令牌验证失败，请核对后重试");
    }
  };

  return (
    <div className="flex min-h-screen w-screen items-center justify-center bg-[#07080b] p-4 font-sans text-gray-200 selection:bg-orange-500/30 selection:text-orange-200">
      {/* Subtle Background Glow */}
      <div className="pointer-events-none fixed inset-0 flex items-center justify-center overflow-hidden">
        <div className="h-[480px] w-[480px] rounded-full bg-orange-600/10 blur-[130px]" />
        <div className="h-[360px] w-[360px] rounded-full bg-blue-600/5 blur-[120px]" />
      </div>

      <div className="relative z-10 w-full max-w-md overflow-hidden rounded-2xl border border-white/10 bg-[#12141a]/95 p-8 shadow-2xl backdrop-blur-xl transition-all">
        {/* Top Header */}
        <div className="mb-6 flex flex-col items-center text-center">
          <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl border border-orange-500/30 bg-orange-500/10 shadow-inner">
            <Lock className="h-7 w-7 text-orange-400" />
          </div>
          <h1 className="text-xl font-bold tracking-tight text-white">
            Blender Shot Preview
          </h1>
          <p className="mt-1.5 text-xs text-gray-400">
            服务访问安全门禁 · Server Access Gate
          </p>
        </div>

        {/* Info Banner */}
        <div className="mb-6 rounded-lg border border-orange-500/20 bg-orange-500/5 p-3 text-xs leading-relaxed text-orange-300/90">
          <div className="flex items-start gap-2">
            <KeyRound className="mt-0.5 h-4 w-4 shrink-0 text-orange-400" />
            <span>
              已启用服务器端安全保护。请输入服务启动时控制台打印的{" "}
              <code className="rounded bg-black/40 px-1 py-0.5 font-mono text-orange-300">
                Access Token
              </code>{" "}
              以建立安全 Session 会话。
            </span>
          </div>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-1.5 block text-xs font-medium text-gray-300">
              访问密钥 (Access Token)
            </label>
            <div className="relative">
              <input
                type={showPassword ? "text" : "password"}
                value={token}
                onChange={(e) => {
                  setToken(e.target.value);
                  if (errorMsg) setErrorMsg("");
                }}
                placeholder="例如: bspe_xxxxxxxxxxxxxxxx..."
                autoFocus
                disabled={loading}
                className="w-full rounded-xl border border-white/10 bg-[#0c0d12] px-4 py-2.5 pr-11 font-mono text-sm text-white placeholder-gray-600 outline-none transition focus:border-orange-500/60 focus:ring-2 focus:ring-orange-500/20 disabled:opacity-50"
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                tabIndex={-1}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-300 transition"
              >
                {showPassword ? (
                  <EyeOff className="h-4 w-4" />
                ) : (
                  <Eye className="h-4 w-4" />
                )}
              </button>
            </div>
          </div>

          {/* Error Message */}
          {errorMsg && (
            <div className="flex items-center gap-2 rounded-lg border border-red-500/30 bg-red-500/10 p-2.5 text-xs text-red-400 animate-in fade-in duration-200">
              <ShieldAlert className="h-4 w-4 shrink-0 text-red-400" />
              <span>{errorMsg}</span>
            </div>
          )}

          {/* Submit Button */}
          <button
            type="submit"
            disabled={loading || !token.trim()}
            className="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-orange-500 to-orange-600 py-2.5 font-medium text-white shadow-lg shadow-orange-500/20 transition hover:from-orange-600 hover:to-orange-700 active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-50"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>验证密钥中...</span>
              </>
            ) : (
              <>
                <span>验证并进入系统</span>
                <ArrowRight className="h-4 w-4" />
              </>
            )}
          </button>
        </form>

        <div className="mt-6 text-center text-[11px] text-gray-500">
          Cookie 鉴权会话有效期为 7 天 · 浏览器退出后可在当前设备保持免登
        </div>
      </div>
    </div>
  );
};
