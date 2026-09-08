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

// Global event target for auth state changes
export const authEvents = new EventTarget();

export interface AuthStatus {
  authenticated: boolean;
  auth_required?: boolean;
}

export async function checkAuthStatus(): Promise<AuthStatus> {
  try {
    const res = await fetch("/api/v0_1/auth/check", {
      method: "GET",
      credentials: "include",
    });
    if (res.ok) {
      const data = await res.json();
      return { authenticated: !!data.authenticated, auth_required: data.auth_required };
    }
    if (res.status === 401) {
      return { authenticated: false, auth_required: true };
    }
    return { authenticated: true, auth_required: false };
  } catch {
    // If backend is not reached or returns error, assume auth not required or handled by caller
    return { authenticated: true, auth_required: false };
  }
}

export async function loginWithToken(token: string): Promise<{ success: boolean; error?: string }> {
  try {
    const res = await fetch("/api/v0_1/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ token: token.trim() }),
      credentials: "include",
    });
    if (!res.ok) {
      let errText = "Invalid access token";
      try {
        const errJson = await res.json();
        if (errJson.error) errText = errJson.error;
      } catch {
        // ignore
      }
      return { success: false, error: errText };
    }
    authEvents.dispatchEvent(new CustomEvent("auth:login"));
    return { success: true };
  } catch (err: unknown) {
    return { success: false, error: (err as Error)?.message || "Network Error" };
  }
}

export async function logout(): Promise<void> {
  try {
    await fetch("/api/v0_1/auth/logout", {
      method: "POST",
      credentials: "include",
    });
  } catch {
    // ignore
  } finally {
    authEvents.dispatchEvent(new CustomEvent("auth:logout"));
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
  const timeoutId = setTimeout(() => controller.abort(), 60000);
  try {
    const response = await fetch(url, {
      ...options,
      headers,
      credentials: "include",
      signal: controller.signal,
    });
    if (!response.ok) {
      let errorData;
      try {
        errorData = await response.json();
      } catch {
        // non-json response
      }

      if (response.status === 401) {
        authEvents.dispatchEvent(new CustomEvent("auth:unauthorized"));
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
