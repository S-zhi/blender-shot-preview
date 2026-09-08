import { request, USE_MOCK_API } from "./client";
import { mockAdapter } from "./mock-adapter";
import {
  ManageLLMKeyRequest,
  ManageLLMKeyResponse,
} from "./types";

export class LLMKeyService {
  /**
   * 对应 idl/v0_1/llm_gateway.thrift:
   * ManageLLMKey(1: ManageLLMKeyRequest request)
   */
  static async manageLLMKey(
    req: ManageLLMKeyRequest
  ): Promise<ManageLLMKeyResponse> {
    if (USE_MOCK_API) {
      return mockAdapter.manageLLMKey(req);
    }

    return request<ManageLLMKeyResponse>("/api/v0_1/llm-gateway/key", {
      method: "POST",
      body: JSON.stringify(req),
    });
  }
}
