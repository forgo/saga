import { Head } from "fresh/runtime";
import { define } from "../../utils.ts";
import UserManager from "../../islands/UserManager.tsx";

export default define.page(function UsersPage() {
  return (
    <>
      <Head>
        <title>Users - Saga Admin</title>
      </Head>

      <div class="mb-8">
        <h1 class="text-2xl font-bold text-gray-900">Users</h1>
        <p class="mt-1 text-sm text-gray-500">
          Manage user accounts, roles, and moderation
        </p>
      </div>

      <UserManager />
    </>
  );
});
