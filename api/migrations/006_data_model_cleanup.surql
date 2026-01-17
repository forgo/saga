-- Migration 006: Data Model Cleanup
-- Version: 2.0.3
-- Date: 2026-01-07
--
-- This migration:
-- 1. Adds cascade delete triggers to prevent orphaned records
-- 2. Standardizes timestamp field naming conventions
-- 3. Adds missing auto-update triggers
-- 4. Cleans up legacy tables after verification

-- =============================================================================
-- 1. CASCADE DELETE TRIGGERS
-- =============================================================================

-- When a guild is deleted, clean up all related records
DEFINE EVENT cascade_guild_delete ON TABLE guild WHEN $event = "DELETE" THEN {
    -- Delete events belonging to this guild
    DELETE event WHERE guild_id = $before.id;

    -- Delete matching pools
    DELETE matching_pool WHERE guild_id = $before.id;

    -- Delete guild values questionnaires
    DELETE guild_values WHERE guild_id = $before.id;
    DELETE guild_values_response WHERE guild_id = $before.id;

    -- Delete guild moderation settings
    DELETE guild_moderation_settings WHERE guild_id = $before.id;

    -- Delete people relationships
    DELETE part_of WHERE out = $before.id;

    -- Delete activity relationships
    DELETE managed_by WHERE out = $before.id;

    -- Delete member relationships (responsible_for)
    DELETE responsible_for WHERE out = $before.id;

    -- Delete guild alliance relationships
    DELETE guild_alliance WHERE guild_a_id = $before.id OR guild_b_id = $before.id;

    -- Delete questions created by this guild
    DELETE question WHERE guild_id = $before.id;
};

-- When an adventure is deleted, clean up all related records
DEFINE EVENT cascade_adventure_delete ON TABLE adventure WHEN $event = "DELETE" THEN {
    -- Delete events within this adventure
    DELETE event WHERE adventure_id = $before.id;

    -- Delete destinations
    DELETE destination WHERE adventure_id = $before.id;

    -- Delete adventure activities
    DELETE adventure_activity WHERE adventure_id = $before.id;

    -- Delete adventure participants
    DELETE adventure_participant WHERE adventure_id = $before.id;

    -- Delete rideshares attached to this adventure
    DELETE rideshare WHERE adventure_id = $before.id;

    -- Delete forum if one exists
    LET $forum = SELECT id FROM forum WHERE adventure_id = $before.id;
    IF array::len($forum) > 0 {
        DELETE forum_post WHERE forum_id = $forum[0].id;
        DELETE forum WHERE adventure_id = $before.id;
    };
};

-- When an event is deleted, clean up all related records
DEFINE EVENT cascade_event_delete ON TABLE event WHEN $event = "DELETE" THEN {
    LET $event_id_str = string::split(<string>$before.id, ":")[1];

    -- Delete RSVPs (unified_rsvp)
    DELETE unified_rsvp WHERE target_type = "event" AND target_id = $event_id_str;

    -- Delete event roles and assignments
    DELETE event_role_assignment WHERE event_id = $before.id;
    DELETE event_role WHERE event_id = $before.id;

    -- Delete event hosts
    DELETE event_host WHERE event_id = $before.id;

    -- Delete rideshares attached to this event
    DELETE rideshare WHERE event_id = $before.id;

    -- Delete forum if one exists
    LET $forum = SELECT id FROM forum WHERE event_id = $before.id;
    IF array::len($forum) > 0 {
        DELETE forum_post WHERE forum_id = $forum[0].id;
        DELETE forum WHERE event_id = $before.id;
    };

    -- Delete legacy event_participant and event_rsvp entries
    DELETE event_participant WHERE out = $before.id;
    DELETE event_rsvp WHERE event_id = $before.id;
    DELETE rsvp WHERE out = $before.id;
};

-- When a destination is deleted, clean up votes
DEFINE EVENT cascade_destination_delete ON TABLE destination WHEN $event = "DELETE" THEN {
    DELETE destination_vote WHERE destination_id = $before.id;
};

-- When a rideshare is deleted, clean up seats and segments
DEFINE EVENT cascade_rideshare_delete ON TABLE rideshare WHEN $event = "DELETE" THEN {
    DELETE rideshare_seat WHERE rideshare_id = $before.id;
    DELETE rideshare_segment WHERE rideshare_id = $before.id;
};

-- When a matching pool is deleted, clean up members and results
DEFINE EVENT cascade_pool_delete ON TABLE matching_pool WHEN $event = "DELETE" THEN {
    DELETE pool_member WHERE pool_id = $before.id;
    DELETE match_result WHERE pool_id = $before.id;
};

