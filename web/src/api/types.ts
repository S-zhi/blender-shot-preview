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

export interface AssetMetadata {
  poly_count?: number;
  vertex_count?: number;
  rig_type?: string;
  has_face_rig?: boolean;
  texture_maps?: string[];
  duration_frames?: number;
  fps?: number;
  fov_range?: string;
  camera_trajectory?: { x: number; y: number; z: number; targetX: number; targetY: number; targetZ: number }[];
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
  metadata?: AssetMetadata;
  preview_images?: { title: string; url: string }[];
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

// Pipeline & Task Types
export type NodeStatusType =
  | "pending"
  | "running"
  | "waiting_confirmation"
  | "succeeded"
  | "failed"
  | "cancelled";

export interface TaskNodeView {
  node_id: string;
  status: NodeStatusType;
  attempts: number;
  input?: string;
  output?: string;
  started_at?: string;
  finished_at?: string;
  failure?: {
    code: string;
    message: string;
    retryable: boolean;
  };
}

export interface TaskArtifactView {
  type: string;
  uri: string;
  name: string;
}

export interface ShotPreviewTaskView {
  task_id: string;
  status: string;
  workflow_id: string;
  workflow_version: string;
  nodes: TaskNodeView[];
  artifacts: TaskArtifactView[];
  created_at: string;
  updated_at: string;
  finished_at?: string;
}

export interface PipelineEvent {
  task_id: string;
  type: string;
  node_id?: string;
  status?: string;
  input?: string;
  output?: string;
  error?: string;
  artifacts?: TaskArtifactView[];
  timestamp: string;
}

export interface ConfirmStepRequest {
  user_id?: string;
  task_id: string;
  node_id: string;
  adjusted_output?: string;
}

export interface AdjustStepRequest {
  user_id?: string;
  task_id: string;
  node_id: string;
  output_json: string;
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
  nodes?: TaskNodeView[];
  artifacts?: TaskArtifactView[];
  waitingNode?: TaskNodeView | null;
  autoConfirm?: boolean;
}

export interface Conversation {
  id: string;
  title: string;
  createdAt: number;
  updatedAt: number;
  messages: Message[];
}
