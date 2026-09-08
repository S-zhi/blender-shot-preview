import { create } from "zustand";

export type ActiveView = "home" | "chat" | "automate" | "custom" | "templates";

interface NavigationState {
  activeView: ActiveView;
  setActiveView: (view: ActiveView) => void;
}

export const useNavigationStore = create<NavigationState>((set) => ({
  activeView: "home",
  setActiveView: (activeView) => set({ activeView }),
}));
