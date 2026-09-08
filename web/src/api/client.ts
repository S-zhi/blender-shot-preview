/**
 * Unified API Client for Shot Preview Frontend
 * Configurable to use real endpoints or mock adapter
 */

export const USE_MOCK_API = false; // Connected to real Go Kitex/HTTP backend

export class ApiClientError extends Error {
  constructor(
    public status: number,
    message: string,
    public data?: unknown
  ) {
    super(message);
    this.name = "ApiClientError";
  }
}

export async function request<T>(
  url: string,
  options: RequestInit = {}
): Promise<T> {
  const headers = {
    "Content-Type": "application/json",
    Accept: "application/json",
    ...options.headers,
  };

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), 30000);
  try {
    const response = await fetch(url, { ...options, headers, signal: controller.signal });
    if (!response.ok) {
      let errorData;
      try {
        errorData = await response.json();
      } catch {
        // non-json response
      }
      throw new ApiClientError(
        response.status,
        `Request failed with status ${response.status}`,
        errorData
      );
    }
    return (await response.json()) as T;
  } catch (err: unknown) {
    if (err instanceof ApiClientError) throw err;
    if ((err as Error)?.name === "AbortError") {
      throw new ApiClientError(408, "Request timed out");
    }
    throw new ApiClientError(500, (err as Error).message || "Network Error");
  } finally {
    clearTimeout(timeoutId);
  }
}
