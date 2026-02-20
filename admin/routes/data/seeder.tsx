import { Head } from "fresh/runtime";
import { define } from "../../utils.ts";
import DataSeeder from "../../islands/DataSeeder.tsx";

export default define.page(function SeederPage() {
  return (
    <>
      <Head>
        <title>Data Seeder - Saga Admin</title>
      </Head>

      <div class="mb-8">
        <h1 class="text-2xl font-bold text-gray-900">Data Seeder</h1>
        <p class="mt-1 text-sm text-gray-500">
          Generate mock users, guilds, and events for testing
        </p>
      </div>

      <DataSeeder />
    </>
  );
});
