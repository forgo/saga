import type { ApiError, ApiResponse } from "./types.ts";

// API configuration
const API_BASE_URL = (globalThis as unknown as {
  Deno?: { env: { get: (key: string) => string | undefined } };
})
  .Deno?.env.get?.("API_BASE_URL") ?? "http://localhost:8080";

interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
  params?: Record<string, string | number | boolean | undefined>;
}

class ApiClient {
  private baseUrl: string;
  private authToken: string | null = null;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl.replace(/\/$/, "");
  }

  setAuthToken(token: string | null) {
    this.authToken = token;
  }

  getAuthToken(): string | null {
    return this.authToken;
  }

  private buildUrl(
    path: string,
    params?: Record<string, string | number | boolean | undefined>,
  ): string {
    const url = new URL(path.startsWith("/") ? path : `/${path}`, this.baseUrl);

    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          url.searchParams.set(key, String(value));
        }
      });
    }

    return url.toString();
  }

  private async request<T>(
    method: string,
    path: string,
    options: RequestOptions = {},
  ): Promise<ApiResponse<T>> {
    const { body, params, headers: customHeaders, ...fetchOptions } = options;

    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...customHeaders,
    };

    if (this.authToken) {
      (headers as Record<string, string>)["Authorization"] =
        `Bearer ${this.authToken}`;
    }

    const url = this.buildUrl(path, params);

    try {
      const response = await fetch(url, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
        ...fetchOptions,
      });

      const contentType = response.headers.get("content-type");
      const isJson = contentType?.includes("application/json");

      if (!response.ok) {
        const errorBody = isJson
          ? await response.json()
          : await response.text();
        const error: ApiError = isJson && typeof errorBody === "object"
          ? {
            code: errorBody.code || `HTTP_${response.status}`,
            message: errorBody.message || response.statusText,
            details: errorBody.details,
          }
          : {
            code: `HTTP_${response.status}`,
            message: typeof errorBody === "string"
              ? errorBody
              : response.statusText,
          };

        return { error };
      }

      if (response.status === 204 || !isJson) {
        return { data: undefined as unknown as T };
      }

      const data = await response.json();
      return { data };
    } catch (err) {
      const error: ApiError = {
        code: "NETWORK_ERROR",
        message: err instanceof Error ? err.message : "Unknown network error",
      };
      return { error };
    }
  }

  get<T>(path: string, options?: RequestOptions): Promise<ApiResponse<T>> {
    return this.request<T>("GET", path, options);
  }

  post<T>(
    path: string,
    body?: unknown,
    options?: RequestOptions,
  ): Promise<ApiResponse<T>> {
    return this.request<T>("POST", path, { ...options, body });
  }

  put<T>(
    path: string,
    body?: unknown,
    options?: RequestOptions,
  ): Promise<ApiResponse<T>> {
    return this.request<T>("PUT", path, { ...options, body });
  }

  patch<T>(
    path: string,
    body?: unknown,
    options?: RequestOptions,
  ): Promise<ApiResponse<T>> {
    return this.request<T>("PATCH", path, { ...options, body });
  }

  delete<T>(path: string, options?: RequestOptions): Promise<ApiResponse<T>> {
    return this.request<T>("DELETE", path, options);
  }
}

// Export singleton instance
export const api = new ApiClient();

// Export class for custom instances
export { ApiClient };

// Helper to check if response has error
export function isError<T>(
  response: ApiResponse<T>,
): response is { error: ApiError } {
  return response.error !== undefined;
}

// Helper to unwrap response or throw
export function unwrap<T>(response: ApiResponse<T>): T {
  if (isError(response)) {
    throw new Error(response.error.message);
  }
  return response.data as T;
}
