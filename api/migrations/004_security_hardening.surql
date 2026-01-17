-- Migration 004: Security Hardening
-- Version: 2.0.1
-- Date: 2026-01-07
--
-- This migration adds security functions and indexes to enforce:
-- 1. Block filtering in discovery
-- 2. Guild visibility access control
-- 3. Shared guild membership checks
-- 4. Alliance membership checks

-- =============================================================================
-- 1. BLOCK CHECKING FUNCTION
-- =============================================================================

-- Check if either user has blocked the other (bidirectional)
-- Used by discovery and profile queries to filter out blocked users
DEFINE FUNCTION fn::is_blocked($user_a: string, $user_b: string) {
    LET $count = SELECT count() FROM block
        WHERE (blocker_user_id = type::record("user", $user_a) AND blocked_user_id = type::record("user", $user_b))
           OR (blocker_user_id = type::record("user", $user_b) AND blocked_user_id = type::record("user", $user_a))
        GROUP ALL;
    RETURN ($count[0].count ?? 0) > 0;
};

-- =============================================================================
-- 2. GUILD MEMBERSHIP FUNCTIONS
-- =============================================================================

-- Check if a user is a member of a specific guild
DEFINE FUNCTION fn::is_guild_member($user_id: string, $guild_id: string) {
    LET $count = SELECT count() FROM member
        WHERE user = type::record("user", $user_id)
        AND id IN (SELECT in FROM responsible_for WHERE out = type::record("guild", $guild_id))
        GROUP ALL;
    RETURN ($count[0].count ?? 0) > 0;
};

-- Check if two users share at least one guild
DEFINE FUNCTION fn::shares_guild($user_a: string, $user_b: string) {
    -- Get guilds for user A
    LET $guilds_a = SELECT out FROM responsible_for
        WHERE in IN (SELECT id FROM member WHERE user = type::record("user", $user_a));

    -- Check if user B is in any of those guilds
    LET $shared = SELECT count() FROM responsible_for
        WHERE in IN (SELECT id FROM member WHERE user = type::record("user", $user_b))
        AND out IN $guilds_a.out
        GROUP ALL;

    RETURN ($shared[0].count ?? 0) > 0;
};

-- =============================================================================
-- 3. GUILD VISIBILITY ACCESS CONTROL
-- =============================================================================

-- Check if a viewer can see a guild based on visibility
DEFINE FUNCTION fn::can_view_guild($viewer_id: string, $guild_id: string) {
    LET $guild = SELECT visibility FROM guild WHERE id = type::record("guild", $guild_id);
    LET $visibility = $guild[0].visibility ?? "private";

    -- Public and guilds visibility are accessible to all authenticated users
    IF $visibility IN ["public", "guilds"] {
        RETURN true;
    };

    -- Private and invite_only require membership
    RETURN fn::is_guild_member($viewer_id, $guild_id);
};

-- =============================================================================
-- 4. ALLIANCE MEMBERSHIP FUNCTIONS
-- =============================================================================

-- Check if two users share an alliance (their guilds are allied)
DEFINE FUNCTION fn::shares_alliance($user_a: string, $user_b: string) {
    -- Get guilds for user A
    LET $guilds_a = SELECT out FROM responsible_for
        WHERE in IN (SELECT id FROM member WHERE user = type::record("user", $user_a));

    -- Get guilds for user B
    LET $guilds_b = SELECT out FROM responsible_for
        WHERE in IN (SELECT id FROM member WHERE user = type::record("user", $user_b));

    -- Check if any guild from A is allied with any guild from B
    LET $alliance_count = SELECT count() FROM guild_alliance
        WHERE status = "active"
        AND (
            (guild_a_id IN $guilds_a.out AND guild_b_id IN $guilds_b.out)
            OR (guild_b_id IN $guilds_a.out AND guild_a_id IN $guilds_b.out)
        )
        GROUP ALL;

    RETURN ($alliance_count[0].count ?? 0) > 0;
};

-- =============================================================================
-- 5. UNIFIED ACCESS CONTROL FUNCTION
-- =============================================================================

