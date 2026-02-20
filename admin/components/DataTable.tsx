import type { ComponentChildren } from "preact";

export interface Column<T> {
  key: keyof T | string;
  header: string;
  render?: (value: unknown, row: T) => ComponentChildren;
  sortable?: boolean;
  width?: string;
}

export interface DataTableProps<T> {
  columns: Column<T>[];
  data: T[];
  keyField: keyof T;
  onRowClick?: (row: T) => void;
  sortColumn?: string;
  sortDirection?: "asc" | "desc";
  onSort?: (column: string) => void;
  loading?: boolean;
  emptyMessage?: string;
  page?: number;
  pageSize?: number;
  total?: number;
  onPageChange?: (page: number) => void;
}

export function DataTable<T extends Record<string, unknown>>({
  columns,
  data,
  keyField,
  onRowClick,
  sortColumn,
  sortDirection,
  onSort,
  loading = false,
  emptyMessage = "No data available",
  page = 1,
  pageSize = 10,
  total,
  onPageChange,
}: DataTableProps<T>) {
  const totalPages = total ? Math.ceil(total / pageSize) : undefined;

  const getValue = (row: T, key: string): unknown => {
    const keys = key.split(".");
    let value: unknown = row;
    for (const k of keys) {
      if (value && typeof value === "object") {
        value = (value as Record<string, unknown>)[k];
      } else {
        return undefined;
      }
    }
    return value;
  };

  return (
    <div class="bg-white rounded-lg border border-gray-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              {columns.map((col) => (
                <th
                  key={String(col.key)}
                  scope="col"
                  class={`px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider ${
                    col.sortable
                      ? "cursor-pointer hover:bg-gray-100 select-none"
                      : ""
                  }`}
                  style={col.width ? { width: col.width } : undefined}
                  onClick={() => col.sortable && onSort?.(String(col.key))}
                >
                  <div class="flex items-center gap-1">
                    {col.header}
                    {col.sortable && (
                      <span class="text-gray-400">
                        {sortColumn === col.key
                          ? (
                            sortDirection === "asc"
                              ? (
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
                                    d="M5 15l7-7 7 7"
                                  />
                                </svg>
                              )
                              : (
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
                                    d="M19 9l-7 7-7-7"
                                  />
                                </svg>
                              )
                          )
                          : (
                            <svg
                              class="w-4 h-4 opacity-0 group-hover:opacity-50"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                stroke-width="2"
                                d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"
                              />
                            </svg>
                          )}
                      </span>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            {loading
              ? (
                <tr>
                  <td colSpan={columns.length} class="px-6 py-12 text-center">
                    <div class="flex items-center justify-center gap-2 text-gray-500">
                      <svg
                        class="w-5 h-5 animate-spin"
                        fill="none"
                        viewBox="0 0 24 24"
                      >
                        <circle
                          class="opacity-25"
                          cx="12"
                          cy="12"
                          r="10"
                          stroke="currentColor"
                          stroke-width="4"
                        />
                        <path
                          class="opacity-75"
                          fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        />
                      </svg>
                      <span>Loading...</span>
                    </div>
                  </td>
                </tr>
              )
              : data.length === 0
              ? (
                <tr>
                  <td
                    colSpan={columns.length}
                    class="px-6 py-12 text-center text-gray-500"
                  >
                    {emptyMessage}
                  </td>
                </tr>
              )
              : (
                data.map((row) => (
                  <tr
                    key={String(row[keyField])}
                    class={onRowClick ? "cursor-pointer hover:bg-gray-50" : ""}
                    onClick={() => onRowClick?.(row)}
                  >
                    {columns.map((col) => {
                      const value = getValue(row, String(col.key));
                      return (
                        <td
                          key={String(col.key)}
                          class="px-6 py-4 whitespace-nowrap text-sm text-gray-900"
                        >
                          {col.render
                            ? col.render(value, row)
                            : String(value ?? "")}
                        </td>
                      );
                    })}
                  </tr>
                ))
              )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages && totalPages > 1 && (
        <div class="bg-gray-50 px-6 py-3 flex items-center justify-between border-t border-gray-200">
          <div class="text-sm text-gray-500">
            Showing {(page - 1) * pageSize + 1} to{" "}
            {Math.min(page * pageSize, total ?? 0)} of {total} results
          </div>
          <div class="flex items-center gap-2">
            <button
              type="button"
              class="px-3 py-1 text-sm rounded border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={page <= 1}
              onClick={() => onPageChange?.(page - 1)}
            >
              Previous
            </button>
            <span class="text-sm text-gray-600">
              Page {page} of {totalPages}
            </span>
            <button
              type="button"
              class="px-3 py-1 text-sm rounded border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={page >= totalPages}
              onClick={() => onPageChange?.(page + 1)}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
