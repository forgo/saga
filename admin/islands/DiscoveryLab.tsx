import { useSignal } from "@preact/signals";
import { useEffect, useRef } from "preact/hooks";
import {
  getCompatibility,
  getUsersWithLocations,
  simulateDiscovery,
} from "../lib/api/admin-discovery.ts";
import type {
  AdminCompatibilityResponse,
  AdminDiscoveryResponse,
  AdminDiscoveryResultItem,
  AdminMapUser,
} from "../lib/api/types.ts";

// MapLibre types — imported dynamically to avoid SSR issues
type MapInstance = {
  addSource(id: string, source: unknown): void;
  addLayer(layer: unknown): void;
  getSource(id: string): { setData(data: unknown): void } | undefined;
  removeLayer(id: string): void;
  removeSource(id: string): void;
  getLayer(id: string): unknown;
  on(event: string, layer: string, handler: (e: unknown) => void): void;
  on(event: string, handler: (e: unknown) => void): void;
  flyTo(options: { center: [number, number]; zoom?: number }): void;
  remove(): void;
  loaded(): boolean;
  once(event: string, handler: () => void): void;
};

function getScoreColor(score: number): string {
  if (score >= 80) return "#16a34a"; // green-600
  if (score >= 60) return "#ca8a04"; // yellow-600
  if (score >= 40) return "#ea580c"; // orange-600
  return "#dc2626"; // red-600
}

function createRadiusGeoJSON(
  centerLat: number,
  centerLng: number,
  radiusKm: number,
): GeoJSON.Feature {
  const points = 64;
  const coords: [number, number][] = [];
  const earthRadiusKm = 6371;

  for (let i = 0; i <= points; i++) {
    const angle = (i / points) * 2 * Math.PI;
    const latOffset = (radiusKm / earthRadiusKm) * (180 / Math.PI) *
      Math.cos(angle);
    const lngOffset =
      ((radiusKm / earthRadiusKm) * (180 / Math.PI) * Math.sin(angle)) /
      Math.cos((centerLat * Math.PI) / 180);
    coords.push([centerLng + lngOffset, centerLat + latOffset]);
  }

  return {
    type: "Feature",
    properties: {},
    geometry: {
      type: "Polygon",
      coordinates: [coords],
    },
  };
}

