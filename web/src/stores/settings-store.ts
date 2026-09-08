import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { LLMProvider, LLMKeyOperation } from "#/api/types";
import { LLMKeyService } from "#/api/llm-gateway-service";
import { testLLMConnection, TestConnectionResult } from "#/api/llm-test";

export interface ProviderItem {
  id: string; // 英文标识，如 "openai", "azure_openai", "deepseek"
  label: string; // 显示名称
  providerType: "openai" | "azure_openai" | "custom";
  modelName: string; // 用户自由指定的模型名称，不硬编码
  apiKey: string;
  baseUrl: string;
  apiVersion?: string; // Azure 专用
  headers?: Record<string, string>;
  savedKeyId?: string;
  lastTestResult?: TestConnectionResult;
}

interface SettingsState {
  isModalOpen: boolean;
  userId: string;
  activeTab: string; // "openai" | "azure_openai" | "render_preferences" | ...
  providers: ProviderItem[];
  selectedProviderId: string;

  // Blender Rendering Preferences
  renderEngine: "eevee" | "cycles";
  renderResolution: "720p" | "1080p" | "4K";

  // Actions
  isSaving: boolean;
  isTesting: boolean;

  setModalOpen: (open: boolean) => void;
  setUserId: (id: string) => void;
  setActiveTab: (tab: string) => void;
  setSelectedProviderId: (id: string) => void;
  setRenderEngine: (engine: "eevee" | "cycles") => void;
  setRenderResolution: (res: "720p" | "1080p" | "4K") => void;

  updateProvider: (id: string, partial: Partial<ProviderItem>) => void;
  addProvider: (provider: ProviderItem) => void;
  removeProvider: (id: string) => void;

  testConnection: (id: string) => Promise<TestConnectionResult>;
  saveCurrentProviderKey: () => Promise<boolean>;
  ensureActiveProviderKey: () => Promise<string | undefined>;

  // Backwards compatibility getters
  provider: LLMProvider;
  modelName: string;
  apiKey: string;
  baseUrl: string;
  savedKeyId: string | null;
  setProvider: (p: LLMProvider) => void;
  setModelName: (name: string) => void;
  setApiKey: (key: string) => void;
  setBaseUrl: (url: string) => void;
  saveKey: () => Promise<boolean>;
}

const DEFAULT_PROVIDERS: ProviderItem[] = [
  {
    id: "openai",
    label: "OpenAI",
    providerType: "openai",
    modelName: "gpt-4o",
    apiKey: "",
    baseUrl: "https://api.openai.com/v1",
  },
  {
    id: "azure_openai",
    label: "Azure OpenAI",
    providerType: "azure_openai",
    modelName: "gpt-4o",
    apiKey: "",
    baseUrl: "https://your-resource.openai.azure.com",
    apiVersion: "2024-02-15-preview",
  },
];

