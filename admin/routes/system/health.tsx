import { Head } from "fresh/runtime";
import { define } from "../../utils.ts";

export default define.page(function HealthPage() {
  return (
    <>
      <Head>
        <title>System Health - Saga Admin</title>
      </Head>

      <div class="mb-8">
        <h1 class="text-2xl font-bold text-gray-900">System Health</h1>
        <p class="mt-1 text-sm text-gray-500">
          Monitor API metrics and system status
        </p>
      </div>

      <div class="bg-white rounded-lg border border-gray-200 p-12 text-center">
        <div class="mx-auto w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center mb-4">
          <svg
            class="w-8 h-8 text-gray-600"
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
        </div>
        <h2 class="text-lg font-semibold text-gray-900 mb-2">Coming Soon</h2>
        <p class="text-sm text-gray-500 max-w-md mx-auto">
          The health dashboard will show server uptime, memory usage, database
          connection status, request rates, error rates, latency percentiles,
          and recent error logs.
        </p>
      </div>
    </>
  );
});
