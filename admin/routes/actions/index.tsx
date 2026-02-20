import { Head } from "fresh/runtime";
import { define } from "../../utils.ts";
import ActionPanel from "../../islands/ActionPanel.tsx";

export default define.page(function ActionsPage() {
  return (
    <>
      <Head>
        <title>Actions - Saga Admin</title>
      </Head>

      <div class="mb-8">
        <h1 class="text-2xl font-bold text-gray-900">Actions</h1>
        <p class="mt-1 text-sm text-gray-500">
          Trigger events as simulated users to test real-time updates
        </p>
      </div>

      <ActionPanel />
    </>
  );
});
