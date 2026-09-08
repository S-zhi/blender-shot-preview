import React from "react";
import { Sidebar } from "#/components/sidebar/sidebar";
import { ChatArea } from "#/components/conversation/chat-area";
import { LLMKeyModal } from "#/components/settings/llm-key-modal";

export const App: React.FC = () => {
  return (
    <div className="flex h-screen w-screen overflow-hidden bg-background text-gray-100">
      {/* OpenHands-style Left Sidebar */}
      <Sidebar />

      {/* Main Conversation & Viewport Area */}
      <main className="flex-1 flex flex-col h-full min-w-0 overflow-hidden relative">
        <ChatArea />
      </main>

      {/* LLM Key / Credential Settings Modal */}
      <LLMKeyModal />
    </div>
  );
};

export default App;
