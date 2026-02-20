import type { ComponentChildren } from "preact";
import { Button } from "../Button.tsx";

interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string | ComponentChildren;
  confirmLabel?: string;
  confirmVariant?: "primary" | "danger";
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = "Confirm",
  confirmVariant = "danger",
  loading = false,
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  if (!open) return null;

  return (
    <div class="fixed inset-0 z-50 overflow-y-auto">
      <div class="flex min-h-full items-center justify-center p-4">
        {/* Backdrop */}
        <div
          class="fixed inset-0 bg-gray-500/75 transition-opacity"
          onClick={onCancel}
        />

        {/* Dialog */}
        <div class="relative bg-white rounded-lg shadow-xl max-w-md w-full p-6">
          <h3 class="text-lg font-semibold text-gray-900 mb-2">{title}</h3>
          <div class="text-sm text-gray-600 mb-6">{message}</div>
          <div class="flex justify-end gap-3">
            <Button
              variant="outline"
              size="sm"
              onClick={onCancel}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button
              variant={confirmVariant}
              size="sm"
              onClick={onConfirm}
              loading={loading}
            >
              {confirmLabel}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
