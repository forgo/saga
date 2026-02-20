import type { Signal } from "@preact/signals";
import { Button } from "../Button.tsx";
import { RoleBadge } from "./RoleBadge.tsx";
import { StatusBadge } from "./StatusBadge.tsx";
import type { AdminUserDetail } from "../../lib/api/types.ts";

interface UserDetailPanelProps {
  user: AdminUserDetail | null;
  loading: boolean;
  onClose: () => void;
  onRoleChange: (userId: string, role: string) => void;
  onSuspend: (userId: string) => void;
  onBan: (userId: string) => void;
  onLiftAction: (actionId: string) => void;
  onDelete: (userId: string, hard: boolean) => void;
  roleLoading: Signal<boolean>;
  actionLoading: Signal<boolean>;
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return "—";
  try {
    return new Date(dateStr).toLocaleString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  } catch {
    return dateStr;
  }
}

function displayName(user: AdminUserDetail): string {
  const parts: string[] = [];
  if (user.firstname) parts.push(user.firstname);
  if (user.lastname) parts.push(user.lastname);
  return parts.length > 0 ? parts.join(" ") : "—";
}

function getUserStatus(user: AdminUserDetail): string {
  if (user.moderation?.is_banned) return "banned";
  if (user.moderation?.is_suspended) return "suspended";
  return "active";
}

