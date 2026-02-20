import { api, isError } from "./client.ts";
import type {
  AdminCompatibilityResponse,
  AdminDiscoveryResponse,
  AdminMapUser,
  DiscoverySimulationRequest,
} from "./types.ts";

// GetUsersWithLocations fetches users that have location data for map display
export async function getUsersWithLocations(
  limit = 200,
): Promise<{ data: AdminMapUser[] | null; error: string | null }> {
  const response = await api.get<{ data: AdminMapUser[] }>(
    "/v1/admin/discovery/users",
    { params: { limit } },
  );

  if (isError(response)) {
    return { data: null, error: response.error.message };
  }

  return { data: response.data?.data ?? null, error: null };
}

// SimulateDiscovery runs the discovery algorithm from a viewer's perspective
export async function simulateDiscovery(
  request: DiscoverySimulationRequest,
): Promise<{ data: AdminDiscoveryResponse | null; error: string | null }> {
  const response = await api.post<{ data: AdminDiscoveryResponse }>(
    "/v1/admin/discovery/simulate",
    request,
  );

  if (isError(response)) {
    return { data: null, error: response.error.message };
  }

  return { data: response.data?.data ?? null, error: null };
}

// GetCompatibility fetches detailed compatibility breakdown between two users
export async function getCompatibility(
  userAId: string,
  userBId: string,
): Promise<{ data: AdminCompatibilityResponse | null; error: string | null }> {
  const response = await api.get<{ data: AdminCompatibilityResponse }>(
    `/v1/admin/discovery/compatibility/${userAId}/${userBId}`,
  );

  if (isError(response)) {
    return { data: null, error: response.error.message };
  }

  return { data: response.data?.data ?? null, error: null };
}
