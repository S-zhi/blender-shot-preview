import { create } from "zustand";
import { Conversation, Message } from "#/api/types";
import { ShotPreviewService } from "#/api/shot-preview-service";
import { mockAdapter } from "#/api/mock-adapter";

interface ConversationState {
  conversations: Conversation[];
  activeConversationId: string | null;
  isGenerating: boolean;

  // Actions
  createNewConversation: () => string;
  selectConversation: (id: string) => void;
  deleteConversation: (id: string) => void;
  updateConversationTitle: (id: string, title: string) => void;
  sendMessage: (prompt: string) => Promise<void>;
  stopGenerating: () => void;
}

const initialConversations: Conversation[] = [
  {
    id: "conv_default_1",
    title: "赛博朋克雨夜街道分镜规划",
    createdAt: Date.now() - 3600 * 1000 * 2,
    updatedAt: Date.now() - 3600 * 1000 * 2,
    messages: [
      {
        id: "msg_1",
        role: "user",
        content: "请为雨夜赛博朋克街道设计一个由远及近的推镜头，重点展示霓虹灯倒影与主角雨伞反光。",
        timestamp: Date.now() - 3600 * 1000 * 2,
      },
      {
        id: "msg_2",
        role: "assistant",
        content: "### 分镜预览设计方案：\n- **镜头编号**：Shot-001\n- **机位构图**：俯角 45° 广角慢推至平视特写\n- **焦距**：35mm -> 85mm 动态平滑变焦\n- **Blender 渲染设置**：开启 SSR 屏幕空间反射，Eevee 实时光照已加载。\n\n分镜预览任务已就绪，可随时在面板中触发渲染。",
        timestamp: Date.now() - 3600 * 1000 * 2 + 1000,
        thoughts: [
          "分析提示词：雨夜、赛博朋克、霓虹倒影、特写推镜头",
          "计算 Blender 摄影机起始坐标 (X: 12.0, Y: -15.0, Z: 8.0) 到终点坐标",
          "构建场景粗模参数完成",
        ],
      },
    ],
  },
  {
    id: "conv_default_2",
    title: "产品展示 360度环绕镜头",
    createdAt: Date.now() - 3600 * 1000 * 26,
    updatedAt: Date.now() - 3600 * 1000 * 26,
    messages: [
      {
        id: "msg_201",
        role: "user",
        content: "生成机械手表 360 度圆周环绕预览，保持对焦在表盘刻度上。",
        timestamp: Date.now() - 3600 * 1000 * 26,
      },
      {
        id: "msg_202",
        role: "assistant",
        content: "已生成环绕圆形轨迹约束 (Damped Track Constraint)，目标对象为刻度中心。总帧数 180 帧。",
        timestamp: Date.now() - 3600 * 1000 * 26 + 1200,
      },
    ],
  },
];

export const useConversationStore = create<ConversationState>((set, get) => ({
  conversations: initialConversations,
  activeConversationId: initialConversations[0].id,
  isGenerating: false,

  createNewConversation: () => {
    const newId = `conv_${Date.now()}`;
    const newConv: Conversation = {
      id: newId,
      title: "新的分镜会话",
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [],
    };
    set((state) => ({
      conversations: [newConv, ...state.conversations],
      activeConversationId: newId,
    }));
    return newId;
  },

  selectConversation: (id: string) => {
    set({ activeConversationId: id });
  },

  deleteConversation: (id: string) => {
    set((state) => {
      const remaining = state.conversations.filter((c) => c.id !== id);
      const nextActiveId =
        state.activeConversationId === id
          ? remaining.length > 0
            ? remaining[0].id
            : null
          : state.activeConversationId;
      return {
        conversations: remaining,
        activeConversationId: nextActiveId,
      };
    });
  },

  updateConversationTitle: (id: string, title: string) => {
    set((state) => ({
      conversations: state.conversations.map((c) =>
        c.id === id ? { ...c, title, updatedAt: Date.now() } : c
      ),
    }));
  },

  sendMessage: async (prompt: string) => {
    const { activeConversationId, conversations } = get();
    let convId = activeConversationId;

    if (!convId || !conversations.find((c) => c.id === convId)) {
      convId = get().createNewConversation();
    }

    const userMsgId = `msg_${Date.now()}_u`;
    const userMessage: Message = {
      id: userMsgId,
      role: "user",
      content: prompt,
      timestamp: Date.now(),
    };

    const assistantMsgId = `msg_${Date.now()}_a`;
    const assistantMessage: Message = {
      id: assistantMsgId,
      role: "assistant",
      content: "",
      timestamp: Date.now(),
      isThinking: true,
      thoughts: [],
    };

    // Append user message & placeholder assistant message
    set((state) => ({
      isGenerating: true,
      conversations: state.conversations.map((c) => {
        if (c.id === convId) {
          const isFirstMessage = c.messages.length === 0;
          const updatedTitle = isFirstMessage
            ? prompt.slice(0, 20) + (prompt.length > 20 ? "..." : "")
            : c.title;
          return {
            ...c,
            title: updatedTitle,
            updatedAt: Date.now(),
            messages: [...c.messages, userMessage, assistantMessage],
          };
        }
        return c;
      }),
    }));

    try {
      // 1. Call ShotPreviewService (RPC Task creation)
      const taskRes = await ShotPreviewService.createShotPreviewTask({
        user_id: "default_user",
        prompt,
        conversation_id: convId,
      });

      // 2. Simulate streaming thought and response like OpenHands
      await mockAdapter.simulateAgentStream(
        prompt,
        (thought) => {
          set((state) => ({
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) =>
                    m.id === assistantMsgId
                      ? {
                          ...m,
                          thoughts: [...(m.thoughts || []), thought],
                        }
                      : m
                  ),
                };
              }
              return c;
            }),
          }));
        },
        (content) => {
          set((state) => ({
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) =>
                    m.id === assistantMsgId
                      ? {
                          ...m,
                          content,
                          isThinking: false,
                          taskId: taskRes.task_id,
                          status: "done",
                        }
                      : m
                  ),
                };
              }
              return c;
            }),
          }));
        }
      );
    } catch {
      set((state) => ({
        conversations: state.conversations.map((c) => {
          if (c.id === convId) {
            return {
              ...c,
              messages: c.messages.map((m) =>
                m.id === assistantMsgId
                  ? {
                      ...m,
                      content: "处理分镜预览任务时发生异常，请检查后端服务连接。",
                      isThinking: false,
                      status: "error",
                    }
                  : m
              ),
            };
          }
          return c;
        }),
      }));
    } finally {
      set({ isGenerating: false });
    }
  },

  stopGenerating: () => {
    set({ isGenerating: false });
  },
}));