export function UserDetailPanel({
  user,
  loading,
  onClose,
  onRoleChange,
  onSuspend,
  onBan,
  onLiftAction,
  onDelete,
  roleLoading,
  actionLoading,
}: UserDetailPanelProps) {
  if (!user && !loading) return null;

  const status = user ? getUserStatus(user) : "active";

  return (
    <div class="fixed inset-0 z-40 flex justify-end">
      {/* Backdrop */}
      <div
        class="fixed inset-0 bg-gray-500/50 transition-opacity"
        onClick={onClose}
      />

      {/* Panel */}
      <div class="relative w-full max-w-lg bg-white shadow-xl overflow-y-auto">
        {/* Header */}
        <div class="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between z-10">
          <h2 class="text-lg font-semibold text-gray-900">User Details</h2>
          <button
            type="button"
            onClick={onClose}
            class="text-gray-400 hover:text-gray-600"
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

        {loading
          ? (
            <div class="p-6 flex items-center justify-center">
              <svg
                class="w-6 h-6 animate-spin text-gray-400"
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
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
            </div>
          )
          : user
          ? (
            <div class="p-6 space-y-6">
              {/* Identity */}
              <section>
                <h3 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">
                  Identity
                </h3>
                <div class="space-y-2">
                  <DetailRow label="Email" value={user.email} />
                  <DetailRow label="Username" value={user.username ?? "—"} />
                  <DetailRow label="Name" value={displayName(user)} />
                  <DetailRow label="ID" value={user.id} mono />
                  <div class="flex items-center justify-between py-1">
                    <span class="text-sm text-gray-500">Role</span>
                    <RoleBadge role={user.role} />
                  </div>
                  <div class="flex items-center justify-between py-1">
                    <span class="text-sm text-gray-500">Status</span>
                    <StatusBadge status={status} />
                  </div>
                  <DetailRow
                    label="Email Verified"
                    value={user.email_verified ? "Yes" : "No"}
                  />
                </div>
              </section>

              {/* Dates */}
              <section>
                <h3 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">
                  Timeline
                </h3>
                <div class="space-y-2">
                  <DetailRow
                    label="Created"
                    value={formatDate(user.created_on)}
                  />
                  <DetailRow
                    label="Updated"
                    value={formatDate(user.updated_on)}
                  />
                  <DetailRow
                    label="Last Login"
                    value={formatDate(user.login_on)}
                  />
                  {user.profile?.last_active && (
                    <DetailRow
                      label="Last Active"
                      value={formatDate(user.profile.last_active)}
                    />
                  )}
                </div>
              </section>

              {/* Profile */}
              {user.profile && (
                <section>
                  <h3 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">
                    Profile
                  </h3>
                  <div class="space-y-2">
                    <DetailRow
                      label="Visibility"
                      value={user.profile.visibility}
                    />
                    {user.profile.city && (
                      <DetailRow
                        label="Location"
                        value={[user.profile.city, user.profile.country].filter(
                          Boolean,
                        ).join(", ")}
                      />
                    )}
                    {user.profile.bio && (
                      <div class="py-1">
                        <span class="text-sm text-gray-500 block mb-1">
                          Bio
                        </span>
                        <p class="text-sm text-gray-900">{user.profile.bio}</p>
                      </div>
                    )}
                  </div>
                </section>
              )}

              {/* Stats */}
              {user.stats && (
                <section>
                  <h3 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">
                    Stats
                  </h3>
                  <div class="grid grid-cols-2 gap-3">
                    <StatBox label="Guilds" value={user.stats.guild_count} />
                    <StatBox label="Events" value={user.stats.event_count} />
                  </div>
                </section>
              )}

              {/* Moderation */}
              {user.moderation && (
                <section>
                  <h3 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">
                    Moderation
                  </h3>
                  <div class="space-y-2 mb-3">
                    <DetailRow
                      label="Reports Against"
                      value={String(user.moderation.report_count)}
                    />
                    <DetailRow
                      label="Recent Reports (30d)"
                      value={String(user.moderation.recent_report_count)}
                    />
                  </div>
                  {user.moderation.active_actions &&
                    user.moderation.active_actions.length > 0 && (
                    <div class="space-y-2">
                      <span class="text-sm font-medium text-gray-700">
                        Active Actions
                      </span>
                      {user.moderation.active_actions.map((action) => (
                        <div
                          key={action.id}
                          class="bg-red-50 border border-red-200 rounded-lg p-3"
                        >
                          <div class="flex items-center justify-between mb-1">
                            <span class="text-xs font-medium text-red-800 uppercase">
                              {action.level}
                            </span>
                            {action.expires_on && (
                              <span class="text-xs text-red-600">
                                Expires: {formatDate(action.expires_on)}
                              </span>
                            )}
                          </div>
                          <p class="text-sm text-red-700">{action.reason}</p>
                          <Button
                            variant="outline"
                            size="sm"
                            class="mt-2"
                            onClick={() => onLiftAction(action.id)}
                            loading={actionLoading.value}
                          >
                            Lift Action
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}
                </section>
              )}

              {/* Actions */}
              <section>
                <h3 class="text-sm font-medium text-gray-500 uppercase tracking-wider mb-3">
                  Actions
                </h3>

                {/* Role change */}
                <div class="mb-4">
                  <label class="block text-sm text-gray-700 mb-1">
                    Change Role
                  </label>
                  <div class="flex items-center gap-2">
                    <select
                      value={user.role}
                      onChange={(e) =>
                        onRoleChange(
                          user.id,
                          (e.target as HTMLSelectElement).value,
                        )}
                      disabled={roleLoading.value}
                      class="rounded-md border-gray-300 shadow-sm focus:border-primary-500 focus:ring-primary-500 sm:text-sm px-3 py-2 border flex-1"
                    >
                      <option value="user">User</option>
                      <option value="moderator">Moderator</option>
                      <option value="admin">Admin</option>
                    </select>
                    {roleLoading.value && (
                      <svg
                        class="w-4 h-4 animate-spin text-gray-400"
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
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        />
                      </svg>
                    )}
                  </div>
                </div>

                {/* Moderation actions */}
                <div class="flex flex-wrap gap-2">
                  {status === "active" && (
                    <>
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onSuspend(user.id)}
                        loading={actionLoading.value}
                      >
                        Suspend
                      </Button>
                      <Button
                        variant="danger"
                        size="sm"
                        onClick={() => onBan(user.id)}
                        loading={actionLoading.value}
                      >
                        Ban
                      </Button>
                    </>
                  )}
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => onDelete(user.id, false)}
                    loading={actionLoading.value}
                  >
                    Soft Delete
                  </Button>
                  <Button
                    variant="danger"
                    size="sm"
                    onClick={() => onDelete(user.id, true)}
                    loading={actionLoading.value}
                  >
                    Hard Delete
                  </Button>
                </div>
              </section>
            </div>
          )
          : null}
      </div>
    </div>
  );
}

function DetailRow(
  { label, value, mono }: { label: string; value: string; mono?: boolean },
) {
  return (
    <div class="flex items-center justify-between py-1">
      <span class="text-sm text-gray-500">{label}</span>
      <span class={`text-sm text-gray-900 ${mono ? "font-mono text-xs" : ""}`}>
        {value}
      </span>
    </div>
  );
}

function StatBox({ label, value }: { label: string; value: number }) {
  return (
    <div class="bg-gray-50 rounded-lg p-3 text-center">
      <div class="text-2xl font-semibold text-gray-900">{value}</div>
      <div class="text-xs text-gray-500">{label}</div>
    </div>
  );
}
