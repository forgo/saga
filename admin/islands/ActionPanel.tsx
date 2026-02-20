import { useComputed, useSignal } from "@preact/signals";
import { useEffect } from "preact/hooks";
import { api, isError } from "../lib/api/client.ts";
import { Button } from "../components/Button.tsx";

type ActionType =
  | "location"
  | "trust-rating"
  | "guild-join"
  | "rsvp"
  | "event-create";

interface User {
  id: string;
  email: string;
  username?: string;
  firstname?: string;
  lastname?: string;
}

interface Guild {
  id: string;
  name: string;
  description?: string;
}

interface Event {
  id: string;
  title: string;
  guild_id?: string;
  starts_at?: string;
}

interface ActionResult {
  success: boolean;
  action: string;
  acting_as: string;
  target?: string;
  data?: Record<string, unknown>;
  timestamp: string;
}

interface ActionLogEntry {
  id: string;
  result: ActionResult;
  timestamp: Date;
}

export default function ActionPanel() {
  const actionType = useSignal<ActionType>("location");
  const isLoading = useSignal(false);
  const error = useSignal<string | null>(null);

  // Data for selectors
  const users = useSignal<User[]>([]);
  const guilds = useSignal<Guild[]>([]);
  const events = useSignal<Event[]>([]);

  // Action log
  const actionLog = useSignal<ActionLogEntry[]>([]);

  // Selected values
  const selectedUserId = useSignal<string>("");
  const targetUserId = useSignal<string>("");
  const selectedGuildId = useSignal<string>("");
  const selectedEventId = useSignal<string>("");

  // Location action
  const lat = useSignal(37.7749);
  const lng = useSignal(-122.4194);
  const city = useSignal("San Francisco");

  // Trust rating action
  const trustLevel = useSignal("medium");
  const trustReview = useSignal("");

  // RSVP action
  const rsvpResponse = useSignal("yes");

  // Event create action
  const eventTitle = useSignal("");

  // Load data on mount
  useEffect(() => {
    loadUsers();
    loadGuilds();
    loadEvents();
  }, []);

  const loadUsers = async () => {
    const response = await api.get<{ data: User[] }>("/v1/admin/actions/users");
    if (!isError(response) && response.data?.data) {
      users.value = response.data.data;
      if (response.data.data.length > 0 && !selectedUserId.value) {
        selectedUserId.value = response.data.data[0].id;
      }
    }
  };

  const loadGuilds = async () => {
    const response = await api.get<{ data: Guild[] }>(
      "/v1/admin/actions/guilds",
    );
    if (!isError(response) && response.data?.data) {
      guilds.value = response.data.data;
      if (response.data.data.length > 0 && !selectedGuildId.value) {
        selectedGuildId.value = response.data.data[0].id;
      }
    }
  };

  const loadEvents = async () => {
    const response = await api.get<{ data: Event[] }>(
      "/v1/admin/actions/events",
    );
    if (!isError(response) && response.data?.data) {
      events.value = response.data.data;
      if (response.data.data.length > 0 && !selectedEventId.value) {
        selectedEventId.value = response.data.data[0].id;
      }
    }
  };

  const addToLog = (result: ActionResult) => {
    const entry: ActionLogEntry = {
      id: crypto.randomUUID(),
      result,
      timestamp: new Date(),
    };
    actionLog.value = [entry, ...actionLog.value].slice(0, 20); // Keep last 20
  };

  const executeAction = async () => {
    if (!selectedUserId.value) {
      error.value = "Please select a user to act as";
      return;
    }

    isLoading.value = true;
    error.value = null;

    let response;

    switch (actionType.value) {
      case "location":
        response = await api.post<{ data: ActionResult }>(
          "/v1/admin/actions/location",
          {
            user_id: selectedUserId.value,
            lat: lat.value,
            lng: lng.value,
            city: city.value,
          },
        );
        break;

      case "trust-rating":
        if (!targetUserId.value) {
          error.value = "Please select a target user for the trust rating";
          isLoading.value = false;
          return;
        }
        response = await api.post<{ data: ActionResult }>(
          "/v1/admin/actions/trust-rating",
          {
            rater_id: selectedUserId.value,
            ratee_id: targetUserId.value,
            trust_level: trustLevel.value,
            trust_review: trustReview.value,
          },
        );
        break;

      case "guild-join":
        if (!selectedGuildId.value) {
          error.value = "Please select a guild to join";
          isLoading.value = false;
          return;
        }
        response = await api.post<{ data: ActionResult }>(
          "/v1/admin/actions/guild-join",
          {
            user_id: selectedUserId.value,
            guild_id: selectedGuildId.value,
          },
        );
        break;

      case "rsvp":
        if (!selectedEventId.value) {
          error.value = "Please select an event to RSVP to";
          isLoading.value = false;
          return;
        }
        response = await api.post<{ data: ActionResult }>(
          "/v1/admin/actions/rsvp",
          {
            user_id: selectedUserId.value,
            event_id: selectedEventId.value,
            response: rsvpResponse.value,
          },
        );
        break;

      case "event-create":
        if (!selectedGuildId.value || !eventTitle.value) {
          error.value = "Please select a guild and enter an event title";
          isLoading.value = false;
          return;
        }
        response = await api.post<{ data: ActionResult }>(
          "/v1/admin/actions/event-create",
          {
            user_id: selectedUserId.value,
            guild_id: selectedGuildId.value,
            title: eventTitle.value,
          },
        );
        // Refresh events list after creating
        loadEvents();
        break;
    }

    isLoading.value = false;

    if (response && isError(response)) {
      error.value = response.error.message;
    } else if (response?.data?.data) {
      addToLog(response.data.data);
    }
  };

  const selectedUser = useComputed(() =>
    users.value.find((u) => u.id === selectedUserId.value)
  );

  const getUserLabel = (user: User) => {
    if (user.firstname && user.lastname) {
      return `${user.firstname} ${user.lastname} (${user.email})`;
    }
    return user.email;
  };

  return (
    <div class="space-y-6">
      {/* Acting As Selector */}
      <div class="bg-white rounded-lg border border-gray-200 p-6">
        <h3 class="text-lg font-medium text-gray-900 mb-4">Acting As</h3>
        <div class="flex items-center gap-4">
          <select
            value={selectedUserId.value}
            onChange={(
              e,
            ) => (selectedUserId.value = (e.target as HTMLSelectElement).value)}
            class="flex-1 rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
          >
            <option value="">Select a user...</option>
            {users.value.map((user) => (
              <option key={user.id} value={user.id}>
                {getUserLabel(user)}
              </option>
            ))}
          </select>
          <Button variant="outline" onClick={loadUsers}>
            Refresh
          </Button>
        </div>
        {selectedUser.value && (
          <p class="mt-2 text-sm text-gray-500">
            ID:{" "}
            <code class="bg-gray-100 px-1 rounded">
              {selectedUser.value.id}
            </code>
          </p>
        )}
      </div>

      {/* Action Type Tabs */}
      <div class="border-b border-gray-200">
        <nav class="-mb-px flex space-x-8">
          {(
            [
              { type: "location", label: "Location" },
              { type: "trust-rating", label: "Trust Rating" },
              { type: "guild-join", label: "Join Guild" },
              { type: "rsvp", label: "RSVP" },
              { type: "event-create", label: "Create Event" },
            ] as const
          ).map(({ type, label }) => (
            <button
              type="button"
              key={type}
              onClick={() => {
                actionType.value = type;
                error.value = null;
              }}
              class={`py-2 px-1 border-b-2 font-medium text-sm ${
                actionType.value === type
                  ? "border-primary-600 text-primary-600"
                  : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
              }`}
            >
              {label}
            </button>
          ))}
        </nav>
      </div>

      {/* Action Configuration */}
      <div class="bg-white rounded-lg border border-gray-200 p-6">
        {actionType.value === "location" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">Update Location</h3>
            <p class="text-sm text-gray-500">
              Move the user to a new location. This will update their profile
              and trigger discovery recalculation.
            </p>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Latitude
                </label>
                <input
                  type="number"
                  step="0.0001"
                  value={lat.value}
                  onInput={(
                    e,
                  ) => (lat.value =
                    parseFloat((e.target as HTMLInputElement).value) || 0)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Longitude
                </label>
                <input
                  type="number"
                  step="0.0001"
                  value={lng.value}
                  onInput={(
                    e,
                  ) => (lng.value =
                    parseFloat((e.target as HTMLInputElement).value) || 0)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  City
                </label>
                <input
                  type="text"
                  value={city.value}
                  onInput={(
                    e,
                  ) => (city.value = (e.target as HTMLInputElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                />
              </div>
            </div>

            <div class="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  lat.value = 37.7749;
                  lng.value = -122.4194;
                  city.value = "San Francisco";
                }}
              >
                SF
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  lat.value = 40.7128;
                  lng.value = -74.006;
                  city.value = "New York";
                }}
              >
                NYC
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  lat.value = 34.0522;
                  lng.value = -118.2437;
                  city.value = "Los Angeles";
                }}
              >
                LA
              </Button>
            </div>
          </div>
        )}

        {actionType.value === "trust-rating" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">
              Create Trust Rating
            </h3>
            <p class="text-sm text-gray-500">
              Rate another user. This will send a real-time notification to the
              target user.
            </p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Target User
                </label>
                <select
                  value={targetUserId.value}
                  onChange={(
                    e,
                  ) => (targetUserId.value =
                    (e.target as HTMLSelectElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="">Select target user...</option>
                  {users.value
                    .filter((u) => u.id !== selectedUserId.value)
                    .map((user) => (
                      <option key={user.id} value={user.id}>
                        {getUserLabel(user)}
                      </option>
                    ))}
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Trust Level
                </label>
                <select
                  value={trustLevel.value}
                  onChange={(
                    e,
                  ) => (trustLevel.value =
                    (e.target as HTMLSelectElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="high">High Trust</option>
                  <option value="medium">Medium Trust</option>
                  <option value="low">Low Trust</option>
                  <option value="distrust">Distrust</option>
                </select>
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700">
                Review (optional)
              </label>
              <textarea
                value={trustReview.value}
                onInput={(
                  e,
                ) => (trustReview.value =
                  (e.target as HTMLTextAreaElement).value)}
                rows={2}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                placeholder="Optional review text..."
              />
            </div>
          </div>
        )}

        {actionType.value === "guild-join" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">Join Guild</h3>
            <p class="text-sm text-gray-500">
              Have the user join a guild. This will broadcast a member_joined
              event.
            </p>

            <div>
              <label class="block text-sm font-medium text-gray-700">
                Guild
              </label>
              <select
                value={selectedGuildId.value}
                onChange={(
                  e,
                ) => (selectedGuildId.value =
                  (e.target as HTMLSelectElement).value)}
                class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
              >
                <option value="">Select a guild...</option>
                {guilds.value.map((guild) => (
                  <option key={guild.id} value={guild.id}>
                    {guild.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        )}

        {actionType.value === "rsvp" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">RSVP to Event</h3>
            <p class="text-sm text-gray-500">
              RSVP to an event. This will notify the event creator.
            </p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Event
                </label>
                <select
                  value={selectedEventId.value}
                  onChange={(
                    e,
                  ) => (selectedEventId.value =
                    (e.target as HTMLSelectElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="">Select an event...</option>
                  {events.value.map((event) => (
                    <option key={event.id} value={event.id}>
                      {event.title}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Response
                </label>
                <select
                  value={rsvpResponse.value}
                  onChange={(
                    e,
                  ) => (rsvpResponse.value =
                    (e.target as HTMLSelectElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="yes">Yes, I'm going</option>
                  <option value="maybe">Maybe</option>
                  <option value="no">No, can't make it</option>
                </select>
              </div>
            </div>
          </div>
        )}

        {actionType.value === "event-create" && (
          <div class="space-y-4">
            <h3 class="text-lg font-medium text-gray-900">Create Event</h3>
            <p class="text-sm text-gray-500">
              Create a new event in a guild. This will broadcast an
              event.created notification.
            </p>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Guild
                </label>
                <select
                  value={selectedGuildId.value}
                  onChange={(
                    e,
                  ) => (selectedGuildId.value =
                    (e.target as HTMLSelectElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                >
                  <option value="">Select a guild...</option>
                  {guilds.value.map((guild) => (
                    <option key={guild.id} value={guild.id}>
                      {guild.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label class="block text-sm font-medium text-gray-700">
                  Event Title
                </label>
                <input
                  type="text"
                  value={eventTitle.value}
                  onInput={(
                    e,
                  ) => (eventTitle.value =
                    (e.target as HTMLInputElement).value)}
                  class="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border"
                  placeholder="Weekly Meetup"
                />
              </div>
            </div>
          </div>
        )}

        {/* Execute Button */}
        <div class="mt-6">
          <Button
            onClick={executeAction}
            disabled={isLoading.value || !selectedUserId.value}
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
                  Executing...
                </span>
              )
              : (
                "Execute Action"
              )}
          </Button>
        </div>

        {/* Error Display */}
        {error.value && (
          <div class="mt-4 bg-red-50 border border-red-200 rounded-lg p-4">
            <p class="text-sm text-red-700">{error.value}</p>
          </div>
        )}
      </div>

      {/* Action Log */}
      <div class="bg-white rounded-lg border border-gray-200 p-6">
        <h3 class="text-lg font-medium text-gray-900 mb-4">Action Log</h3>

        {actionLog.value.length === 0
          ? <p class="text-sm text-gray-500">No actions executed yet.</p>
          : (
            <div class="space-y-3">
              {actionLog.value.map((entry) => (
                <div
                  key={entry.id}
                  class={`p-3 rounded-lg border ${
                    entry.result.success
                      ? "bg-green-50 border-green-200"
                      : "bg-red-50 border-red-200"
                  }`}
                >
                  <div class="flex items-center justify-between">
                    <span class="font-medium text-sm">
                      {entry.result.action.replace(/_/g, " ").replace(
                        /\b\w/g,
                        (l) => l.toUpperCase(),
                      )}
                    </span>
                    <span class="text-xs text-gray-500">
                      {entry.timestamp.toLocaleTimeString()}
                    </span>
                  </div>
                  <div class="mt-1 text-xs text-gray-600">
                    Acting as:{" "}
                    <code class="bg-white/50 px-1 rounded">
                      {entry.result.acting_as}
                    </code>
                    {entry.result.target && (
                      <>
                        {" → "}
                        <code class="bg-white/50 px-1 rounded">
                          {entry.result.target}
                        </code>
                      </>
                    )}
                  </div>
                  {entry.result.data && (
                    <div class="mt-1 text-xs text-gray-500">
                      {JSON.stringify(entry.result.data)}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
      </div>
    </div>
  );
}
