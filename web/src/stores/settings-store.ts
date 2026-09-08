import { create } from "zustand";
import { LLMProvider, LLMKeyOperation } from "#/api/types";
import { LLMKeyService } from "#/api/llm-gateway-service";

interface SettingsState {
  isModalOpen: boolean;
  userId: string;
  provider: LLMProvider;
  modelName: string;
  apiKey: string;
  baseUrl: string;
  isSaving: boolean;
  savedKeyId: string | null;

  setModalOpen: (open: boolean) => void;
  setUserId: (id: string) => void;
  setProvider: (p: LLMProvider) => void;
  setModelName: (name: string) => void;
  setApiKey: (key: string) => void;
  setBaseUrl: (url: string) => void;

  saveKey: () => Promise<boolean>;
}

export const useSettingsStore = create<SettingsState>((set, get) => ({
  isModalOpen: false,
  userId: "user_default_001",
  provider: LLMProvider.OPENAI,
  modelName: "gpt-4o",
  apiKey: "",
  baseUrl: "https://api.openai.com/v1",
  isSaving: false,
  savedKeyId: null,

  setModalOpen: (isModalOpen) => set({ isModalOpen }),
  setUserId: (userId) => set({ userId }),
  setProvider: (provider) => {
    const defaultUrl =
      provider === LLMProvider.OPENAI
        ? "https://api.openai.com/v1"
        : "https://api.anthropic.com/v1";
    set({ provider, baseUrl: defaultUrl });
  },
  setModelName: (modelName) => set({ modelName }),
  setApiKey: (apiKey) => set({ apiKey }),
  setBaseUrl: (baseUrl) => set({ baseUrl }),

  saveKey: async () => {
    const { userId, provider, modelName, apiKey, baseUrl, savedKeyId } = get();
    set({ isSaving: true });
    try {
      const res = await LLMKeyService.manageLLMKey({
        user_id: userId,
        operation: savedKeyId ? LLMKeyOperation.UPDATE : LLMKeyOperation.SAVE,
        key_id: savedKeyId || undefined,
        provider,
        name: modelName,
        api_key: apiKey,
        base_url: baseUrl,
      });

      set({ savedKeyId: res.key_id, isSaving: false, isModalOpen: false });
      return true;
    } catch {
      set({ isSaving: false });
      return false;
    }
  },
}));
