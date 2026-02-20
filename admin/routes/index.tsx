import { Head } from "fresh/runtime";
import { define } from "../utils.ts";
import { StatCard } from "../components/StatCard.tsx";
import { Button } from "../components/Button.tsx";

export default define.page(function Dashboard() {
  return (
    <>
      <Head>
        <title>Dashboard - Saga Admin</title>
      </Head>

      {/* Page header */}
      <div class="mb-8">
        <h1 class="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p class="mt-1 text-sm text-gray-500">
          Overview of your Saga application
        </p>
      </div>

      {/* Stats grid */}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <StatCard
          title="Total Users"
          value="--"
          subtitle="Registered accounts"
          icon={
            <svg
              class="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
              />
            </svg>
          }
        />
        <StatCard
          title="Active Guilds"
          value="--"
          subtitle="Community groups"
          icon={
            <svg
              class="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
              />
            </svg>
          }
        />
        <StatCard
          title="Events"
          value="--"
          subtitle="Upcoming events"
          icon={
            <svg
              class="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          }
        />
        <StatCard
          title="API Health"
          value="--"
          subtitle="System status"
          icon={
            <svg
              class="w-6 h-6"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
          }
        />
      </div>

      {/* Quick actions */}
      <div class="bg-white rounded-lg border border-gray-200 p-6 mb-8">
        <h2 class="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
        <div class="flex flex-wrap gap-3">
          <Button variant="primary" size="md">
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
              />
            </svg>
            Seed Test Data
          </Button>
          <Button variant="outline" size="md">
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
              />
            </svg>
            View Users
          </Button>
          <Button variant="outline" size="md">
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M13 10V3L4 14h7v7l9-11h-7z"
              />
            </svg>
            Trigger Action
          </Button>
          <Button variant="outline" size="md">
            <svg
              class="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"
              />
            </svg>
            Discovery Lab
          </Button>
        </div>
      </div>

      {/* Feature sections */}
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Data Tools */}
        <div class="bg-white rounded-lg border border-gray-200 p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 bg-purple-100 rounded-lg">
              <svg
                class="w-5 h-5 text-purple-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z"
                />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Data Tools</h3>
          </div>
          <p class="text-sm text-gray-500 mb-4">
            Generate mock users, guilds, and events for testing. Configure
            parameters like location bounds and activity distribution.
          </p>
          <a
            href="/data/seeder"
            class="text-sm font-medium text-primary-600 hover:text-primary-700"
          >
            Go to Seeder &rarr;
          </a>
        </div>

        {/* User Management */}
        <div class="bg-white rounded-lg border border-gray-200 p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 bg-blue-100 rounded-lg">
              <svg
                class="w-5 h-5 text-blue-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">User Management</h3>
          </div>
          <p class="text-sm text-gray-500 mb-4">
            View and edit users, manage roles, impersonate for testing. Search
            and filter through all registered accounts.
          </p>
          <a
            href="/users"
            class="text-sm font-medium text-primary-600 hover:text-primary-700"
          >
            View Users &rarr;
          </a>
        </div>

        {/* Action Triggers */}
        <div class="bg-white rounded-lg border border-gray-200 p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 bg-yellow-100 rounded-lg">
              <svg
                class="w-5 h-5 text-yellow-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Action Triggers</h3>
          </div>
          <p class="text-sm text-gray-500 mb-4">
            Trigger events as simulated users to test real-time updates. Send
            messages, RSVPs, location updates, and more.
          </p>
          <a
            href="/actions"
            class="text-sm font-medium text-primary-600 hover:text-primary-700"
          >
            Open Actions &rarr;
          </a>
        </div>

        {/* Discovery Lab */}
        <div class="bg-white rounded-lg border border-gray-200 p-6">
          <div class="flex items-center gap-3 mb-4">
            <div class="p-2 bg-green-100 rounded-lg">
              <svg
                class="w-5 h-5 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-.553-.894L15 4m0 13V4m0 0L9 7"
                />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900">Discovery Lab</h3>
          </div>
          <p class="text-sm text-gray-500 mb-4">
            Test the discovery algorithm visually on a map. See compatibility
            scores and adjust radius/filters.
          </p>
          <a
            href="/discovery"
            class="text-sm font-medium text-primary-600 hover:text-primary-700"
          >
            Open Map &rarr;
          </a>
        </div>
      </div>
    </>
  );
});
