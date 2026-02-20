import { useComputed, useSignal } from "@preact/signals";
import { api, isError } from "../lib/api/client.ts";
import type { SeedResult } from "../lib/api/types.ts";
import { Button } from "../components/Button.tsx";

type SeedType = "users" | "guilds" | "events" | "scenario";

interface Scenario {
  id: string;
  name: string;
  description: string;
}

interface BoundingBox {
  min_lat: number;
  max_lat: number;
  min_lng: number;
  max_lng: number;
}

const PRESET_REGIONS: Record<string, BoundingBox> = {
  sf: {
    min_lat: 37.7079,
    max_lat: 37.8324,
    min_lng: -122.5149,
    max_lng: -122.357,
  },
  nyc: {
    min_lat: 40.4961,
    max_lat: 40.9155,
    min_lng: -74.2557,
    max_lng: -73.7004,
  },
  la: {
    min_lat: 33.7037,
    max_lat: 34.3373,
    min_lng: -118.6682,
    max_lng: -118.1553,
  },
};

export default function DataSeeder() {
  const seedType = useSignal<SeedType>("users");
  const isLoading = useSignal(false);
  const result = useSignal<SeedResult | null>(null);
  const error = useSignal<string | null>(null);

  // User seeding options
  const userCount = useSignal(10);
  const userRegion = useSignal<string>("sf");
  const userPrefix = useSignal("seed_");
  const activeNow = useSignal(20);
  const activeToday = useSignal(30);
  const activeThisWeek = useSignal(30);
  const away = useSignal(20);

  // Guild seeding options
  const guildCount = useSignal(5);
  const membersPerGuild = useSignal(5);
  const guildVisibility = useSignal<"public" | "private">("public");
  const guildPrefix = useSignal("seed_");

  // Event seeding options
  const eventCount = useSignal(5);
  const eventGuildId = useSignal("");
  const eventStatus = useSignal("published");
  const eventPrefix = useSignal("seed_");

  // Scenario options
  const selectedScenario = useSignal("sf_discovery_pool");
  const scenarios = useSignal<Scenario[]>([
    {
      id: "sf_discovery_pool",
      name: "SF Discovery Pool",
      description: "20 users in San Francisco for discovery testing",
    },
    {
      id: "active_guild",
      name: "Active Guild",
      description: "A guild with 10 members and 5 upcoming events",
    },
    {
      id: "event_with_attendees",
      name: "Event with Attendees",
      description: "An event with 20 attendees for RSVP testing",
    },
  ]);

  // Cleanup options
  const cleanupPrefix = useSignal("seed_");

  const activityTotal = useComputed(
    () =>
      activeNow.value + activeToday.value + activeThisWeek.value + away.value,
  );

  const activityError = useComputed(() =>
    activityTotal.value !== 100
      ? `Activity distribution must total 100% (currently ${activityTotal.value}%)`
      : null
  );

  const seedUsers = async () => {
    if (activityError.value) {
      error.value = activityError.value;
      return;
    }

    isLoading.value = true;
    error.value = null;
    result.value = null;

    const region = PRESET_REGIONS[userRegion.value];
    const response = await api.post<{ data: SeedResult }>(
      "/v1/admin/seed/users",
      {
        count: userCount.value,
        region,
        activity_distribution: {
          active_now: activeNow.value,
          active_today: activeToday.value,
          active_this_week: activeThisWeek.value,
          away: away.value,
        },
        prefix: userPrefix.value,
      },
    );

    isLoading.value = false;

    if (isError(response)) {
      error.value = response.error.message;
    } else {
      result.value = response.data?.data ?? null;
    }
  };

  const seedGuilds = async () => {
    isLoading.value = true;
    error.value = null;
    result.value = null;

    const response = await api.post<{ data: SeedResult }>(
      "/v1/admin/seed/guilds",
      {
        count: guildCount.value,
        members_per_guild: membersPerGuild.value,
        visibility: guildVisibility.value,
        prefix: guildPrefix.value,
      },
    );

    isLoading.value = false;

    if (isError(response)) {
      error.value = response.error.message;
    } else {
      result.value = response.data?.data ?? null;
    }
  };

  const seedEvents = async () => {
    isLoading.value = true;
    error.value = null;
    result.value = null;

    const response = await api.post<{ data: SeedResult }>(
      "/v1/admin/seed/events",
      {
        count: eventCount.value,
        guild_id: eventGuildId.value || undefined,
        status: eventStatus.value,
        prefix: eventPrefix.value,
      },
    );

    isLoading.value = false;

    if (isError(response)) {
      error.value = response.error.message;
    } else {
      result.value = response.data?.data ?? null;
    }
  };

  const runScenario = async () => {
    isLoading.value = true;
    error.value = null;
    result.value = null;

    const response = await api.post<{ data: SeedResult }>(
      "/v1/admin/seed/scenario",
      {
        scenario: selectedScenario.value,
      },
    );

    isLoading.value = false;

    if (isError(response)) {
      error.value = response.error.message;
    } else {
      result.value = response.data?.data ?? null;
    }
  };

  const cleanup = async () => {
    if (
      !confirm(
        `This will delete all data with prefix "${cleanupPrefix.value}". Continue?`,
      )
    ) {
      return;
    }

    isLoading.value = true;
    error.value = null;
    result.value = null;

    const response = await api.delete<
      { data: { deleted: number; duration_ms: number } }
    >(
      `/v1/admin/seed/cleanup?prefix=${
        encodeURIComponent(cleanupPrefix.value)
      }`,
    );

    isLoading.value = false;

    if (isError(response)) {
      error.value = response.error.message;
    } else if (response.data?.data) {
      result.value = {
        created: response.data.data.deleted,
        ids: [],
        duration: response.data.data.duration_ms,
      };
    }
  };

  const handleSeed = () => {
    switch (seedType.value) {
      case "users":
        seedUsers();
        break;
      case "guilds":
        seedGuilds();
        break;
      case "events":
        seedEvents();
        break;
      case "scenario":
        runScenario();
        break;
    }
  };

  return (
    <div class="space-y-6">
      {/* Seed Type Tabs */}
      <div class="border-b border-gray-200">
        <nav class="-mb-px flex space-x-8">
          {(["users", "guilds", "events", "scenario"] as const).map((type) => (
            <button
              type="button"
              key={type}
              onClick={() => {
                seedType.value = type;
                result.value = null;
                error.value = null;
              }}
              class={`py-2 px-1 border-b-2 font-medium text-sm capitalize ${
                seedType.value === type
                  ? "border-primary-600 text-primary-600"
                  : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
              }`}
            >
              {type === "scenario" ? "Scenarios" : type}
            </button>
          ))}
        </nav>
      </div>

      {/* Configuration Form */}
      <div class="bg-white rounded-lg border border-gray-200 p-6">
        {seedType.value === "users" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">Seed Users</h3>
            <p class="text-sm text-gray-500">
              Generate mock users with profiles, locations, and activity status.
            </p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Count
                </label>
                <input
                  type="number"
                  min="1"
                  max="1000"
                  value={userCount.value}
                  onInput={(
                    e,
                  ) => (userCount.value =
                    parseInt((e.target as HTMLInputElement).value) || 1)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Region
                </label>
                <select
                  value={userRegion.value}
                  onChange={(
                    e,
                  ) => (userRegion.value =
                    (e.target as HTMLSelectElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="sf">San Francisco</option>
                  <option value="nyc">New York City</option>
                  <option value="la">Los Angeles</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Prefix
                </label>
                <input
                  type="text"
                  value={userPrefix.value}
                  onInput={(
                    e,
                  ) => (userPrefix.value =
                    (e.target as HTMLInputElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  placeholder="seed_"
                />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">
                Activity Distribution (must total 100%)
              </label>
              {activityError.value && (
                <p class="text-sm text-red-600 mb-2">{activityError.value}</p>
              )}
              <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div>
                  <label class="block text-xs text-gray-500">Active Now</label>
                  <input
                    type="number"
                    min="0"
                    max="100"
                    value={activeNow.value}
                    onInput={(
                      e,
                    ) => (activeNow.value =
                      parseInt((e.target as HTMLInputElement).value) || 0)}
                    class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  />
                </div>
                <div>
                  <label class="block text-xs text-gray-500">
                    Active Today
                  </label>
                  <input
                    type="number"
                    min="0"
                    max="100"
                    value={activeToday.value}
                    onInput={(
                      e,
                    ) => (activeToday.value =
                      parseInt((e.target as HTMLInputElement).value) || 0)}
                    class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  />
                </div>
                <div>
                  <label class="block text-xs text-gray-500">
                    Active This Week
                  </label>
                  <input
                    type="number"
                    min="0"
                    max="100"
                    value={activeThisWeek.value}
                    onInput={(
                      e,
                    ) => (activeThisWeek.value =
                      parseInt((e.target as HTMLInputElement).value) || 0)}
                    class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  />
                </div>
                <div>
                  <label class="block text-xs text-gray-500">Away</label>
                  <input
                    type="number"
                    min="0"
                    max="100"
                    value={away.value}
                    onInput={(
                      e,
                    ) => (away.value =
                      parseInt((e.target as HTMLInputElement).value) || 0)}
                    class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {seedType.value === "guilds" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">Seed Guilds</h3>
            <p class="text-sm text-gray-500">
              Generate mock guilds with members. Will create users if needed.
            </p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Count
                </label>
                <input
                  type="number"
                  min="1"
                  max="100"
                  value={guildCount.value}
                  onInput={(
                    e,
                  ) => (guildCount.value =
                    parseInt((e.target as HTMLInputElement).value) || 1)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Members per Guild
                </label>
                <input
                  type="number"
                  min="1"
                  max="50"
                  value={membersPerGuild.value}
                  onInput={(
                    e,
                  ) => (membersPerGuild.value =
                    parseInt((e.target as HTMLInputElement).value) || 1)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Visibility
                </label>
                <select
                  value={guildVisibility.value}
                  onChange={(
                    e,
                  ) => (guildVisibility.value = (e.target as HTMLSelectElement)
                    .value as "public" | "private")}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="public">Public</option>
                  <option value="private">Private</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Prefix
                </label>
                <input
                  type="text"
                  value={guildPrefix.value}
                  onInput={(
                    e,
                  ) => (guildPrefix.value =
                    (e.target as HTMLInputElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  placeholder="seed_"
                />
              </div>
            </div>
          </div>
        )}

        {seedType.value === "events" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">Seed Events</h3>
            <p class="text-sm text-gray-500">
              Generate mock events. Will create guilds if needed.
            </p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Count
                </label>
                <input
                  type="number"
                  min="1"
                  max="100"
                  value={eventCount.value}
                  onInput={(
                    e,
                  ) => (eventCount.value =
                    parseInt((e.target as HTMLInputElement).value) || 1)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Status
                </label>
                <select
                  value={eventStatus.value}
                  onChange={(
                    e,
                  ) => (eventStatus.value =
                    (e.target as HTMLSelectElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="published">Published</option>
                  <option value="draft">Draft</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Guild ID (optional)
                </label>
                <input
                  type="text"
                  value={eventGuildId.value}
                  onInput={(
                    e,
                  ) => (eventGuildId.value =
                    (e.target as HTMLInputElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  placeholder="Leave empty to create across guilds"
                />
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Prefix
                </label>
                <input
                  type="text"
                  value={eventPrefix.value}
                  onInput={(
                    e,
                  ) => (eventPrefix.value =
                    (e.target as HTMLInputElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  placeholder="seed_"
                />
              </div>
            </div>
          </div>
        )}

        {seedType.value === "scenario" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">Run Scenario</h3>
            <p class="text-sm text-gray-500">
              Deploy predefined test setups with one click.
            </p>

            <div class="space-y-3">
              {scenarios.value.map((scenario) => (
                <label
                  key={scenario.id}
                  class={`flex items-start p-4 border rounded-lg cursor-pointer transition-colors ${
                    selectedScenario.value === scenario.id
                      ? "border-primary-500 bg-primary-50"
                      : "border-gray-200 hover:border-gray-300"
                  }`}
                >
                  <input
                    type="radio"
                    name="scenario"
                    value={scenario.id}
                    checked={selectedScenario.value === scenario.id}
                    onChange={(
                      e,
                    ) => (selectedScenario.value =
                      (e.target as HTMLInputElement).value)}
                    class="mt-1 h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300"
                  />
                  <div class="ml-3">
                    <span class="block text-sm font-medium text-gray-900">
                      {scenario.name}
                    </span>
                    <span class="block text-sm text-gray-500">
                      {scenario.description}
                    </span>
                  </div>
                </label>
              ))}
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div class="mt-6 flex items-center gap-4">
          <Button
            onClick={handleSeed}
            disabled={isLoading.value ||
              (seedType.value === "users" && !!activityError.value)}
          >
            {isLoading.value
              ? (
                <span class="flex items-center gap-2">
                  <svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
                    <circle
                      class="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      stroke-width="4"
                      fill="none"
                    />
                    <path
                      class="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  Seeding...
                </span>
              )
              : seedType.value === "scenario"
              ? (
                "Run Scenario"
              )
              : (
                `Seed ${
                  seedType.value.charAt(0).toUpperCase() +
                  seedType.value.slice(1)
                }`
              )}
          </Button>
        </div>
      </div>

      {/* Results */}
      {result.value && (
        <div class="bg-green-50 border border-green-200 rounded-lg p-4">
          <div class="flex items-center gap-2 mb-2">
            <svg
              class="h-5 w-5 text-green-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M5 13l4 4L19 7"
              />
            </svg>
            <h4 class="text-sm font-medium text-green-800">Success</h4>
          </div>
          <div class="text-sm text-green-700">
            <p>
              Created {result.value.created} records in{" "}
              {result.value.duration}ms
            </p>
            {result.value.ids.length > 0 && result.value.ids.length <= 10 && (
              <details class="mt-2">
                <summary class="cursor-pointer text-green-600 hover:text-green-800">
                  View IDs ({result.value.ids.length})
                </summary>
                <ul class="mt-1 space-y-1 font-mono text-xs">
                  {result.value.ids.map((id) => <li key={id}>{id}</li>)}
                </ul>
              </details>
            )}
          </div>
        </div>
      )}

      {error.value && (
        <div class="bg-red-50 border border-red-200 rounded-lg p-4">
          <div class="flex items-center gap-2 mb-2">
            <svg
              class="h-5 w-5 text-red-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
            <h4 class="text-sm font-medium text-red-800">Error</h4>
          </div>
          <p class="text-sm text-red-700">{error.value}</p>
        </div>
      )}

      {/* Cleanup Section */}
      <div class="bg-white rounded-lg border border-gray-200 p-6">
        <h3 class="text-lg font-medium text-gray-900">Cleanup</h3>
        <p class="text-sm text-gray-500 mt-1">
          Remove all seeded data matching a prefix.
        </p>

        <div class="mt-4 flex items-end gap-4">
          <div class="flex-1 max-w-xs">
            <label class="block text-sm font-medium text-gray-700">
              Prefix
            </label>
            <input
              type="text"
              value={cleanupPrefix.value}
              onInput={(
                e,
              ) => (cleanupPrefix.value = (e.target as HTMLInputElement).value)}
              class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
              placeholder="seed_"
            />
          </div>
          <Button variant="danger" onClick={cleanup} disabled={isLoading.value}>
            {isLoading.value ? "Cleaning..." : "Cleanup Data"}
          </Button>
        </div>
      </div>
    </div>
  );
}
