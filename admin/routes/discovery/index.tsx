import { Head } from "fresh/runtime";
import { define } from "../../utils.ts";
import DiscoveryLab from "../../islands/DiscoveryLab.tsx";

export default define.page(function DiscoveryPage() {
  return (
    <>
      <Head>
        <title>Discovery Lab - Saga Admin</title>
        <link
          rel="stylesheet"
          href="https://unpkg.com/maplibre-gl@5.1.0/dist/maplibre-gl.css"
        />
      </Head>

      <div class="mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Discovery Lab</h1>
        <p class="mt-1 text-sm text-gray-500">
          Test the discovery algorithm visually — select a viewer, set radius,
          and simulate discovery
        </p>
      </div>

      <DiscoveryLab />
    </>
  );
});
