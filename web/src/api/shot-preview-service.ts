import { request, USE_MOCK_API } from "./client";
import { mockAdapter } from "./mock-adapter";
import {
  CreateShotPreviewTaskRequest,
  CreateShotPreviewTaskResponse,
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
}
