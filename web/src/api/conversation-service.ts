import { request } from "./client";
import { Conversation, Message, TaskArtifactView, TaskNodeView } from "./types";

interface ServerMessage {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  status?: string;
  task_id?: string;
  thoughts?: string[];
  nodes?: TaskNodeView[];
  artifacts?: TaskArtifactView[];
  waiting_node?: TaskNodeView | null;
  created_at: string;
  updated_at: string;
}

interface ServerConversation {
  id: string;
  user_id: string;
  title: string;
  created_at: string;
  updated_at: string;
  messages: ServerMessage[];
}

const mapMessage = (message: ServerMessage): Message => ({
  id: message.id,
  role: message.role,
  content: message.content,
  timestamp: Date.parse(message.created_at),
  thoughts: message.thoughts,
  taskId: message.task_id || undefined,
  status: message.status === "done" || message.status === "error" ? message.status : "thought",
  nodes: message.nodes,
  artifacts: message.artifacts,
  waitingNode: message.waiting_node,
  isThinking: message.status === "thought",
  autoConfirm: false,
});

const mapConversation = (conversation: ServerConversation): Conversation => ({
  id: conversation.id,
  title: conversation.title,
  createdAt: Date.parse(conversation.created_at),
  updatedAt: Date.parse(conversation.updated_at),
  messages: conversation.messages.map(mapMessage),
});

export class ConversationService {
  static async list(userId = "default_user_001"): Promise<Conversation[]> {
    const response = await request<{ conversations: ServerConversation[] }>(
      `/api/v0_1/conversations?user_id=${encodeURIComponent(userId)}`
    );
    return (response.conversations || []).map(mapConversation);
  }

  static async get(id: string, userId = "default_user_001"): Promise<Conversation> {
    const response = await request<ServerConversation>(
      `/api/v0_1/conversations?user_id=${encodeURIComponent(userId)}&conversation_id=${encodeURIComponent(id)}`
    );
    return mapConversation(response);
  }

  static async remove(id: string, userId = "default_user_001"): Promise<void> {
    await request(`/api/v0_1/conversations?user_id=${encodeURIComponent(userId)}&conversation_id=${encodeURIComponent(id)}`, { method: "DELETE" });
  }
}