// Multiple open tabs can write the whole persisted Zustand state. Merge the
// provider list at the storage boundary so an older tab cannot overwrite a
// provider created in a newer tab with its stale default-only list.
const settingsStorage: Storage = {
  get length() {
    return localStorage.length;
  },
  clear: () => localStorage.clear(),
  getItem: (name) => localStorage.getItem(name),
  key: (index) => localStorage.key(index),
  removeItem: (name) => localStorage.removeItem(name),
  setItem: (name, value) => {
    try {
      const incoming = JSON.parse(value);
      const existing = JSON.parse(localStorage.getItem(name) || "null");
      const existingProviders = existing?.state?.providers;
      const incomingProviders = incoming?.state?.providers;

      if (Array.isArray(existingProviders) && Array.isArray(incomingProviders)) {
        const providers = new Map<string, ProviderItem>();
        for (const provider of existingProviders) {
          if (provider && typeof provider.id === "string") providers.set(provider.id, provider);
        }
        for (const provider of incomingProviders) {
          if (provider && typeof provider.id === "string") providers.set(provider.id, provider);
        }
        incoming.state.providers = [...providers.values()];
      }

      localStorage.setItem(name, JSON.stringify(incoming));
    } catch {
      // Fall back to the normal write for malformed or unavailable storage.
      localStorage.setItem(name, value);
    }
  },
};

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set, get) => ({
      isModalOpen: false,
      userId: "default_user_001",
      activeTab: "openai",
      providers: DEFAULT_PROVIDERS,
      selectedProviderId: "openai",

      renderEngine: "eevee",
      renderResolution: "1080p",

      isSaving: false,
      isTesting: false,

      setModalOpen: (isModalOpen) => set({ isModalOpen }),
      setUserId: (userId) => set({ userId }),
      setActiveTab: (activeTab) => set({ activeTab }),
      setSelectedProviderId: (id) => {
        set({ selectedProviderId: id, activeTab: id });
      },
      setRenderEngine: (renderEngine) => set({ renderEngine }),
      setRenderResolution: (renderResolution) => set({ renderResolution }),

      updateProvider: (id, partial) => {
        const { providers } = get();
        const next = providers.map((p) => (p.id === id ? { ...p, ...partial } : p));
        set({ providers: next });
      },

      addProvider: (newProvider) => {
        const { providers } = get();
        // Prevent duplicate IDs
        const exists = providers.some((p) => p.id === newProvider.id);
        const id = exists ? `${newProvider.id}_${Date.now().toString().slice(-4)}` : newProvider.id;
        const item = { ...newProvider, id };
        set({
          providers: [...providers, item],
          selectedProviderId: item.id,
          activeTab: item.id,
        });
      },

      removeProvider: (id) => {
        const { providers, selectedProviderId } = get();
        if (providers.length <= 1) return;
        const filtered = providers.filter((p) => p.id !== id);
        const nextSelected = selectedProviderId === id ? filtered[0].id : selectedProviderId;
        set({
          providers: filtered,
          selectedProviderId: nextSelected,
          activeTab: nextSelected,
        });
      },

      testConnection: async (id) => {
        const { providers } = get();
        const target = providers.find((p) => p.id === id);
        if (!target) {
          return { success: false, message: "未找到对应的提供商配置" };
        }

        set({ isTesting: true });
        const result = await testLLMConnection({
          providerType: target.providerType,
          baseUrl: target.baseUrl,
          apiKey: target.apiKey,
          modelName: target.modelName,
          apiVersion: target.apiVersion,
        });

        // Update target test result
        const next = providers.map((p) =>
          p.id === id ? { ...p, lastTestResult: result } : p
        );
        set({ providers: next, isTesting: false });
        return result;
      },

      saveCurrentProviderKey: async () => {
        const { userId, providers, selectedProviderId } = get();
        const current = providers.find((p) => p.id === selectedProviderId);
        if (!current) return false;

        set({ isSaving: true });
        try {
          let res;
          try {
            res = await LLMKeyService.manageLLMKey({
              user_id: userId,
              operation: current.savedKeyId ? LLMKeyOperation.UPDATE : LLMKeyOperation.SAVE,
              key_id: current.savedKeyId || undefined,
              provider: LLMProvider.OPENAI, // Both OpenAI and Azure use OpenAI compatible protocol
              name: current.modelName || current.id,
              api_key: current.apiKey,
              base_url: current.baseUrl,
            });
          } catch (updateErr) {
            // If UPDATE failed (e.g. server restarted and cleared in-memory store), retry with SAVE
            if (current.savedKeyId) {
              res = await LLMKeyService.manageLLMKey({
                user_id: userId,
                operation: LLMKeyOperation.SAVE,
                provider: LLMProvider.OPENAI,
                name: current.modelName || current.id,
                api_key: current.apiKey,
                base_url: current.baseUrl,
              });
            } else {
              throw updateErr;
            }
          }

          const next = get().providers.map((p) =>
            p.id === selectedProviderId ? { ...p, savedKeyId: res.key_id } : p
          );
          set({ providers: next, isSaving: false, isModalOpen: false });
          return true;
        } catch {
          set({ isSaving: false });
          return false;
        }
      },

      ensureActiveProviderKey: async () => {
        const { userId, providers, selectedProviderId } = get();
        const current = providers.find((p) => p.id === selectedProviderId);
        if (!current || !current.apiKey) return undefined;

        try {
          let res;
          try {
            res = await LLMKeyService.manageLLMKey({
              user_id: userId,
              operation: current.savedKeyId ? LLMKeyOperation.UPDATE : LLMKeyOperation.SAVE,
              key_id: current.savedKeyId || undefined,
              provider: LLMProvider.OPENAI,
              name: current.modelName || current.id,
              api_key: current.apiKey,
              base_url: current.baseUrl,
            });
          } catch {
            res = await LLMKeyService.manageLLMKey({
              user_id: userId,
              operation: LLMKeyOperation.SAVE,
              provider: LLMProvider.OPENAI,
              name: current.modelName || current.id,
              api_key: current.apiKey,
              base_url: current.baseUrl,
            });
          }

          if (res?.key_id) {
            const next = get().providers.map((p) =>
              p.id === selectedProviderId ? { ...p, savedKeyId: res.key_id } : p
            );
            set({ providers: next });
            return res.key_id;
          }
        } catch (err) {
          console.warn("Auto-sync provider key to backend encountered error:", err);
        }
        return current.savedKeyId;
      },

      // Backwards compatibility mappings for older components
      get provider() {
        return LLMProvider.OPENAI;
      },
      get modelName() {
        const current = get().providers.find((p) => p.id === get().selectedProviderId);
        return current?.modelName || "gpt-4o";
      },
      get apiKey() {
        const current = get().providers.find((p) => p.id === get().selectedProviderId);
        return current?.apiKey || "";
      },
      get baseUrl() {
        const current = get().providers.find((p) => p.id === get().selectedProviderId);
        return current?.baseUrl || "https://api.openai.com/v1";
      },
      get savedKeyId() {
        const current = get().providers.find((p) => p.id === get().selectedProviderId);
        return current?.savedKeyId || null;
      },

      setProvider: (_p) => {},
      setModelName: (name) => {
        const { selectedProviderId, updateProvider } = get();
        updateProvider(selectedProviderId, { modelName: name });
      },
      setApiKey: (key) => {
        const { selectedProviderId, updateProvider } = get();
        updateProvider(selectedProviderId, { apiKey: key });
      },
      setBaseUrl: (url) => {
        const { selectedProviderId, updateProvider } = get();
        updateProvider(selectedProviderId, { baseUrl: url });
      },
      saveKey: async () => {
        return get().saveCurrentProviderKey();
      },
    }),
    {
      name: "bspe_settings_storage",
      storage: createJSONStorage(() => settingsStorage),
      partialize: (state) => ({
        userId: state.userId,
        providers: state.providers,
        selectedProviderId: state.selectedProviderId,
        activeTab: state.activeTab,
        renderEngine: state.renderEngine,
        renderResolution: state.renderResolution,
      }),
      merge: (persistedState: any, currentState) => {
        if (!persistedState) return currentState;
        // Merge persisted providers with default providers if needed
        const persistedProviders = Array.isArray(persistedState.providers)
          ? persistedState.providers
          : currentState.providers;
        return {
          ...currentState,
          ...persistedState,
          providers: persistedProviders.length > 0 ? persistedProviders : currentState.providers,
        };
      },
    }
  )
);
