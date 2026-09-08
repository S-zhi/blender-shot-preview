export interface TestConnectionParams {
  providerType: "openai" | "azure_openai" | "custom";
  baseUrl: string;
  apiKey: string;
  modelName: string;
  apiVersion?: string;
}

export interface TestConnectionResult {
  success: boolean;
  message: string;
  latencyMs?: number;
}

/**
 * Perform a lightweight connection test for OpenAI / Azure OpenAI protocol endpoints.
 *
 * The request is proxied through the backend (/api/v0_1/llm-probe) to avoid
 * browser CORS restrictions when the LLM endpoint does not allow cross-origin
 * requests from localhost.
 */
export async function testLLMConnection(params: TestConnectionParams): Promise<TestConnectionResult> {
  const { providerType, baseUrl, apiKey, modelName, apiVersion } = params;
  const cleanBaseUrl = baseUrl.trim().replace(/\/+$/, "");

  if (!cleanBaseUrl) {
    return { success: false, message: "Base URL 不能为空" };
  }

  const startTime = Date.now();
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 14000); // slightly longer than backend's 12s

  try {
    const response = await fetch("/api/v0_1/llm-probe", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        provider_type: providerType,
        base_url: cleanBaseUrl,
        api_key: apiKey,
        model_name: modelName,
        api_version: apiVersion ?? "",
      }),
      signal: controller.signal,
    });

    clearTimeout(timeoutId);
    const latencyMs = Date.now() - startTime;

    if (!response.ok) {
      return {
        success: false,
        message: `探针服务异常 (HTTP ${response.status})`,
        latencyMs,
      };
    }

    const result = await response.json() as { success: boolean; message: string; latency_ms?: number };
    return {
      success: result.success,
      message: result.message,
      latencyMs: result.latency_ms ?? latencyMs,
    };
  } catch (err: unknown) {
    clearTimeout(timeoutId);
    const latencyMs = Date.now() - startTime;
    if (err instanceof Error) {
      if (err.name === "AbortError") {
        return { success: false, message: "探针请求超时 (14秒无响应)", latencyMs };
      }
      return { success: false, message: `探针请求失败: ${err.message}`, latencyMs };
    }
    return { success: false, message: "未知的探针请求错误", latencyMs };
  }
}

