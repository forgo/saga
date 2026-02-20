import type { ComponentChildren } from "preact";

interface NavItem {
  label: string;
  href: string;
  icon: ComponentChildren;
}

interface NavSection {
  title?: string;
  items: NavItem[];
}

const navigation: NavSection[] = [
  {
    items: [
      {
        label: "Dashboard",
        href: "/",
        icon: (
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
              d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
            />
          </svg>
        ),
      },
    ],
  },
  {
    title: "Data Tools",
    items: [
      {
        label: "Seeder",
        href: "/data/seeder",
        icon: (
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
              d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
            />
          </svg>
        ),
      },
    ],
  },
  {
    title: "Management",
    items: [
      {
        label: "Users",
        href: "/users",
        icon: (
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
              d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z"
            />
          </svg>
        ),
      },
      {
        label: "Actions",
        href: "/actions",
        icon: (
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
              d="M13 10V3L4 14h7v7l9-11h-7z"
            />
          </svg>
        ),
      },
    ],
  },
  {
    title: "Testing",
    items: [
      {
        label: "Discovery Lab",
        href: "/discovery",
        icon: (
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
              d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"
            />
          </svg>
        ),
      },
    ],
  },
  {
    title: "System",
    items: [
      {
        label: "Health",
        href: "/system/health",
        icon: (
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
              d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
            />
          </svg>
        ),
      },
    ],
  },
];

interface SidebarProps {
  collapsed?: boolean;
  currentPath?: string;
}

export function Sidebar(
  { collapsed = false, currentPath = "/" }: SidebarProps,
) {
  const isActive = (href: string) => {
    if (href === "/") return currentPath === "/";
    return currentPath.startsWith(href);
  };

  return (
    <aside
      class={`bg-sidebar text-gray-300 flex flex-col sidebar-transition scrollbar-thin overflow-y-auto ${
        collapsed ? "w-16" : "w-64"
      }`}
    >
      {/* Logo */}
      <div class="h-16 flex items-center px-4 border-b border-gray-800">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center flex-shrink-0">
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
          {!collapsed && (
            <span class="font-semibold text-white text-lg">Saga Admin</span>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav class="flex-1 py-4">
        {navigation.map((section, idx) => (
          <div key={idx} class={section.title ? "mt-4" : ""}>
            {section.title && !collapsed && (
              <h3 class="px-4 mb-2 text-xs font-semibold text-gray-500 uppercase tracking-wider">
                {section.title}
              </h3>
            )}
            {section.title && collapsed && (
              <div class="border-t border-gray-800 mx-3 my-2" />
            )}
            <ul class="space-y-1 px-2">
              {section.items.map((item) => (
                <li key={item.href}>
                  <a
                    href={item.href}
                    class={`flex items-center gap-3 px-3 py-2 rounded-lg transition-colors ${
                      isActive(item.href)
                        ? "bg-sidebar-active text-white"
                        : "hover:bg-sidebar-hover hover:text-white"
                    } ${collapsed ? "justify-center" : ""}`}
                    title={collapsed ? item.label : undefined}
                  >
                    {item.icon}
                    {!collapsed && <span>{item.label}</span>}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </nav>

      {/* Footer */}
      <div class="border-t border-gray-800 p-4">
        {!collapsed && (
          <div class="text-xs text-gray-500">
            <p>Saga Admin v0.1.0</p>
          </div>
        )}
      </div>
    </aside>
  );
}
