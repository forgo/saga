// Common API response types

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ApiResponse<T> {
  data?: T;
  error?: ApiError;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  hasMore: boolean;
}

// User types
export interface User {
  id: string;
  email: string;
  name: string;
  displayName?: string;
  avatarUrl?: string;
  role: UserRole;
  status: UserStatus;
  emailVerified: boolean;
  createdAt: string;
  updatedAt: string;
  lastActiveAt?: string;
  location?: Location;
}

export type UserRole = "user" | "moderator" | "admin";
export type UserStatus = "active" | "inactive" | "suspended" | "deleted";

export interface Location {
  latitude: number;
  longitude: number;
  accuracy?: number;
  updatedAt?: string;
}

// Guild types
export interface Guild {
  id: string;
  name: string;
  description?: string;
  iconUrl?: string;
  memberCount: number;
  createdAt: string;
  createdBy: string;
}

// Event types
export interface Event {
  id: string;
  title: string;
  description?: string;
  location?: Location;
  address?: string;
  startTime: string;
  endTime?: string;
  hostId: string;
  guildId?: string;
  attendeeCount: number;
  maxAttendees?: number;
  status: EventStatus;
  createdAt: string;
}

export type EventStatus = "draft" | "published" | "cancelled" | "completed";

// Seeding types
export interface SeedConfig {
  count: number;
  region?: {
    minLat: number;
    maxLat: number;
    minLng: number;
    maxLng: number;
  };
  activityDistribution?: {
    active: number;
    inactive: number;
  };
}

export interface SeedResult {
  created: number;
  ids: string[];
  duration: number;
}

// Action trigger types
export interface ActionLog {
  id: string;
  type: ActionType;
  actingUserId: string;
  targetUserId?: string;
  targetEventId?: string;
  targetGuildId?: string;
  payload?: Record<string, unknown>;
  createdAt: string;
  success: boolean;
  error?: string;
}

export type ActionType =
  | "message"
  | "rsvp"
  | "location_update"
  | "trust_rating"
  | "guild_join"
  | "event_create";

// Health/metrics types
export interface SystemHealth {
  status: "healthy" | "degraded" | "unhealthy";
  uptime: number;
  version: string;
  checks: HealthCheck[];
}

export interface HealthCheck {
  name: string;
  status: "pass" | "warn" | "fail";
  message?: string;
  duration?: number;
}

export interface SystemMetrics {
  memory: {
    allocated: number;
    total: number;
    used: number;
  };
  goroutines: number;
  connections: {
    active: number;
    idle: number;
    total: number;
  };
  requests: {
    total: number;
    rate: number;
    errorRate: number;
    p50Latency: number;
    p95Latency: number;
    p99Latency: number;
  };
}

// Admin User Management types
export interface AdminUserItem {
  id: string;
  email: string;
  username?: string;
  firstname?: string;
  lastname?: string;
  role: UserRole;
  email_verified: boolean;
  created_on: string;
  updated_on: string;
  login_on?: string;
  status: "active" | "suspended" | "banned";
}

export interface AdminUserProfile {
  bio?: string;
  tagline?: string;
  city?: string;
  country?: string;
  visibility: string;
  last_active?: string;
}

export interface AdminUserStats {
  guild_count: number;
  event_count: number;
}

export interface ModerationAction {
  id: string;
  user_id: string;
  level: "nudge" | "warning" | "suspension" | "ban";
  reason: string;
  report_id?: string;
  admin_user_id?: string;
  duration_days?: number;
  expires_on?: string;
  is_active: boolean;
  restrictions?: string[];
  created_on: string;
  lifted_on?: string;
  lifted_by_id?: string;
  lift_reason?: string;
}

export interface UserModerationStatus {
  user_id: string;
  is_banned: boolean;
  is_suspended: boolean;
  suspension_ends_on?: string;
  has_warning: boolean;
  warning_expires_on?: string;
  restrictions?: string[];
  active_actions?: ModerationAction[];
  report_count: number;
  recent_report_count: number;
}

export interface AdminUserDetail {
  id: string;
  email: string;
  username?: string;
  firstname?: string;
  lastname?: string;
  role: UserRole;
  email_verified: boolean;
  created_on: string;
  updated_on: string;
  login_on?: string;
  profile?: AdminUserProfile;
  moderation?: UserModerationStatus;
  stats?: AdminUserStats;
}

export interface ListUsersResponse {
  users: AdminUserItem[];
  total: number;
  page: number;
  page_size: number;
}

export interface ListUsersParams {
  page?: number;
  page_size?: number;
  search?: string;
  role?: string;
  sort_by?: string;
  sort_dir?: string;
}

// Discovery Lab types
export interface AdminMapUser {
  id: string;
  email: string;
  username?: string;
  firstname?: string;
  lat: number;
  lng: number;
  city?: string;
  has_location: boolean;
}

export interface DiscoverySimulationRequest {
  viewer_id: string;
  radius_km?: number;
  min_compatibility?: number;
  require_shared_answer?: boolean;
  limit?: number;
}

export interface AdminDiscoveryResultItem {
  user_id: string;
  email?: string;
  username?: string;
  firstname?: string;
  lat: number;
  lng: number;
  city?: string;
  compatibility_score: number;
  match_score: number;
  distance_km: number;
  shared_interests?: SharedInterestBrief[];
}

export interface SharedInterestBrief {
  interest_id: string;
  interest_name: string;
  category: string;
  teach_learn_match?: boolean;
}

export interface AdminDiscoveryResponse {
  results: AdminDiscoveryResultItem[];
  total_count: number;
  viewer_lat: number;
  viewer_lng: number;
  radius_km: number;
}

export interface CompatibilityScore {
  user_a_id: string;
  user_b_id: string;
  score: number;
  a_to_b_score: number;
  b_to_a_score: number;
  shared_count: number;
  deal_breaker: boolean;
}

export interface DealBreakerViolation {
  question_id: string;
  question_text: string;
  user_answer: string;
  partner_answer: string;
}

export interface CompatibilityBreakdown extends CompatibilityScore {
  category_scores?: Record<string, number>;
  deal_breakers?: DealBreakerViolation[];
}

export interface YikesSummary {
  has_yikes: boolean;
  yikes_count: number;
  categories?: string[];
  severity?: string;
}

export interface AdminCompatibilityResponse {
  breakdown: CompatibilityBreakdown;
  yikes: YikesSummary;
}
