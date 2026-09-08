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

// Types matching idl/v0_1/asset_service.thrift
export enum AssetType {
  MODEL_3D = 1,
  SHOT_PRESET = 2,
  MATERIAL = 3,
  ANIMATION = 4,
}

export enum AssetStatus {
  AVAILABLE = 1,
  PROCESSING = 2,
  ARCHIVED = 3,
}

export interface AssetView {
  asset_id: string;
  name: string;
  asset_type: AssetType;
  file_format: string;
  file_size_bytes: number;
  storage_uri: string;
  thumbnail_uri?: string;
  status: AssetStatus;
  tags?: string[];
  created_at: string;
  updated_at: string;
  description?: string;
}

export interface ListAssetsRequest {
  user_id: string;
  asset_type?: AssetType;
  query_keyword?: string;
  page_size?: number;
  page_num?: number;
}

export interface ListAssetsResponse {
  assets: AssetView[];
  total_count: number;
}

export interface RegisterAssetRequest {
  user_id: string;
  name: string;
  asset_type: AssetType;
  file_format: string;
  file_size_bytes: number;
  storage_uri: string;
  description?: string;
  tags?: string[];
}

export interface RegisterAssetResponse {
  asset_id: string;
  status: AssetStatus;
}

export interface DeleteAssetRequest {
  user_id: string;
  asset_id: string;
}

export interface DeleteAssetResponse {
  success: boolean;
}

export interface GetAssetStatsRequest {
  user_id: string;
}

export interface GetAssetStatsResponse {
  total_assets: number;
  total_models: number;
  total_presets: number;
  total_storage_bytes: number;
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
