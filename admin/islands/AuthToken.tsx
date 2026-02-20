import { useComputed, useSignal } from "@preact/signals";
import {
  authToken,
  clearAuthToken,
  isTokenExpired,
  parseToken,
  setAuthToken,
} from "../lib/signals/auth.ts";
import { Button } from "../components/Button.tsx";

export default function AuthToken() {
  const inputToken = useSignal("");
  const showInput = useSignal(!authToken.value);
  const error = useSignal<string | null>(null);

  const tokenInfo = useComputed(() => {
    if (!authToken.value) return null;
    const info = parseToken(authToken.value);
    if (!info) return null;

    const expired = isTokenExpired(authToken.value);
    const expiresAt = info.exp ? new Date(info.exp * 1000) : null;

    return {
      email: info.email,
      role: info.role,
      userId: info.user_id,
      expiresAt,
      expired,
    };
  });

  const handleSetToken = () => {
    const token = inputToken.value.trim();
    if (!token) {
      error.value = "Please enter a token";
      return;
    }

    const info = parseToken(token);
    if (!info) {
      error.value = "Invalid token format";
      return;
    }

    if (info.role !== "admin") {
      error.value = "Token does not have admin role";
      return;
    }

    if (isTokenExpired(token)) {
      error.value = "Token is expired";
      return;
    }

    error.value = null;
    setAuthToken(token);
    inputToken.value = "";
    showInput.value = false;
  };

  const handleClearToken = () => {
    clearAuthToken();
    showInput.value = true;
  };

  if (authToken.value && !showInput.value) {
    const info = tokenInfo.value;
    return (
      <div class="bg-white rounded-lg border border-gray-200 p-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div
              class={`w-3 h-3 rounded-full ${
                info?.expired ? "bg-red-500" : "bg-green-500"
              }`}
            />
            <div>
              <p class="text-sm font-medium text-gray-900">
                {info?.email || "Admin"}
                {info?.role && (
                  <span class="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary-100 text-primary-800">
                    {info.role}
                  </span>
                )}
              </p>
              {info?.expiresAt && (
                <p
                  class={`text-xs ${
                    info.expired ? "text-red-600" : "text-gray-500"
                  }`}
                >
                  {info.expired
                    ? "Expired"
                    : `Expires: ${info.expiresAt.toLocaleString()}`}
                </p>
              )}
            </div>
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              onClick={() => {
                showInput.value = true;
              }}
              class="text-sm text-gray-500 hover:text-gray-700"
            >
              Change
            </button>
            <button
              type="button"
              onClick={handleClearToken}
              class="text-sm text-red-600 hover:text-red-800"
            >
              Clear
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
      <h3 class="text-sm font-medium text-yellow-800 mb-2">
        Admin Token Required
      </h3>
      <p class="text-xs text-yellow-700 mb-3">
        Generate an admin token using{" "}
        <code class="bg-yellow-100 px-1 py-0.5 rounded">make admin-token</code>
        {" "}
        in the api directory.
      </p>

      <div class="space-y-3">
        <div>
          <label class="sr-only">Admin Token</label>
          <textarea
            value={inputToken.value}
            onInput={(e) => {
              inputToken.value = (e.target as HTMLTextAreaElement).value;
              error.value = null;
            }}
            placeholder="Paste your admin token here..."
            rows={3}
            class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 text-xs font-mono px-3 py-2 border"
          />
        </div>

        {error.value && <p class="text-sm text-red-600">{error.value}</p>}

        <div class="flex items-center gap-2">
          <Button onClick={handleSetToken} size="sm">
            Set Token
          </Button>
          {authToken.value && (
            <button
              type="button"
              onClick={() => {
                showInput.value = false;
              }}
              class="text-sm text-gray-500 hover:text-gray-700"
            >
              Cancel
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
