import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import { UserListFilters } from "../components/users/UserListFilters.tsx";
import { UserTable } from "../components/users/UserTable.tsx";
import { UserDetailPanel } from "../components/users/UserDetailPanel.tsx";
import { ConfirmDialog } from "../components/users/ConfirmDialog.tsx";
import {
  deleteUser,
  getUser,
  liftModAction,
  listUsers,
  takeModAction,
  updateUserRole,
} from "../lib/api/admin-users.ts";
import type { AdminUserDetail, AdminUserItem } from "../lib/api/types.ts";

export default function UserManager() {
  // List state
  const users = useSignal<AdminUserItem[]>([]);
  const total = useSignal(0);
  const page = useSignal(1);
  const pageSize = 20;
  const loading = useSignal(false);
  const error = useSignal<string | null>(null);

  // Filters
  const search = useSignal("");
  const roleFilter = useSignal("");
  const sortColumn = useSignal("created_on");
  const sortDirection = useSignal<"asc" | "desc">("desc");

  // Detail panel
  const selectedUser = useSignal<AdminUserDetail | null>(null);
  const detailLoading = useSignal(false);
  const panelOpen = useSignal(false);

  // Action states
  const roleLoading = useSignal(false);
  const actionLoading = useSignal(false);

  // Confirm dialog
  const confirmOpen = useSignal(false);
  const confirmTitle = useSignal("");
  const confirmMessage = useSignal("");
  const confirmAction = useRef<(() => void) | null>(null);

  // Debounce timer
  const searchTimer = useRef<number | null>(null);

  // Load users
  const loadUsers = async () => {
    loading.value = true;
    error.value = null;

    const result = await listUsers({
      page: page.value,
      page_size: pageSize,
      search: search.value || undefined,
      role: roleFilter.value || undefined,
      sort_by: sortColumn.value,
      sort_dir: sortDirection.value,
    });

    if (result.error) {
      error.value = result.error;
      users.value = [];
      total.value = 0;
    } else if (result.data) {
      users.value = result.data.users ?? [];
      total.value = result.data.total;
    }

    loading.value = false;
  };

  // Load user detail
  const loadUserDetail = async (userId: string) => {
    detailLoading.value = true;
    panelOpen.value = true;

    const result = await getUser(userId);

    if (result.error) {
      error.value = result.error;
      selectedUser.value = null;
    } else {
      selectedUser.value = result.data;
    }

    detailLoading.value = false;
  };

  // Initial load
  useEffect(() => {
    loadUsers();
  }, []);

  // Debounced search
  const handleSearchChange = (value: string) => {
    search.value = value;
    if (searchTimer.current !== null) {
      clearTimeout(searchTimer.current);
    }
    searchTimer.current = setTimeout(() => {
      page.value = 1;
      loadUsers();
    }, 300) as unknown as number;
  };

  // Role filter change
  const handleRoleChange = (value: string) => {
    roleFilter.value = value;
    page.value = 1;
    loadUsers();
  };

  // Clear filters
  const handleClearFilters = () => {
    search.value = "";
    roleFilter.value = "";
    page.value = 1;
    loadUsers();
  };

  // Sort
  const handleSort = (column: string) => {
    if (sortColumn.value === column) {
      sortDirection.value = sortDirection.value === "asc" ? "desc" : "asc";
    } else {
      sortColumn.value = column;
      sortDirection.value = "asc";
    }
    loadUsers();
  };

  // Page change
  const handlePageChange = (newPage: number) => {
    page.value = newPage;
    loadUsers();
  };

  // Row click
  const handleRowClick = (user: AdminUserItem) => {
    loadUserDetail(user.id);
  };

  // Close panel
  const handleClosePanel = () => {
    panelOpen.value = false;
    selectedUser.value = null;
  };

  // Role change
  const handleRoleUpdate = async (userId: string, role: string) => {
    if (selectedUser.value && role === selectedUser.value.role) return;

    roleLoading.value = true;
    const result = await updateUserRole(userId, role);
    roleLoading.value = false;

    if (result.error) {
      error.value = result.error;
      return;
    }

    // Refresh both list and detail
    loadUsers();
    loadUserDetail(userId);
  };

  // Show confirm dialog helper
  const showConfirm = (title: string, message: string, action: () => void) => {
    confirmTitle.value = title;
    confirmMessage.value = message;
    confirmAction.current = action;
    confirmOpen.value = true;
  };

  // Suspend user
  const handleSuspend = (userId: string) => {
    showConfirm(
      "Suspend User",
      "This will temporarily suspend the user for 30 days. They will not be able to log in during the suspension.",
      async () => {
        actionLoading.value = true;
        confirmOpen.value = false;
        const result = await takeModAction(
          userId,
          "suspension",
          "Suspended by admin",
          30,
        );
        actionLoading.value = false;

        if (result.error) {
          error.value = result.error;
          return;
        }

        loadUsers();
        loadUserDetail(userId);
      },
    );
  };

  // Ban user
  const handleBan = (userId: string) => {
    showConfirm(
      "Ban User",
      "This will permanently ban the user. This action can be reversed by lifting the ban.",
      async () => {
        actionLoading.value = true;
        confirmOpen.value = false;
        const result = await takeModAction(userId, "ban", "Banned by admin");
        actionLoading.value = false;

        if (result.error) {
          error.value = result.error;
          return;
        }

        loadUsers();
        loadUserDetail(userId);
      },
    );
  };

  // Lift moderation action
  const handleLiftAction = (actionId: string) => {
    showConfirm(
      "Lift Moderation Action",
      "This will remove the active moderation action and restore the user's access.",
      async () => {
        actionLoading.value = true;
        confirmOpen.value = false;
        const result = await liftModAction(actionId, "Lifted by admin");
        actionLoading.value = false;

        if (result.error) {
          error.value = result.error;
          return;
        }

        if (selectedUser.value) {
          loadUsers();
          loadUserDetail(selectedUser.value.id);
        }
      },
    );
  };

  // Delete user
  const handleDelete = (userId: string, hard: boolean) => {
    const title = hard ? "Permanently Delete User" : "Soft Delete User";
    const message = hard
      ? "This will permanently delete the user and all their data. This action cannot be undone."
      : "This will ban the user, effectively soft-deleting their account. The ban can be lifted later.";

    showConfirm(title, message, async () => {
      actionLoading.value = true;
      confirmOpen.value = false;
      const result = await deleteUser(userId, hard);
      actionLoading.value = false;

      if (result.error) {
        error.value = result.error;
        return;
      }

      handleClosePanel();
      loadUsers();
    });
  };

  return (
    <div>
      {/* Error banner */}
      {error.value && (
        <div class="mb-4 bg-red-50 border border-red-200 rounded-lg p-4 flex items-center justify-between">
          <p class="text-sm text-red-700">{error.value}</p>
          <button
            type="button"
            onClick={() => {
              error.value = null;
            }}
            class="text-red-500 hover:text-red-700"
          >
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>
        </div>
      )}

      {/* Filters */}
      <UserListFilters
        search={search}
        roleFilter={roleFilter}
        onSearchChange={handleSearchChange}
        onRoleChange={handleRoleChange}
        onClear={handleClearFilters}
      />

      {/* Table */}
      <UserTable
        users={users.value}
        loading={loading.value}
        page={page.value}
        pageSize={pageSize}
        total={total.value}
        sortColumn={sortColumn}
        sortDirection={sortDirection}
        onRowClick={handleRowClick}
        onPageChange={handlePageChange}
        onSort={handleSort}
      />

      {/* Detail panel */}
      {panelOpen.value && (
        <UserDetailPanel
          user={selectedUser.value}
          loading={detailLoading.value}
          onClose={handleClosePanel}
          onRoleChange={handleRoleUpdate}
          onSuspend={handleSuspend}
          onBan={handleBan}
          onLiftAction={handleLiftAction}
          onDelete={handleDelete}
          roleLoading={roleLoading}
          actionLoading={actionLoading}
        />
      )}

      {/* Confirm dialog */}
      <ConfirmDialog
        open={confirmOpen.value}
        title={confirmTitle.value}
        message={confirmMessage.value}
        confirmLabel="Confirm"
        confirmVariant="danger"
        loading={actionLoading.value}
        onConfirm={() => confirmAction.current?.()}
        onCancel={() => {
          confirmOpen.value = false;
        }}
      />
    </div>
  );
}