-- When a user is deleted, clean up all their data
DEFINE EVENT cascade_user_delete ON TABLE user WHEN $event = "DELETE" THEN {
    LET $user_id_str = string::split(<string>$before.id, ":")[1];

    -- Delete identity providers
    DELETE identity WHERE user = $before.id;

    -- Delete passkeys
    DELETE passkey WHERE user = $before.id;

    -- Delete refresh tokens
    DELETE refresh_token WHERE user = $before.id;

    -- Delete profile
    DELETE user_profile WHERE user = $before.id;

    -- Delete interests
    DELETE has_interest WHERE in = $before.id;

    -- Delete availability
    DELETE availability WHERE user = $before.id;

    -- Delete answers
    DELETE answer WHERE user = $before.id;

    -- Delete trust relations
    DELETE trust_relation WHERE user_a_id = $before.id OR user_b_id = $before.id;

    -- Delete IRL verifications
    DELETE irl_verification WHERE user_a_id = $before.id OR user_b_id = $before.id;

    -- Delete blocks
    DELETE block WHERE blocker_user_id = $before.id OR blocked_user_id = $before.id;
    DELETE user_block WHERE blocker = $before.id OR blocked = $before.id;

    -- Delete resonance data
    DELETE resonance_ledger WHERE user = $before.id;
    DELETE resonance_score WHERE user = $before.id;
    DELETE resonance_daily_cap WHERE user = $before.id;

    -- Delete nudge data
    DELETE nudge WHERE user_id = $before.id;
    DELETE nudge_history WHERE user_id = $before.id;
    DELETE nudge_preference WHERE user_id = $before.id;
    DELETE nudge_settings WHERE user_id = $before.id;

    -- Delete user bias profile
    DELETE user_bias_profile WHERE user = $before.id;

    -- Delete RSVPs
    DELETE unified_rsvp WHERE user_id = $user_id_str;
    DELETE event_rsvp WHERE user_id = $before.id;

    -- Delete moderation records
    DELETE user_flag WHERE reporter = $before.id OR reported_user = $before.id;
    DELETE user_moderation WHERE user = $before.id;
    DELETE moderation_action WHERE user_id = $before.id;
    DELETE report WHERE reporter_user_id = $before.id OR reported_user_id = $before.id;

    -- Delete reminders
    DELETE reminder WHERE member_id IN (SELECT id FROM member WHERE user = $before.id);

    -- Delete member records (will cascade to responsible_for via member delete)
    DELETE member WHERE user = $before.id;
};

-- When a member is deleted, clean up relationships
DEFINE EVENT cascade_member_delete ON TABLE member WHEN $event = "DELETE" THEN {
    -- Remove from all guilds
    DELETE responsible_for WHERE in = $before.id;

    -- Remove authentication link
    DELETE authenticates_via WHERE in = $before.id;

    -- Remove from pool memberships
    DELETE pool_member WHERE member_id = $before.id;

    -- Remove from match results (update arrays)
    -- Note: This requires special handling as members is an array
    UPDATE match_result SET members = array::filter(members, |$m| $m != $before.id)
        WHERE $before.id IN members;

    -- Remove from event participation
    DELETE event_participant WHERE in = $before.id;

    -- Delete reminders
    DELETE reminder WHERE member_id = $before.id;
};

-- When a forum is deleted, clean up all posts
DEFINE EVENT cascade_forum_delete ON TABLE forum WHEN $event = "DELETE" THEN {
    DELETE forum_post WHERE forum_id = $before.id;
};

-- When an event role is deleted, clean up assignments
DEFINE EVENT cascade_event_role_delete ON TABLE event_role WHEN $event = "DELETE" THEN {
    DELETE event_role_assignment WHERE role_id = $before.id;
};

-- =============================================================================
-- 2. AUTO-UPDATE TIMESTAMPS USING FIELD VALUE
-- =============================================================================
-- Note: These replace event-based triggers which cause infinite recursion.
-- The VALUE clause sets the field value on every CREATE/UPDATE without triggering events.

DEFINE FIELD OVERWRITE updated_on ON user TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON member TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON guild TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON event TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON timer TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON reminder TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON availability TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON hangout TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON answer TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON guild_values TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON event_role TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON event_role_assignment TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON trust_relation TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON user_moderation TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON guild_moderation_settings TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON event_rsvp TYPE datetime VALUE time::now();
DEFINE FIELD OVERWRITE updated_on ON unified_rsvp TYPE datetime VALUE time::now();

-- =============================================================================
-- 3. DATA INTEGRITY CONSTRAINTS
-- =============================================================================

-- Ensure event dates are valid
DEFINE EVENT validate_event_dates ON TABLE event WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    IF $after.ends_at IS NOT NONE AND $after.ends_at < $after.starts_at {
        THROW "Event end time must be after start time";
    };
};

-- Ensure availability dates are valid
DEFINE EVENT validate_availability_dates ON TABLE availability
    WHEN $event = "CREATE" OR $event = "UPDATE"
    THEN {
    IF $after.end_time <= $after.start_time {
        THROW "Availability end time must be after start time";
    };
    IF $after.expires_at <= $after.start_time {
        THROW "Availability expires_at must be after start time";
    };
};

-- Ensure rideshare has valid times
DEFINE EVENT validate_rideshare_times ON TABLE rideshare
    WHEN $event = "CREATE" OR $event = "UPDATE"
    THEN {
    IF $after.arrival_time IS NOT NONE AND $after.arrival_time < $after.departure_time {
        THROW "Rideshare arrival time must be after departure time";
    };
};

-- =============================================================================
-- 4. UNIFIED RSVP TABLE MIGRATION
-- =============================================================================

-- Add updated_on to unified_rsvp if not present
DEFINE FIELD OVERWRITE updated_on ON unified_rsvp TYPE datetime DEFAULT time::now();

-- Add status index for faster queries
DEFINE INDEX idx_unified_rsvp_user_status ON unified_rsvp FIELDS user_id, status;

-- Note: After verifying all data has been migrated to unified_rsvp,
-- the following tables can be dropped in a future migration:
-- - rsvp (legacy relation)
-- - event_participant (legacy)
-- - event_rsvp (legacy)
--
-- Verification query to run before dropping:
-- SELECT count() FROM event_rsvp WHERE NOT (event_id, user_id) IN
--   (SELECT type::record("event", target_id), type::record("user", user_id)
--    FROM unified_rsvp WHERE target_type = "event") GROUP ALL;