export default function DiscoveryLab() {
  const mapContainer = useRef<HTMLDivElement>(null);
  const mapRef = useRef<MapInstance | null>(null);
  const popupRef = useRef<{ remove(): void } | null>(null);

  // Data state
  const allUsers = useSignal<AdminMapUser[]>([]);
  const loadingUsers = useSignal(false);
  const loadError = useSignal<string | null>(null);

  // Simulation controls
  const selectedViewerId = useSignal("");
  const radiusKm = useSignal(25);
  const minCompatibility = useSignal(0);
  const requireSharedAnswer = useSignal(false);
  const resultLimit = useSignal(50);

  // Simulation results
  const discoveryResults = useSignal<AdminDiscoveryResponse | null>(null);
  const simulating = useSignal(false);
  const simError = useSignal<string | null>(null);

  // Compatibility modal
  const showCompatModal = useSignal(false);
  const compatData = useSignal<AdminCompatibilityResponse | null>(null);
  const compatLoading = useSignal(false);
  const compatUserA = useSignal("");
  const compatUserB = useSignal("");

  // Load users with locations on mount
  useEffect(() => {
    async function loadUsers() {
      loadingUsers.value = true;
      loadError.value = null;
      const { data, error } = await getUsersWithLocations(500);
      if (error) {
        loadError.value = error;
      } else if (data) {
        allUsers.value = data;
      }
      loadingUsers.value = false;
    }
    loadUsers();
  }, []);

  // Initialize map once
  useEffect(() => {
    if (!mapContainer.current || mapRef.current) return;

    let cancelled = false;

    async function initMap() {
      const maplibregl = await import("maplibre-gl");
      if (cancelled) return;

      const map = new maplibregl.Map({
        container: mapContainer.current!,
        style: "https://basemaps.cartocdn.com/gl/positron-gl-style/style.json",
        center: [-98, 39], // Center of US
        zoom: 3,
      }) as unknown as MapInstance;

      mapRef.current = map;

      map.once("load", () => {
        if (cancelled) return;

        // All users layer (gray dots)
        map.addSource("all-users", {
          type: "geojson",
          data: { type: "FeatureCollection", features: [] },
        });
        map.addLayer({
          id: "all-users-layer",
          type: "circle",
          source: "all-users",
          paint: {
            "circle-radius": 5,
            "circle-color": "#9ca3af",
            "circle-stroke-width": 1,
            "circle-stroke-color": "#6b7280",
            "circle-opacity": 0.7,
          },
        });

        // Radius circle layer
        map.addSource("radius-circle", {
          type: "geojson",
          data: { type: "FeatureCollection", features: [] },
        });
        map.addLayer({
          id: "radius-circle-fill",
          type: "fill",
          source: "radius-circle",
          paint: {
            "fill-color": "#3b82f6",
            "fill-opacity": 0.05,
          },
        });
        map.addLayer({
          id: "radius-circle-line",
          type: "line",
          source: "radius-circle",
          paint: {
            "line-color": "#3b82f6",
            "line-width": 2,
            "line-dasharray": [4, 4],
            "line-opacity": 0.6,
          },
        });

        // Viewer layer (blue dot)
        map.addSource("viewer", {
          type: "geojson",
          data: { type: "FeatureCollection", features: [] },
        });
        map.addLayer({
          id: "viewer-layer",
          type: "circle",
          source: "viewer",
          paint: {
            "circle-radius": 10,
            "circle-color": "#3b82f6",
            "circle-stroke-width": 3,
            "circle-stroke-color": "#1d4ed8",
          },
        });

        // Results layer (color-coded dots)
        map.addSource("results", {
          type: "geojson",
          data: { type: "FeatureCollection", features: [] },
        });
        map.addLayer({
          id: "results-layer",
          type: "circle",
          source: "results",
          paint: {
            "circle-radius": 7,
            "circle-color": ["get", "color"],
            "circle-stroke-width": 2,
            "circle-stroke-color": "#ffffff",
          },
        });

        // Click handler for result pins
        map.on(
          "click",
          "results-layer",
          (e: { features?: Array<{ properties?: { user_id?: string } }> }) => {
            const feature = e.features?.[0];
            if (feature?.properties?.user_id && selectedViewerId.value) {
              loadCompatibility(
                selectedViewerId.value,
                feature.properties.user_id,
              );
            }
          },
        );

        // Hover popup for results
        map.on(
          "mouseenter",
          "results-layer",
          (
            e: {
              features?: Array<{
                properties?: { name?: string; match_score?: number };
                geometry?: { coordinates?: [number, number] };
              }>;
            },
          ) => {
            const feature = e.features?.[0];
            if (!feature?.properties) return;

            const coords = feature.geometry?.coordinates;
            if (!coords) return;

            const name = feature.properties.name || "Unknown";
            const score = feature.properties.match_score
              ? Math.round(feature.properties.match_score as number)
              : 0;

            const popup = new maplibregl.Popup({
              closeButton: false,
              closeOnClick: false,
            })
              .setLngLat(coords as [number, number])
              .setHTML(
                `<div class="text-sm font-medium">${name}</div><div class="text-xs text-gray-500">Match: ${score}%</div>`,
              )
              .addTo(map as unknown as maplibregl.Map);

            popupRef.current = popup;
          },
        );

        map.on("mouseleave", "results-layer", () => {
          if (popupRef.current) {
            popupRef.current.remove();
            popupRef.current = null;
          }
        });

        // Update all-users layer if data already loaded
        if (allUsers.value.length > 0) {
          updateAllUsersLayer();
        }
      });
    }

    initMap();

    return () => {
      cancelled = true;
      if (mapRef.current) {
        mapRef.current.remove();
        mapRef.current = null;
      }
    };
  }, []);

  // Update map when allUsers change
  useEffect(() => {
    if (allUsers.value.length > 0) {
      updateAllUsersLayer();
    }
  }, [allUsers.value]);

  function updateAllUsersLayer() {
    const map = mapRef.current;
    if (!map) return;

    const source = map.getSource("all-users");
    if (!source) return;

    const features = allUsers.value
      .filter((u) => u.has_location)
      .map((u) => ({
        type: "Feature" as const,
        properties: { id: u.id, name: u.firstname || u.email },
        geometry: {
          type: "Point" as const,
          coordinates: [u.lng, u.lat],
        },
      }));

    source.setData({ type: "FeatureCollection", features });
  }

  function updateViewerLayer(viewer: AdminMapUser) {
    const map = mapRef.current;
    if (!map) return;

    const viewerSource = map.getSource("viewer");
    if (viewerSource) {
      viewerSource.setData({
        type: "FeatureCollection",
        features: [
          {
            type: "Feature",
            properties: { id: viewer.id },
            geometry: {
              type: "Point",
              coordinates: [viewer.lng, viewer.lat],
            },
          },
        ],
      });
    }

    // Update radius circle
    const radiusSource = map.getSource("radius-circle");
    if (radiusSource) {
      radiusSource.setData({
        type: "FeatureCollection",
        features: [
          createRadiusGeoJSON(viewer.lat, viewer.lng, radiusKm.value),
        ],
      });
    }

    map.flyTo({ center: [viewer.lng, viewer.lat], zoom: 10 });
  }

  function updateResultsLayer(results: AdminDiscoveryResultItem[]) {
    const map = mapRef.current;
    if (!map) return;

    const source = map.getSource("results");
    if (!source) return;

    const features = results.map((r) => ({
      type: "Feature" as const,
      properties: {
        user_id: r.user_id,
        name: r.firstname || r.username || r.email || "Unknown",
        match_score: r.match_score,
        compatibility_score: r.compatibility_score,
        color: getScoreColor(r.match_score),
      },
      geometry: {
        type: "Point" as const,
        coordinates: [r.lng, r.lat],
      },
    }));

    source.setData({ type: "FeatureCollection", features });
  }

  function clearResults() {
    discoveryResults.value = null;
    const map = mapRef.current;
    if (!map) return;
    const source = map.getSource("results");
    if (source) {
      source.setData({ type: "FeatureCollection", features: [] });
    }
  }

  function handleViewerChange(userId: string) {
    selectedViewerId.value = userId;
    clearResults();

    if (!userId) {
      const map = mapRef.current;
      if (map) {
        const viewerSource = map.getSource("viewer");
        if (viewerSource) {
          viewerSource.setData({ type: "FeatureCollection", features: [] });
        }
        const radiusSource = map.getSource("radius-circle");
        if (radiusSource) {
          radiusSource.setData({ type: "FeatureCollection", features: [] });
        }
      }
      return;
    }

    const viewer = allUsers.value.find((u) => u.id === userId);
    if (viewer && viewer.has_location) {
      updateViewerLayer(viewer);
    }
  }

  async function runSimulation() {
    if (!selectedViewerId.value) return;

    simulating.value = true;
    simError.value = null;

    const { data, error } = await simulateDiscovery({
      viewer_id: selectedViewerId.value,
      radius_km: radiusKm.value,
      min_compatibility: minCompatibility.value,
      require_shared_answer: requireSharedAnswer.value,
      limit: resultLimit.value,
    });

    if (error) {
      simError.value = error;
    } else if (data) {
      discoveryResults.value = data;
      updateResultsLayer(data.results);
    }

    simulating.value = false;
  }

  async function loadCompatibility(userAId: string, userBId: string) {
    compatLoading.value = true;
    compatUserA.value = userAId;
    compatUserB.value = userBId;
    showCompatModal.value = true;
    compatData.value = null;

    const { data, error } = await getCompatibility(userAId, userBId);
    if (error) {
      compatData.value = null;
    } else if (data) {
      compatData.value = data;
    }
    compatLoading.value = false;
  }

  const avgMatchScore =
    discoveryResults.value && discoveryResults.value.results.length > 0
      ? Math.round(
        discoveryResults.value.results.reduce(
          (sum, r) => sum + r.match_score,
          0,
        ) / discoveryResults.value.results.length,
      )
      : 0;

  return (
    <div class="space-y-4">
      {/* Controls Bar */}
      <div class="bg-white rounded-lg border border-gray-200 p-4">
        <div class="flex flex-wrap items-end gap-4">
          <div class="flex-1 min-w-[200px]">
            <label class="block text-xs font-medium text-gray-600 mb-1">
              Viewer
            </label>
            <select
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              value={selectedViewerId.value}
              onChange={(e) =>
                handleViewerChange((e.target as HTMLSelectElement).value)}
            >
              <option value="">Select a user...</option>
              {allUsers.value.map((u) => (
                <option key={u.id} value={u.id}>
                  {u.firstname || u.username || u.email}
                  {u.city ? ` (${u.city})` : ""}
                  {!u.has_location ? " [no location]" : ""}
                </option>
              ))}
            </select>
          </div>

          <div class="w-24">
            <label class="block text-xs font-medium text-gray-600 mb-1">
              Radius (km)
            </label>
            <input
              type="number"
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              value={radiusKm.value}
              min={1}
              max={100}
              onInput={(e) => {
                radiusKm.value = parseInt(
                  (e.target as HTMLInputElement).value,
                  10,
                ) || 25;
              }}
            />
          </div>

          <div class="w-28">
            <label class="block text-xs font-medium text-gray-600 mb-1">
              Min Compat (%)
            </label>
            <input
              type="number"
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              value={minCompatibility.value}
              min={0}
              max={100}
              onInput={(e) => {
                minCompatibility.value = parseInt(
                  (e.target as HTMLInputElement).value,
                  10,
                ) || 0;
              }}
            />
          </div>

          <div class="w-20">
            <label class="block text-xs font-medium text-gray-600 mb-1">
              Limit
            </label>
            <input
              type="number"
              class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
              value={resultLimit.value}
              min={1}
              max={100}
              onInput={(e) => {
                resultLimit.value = parseInt(
                  (e.target as HTMLInputElement).value,
                  10,
                ) || 50;
              }}
            />
          </div>

          <label class="flex items-center gap-2 pb-2 cursor-pointer">
            <input
              type="checkbox"
              class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              checked={requireSharedAnswer.value}
              onChange={(e) => {
                requireSharedAnswer.value = (
                  e.target as HTMLInputElement
                ).checked;
              }}
            />
            <span class="text-xs text-gray-600">Require Shared Answer</span>
          </label>

          <button
            type="button"
            class="inline-flex items-center px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            onClick={runSimulation}
            disabled={!selectedViewerId.value || simulating.value}
          >
            {simulating.value
              ? (
                <>
                  <svg
                    class="animate-spin -ml-1 mr-2 h-4 w-4"
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
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                  </svg>
                  Running...
                </>
              )
              : (
                "Run Simulation"
              )}
          </button>
        </div>

        {loadingUsers.value && (
          <p class="text-xs text-gray-400 mt-2">Loading users...</p>
        )}
        {loadError.value && (
          <p class="text-xs text-red-500 mt-2">{loadError.value}</p>
        )}
        {simError.value && (
          <p class="text-xs text-red-500 mt-2">{simError.value}</p>
        )}
      </div>

      {/* Map */}
      <div class="bg-white rounded-lg border border-gray-200 overflow-hidden">
        <div ref={mapContainer} style={{ height: "500px", width: "100%" }} />
      </div>

      {/* Legend + Stats */}
      {discoveryResults.value && (
        <div class="bg-white rounded-lg border border-gray-200 px-4 py-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-6 text-xs text-gray-500">
              <span class="flex items-center gap-1.5">
                <span
                  class="inline-block w-2.5 h-2.5 rounded-full bg-gray-400"
                  style={{ opacity: 0.7 }}
                />
                All Users
              </span>
              <span class="flex items-center gap-1.5">
                <span class="inline-block w-3 h-3 rounded-full bg-blue-500 border-2 border-blue-700" />
                Viewer
              </span>
              <span class="flex items-center gap-1.5">
                <span class="inline-block w-2.5 h-2.5 rounded-full bg-green-600" />
                High Match
              </span>
              <span class="flex items-center gap-1.5">
                <span class="inline-block w-2.5 h-2.5 rounded-full bg-yellow-600" />
                Medium
              </span>
              <span class="flex items-center gap-1.5">
                <span class="inline-block w-2.5 h-2.5 rounded-full bg-red-600" />
                Low
              </span>
              <span class="flex items-center gap-1.5">
                <span
                  class="inline-block w-4 h-0 border border-dashed border-blue-500"
                  style={{ borderStyle: "dashed" }}
                />
                Radius
              </span>
            </div>
            <div class="flex items-center gap-4 text-sm">
              <span class="text-gray-600">
                Found{" "}
                <span class="font-semibold text-gray-900">
                  {discoveryResults.value.total_count}
                </span>
              </span>
              <span class="text-gray-600">
                Radius{" "}
                <span class="font-semibold text-gray-900">
                  {discoveryResults.value.radius_km}km
                </span>
              </span>
              <span class="text-gray-600">
                Avg Match{" "}
                <span class="font-semibold text-gray-900">
                  {avgMatchScore}%
                </span>
              </span>
            </div>
          </div>
        </div>
      )}

      {/* Results Table */}
      {discoveryResults.value && discoveryResults.value.results.length > 0 && (
        <div class="bg-white rounded-lg border border-gray-200 overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-gray-200 bg-gray-50">
                  <th class="text-left px-4 py-3 font-medium text-gray-600">
                    User
                  </th>
                  <th class="text-right px-4 py-3 font-medium text-gray-600">
                    Match
                  </th>
                  <th class="text-right px-4 py-3 font-medium text-gray-600">
                    Compat
                  </th>
                  <th class="text-right px-4 py-3 font-medium text-gray-600">
                    Distance
                  </th>
                  <th class="text-left px-4 py-3 font-medium text-gray-600">
                    Shared Interests
                  </th>
                </tr>
              </thead>
              <tbody>
                {discoveryResults.value.results.map((r) => (
                  <tr
                    key={r.user_id}
                    class="border-b border-gray-100 hover:bg-gray-50 cursor-pointer transition-colors"
                    onClick={() =>
                      loadCompatibility(selectedViewerId.value, r.user_id)}
                  >
                    <td class="px-4 py-3">
                      <div class="font-medium text-gray-900">
                        {r.firstname || r.username || "—"}
                      </div>
                      <div class="text-xs text-gray-500">{r.email}</div>
                    </td>
                    <td class="text-right px-4 py-3">
                      <span
                        class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium"
                        style={{
                          backgroundColor: getScoreColor(r.match_score) + "20",
                          color: getScoreColor(r.match_score),
                        }}
                      >
                        {Math.round(r.match_score)}%
                      </span>
                    </td>
                    <td class="text-right px-4 py-3 text-gray-700">
                      {Math.round(r.compatibility_score)}%
                    </td>
                    <td class="text-right px-4 py-3 text-gray-700">
                      {r.distance_km.toFixed(1)} km
                    </td>
                    <td class="px-4 py-3">
                      {r.shared_interests && r.shared_interests.length > 0
                        ? (
                          <div class="flex flex-wrap gap-1">
                            {r.shared_interests.slice(0, 3).map((si) => (
                              <span
                                key={si.interest_id}
                                class="inline-flex items-center px-1.5 py-0.5 rounded text-xs bg-gray-100 text-gray-600"
                              >
                                {si.interest_name}
                                {si.teach_learn_match && " \u2194"}
                              </span>
                            ))}
                            {r.shared_interests.length > 3 && (
                              <span class="text-xs text-gray-400">
                                +{r.shared_interests.length - 3}
                              </span>
                            )}
                          </div>
                        )
                        : <span class="text-xs text-gray-400">None</span>}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {discoveryResults.value &&
        discoveryResults.value.results.length === 0 && (
        <div class="bg-white rounded-lg border border-gray-200 p-8 text-center">
          <p class="text-sm text-gray-500">
            No results found. Discovery requires users to have active
            availability records. Use the Data Seeder and Actions panel to
            create availability for seeded users first.
          </p>
        </div>
      )}

      {/* Compatibility Modal */}
      {showCompatModal.value && (
        <div
          class="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
          onClick={(e) => {
            if (e.target === e.currentTarget) {
              showCompatModal.value = false;
            }
          }}
        >
          <div class="bg-white rounded-xl shadow-xl max-w-lg w-full mx-4 max-h-[80vh] overflow-y-auto">
            <div class="flex items-center justify-between px-6 py-4 border-b border-gray-200">
              <h3 class="text-lg font-semibold text-gray-900">
                Compatibility Detail
              </h3>
              <button
                type="button"
                class="text-gray-400 hover:text-gray-600"
                onClick={() => (showCompatModal.value = false)}
              >
                <svg
                  class="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width="2"
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>

            <div class="px-6 py-4 space-y-4">
              {compatLoading.value && (
                <div class="flex items-center justify-center py-8">
                  <svg
                    class="animate-spin h-6 w-6 text-blue-600"
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
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
                    />
                  </svg>
                </div>
              )}

              {!compatLoading.value && compatData.value && (
                <>
                  {/* User IDs */}
                  <div class="text-xs text-gray-400 font-mono">
                    {compatUserA.value} ↔ {compatUserB.value}
                  </div>

                  {/* Overall Score */}
                  <div class="text-center py-2">
                    <div
                      class="text-4xl font-bold"
                      style={{
                        color: getScoreColor(
                          compatData.value.breakdown.score,
                        ),
                      }}
                    >
                      {Math.round(compatData.value.breakdown.score)}%
                    </div>
                    <div class="text-xs text-gray-500 mt-1">
                      Overall Compatibility
                    </div>
                  </div>

                  {/* Directional Scores */}
                  <div class="grid grid-cols-2 gap-4">
                    <div class="bg-gray-50 rounded-lg p-3 text-center">
                      <div class="text-lg font-semibold text-gray-900">
                        {Math.round(
                          compatData.value.breakdown.a_to_b_score,
                        )}
                        %
                      </div>
                      <div class="text-xs text-gray-500">A → B</div>
                    </div>
                    <div class="bg-gray-50 rounded-lg p-3 text-center">
                      <div class="text-lg font-semibold text-gray-900">
                        {Math.round(
                          compatData.value.breakdown.b_to_a_score,
                        )}
                        %
                      </div>
                      <div class="text-xs text-gray-500">B → A</div>
                    </div>
                  </div>

                  {/* Shared Questions */}
                  <div class="text-sm text-gray-600">
                    Shared Questions:{" "}
                    <span class="font-medium">
                      {compatData.value.breakdown.shared_count}
                    </span>
                  </div>

                  {/* Category Scores */}
                  {compatData.value.breakdown.category_scores &&
                    Object.keys(
                        compatData.value.breakdown.category_scores,
                      ).length > 0 &&
                    (
                      <div>
                        <h4 class="text-sm font-medium text-gray-700 mb-2">
                          Category Breakdown
                        </h4>
                        <div class="space-y-2">
                          {Object.entries(
                            compatData.value.breakdown.category_scores,
                          ).map(([cat, score]) => (
                            <div key={cat}>
                              <div class="flex items-center justify-between text-xs mb-0.5">
                                <span class="text-gray-600 capitalize">
                                  {cat}
                                </span>
                                <span class="font-medium text-gray-900">
                                  {Math.round(score)}%
                                </span>
                              </div>
                              <div class="w-full bg-gray-200 rounded-full h-2">
                                <div
                                  class="h-2 rounded-full transition-all"
                                  style={{
                                    width: `${Math.round(score)}%`,
                                    backgroundColor: getScoreColor(score),
                                  }}
                                />
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}

                  {/* Dealbreakers */}
                  {compatData.value.breakdown.deal_breakers &&
                    compatData.value.breakdown.deal_breakers.length > 0 && (
                    <div>
                      <h4 class="text-sm font-medium text-red-600 mb-2">
                        Dealbreaker Violations
                      </h4>
                      <div class="space-y-2">
                        {compatData.value.breakdown.deal_breakers.map(
                          (db) => (
                            <div
                              key={db.question_id}
                              class="bg-red-50 border border-red-200 rounded-lg p-3 text-xs"
                            >
                              <div class="font-medium text-red-800">
                                {db.question_text}
                              </div>
                              <div class="mt-1 text-red-600">
                                Required: "{db.user_answer}" — Got: "
                                {db.partner_answer}"
                              </div>
                            </div>
                          ),
                        )}
                      </div>
                    </div>
                  )}

                  {/* Deal Breaker Flag */}
                  {compatData.value.breakdown.deal_breaker && (
                    <div class="bg-red-100 text-red-800 text-sm font-medium px-3 py-2 rounded-lg text-center">
                      Dealbreaker Violated — Score Zeroed
                    </div>
                  )}

                  {/* Yikes Summary */}
                  {compatData.value.yikes.has_yikes && (
                    <div class="bg-amber-50 border border-amber-200 rounded-lg p-3">
                      <h4 class="text-sm font-medium text-amber-800 mb-1">
                        Yikes Summary
                      </h4>
                      <div class="text-xs text-amber-700 space-y-1">
                        <div>
                          Count:{" "}
                          <span class="font-medium">
                            {compatData.value.yikes.yikes_count}
                          </span>
                        </div>
                        <div>
                          Severity:{" "}
                          <span
                            class={`font-medium ${
                              compatData.value.yikes.severity === "severe"
                                ? "text-red-600"
                                : compatData.value.yikes.severity ===
                                    "moderate"
                                ? "text-amber-600"
                                : "text-yellow-600"
                            }`}
                          >
                            {compatData.value.yikes.severity}
                          </span>
                        </div>
                        {compatData.value.yikes.categories &&
                          compatData.value.yikes.categories.length > 0 && (
                          <div>
                            Categories:{" "}
                            {compatData.value.yikes.categories.join(", ")}
                          </div>
                        )}
                      </div>
                    </div>
                  )}

                  {!compatData.value.yikes.has_yikes && (
                    <div class="bg-green-50 text-green-700 text-xs px-3 py-2 rounded-lg text-center">
                      No yikes detected
                    </div>
                  )}
                </>
              )}

              {!compatLoading.value && !compatData.value && (
                <div class="text-center py-8 text-sm text-gray-500">
                  No shared questions between these users.
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
