// Types matching idl/v0_1/shot_preview.thrift
export enum TaskStatus {
  ACCEPTED = 1,
  REJECTED = 2,
}

export interface CreateShotPreviewTaskRequest {
  user_id: string;
  prompt: string;
  conversation_id?: string;
  request_id?: string;
}

export interface CreateShotPreviewTaskResponse {
  task_id: string;
  status: TaskStatus;
  request_id?: string;
}

// Types matching idl/v0_1/llm_gateway.thrift
export enum LLMProvider {
  OPENAI = 1,
  ANTHROPIC = 2,
}

export enum LLMKeyOperation {
  SAVE = 1,
  UPDATE = 2,
  DELETE = 3,
}

export interface ManageLLMKeyRequest {
  user_id: string;
  operation: LLMKeyOperation;
  key_id?: string;
  provider?: LLMProvider;
  name?: string;
  api_key?: string;
  base_url?: string;
}

export interface ManageLLMKeyResponse {
  key_id: string;
  provider?: LLMProvider;
}

// Chat UI Domain Models
export interface Message {
  id: string;
  role: "user" | "assistant" | "system";
  content: string;
  timestamp: number;
  thoughts?: string[];
  isThinking?: boolean;
  taskId?: string;
  status?: "sending" | "thought" | "done" | "error";
}

export interface Conversation {
  id: string;
  title: string;
  createdAt: number;
  updatedAt: number;
  messages: Message[];
}
