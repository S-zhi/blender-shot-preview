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
 */
export async function testLLMConnection(params: TestConnectionParams): Promise<TestConnectionResult> {
  const { providerType, baseUrl, apiKey, modelName, apiVersion } = params;
  const cleanBaseUrl = baseUrl.trim().replace(/\/+$/, "");

  if (!cleanBaseUrl) {
    return { success: false, message: "Base URL 不能为空" };
  }

  const startTime = Date.now();
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 12000); // 12s timeout

  try {
    let testUrl = "";
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (providerType === "azure_openai") {
      // Azure OpenAI chat completions URL format:
      // {endpoint}/openai/deployments/{deployment-id}/chat/completions?api-version={api-version}
      const version = (apiVersion || "2024-02-15-preview").trim();
      const deployment = modelName.trim() || "gpt-4o";
      testUrl = `${cleanBaseUrl}/openai/deployments/${deployment}/chat/completions?api-version=${version}`;
      if (apiKey) {
        headers["api-key"] = apiKey.trim();
      }
    } else {
      // Standard OpenAI compatible format:
      testUrl = `${cleanBaseUrl}/chat/completions`;
      if (apiKey) {
        headers["Authorization"] = `Bearer ${apiKey.trim()}`;
      }
    }

    // Send minimal test prompt
    const response = await fetch(testUrl, {
      method: "POST",
      headers,
      body: JSON.stringify({
        model: modelName.trim() || "gpt-4o",
        messages: [{ role: "user", content: "ping" }],
        max_tokens: 1,
      }),
      signal: controller.signal,
    });

    clearTimeout(timeoutId);
    const latencyMs = Date.now() - startTime;

    if (response.ok) {
      return {
        success: true,
        message: `连接成功 (HTTP ${response.status} · 耗时 ${latencyMs}ms)`,
        latencyMs,
      };
    }

    // Try reading error message
    let errDetail = `HTTP ${response.status} ${response.statusText}`;
    try {
      const errBody = await response.json();
      if (errBody?.error?.message) {
        errDetail += `: ${errBody.error.message}`;
      }
    } catch {
      // Ignore JSON parse error on error response
    }

    return {
      success: false,
      message: `连通性检测未通过: ${errDetail}`,
      latencyMs,
    };
  } catch (err: unknown) {
    clearTimeout(timeoutId);
    const latencyMs = Date.now() - startTime;
    if (err instanceof Error) {
      if (err.name === "AbortError") {
        return { success: false, message: "连接超时 (12秒无响应)", latencyMs };
      }
      return { success: false, message: `网络或跨域错误: ${err.message}`, latencyMs };
    }
    return { success: false, message: "未知的网络请求错误", latencyMs };
  }
}
