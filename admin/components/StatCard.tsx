import type { ComponentChildren } from "preact";

export interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  icon?: ComponentChildren;
  trend?: {
    value: number;
    label?: string;
  };
  loading?: boolean;
}

export function StatCard(
  { title, value, subtitle, icon, trend, loading = false }: StatCardProps,
) {
  const trendPositive = trend && trend.value >= 0;

  return (
    <div class="bg-white rounded-lg border border-gray-200 p-6">
      <div class="flex items-start justify-between">
        <div class="flex-1">
          <p class="text-sm font-medium text-gray-500">{title}</p>
          {loading
            ? <div class="mt-2 h-8 w-24 bg-gray-200 rounded animate-pulse" />
            : <p class="mt-2 text-3xl font-semibold text-gray-900">{value}</p>}
          {subtitle && <p class="mt-1 text-sm text-gray-500">{subtitle}</p>}
          {trend && !loading && (
            <div class="mt-2 flex items-center gap-1">
              <span
                class={`inline-flex items-center text-sm font-medium ${
                  trendPositive ? "text-green-600" : "text-red-600"
                }`}
              >
                {trendPositive
                  ? (
                    <svg
                      class="w-4 h-4 mr-0.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M5 10l7-7m0 0l7 7m-7-7v18"
                      />
                    </svg>
                  )
                  : (
                    <svg
                      class="w-4 h-4 mr-0.5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="2"
                        d="M19 14l-7 7m0 0l-7-7m7 7V3"
                      />
                    </svg>
                  )}
                {Math.abs(trend.value)}%
              </span>
              {trend.label && (
                <span class="text-sm text-gray-500">{trend.label}</span>
              )}
            </div>
          )}
        </div>
        {icon && (
          <div class="p-3 bg-primary-100 rounded-lg text-primary-600">
            {icon}
          </div>
        )}
      </div>
    </div>
  );
}