-- Master access control function for visibility-based content
-- Returns true if viewer can access content with given visibility from owner
DEFINE FUNCTION fn::can_access($viewer_id: string, $owner_id: string, $visibility: string) {
    -- Step 1: Always check blocks first
    IF fn::is_blocked($viewer_id, $owner_id) {
        RETURN false;
    };

    -- Step 2: Check visibility level
    IF $visibility = "public" {
        RETURN true;
    };

    IF $visibility = "private" {
        RETURN $viewer_id = $owner_id;
    };

    IF $visibility = "guilds" {
        RETURN fn::shares_guild($viewer_id, $owner_id);
    };

    IF $visibility = "alliance" {
        RETURN fn::shares_guild($viewer_id, $owner_id) OR fn::shares_alliance($viewer_id, $owner_id);
    };

    IF $visibility = "invite_only" {
        -- Invite-only requires explicit invitation (handled at application layer)
        RETURN false;
    };

    RETURN false;
};

-- =============================================================================
-- 6. SAFE PROFILE QUERY FUNCTION
-- =============================================================================

-- Get a safe public profile view with block and visibility checks
-- Returns NONE if viewer cannot see the profile
DEFINE FUNCTION fn::get_safe_profile($viewer_id: string, $target_id: string) {
    -- Check if blocked
    IF fn::is_blocked($viewer_id, $target_id) {
        RETURN NONE;
    };

    -- Get profile
    LET $profile = SELECT * FROM user_profile WHERE user = type::record("user", $target_id);
    IF array::len($profile) = 0 {
        RETURN NONE;
    };

    LET $p = $profile[0];

    -- Check visibility
    IF $p.visibility = "private" {
        RETURN NONE;
    };

    IF $p.visibility = "guilds" AND !fn::shares_guild($viewer_id, $target_id) {
        RETURN NONE;
    };

    -- Return safe profile (without exact location)
    RETURN {
        user_id: $target_id,
        bio: $p.bio,
        tagline: $p.tagline,
        languages: $p.languages,
        city: $p.location.city,
        country: $p.location.country,
        visibility: $p.visibility,
        last_active: $p.last_active
    };
};

-- =============================================================================
-- 7. DISCOVERY HELPER FUNCTIONS
-- =============================================================================

-- Get blocked user IDs for a user (for batch filtering)
DEFINE FUNCTION fn::get_blocked_users($user_id: string) {
    LET $blocked_by_me = SELECT blocked_user_id FROM block
        WHERE blocker_user_id = type::record("user", $user_id);
    LET $blocked_me = SELECT blocker_user_id FROM block
        WHERE blocked_user_id = type::record("user", $user_id);

    RETURN array::union(
        $blocked_by_me.map(|$b| string::split(<string>$b.blocked_user_id, ":")[1]),
        $blocked_me.map(|$b| string::split(<string>$b.blocker_user_id, ":")[1])
    );
};

-- =============================================================================
-- 8. ADDITIONAL INDEXES FOR SECURITY QUERIES
-- =============================================================================

-- Index for faster block lookups (already have individual indexes, add composite)
DEFINE INDEX idx_block_lookup ON block FIELDS blocker_user_id, blocked_user_id;

-- Index for faster member guild lookups
DEFINE INDEX idx_member_user_lookup ON member FIELDS user;

-- =============================================================================
-- 9. SELF-BLOCK PREVENTION FOR user_block TABLE
-- =============================================================================

-- Prevent users from blocking themselves (for user_block table, block table already has this)
DEFINE EVENT OVERWRITE user_block_self_check ON TABLE user_block WHEN $event = "CREATE" THEN {
    IF $after.blocker = $after.blocked {
        THROW "Cannot block yourself";
    };
};

-- =============================================================================
-- 10. TIMESTAMP CONSISTENCY
-- =============================================================================

-- Add missing auto-update triggers for tables that need them
-- Guard condition prevents infinite recursion by only updating when updated_on wasn't already changed
DEFINE EVENT user_profile_updated ON TABLE user_profile WHEN $event = "UPDATE" AND $before.updated_on = $after.updated_on THEN {
    UPDATE user_profile SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT person_updated ON TABLE person WHEN $event = "UPDATE" AND $before.updated_on = $after.updated_on THEN {
    UPDATE person SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT activity_updated ON TABLE activity WHEN $event = "UPDATE" AND $before.updated_on = $after.updated_on THEN {
    UPDATE activity SET updated_on = time::now() WHERE id = $after.id;
};
