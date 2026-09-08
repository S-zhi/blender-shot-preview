import { create } from "zustand";
import { Conversation, Message, TaskNodeView, TaskStatus } from "#/api/types";
import { ShotPreviewService } from "#/api/shot-preview-service";
import { ConversationService } from "#/api/conversation-service";
import { useSettingsStore } from "#/stores/settings-store";


interface ConversationState {
  conversations: Conversation[];
  activeConversationId: string | null;
  isGenerating: boolean;

  // Actions
  createNewConversation: () => string;
  hydrateConversations: () => Promise<void>;
  selectConversation: (id: string) => void;
  deleteConversation: (id: string) => void;
  updateConversationTitle: (id: string, title: string) => void;
  sendMessage: (prompt: string) => Promise<void>;
  confirmNodeStep: (messageId: string, nodeId: string, adjustedOutput?: string) => Promise<void>;
  adjustNodeStep: (messageId: string, nodeId: string, outputJson: string) => Promise<void>;
  toggleAutoConfirm: (messageId: string) => void;
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

const activeEventSources = new Map<string, EventSource>();
const activeStreamClosers = new Map<string, () => void>();

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

  hydrateConversations: async () => {
    try {
      const conversations = await ConversationService.list();
      set((state) => ({
        conversations,
        activeConversationId:
          state.activeConversationId && conversations.some((c) => c.id === state.activeConversationId)
            ? state.activeConversationId
            : conversations[0]?.id || null,
      }));
    } catch (error) {
      // Keep the local demo state when the backend is unavailable. Once the
      // SQLite API is online, this branch is replaced by durable history.
      console.warn("Failed to hydrate conversation history:", error);
    }
  },

  selectConversation: (id: string) => {
    set({ activeConversationId: id });
  },

