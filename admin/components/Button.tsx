import type { ComponentChildren, JSX } from "preact";

export type ButtonVariant =
  | "primary"
  | "secondary"
  | "outline"
  | "ghost"
  | "danger";
export type ButtonSize = "sm" | "md" | "lg";

export interface ButtonProps
  extends Omit<JSX.HTMLAttributes<HTMLButtonElement>, "size"> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  disabled?: boolean;
  icon?: ComponentChildren;
  children?: ComponentChildren;
}

const variantClasses: Record<ButtonVariant, string> = {
  primary: "bg-primary-600 text-white hover:bg-primary-700 border-transparent",
  secondary: "bg-gray-100 text-gray-900 hover:bg-gray-200 border-transparent",
  outline: "bg-white text-gray-700 hover:bg-gray-50 border-gray-300",
  ghost: "bg-transparent text-gray-700 hover:bg-gray-100 border-transparent",
  danger: "bg-red-600 text-white hover:bg-red-700 border-transparent",
};

const sizeClasses: Record<ButtonSize, string> = {
  sm: "px-2.5 py-1.5 text-xs",
  md: "px-4 py-2 text-sm",
  lg: "px-6 py-3 text-base",
};

export function Button({
  variant = "primary",
  size = "md",
  loading = false,
  icon,
  disabled,
  children,
  class: className,
  ...props
}: ButtonProps) {
  const isDisabled = disabled || loading;

  return (
    <button
      {...props}
      disabled={isDisabled}
      class={`inline-flex items-center justify-center gap-2 font-medium rounded-lg border transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed ${
        variantClasses[variant]
      } ${sizeClasses[size]} ${className ?? ""}`}
    >
      {loading
        ? (
          <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
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
        )
        : icon
        ? icon
        : null}
      {children}
    </button>
  );
}
