import React from "react";
import { Sidebar } from "#/components/sidebar/sidebar";
import { ChatArea } from "#/components/conversation/chat-area";
import { AssetDashboard } from "#/components/assets/asset-dashboard";
import { SettingsView } from "#/components/settings/settings-view";
import { useNavigationStore } from "#/stores/navigation-store";

export const App: React.FC = () => {
  const { activeView } = useNavigationStore();

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
