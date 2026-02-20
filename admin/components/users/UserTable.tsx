import type { Signal } from "@preact/signals";
import { type Column, DataTable } from "../DataTable.tsx";
import { RoleBadge } from "./RoleBadge.tsx";
import { StatusBadge } from "./StatusBadge.tsx";
import type { AdminUserItem, UserRole } from "../../lib/api/types.ts";

type Row = Record<string, unknown>;

interface UserTableProps {
  users: AdminUserItem[];
  loading: boolean;
  page: number;
  pageSize: number;
  total: number;
  sortColumn: Signal<string>;
  sortDirection: Signal<"asc" | "desc">;
  onRowClick: (user: AdminUserItem) => void;
  onPageChange: (page: number) => void;
  onSort: (column: string) => void;
}

function formatDate(dateStr: string): string {
  if (!dateStr) return "—";
  try {
    return new Date(dateStr).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  } catch {
    return dateStr;
  }
}

function displayName(row: Row): string {
  const parts: string[] = [];
  if (row.firstname) parts.push(row.firstname as string);
  if (row.lastname) parts.push(row.lastname as string);
  return parts.length > 0 ? parts.join(" ") : "—";
}

const columns: Column<Row>[] = [
  {
    key: "email",
    header: "Email",
    sortable: true,
  },
  {
    key: "username",
    header: "Username",
    sortable: true,
    render: (val) => (val as string) || "—",
  },
  {
    key: "firstname",
    header: "Name",
    render: (_val, row) => displayName(row),
  },
  {
    key: "role",
    header: "Role",
    sortable: true,
    render: (val) => <RoleBadge role={(val as UserRole) || "user"} />,
  },
  {
    key: "status",
    header: "Status",
    render: (val) => <StatusBadge status={(val as string) || "active"} />,
  },
  {
    key: "created_on",
    header: "Created",
    sortable: true,
    render: (val) => formatDate(val as string),
  },
];

export function UserTable({
  users,
  loading,
  page,
  pageSize,
  total,
  sortColumn,
  sortDirection,
  onRowClick,
  onPageChange,
  onSort,
}: UserTableProps) {
  return (
    <DataTable
      columns={columns}
      data={users as unknown as Row[]}
      keyField="id"
      onRowClick={(row) => onRowClick(row as unknown as AdminUserItem)}
      loading={loading}
      emptyMessage="No users found"
      page={page}
      pageSize={pageSize}
      total={total}
      onPageChange={onPageChange}
      sortColumn={sortColumn.value}
      sortDirection={sortDirection.value}
      onSort={onSort}
    />
  );
}
