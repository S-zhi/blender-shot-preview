import React from "react";
import { Sidebar } from "#/components/sidebar/sidebar";
import { HomeCanvas } from "#/components/canvas/home-canvas";
import { AutomateDashboard } from "#/components/automate/automate-dashboard";
import { ChatArea } from "#/components/conversation/chat-area";
import { LLMKeyModal } from "#/components/settings/llm-key-modal";
import { useNavigationStore } from "#/stores/navigation-store";

export const App: React.FC = () => {
  const { activeView } = useNavigationStore();

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[#0a0c0e] text-content font-sans">
      {/* 1. OpenHands-style Left Sidebar (Pixel-matched with Screenshot 1) */}
      <Sidebar />

      {/* 2. Main Viewport Area */}
      <main className="flex-1 flex flex-col h-full min-w-0 overflow-hidden relative">
        {activeView === "home" && <HomeCanvas />}
        {activeView === "automate" && <AutomateDashboard />}
        {activeView === "chat" && <ChatArea />}
        {activeView === "custom" && <HomeCanvas />}
        {activeView === "templates" && <AutomateDashboard />}
      </main>

      {/* 3. LLM Key / Credential Settings Modal */}
      <LLMKeyModal />
    </div>
  );
};

export default App;
