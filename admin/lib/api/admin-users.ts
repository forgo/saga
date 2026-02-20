import { api, isError } from "./client.ts";
import type {
  AdminUserDetail,
  ListUsersParams,
  ListUsersResponse,
} from "./types.ts";

// ListUsers fetches a paginated list of users
export async function listUsers(
  params: ListUsersParams = {},
): Promise<{ data: ListUsersResponse | null; error: string | null }> {
  const response = await api.get<{ data: ListUsersResponse }>(
    "/v1/admin/users",
    {
      params: {
        page: params.page,
        page_size: params.page_size,
        search: params.search,
        role: params.role,
        sort_by: params.sort_by,
        sort_dir: params.sort_dir,
      },
    },
  );

  if (isError(response)) {
    return { data: null, error: response.error.message };
  }

  return { data: response.data?.data ?? null, error: null };
}

// GetUser fetches detailed information about a single user
export async function getUser(
  userId: string,
): Promise<{ data: AdminUserDetail | null; error: string | null }> {
  const response = await api.get<{ data: AdminUserDetail }>(
    `/v1/admin/users/${userId}`,
  );

  if (isError(response)) {
    return { data: null, error: response.error.message };
  }

  return { data: response.data?.data ?? null, error: null };
}

// UpdateUserRole changes a user's role
export async function updateUserRole(
  userId: string,
  role: string,
): Promise<{ error: string | null }> {
  const response = await api.patch(`/v1/admin/users/${userId}/role`, { role });

  if (isError(response)) {
    return { error: response.error.message };
  }

  return { error: null };
}

// DeleteUser soft-deletes (bans) or hard-deletes a user
export async function deleteUser(
  userId: string,
  hard = false,
): Promise<{ error: string | null }> {
  const response = await api.delete(`/v1/admin/users/${userId}`, {
    params: hard ? { hard: "true" } : undefined,
  });

  if (isError(response)) {
    return { error: response.error.message };
  }

  return { error: null };
}

// TakeModAction creates a moderation action against a user
export async function takeModAction(
  userId: string,
  level: string,
  reason: string,
  durationDays?: number,
): Promise<{ error: string | null }> {
  const body: Record<string, unknown> = {
    user_id: userId,
    level,
    reason,
  };
  if (durationDays !== undefined) {
    body.duration_days = durationDays;
  }

  const response = await api.post("/v1/moderation/actions", body);

  if (isError(response)) {
    return { error: response.error.message };
  }

  return { error: null };
}

// LiftModAction lifts a moderation action
export async function liftModAction(
  actionId: string,
  reason: string,
): Promise<{ error: string | null }> {
  const response = await api.post(
    `/v1/moderation/actions/${actionId}/lift`,
    { reason },
  );

  if (isError(response)) {
    return { error: response.error.message };
  }

  return { error: null };
}
