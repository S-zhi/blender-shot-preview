import { create } from "zustand";

interface SidebarState {
  isOpen: boolean;
  searchQuery: string;
  toggleSidebar: () => void;
  setIsOpen: (isOpen: boolean) => void;
  setSearchQuery: (query: string) => void;
}

export const useSidebarStore = create<SidebarState>((set) => ({
  isOpen: true,
  searchQuery: "",
  toggleSidebar: () => set((state) => ({ isOpen: !state.isOpen })),
  setIsOpen: (isOpen: boolean) => set({ isOpen }),
  setSearchQuery: (searchQuery: string) => set({ searchQuery }),
}));
