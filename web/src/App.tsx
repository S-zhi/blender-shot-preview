import React, { useEffect, useState } from "react";
import { Sidebar } from "#/components/sidebar/sidebar";
import { ChatArea } from "#/components/conversation/chat-area";
import { AssetDashboard } from "#/components/assets/asset-dashboard";
import { SettingsView } from "#/components/settings/settings-view";
import { LoginPage } from "#/components/auth/login-page";
import { useNavigationStore } from "#/stores/navigation-store";
import { useConversationStore } from "#/stores/conversation-store";
import { checkAuthStatus, authEvents } from "#/api/client";
import { Loader2 } from "lucide-react";

export const App: React.FC = () => {
  const { activeView } = useNavigationStore();
  const hydrateConversations = useConversationStore((state) => state.hydrateConversations);

  const [authChecked, setAuthChecked] = useState(false);
  const [authenticated, setAuthenticated] = useState(false);

  useEffect(() => {
    let active = true;

    const runAuthCheck = async () => {
      const status = await checkAuthStatus();
      if (!active) return;
      setAuthenticated(status.authenticated);
      setAuthChecked(true);
      if (status.authenticated) {
        void hydrateConversations();
      }
    };

    void runAuthCheck();

    const handleUnauthorized = () => {
      setAuthenticated(false);
    };

    const handleLogin = () => {
      setAuthenticated(true);
      void hydrateConversations();
    };

    const handleLogout = () => {
      setAuthenticated(false);
    };

    authEvents.addEventListener("auth:unauthorized", handleUnauthorized);
    authEvents.addEventListener("auth:login", handleLogin);
    authEvents.addEventListener("auth:logout", handleLogout);

    return () => {
      active = false;
      authEvents.removeEventListener("auth:unauthorized", handleUnauthorized);
      authEvents.removeEventListener("auth:login", handleLogin);
      authEvents.removeEventListener("auth:logout", handleLogout);
    };
  }, [hydrateConversations]);

  if (!authChecked) {
    return (
      <div className="flex h-screen w-screen items-center justify-center bg-[#07080b] text-gray-400 font-sans">
        <div className="flex flex-col items-center gap-3">
          <Loader2 className="h-8 w-8 animate-spin text-orange-500" />
          <span className="text-xs tracking-wider uppercase text-gray-500">正在检查访问凭证...</span>
        </div>
      </div>
    );
  }

  if (!authenticated) {
    return <LoginPage onSuccess={() => setAuthenticated(true)} />;
  }

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[#090b0e] text-content font-sans">
      {/* 1. OpenHands-style Left Sidebar (Streamlined to 3 destinations) */}
      <Sidebar />

      {/* 2. Main Workspace: Exactly 3 Views (Chat, Assets, Settings) */}
      <main className="flex-1 flex flex-col h-full min-w-0 overflow-hidden relative">
        {activeView === "chat" && <ChatArea />}
        {activeView === "assets" && <AssetDashboard />}
        {activeView === "settings" && <SettingsView />}
      </main>
    </div>
  );
};

export default App;
