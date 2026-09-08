import { request, USE_MOCK_API } from "./client";
import { mockAdapter } from "./mock-adapter";
import {
  CreateShotPreviewTaskRequest,
  CreateShotPreviewTaskResponse,
  ConfirmStepRequest,
  AdjustStepRequest,
} from "./types";

export class ShotPreviewService {
  /**
   * 对应 idl/v0_1/shot_preview.thrift:
   * CreateShotPreviewTask(1: CreateShotPreviewTaskRequest request)
   */
  static async createShotPreviewTask(
    req: CreateShotPreviewTaskRequest
  ): Promise<CreateShotPreviewTaskResponse> {
    if (USE_MOCK_API) {
      return mockAdapter.createShotPreviewTask(req);
    }

    return request<CreateShotPreviewTaskResponse>(
      "/api/v0_1/shot-preview/task",
      {
        method: "POST",
        body: JSON.stringify(req),
      }
    );
  }

  static async confirmStep(req: ConfirmStepRequest): Promise<{ status: string }> {
    return request<{ status: string }>("/api/v0_1/shot-preview/task/node/confirm", {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  static async adjustStep(req: AdjustStepRequest): Promise<{ status: string }> {
    return request<{ status: string }>("/api/v0_1/shot-preview/task/node/adjust", {
      method: "POST",
      body: JSON.stringify(req),
    });
  }

  static getTaskStreamUrl(taskId: string, userId: string = "default_user_001"): string {
    return `/api/v0_1/shot-preview/task/stream?task_id=${encodeURIComponent(
      taskId
    )}&user_id=${encodeURIComponent(userId)}`;
  }

  static getArtifactDownloadUrl(taskId: string, name: string = "shot-preview.mp4"): string {
    return `/api/v0_1/shot-preview/artifacts/download?task_id=${encodeURIComponent(
      taskId
    )}&name=${encodeURIComponent(name)}`;
  }
}
