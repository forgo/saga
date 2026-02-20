import type { Signal } from "@preact/signals";

interface UserListFiltersProps {
  search: Signal<string>;
  roleFilter: Signal<string>;
  onSearchChange: (value: string) => void;
  onRoleChange: (value: string) => void;
  onClear: () => void;
}

export function UserListFilters({
  search,
  roleFilter,
  onSearchChange,
  onRoleChange,
  onClear,
}: UserListFiltersProps) {
  const hasFilters = search.value !== "" || roleFilter.value !== "";

  return (
    <div class="flex flex-wrap items-center gap-3 mb-4">
      {/* Search */}
      <div class="flex-1 min-w-[200px]">
        <input
          type="text"
          placeholder="Search by email, username, or name..."
          value={search.value}
          onInput={(e) => onSearchChange((e.target as HTMLInputElement).value)}
          class="block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
        />
      </div>

      {/* Role filter */}
      <select
        value={roleFilter.value}
        onChange={(e) => onRoleChange((e.target as HTMLSelectElement).value)}
        class="rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
      >
        <option value="">All roles</option>
        <option value="admin">Admin</option>
        <option value="moderator">Moderator</option>
        <option value="user">User</option>
      </select>

      {/* Clear filters */}
      {hasFilters && (
        <button
          type="button"
          onClick={onClear}
          class="text-sm text-gray-500 hover:text-gray-700"
        >
          Clear filters
        </button>
      )}
    </div>
  );
}
