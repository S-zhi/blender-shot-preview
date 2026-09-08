import { create } from "zustand";

export type ActiveView = "chat" | "assets" | "settings";

interface NavigationState {
  activeView: ActiveView;
  setActiveView: (view: ActiveView) => void;
}

export const useNavigationStore = create<NavigationState>((set) => ({
  // Default entry: directly enters "chat" (新建对话) as requested
  activeView: "chat",
  setActiveView: (activeView) => set({ activeView }),
}));
