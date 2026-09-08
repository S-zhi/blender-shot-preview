/**
 * Unified API Client for Shot Preview Frontend
 * Configurable to use real endpoints or mock adapter
 */

export const USE_MOCK_API = true; // Easily toggleable or via import.meta.env.VITE_USE_MOCK

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

  try {
    const response = await fetch(url, { ...options, headers });
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
    throw new ApiClientError(500, (err as Error).message || "Network Error");
  }
}
