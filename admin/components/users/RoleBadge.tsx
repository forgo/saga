import type { UserRole } from "../../lib/api/types.ts";

const roleStyles: Record<UserRole, string> = {
  admin: "bg-purple-100 text-purple-800",
  moderator: "bg-blue-100 text-blue-800",
  user: "bg-gray-100 text-gray-800",
};

export function RoleBadge({ role }: { role: UserRole }) {
  return (
    <span
      class={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
        roleStyles[role] ?? roleStyles.user
      }`}
    >
      {role}
    </span>
  );
}
