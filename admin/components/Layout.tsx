import type { ComponentChildren } from "preact";
import { Sidebar } from "./Sidebar.tsx";
import AuthToken from "../islands/AuthToken.tsx";

interface LayoutProps {
  children: ComponentChildren;
  currentPath?: string;
}

export function Layout({ children, currentPath = "/" }: LayoutProps) {
  return (
    <div class="flex h-screen bg-gray-100">
      {/* Sidebar - hidden on mobile by default */}
      <div class="hidden md:flex">
        <Sidebar currentPath={currentPath} />
      </div>

      {/* Main content */}
      <div class="flex-1 flex flex-col overflow-hidden">
        {/* Top header */}
        <header class="bg-white border-b border-gray-200 flex items-center justify-between px-4 md:px-6 py-3">
          {/* Mobile menu button */}
          <button
            type="button"
            class="md:hidden p-2 rounded-lg text-gray-500 hover:bg-gray-100"
            aria-label="Open menu"
          >
            <svg
              class="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M4 6h16M4 12h16M4 18h16"
              />
            </svg>
          </button>

          {/* Mobile logo */}
          <div class="md:hidden flex items-center gap-2">
            <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
              <svg
                class="w-5 h-5 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            <span class="font-semibold text-gray-900">Saga Admin</span>
          </div>

          {/* Breadcrumb / page title area */}
          <div class="hidden md:block">
            <nav class="flex items-center text-sm text-gray-500">
              <a href="/" class="hover:text-gray-700">
                Home
              </a>
              {currentPath !== "/" && (
                <>
                  <svg
                    class="w-4 h-4 mx-2"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                  <span class="text-gray-900 font-medium">
                    {getPageTitle(currentPath)}
                  </span>
                </>
              )}
            </nav>
          </div>

          {/* Auth token display */}
          <div class="flex-1 max-w-md mx-4 hidden sm:block">
            <AuthToken />
          </div>

          {/* Right side actions */}
          <div class="flex items-center gap-3">
            {/* User menu placeholder */}
            <button
              type="button"
              class="w-8 h-8 rounded-full bg-gray-300 flex items-center justify-center text-gray-600 hover:bg-gray-400 transition-colors"
            >
              <svg
                class="w-5 h-5"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                />
              </svg>
            </button>
          </div>
        </header>

        {/* Mobile auth token (shown below header) */}
        <div class="sm:hidden px-4 py-2 bg-white border-b border-gray-200">
          <AuthToken />
        </div>

        {/* Page content */}
        <main class="flex-1 overflow-y-auto p-4 md:p-6">{children}</main>
      </div>
    </div>
  );
}

function getPageTitle(path: string): string {
  const titles: Record<string, string> = {
    "/data/seeder": "Data Seeder",
    "/data/scenarios": "Scenarios",
    "/users": "Users",
    "/actions": "Actions",
    "/discovery": "Discovery Lab",
    "/system/health": "System Health",
    "/system/logs": "Logs",
  };

  // Check for dynamic routes like /users/[id]
  if (path.startsWith("/users/") && path !== "/users") {
    return "User Details";
  }

  return titles[path] || "Page";
}