  deleteConversation: (id: string) => {
    ConversationService.remove(id).catch((error) => console.warn("Failed to delete conversation:", error));
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
      status: "connecting",
      thoughts: [],
      nodes: [],
      autoConfirm: false,
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
      // Read the active provider from settings so user_id and key_id align with
      // the credential that was saved via the LLM Key RPC.
      const { userId: settingsUserId, providers, selectedProviderId } = useSettingsStore.getState();
      const activeProvider = providers.find((p) => p.id === selectedProviderId);
      const taskRes = await ShotPreviewService.createShotPreviewTask({
        user_id: settingsUserId,
        prompt,
        conversation_id: convId,
        key_id: activeProvider?.savedKeyId ?? "",
      });
      if (!taskRes.task_id || taskRes.status === TaskStatus.REJECTED) {
        throw new Error("LLM / 后端拒绝创建分镜任务");
      }

      // Update message with taskId
      set((state) => ({
        conversations: state.conversations.map((c) => {
          if (c.id === convId) {
            return {
              ...c,
              messages: c.messages.map((m) =>
                m.id === assistantMsgId
                  ? {
                      ...m,
                      taskId: taskRes.task_id,
                      status: "running",
                      thoughts: [
                        `分镜任务已创建 (Task ID: ${taskRes.task_id})`,
                        "已连接实时执行流，等待 Agent 工作流推进...",
                      ],
                    }
                  : m
              ),
            };
          }
          return c;
        }),
      }));

      // 2. Open Real SSE Stream
      const streamUrl = ShotPreviewService.getTaskStreamUrl(taskRes.task_id);
      const eventSource = new EventSource(streamUrl, { withCredentials: true });
      activeEventSources.set(assistantMsgId, eventSource);
      let settled = false;
      let inactivityTimer: ReturnType<typeof setTimeout> | undefined;

      const closeStream = () => {
        if (settled) return;
        settled = true;
        if (inactivityTimer) clearTimeout(inactivityTimer);
        eventSource.close();
        activeEventSources.delete(assistantMsgId);
        activeStreamClosers.delete(assistantMsgId);
        set({ isGenerating: false });
      };

      const markActivity = () => {
        if (inactivityTimer) clearTimeout(inactivityTimer);
        inactivityTimer = setTimeout(() => {
          if (settled) return;
          set((state) => ({
            conversations: state.conversations.map((c) =>
              c.id !== convId
                ? c
                : {
                    ...c,
                    messages: c.messages.map((m) =>
                      m.id !== assistantMsgId
                        ? m
                        : {
                            ...m,
                            status: "timeout",
                            isThinking: false,
                            content: "### ❌ 分镜生成任务超时\n\n模型服务或实时执行流在规定时间内没有响应。",
                            thoughts: [...(m.thoughts || []), "✖ 实时执行流超时，已停止等待"],
                          }
                    ),
                  }
            ),
          }));
          closeStream();
        }, 120000);
      };

      const pauseInactivityTimeout = () => {
        if (inactivityTimer) clearTimeout(inactivityTimer);
        inactivityTimer = undefined;
      };

      const failStream = (status: "error" | "disconnected", content: string) => {
        if (settled) return;
        set((state) => ({
          conversations: state.conversations.map((c) =>
            c.id !== convId
              ? c
              : {
                  ...c,
                  messages: c.messages.map((m) =>
                    m.id !== assistantMsgId
                      ? m
                      : {
                          ...m,
                          status,
                          isThinking: false,
                          content,
                          thoughts: [...(m.thoughts || []), `✖ ${content.replace(/^### .*\n\n/, "")}`],
                        }
                  ),
                }
          ),
        }));
        closeStream();
      };

      activeStreamClosers.set(assistantMsgId, closeStream);
      markActivity();

      eventSource.addEventListener("task_snapshot", (e) => {
        try {
          markActivity();
          const payload = JSON.parse(e.data);
          const task = payload.task;
          if (!task) return;
          if (task.nodes?.some((n: TaskNodeView) => n.status === "waiting_confirmation")) {
            pauseInactivityTimeout();
          }

          set((state) => ({
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) => {
                    if (m.id === assistantMsgId) {
                      const waitingNode =
                        task.nodes?.find(
                          (n: TaskNodeView) => n.status === "waiting_confirmation"
                        ) || null;
                      const isDone = task.status === "succeeded";
                      const isFailed = task.status === "failed" || task.status === "cancelled";
                      const isWaiting = Boolean(waitingNode);
                      const failureMessage = task.failure?.message || (task.status === "cancelled" ? "任务已取消" : "Agent 或模型服务执行失败");
                      return {
                        ...m,
                        nodes: task.nodes,
                        artifacts: task.artifacts,
                        waitingNode,
                        status: isDone ? "done" : isFailed ? "error" : isWaiting ? "waiting_confirmation" : "running",
                        isThinking: !isDone && !isFailed && !isWaiting,
                        content: isDone
                          ? "### 🎬 分镜预览工作流已全部执行完成\n\n已成功生成分镜视频制品，可点击下方按钮下载。"
                          : isFailed
                            ? `### ❌ 分镜生成任务失败\n\n**错误详情**：${failureMessage}`
                            : m.content,
                      };
                    }
                    return m;
                  }),
                };
              }
              return c;
            }),
          }));
          if (task.status === "succeeded" || task.status === "failed" || task.status === "cancelled") {
            closeStream();
          }
        } catch (err) {
          console.error("Failed to parse task_snapshot", err);
        }
      });

      eventSource.addEventListener("node_started", (e) => {
        try {
          markActivity();
          const payload = JSON.parse(e.data);
          const nodeId = payload.node_id;
          set((state) => ({
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) => {
                    if (m.id === assistantMsgId) {
                      const updatedNodes = m.nodes
                        ? m.nodes.map((n) =>
                            n.node_id === nodeId
                              ? { ...n, status: "running" as const, input: payload.input || n.input }
                              : n
                          )
                        : [{ node_id: nodeId, status: "running" as const, attempts: 1, input: payload.input }];
                      return {
                        ...m,
                        status: "running",
                        isThinking: true,
                        nodes: updatedNodes,
                        thoughts: [...(m.thoughts || []), `▶ 节点 [${nodeId}] 开始执行...`],
                      };
                    }
                    return m;
                  }),
                };
              }
              return c;
            }),
          }));
        } catch (err) {
          console.error("Failed to parse node_started", err);
        }
      });

      eventSource.addEventListener("node_waiting_confirmation", (e) => {
        try {
          markActivity();
          const payload = JSON.parse(e.data);
          const nodeId = payload.node_id;
          pauseInactivityTimeout();
          const waitingNode: TaskNodeView = {
            node_id: nodeId,
            status: "waiting_confirmation",
            attempts: 1,
            input: payload.input,
            output: payload.output,
          };

          set((state) => ({
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) => {
                    if (m.id === assistantMsgId) {
                      const updatedNodes = m.nodes
                        ? m.nodes.map((n) =>
                            n.node_id === nodeId
                              ? {
                                  ...n,
                                  status: "waiting_confirmation" as const,
                                  input: payload.input,
                                  output: payload.output,
                                }
                              : n
                          )
                        : [waitingNode];

                      // Check auto-confirm
                      if (m.autoConfirm && m.taskId) {
                        ShotPreviewService.confirmStep({
                          task_id: m.taskId,
                          node_id: nodeId,
                        }).catch(console.error);
                      }

                      return {
                        ...m,
                        status: m.autoConfirm ? "running" : "waiting_confirmation",
                        isThinking: Boolean(m.autoConfirm),
                        nodes: updatedNodes,
                        waitingNode: m.autoConfirm ? null : waitingNode,
                        thoughts: [
                          ...(m.thoughts || []),
                          `⏸ 节点 [${nodeId}] 执行完成，已暂停并等待确认`,
                        ],
                      };
                    }
                    return m;
                  }),
                };
              }
              return c;
            }),
          }));
        } catch (err) {
          console.error("Failed to parse node_waiting_confirmation", err);
        }
      });

      eventSource.addEventListener("node_output_adjusted", (e) => {
        try {
          markActivity();
          const payload = JSON.parse(e.data);
          const nodeId = payload.node_id;
          set((state) => ({
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) => {
                    if (m.id === assistantMsgId) {
                      return {
                        ...m,
                        nodes: m.nodes?.map((n) =>
                          n.node_id === nodeId ? { ...n, output: payload.output } : n
                        ),
                        thoughts: [...(m.thoughts || []), `✏ 节点 [${nodeId}] 输出数据已更新调整`],
                      };
                    }
                    return m;
                  }),
                };
              }
              return c;
            }),
          }));
        } catch (err) {
          console.error("Failed to parse node_output_adjusted", err);
        }
      });

      eventSource.addEventListener("node_succeeded", (e) => {
        try {
          markActivity();
          const payload = JSON.parse(e.data);
          const nodeId = payload.node_id;
          set((state) => ({
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) => {
                    if (m.id === assistantMsgId) {
                      const updatedNodes = m.nodes?.map((n) =>
                        n.node_id === nodeId
                          ? { ...n, status: "succeeded" as const, output: payload.output || n.output }
                          : n
                      );
                      return {
                        ...m,
                        nodes: updatedNodes,
                        waitingNode: m.waitingNode?.node_id === nodeId ? null : m.waitingNode,
                        thoughts: [...(m.thoughts || []), `✔ 节点 [${nodeId}] 确认通过，已完成`],
                      };
                    }
                    return m;
                  }),
                };
              }
              return c;
            }),
          }));
        } catch (err) {
          console.error("Failed to parse node_succeeded", err);
        }
      });

      eventSource.addEventListener("task_succeeded", (e) => {
        try {
          markActivity();
          const payload = JSON.parse(e.data);
          set((state) => ({
            isGenerating: false,
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) => {
                    if (m.id === assistantMsgId) {
                      return {
                        ...m,
                        status: "done",
                        isThinking: false,
                        waitingNode: null,
                        artifacts: payload.artifacts || m.artifacts || [],
                        content:
                          "### 🎬 分镜预览工作流已全部执行完成\n\n已成功编排各 Agent 节点并完成渲染与视频转码，最终视频文件已生成。请在下方下载查看。",
                        thoughts: [...(m.thoughts || []), "🎉 全流程执行成功，已发布分镜视频制品！"],
                      };
                    }
                    return m;
                  }),
                };
              }
              return c;
            }),
          }));
          closeStream();
        } catch (err) {
          console.error("Failed to parse task_succeeded", err);
          closeStream();
        }
      });

      eventSource.addEventListener("task_failed", (e) => {
        try {
          markActivity();
          const payload = JSON.parse(e.data);
          set((state) => ({
            isGenerating: false,
            conversations: state.conversations.map((c) => {
              if (c.id === convId) {
                return {
                  ...c,
                  messages: c.messages.map((m) => {
                    if (m.id === assistantMsgId) {
                      return {
                        ...m,
                        status: "error",
                        isThinking: false,
                        content: `### ❌ 分镜生成任务失败\n\n**错误详情**：${payload.error || "执行过程中发生异常"}`,
                        thoughts: [...(m.thoughts || []), `✖ 任务失败${payload.error_code ? ` [${payload.error_code}]` : ""}: ${payload.error || "未知异常"}`],
                      };
                    }
                    return m;
                  }),
                };
              }
              return c;
            }),
          }));
          closeStream();
        } catch (err) {
          console.error("Failed to parse task_failed", err);
          closeStream();
        }
      });

      eventSource.onerror = (err) => {
        console.warn("EventSource error:", err);
        failStream("disconnected", "### ❌ 实时执行流已断开\n\n无法继续接收 Agent 状态，请检查后端服务和模型连接。 ");
      };
    } catch {
      set((state) => ({
        isGenerating: false,
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
      // isGenerating will be closed by closeStream when task finishes or on error
    }
  },

  confirmNodeStep: async (messageId: string, nodeId: string, adjustedOutput?: string) => {
    const { conversations, activeConversationId } = get();
    const conv = conversations.find((c) => c.id === activeConversationId);
    const msg = conv?.messages.find((m) => m.id === messageId);
    if (!msg || !msg.taskId) return;

    await ShotPreviewService.confirmStep({
      task_id: msg.taskId,
      node_id: nodeId,
      adjusted_output: adjustedOutput,
    });

    set((state) => ({
      conversations: state.conversations.map((c) => {
        if (c.id === activeConversationId) {
          return {
            ...c,
            messages: c.messages.map((m) => {
              if (m.id === messageId) {
                return {
                  ...m,
                  waitingNode: null,
                  nodes: m.nodes?.map((n) =>
                    n.node_id === nodeId ? { ...n, status: "succeeded" as const } : n
                  ),
                };
              }
              return m;
            }),
          };
        }
        return c;
      }),
    }));
  },

  adjustNodeStep: async (messageId: string, nodeId: string, outputJson: string) => {
    const { conversations, activeConversationId } = get();
    const conv = conversations.find((c) => c.id === activeConversationId);
    const msg = conv?.messages.find((m) => m.id === messageId);
    if (!msg || !msg.taskId) return;

    await ShotPreviewService.adjustStep({
      task_id: msg.taskId,
      node_id: nodeId,
      output_json: outputJson,
    });

    set((state) => ({
      conversations: state.conversations.map((c) => {
        if (c.id === activeConversationId) {
          return {
            ...c,
            messages: c.messages.map((m) => {
              if (m.id === messageId) {
                return {
                  ...m,
                  nodes: m.nodes?.map((n) =>
                    n.node_id === nodeId ? { ...n, output: outputJson } : n
                  ),
                  waitingNode:
                    m.waitingNode?.node_id === nodeId
                      ? { ...m.waitingNode, output: outputJson }
                      : m.waitingNode,
                };
              }
              return m;
            }),
          };
        }
        return c;
      }),
    }));
  },

  toggleAutoConfirm: (messageId: string) => {
    set((state) => ({
      conversations: state.conversations.map((c) => {
        if (c.id === state.activeConversationId) {
          return {
            ...c,
            messages: c.messages.map((m) => {
              if (m.id === messageId) {
                const nextAuto = !m.autoConfirm;
                if (nextAuto && m.waitingNode && m.taskId) {
                  ShotPreviewService.confirmStep({
                    task_id: m.taskId,
                    node_id: m.waitingNode.node_id,
                  }).catch(console.error);
                }
                return { ...m, autoConfirm: nextAuto };
              }
              return m;
            }),
          };
        }
        return c;
      }),
    }));
  },

  stopGenerating: () => {
    activeStreamClosers.forEach((close) => close());
    activeEventSources.forEach((source) => source.close());
    activeEventSources.clear();
    activeStreamClosers.clear();
    set({ isGenerating: false });
  },
}));
