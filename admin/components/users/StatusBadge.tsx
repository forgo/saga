const statusStyles: Record<string, string> = {
  active: "bg-green-100 text-green-800",
  suspended: "bg-yellow-100 text-yellow-800",
  banned: "bg-red-100 text-red-800",
};

export function StatusBadge({ status }: { status: string }) {
  return (
    <span
      class={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
        statusStyles[status] ?? statusStyles.active
      }`}
    >
      {status}
    </span>
  );
}
