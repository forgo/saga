-- Migration 005: Performance Indexes
-- Version: 2.0.2
-- Date: 2026-01-07
--
-- This migration adds:
-- 1. Missing indexes for common query patterns
-- 2. Removes redundant indexes
-- 3. Adds denormalized count fields with triggers
-- 4. Cursor-based pagination functions

-- =============================================================================
-- 1. MISSING CRITICAL INDEXES
-- =============================================================================

-- Forum thread queries (currently full scan for threaded discussions)
DEFINE INDEX idx_forum_post_thread ON forum_post FIELDS reply_to_id, created_on;
DEFINE INDEX idx_forum_post_visible ON forum_post FIELDS forum_id, deleted_on;

-- Rideshare availability queries
DEFINE INDEX idx_rideshare_available ON rideshare FIELDS status, departure_time;
DEFINE INDEX idx_seat_rideshare_status ON rideshare_seat FIELDS rideshare_id, status;

-- Adventure destination voting
DEFINE INDEX idx_destination_vote_dest ON destination_vote FIELDS destination_id;

-- RSVP target status (for "all RSVPs for this event" queries)
DEFINE INDEX idx_unified_rsvp_target_status ON unified_rsvp FIELDS target_type, target_id, status;

-- Event discovery composite index
DEFINE INDEX OVERWRITE idx_event_discover ON event FIELDS visibility, status, starts_at;

-- Adventure discovery composite index
DEFINE INDEX OVERWRITE idx_adventure_discover ON adventure FIELDS visibility, status, start_date;

-- Profile discovery indexes
DEFINE INDEX OVERWRITE idx_profile_eligible ON user_profile FIELDS visibility, last_active;

-- =============================================================================
-- 2. REMOVE REDUNDANT INDEXES
-- =============================================================================

-- These are redundant (covered by composite indexes):
-- Note: SurrealDB may error if these don't exist, which is fine
-- REMOVE INDEX idx_unified_rsvp_user ON unified_rsvp;  -- Covered by user_type_status
-- REMOVE INDEX idx_unified_rsvp_status ON unified_rsvp;  -- Always filtered with user/target

-- =============================================================================
-- 3. DENORMALIZED COUNT FIELDS
-- =============================================================================

-- Event approval breakdown counts
DEFINE FIELD approved_count ON event TYPE int DEFAULT 0;
DEFINE FIELD pending_count ON event TYPE int DEFAULT 0;
DEFINE FIELD waitlist_count ON event TYPE int DEFAULT 0;
DEFINE FIELD OVERWRITE attendee_count ON event TYPE int DEFAULT 0;

-- Trigger to update event counts when RSVP changes
DEFINE EVENT update_event_counts ON TABLE unified_rsvp
    WHEN ($event = "CREATE" OR $event = "UPDATE" OR $event = "DELETE")
    AND ($before.target_type = "event" OR $after.target_type = "event")
    THEN {
    LET $event_id = IF $event = "DELETE" THEN $before.target_id ELSE $after.target_id END;
    LET $type = IF $event = "DELETE" THEN $before.target_type ELSE $after.target_type END;

    IF $type = "event" {
        -- Recalculate all counts
        LET $approved = SELECT count() FROM unified_rsvp
            WHERE target_type = "event" AND target_id = $event_id AND status = "approved" GROUP ALL;
        LET $pending = SELECT count() FROM unified_rsvp
            WHERE target_type = "event" AND target_id = $event_id AND status = "pending" GROUP ALL;
        LET $waitlist = SELECT count() FROM unified_rsvp
            WHERE target_type = "event" AND target_id = $event_id AND status = "waitlist" GROUP ALL;

        UPDATE type::record("event", $event_id) SET
            approved_count = $approved[0].count ?? 0,
            pending_count = $pending[0].count ?? 0,
            waitlist_count = $waitlist[0].count ?? 0,
            attendee_count = ($approved[0].count ?? 0) + ($pending[0].count ?? 0);
    };
};

-- Forum reply stats
DEFINE FIELD reply_count ON forum_post TYPE int DEFAULT 0;
DEFINE FIELD last_reply_on ON forum_post TYPE option<datetime>;

-- Trigger to update reply counts
DEFINE EVENT update_forum_reply_count ON TABLE forum_post
    WHEN $event = "CREATE" AND $after.reply_to_id IS NOT NONE
    THEN {
    UPDATE $after.reply_to_id SET
        reply_count = reply_count + 1,
        last_reply_on = time::now();
};

DEFINE EVENT decrement_forum_reply_count ON TABLE forum_post
    WHEN $event = "DELETE" AND $before.reply_to_id IS NOT NONE
    THEN {
    UPDATE $before.reply_to_id SET
        reply_count = math::max(0, reply_count - 1);
};

-- Destination vote counts
DEFINE FIELD vote_count ON destination TYPE int DEFAULT 0;
DEFINE FIELD total_rank_score ON destination TYPE int DEFAULT 0;

-- Trigger to update destination vote stats
DEFINE EVENT update_destination_votes ON TABLE destination_vote
    WHEN $event = "CREATE" OR $event = "DELETE"
    THEN {
    LET $dest_id = IF $event = "DELETE" THEN $before.destination_id ELSE $after.destination_id END;

    LET $votes = SELECT count(), math::sum(rank) as total_rank FROM destination_vote
        WHERE destination_id = $dest_id GROUP ALL;

    UPDATE $dest_id SET
        vote_count = $votes[0].count ?? 0,
        total_rank_score = $votes[0].total_rank ?? 0;
};

-- Adventure participant count
DEFINE FIELD OVERWRITE participant_count ON adventure TYPE int DEFAULT 0;

-- Trigger to update adventure participant count
DEFINE EVENT OVERWRITE update_adventure_participant_count ON TABLE adventure_participant
    WHEN $event = "CREATE" OR $event = "DELETE"
    THEN {
    LET $adv_id = IF $event = "DELETE" THEN $before.adventure_id ELSE $after.adventure_id END;

    LET $count = SELECT count() FROM adventure_participant
        WHERE adventure_id = $adv_id AND status IN ["interested", "committed"]
        GROUP ALL;

    UPDATE $adv_id SET participant_count = $count[0].count ?? 0;
};

-- Guild member count
DEFINE FIELD OVERWRITE member_count ON guild TYPE int DEFAULT 0;

-- Trigger to update guild member count
DEFINE EVENT OVERWRITE update_guild_member_count ON TABLE responsible_for
    WHEN $event = "CREATE" OR $event = "DELETE"
    THEN {
    LET $guild_id = IF $event = "DELETE" THEN $before.out ELSE $after.out END;

    LET $count = SELECT count() FROM responsible_for
        WHERE out = $guild_id GROUP ALL;

    UPDATE $guild_id SET member_count = $count[0].count ?? 0;
};

-- Interest user count
DEFINE FIELD OVERWRITE user_count ON interest TYPE int DEFAULT 0;

-- Trigger to update interest user count
DEFINE EVENT OVERWRITE update_interest_user_count ON TABLE has_interest
    WHEN $event = "CREATE" OR $event = "DELETE"
    THEN {
    LET $int_id = IF $event = "DELETE" THEN $before.out ELSE $after.out END;

    LET $count = SELECT count() FROM has_interest
        WHERE out = $int_id GROUP ALL;

    UPDATE $int_id SET user_count = $count[0].count ?? 0;
};

-- =============================================================================
-- 4. CURSOR-BASED PAGINATION FUNCTIONS
-- =============================================================================

-- Paginate events for a guild (cursor-based for efficiency)
DEFINE FUNCTION fn::paginate_guild_events($guild_id: string, $cursor: option<datetime>, $limit: int) {
    IF $cursor IS NONE {
        RETURN SELECT * FROM event
            WHERE guild_id = type::record("guild", $guild_id)
            ORDER BY created_on DESC
            LIMIT $limit;
    };
    RETURN SELECT * FROM event
        WHERE guild_id = type::record("guild", $guild_id)
        AND created_on < $cursor
        ORDER BY created_on DESC
        LIMIT $limit;
};

-- Paginate upcoming events (visibility-aware)
DEFINE FUNCTION fn::paginate_upcoming_events($viewer_id: string, $cursor: option<datetime>, $limit: int) {
    IF $cursor IS NONE {
        RETURN SELECT * FROM event
            WHERE visibility IN ["guilds", "public"]
            AND status = "published"
            AND starts_at > time::now()
            ORDER BY starts_at ASC
            LIMIT $limit;
    };
    RETURN SELECT * FROM event
        WHERE visibility IN ["guilds", "public"]
        AND status = "published"
        AND starts_at > $cursor
        ORDER BY starts_at ASC
        LIMIT $limit;
};

-- Paginate forum posts (threaded view)
DEFINE FUNCTION fn::paginate_forum_posts($forum_id: string, $cursor: option<datetime>, $limit: int) {
    IF $cursor IS NONE {
        -- Get top-level posts first
        RETURN SELECT *, (SELECT * FROM forum_post WHERE reply_to_id = $parent.id ORDER BY created_on ASC LIMIT 5) as replies
            FROM forum_post
            WHERE forum_id = type::record("forum", $forum_id)
            AND reply_to_id IS NONE
            AND deleted_on IS NONE
            ORDER BY is_pinned DESC, created_on DESC
            LIMIT $limit;
    };
    RETURN SELECT *, (SELECT * FROM forum_post WHERE reply_to_id = $parent.id ORDER BY created_on ASC LIMIT 5) as replies
        FROM forum_post
        WHERE forum_id = type::record("forum", $forum_id)
        AND reply_to_id IS NONE
        AND deleted_on IS NONE
        AND created_on < $cursor
        ORDER BY is_pinned DESC, created_on DESC
        LIMIT $limit;
};

-- Paginate discovery results
DEFINE FUNCTION fn::paginate_availability($viewer_lat: float, $viewer_lng: float, $radius_km: float, $cursor: option<datetime>, $limit: int) {
    LET $lat_range = $radius_km / 111.0;
    -- Convert degrees to radians: radians = degrees * (PI / 180)
    LET $lat_radians = $viewer_lat * 0.01745329251994;
    LET $lng_range = $radius_km / (111.0 * math::cos($lat_radians));

    IF $cursor IS NONE {
        RETURN SELECT * FROM availability
            WHERE status = "available"
            AND visibility IN ["guilds", "public"]
            AND expires_at > time::now()
            AND location.lat > ($viewer_lat - $lat_range)
            AND location.lat < ($viewer_lat + $lat_range)
            AND location.lng > ($viewer_lng - $lng_range)
            AND location.lng < ($viewer_lng + $lng_range)
            ORDER BY start_time ASC
            LIMIT $limit;
    };
    RETURN SELECT * FROM availability
        WHERE status = "available"
        AND visibility IN ["guilds", "public"]
        AND expires_at > time::now()
        AND start_time > $cursor
        AND location.lat > ($viewer_lat - $lat_range)
        AND location.lat < ($viewer_lat + $lat_range)
        AND location.lng > ($viewer_lng - $lng_range)
        AND location.lng < ($viewer_lng + $lng_range)
        ORDER BY start_time ASC
        LIMIT $limit;
};

-- =============================================================================
-- 5. BACKFILL EXISTING COUNTS
-- =============================================================================

-- Backfill guild member counts
UPDATE guild SET member_count = (
    SELECT count() FROM responsible_for WHERE out = $parent.id GROUP ALL
)[0].count ?? 0;

-- Backfill interest user counts
UPDATE interest SET user_count = (
    SELECT count() FROM has_interest WHERE out = $parent.id GROUP ALL
)[0].count ?? 0;

-- Backfill destination vote counts
UPDATE destination SET
    vote_count = (SELECT count() FROM destination_vote WHERE destination_id = $parent.id GROUP ALL)[0].count ?? 0,
    total_rank_score = (SELECT math::sum(rank) FROM destination_vote WHERE destination_id = $parent.id GROUP ALL)[0] ?? 0;

-- Backfill adventure participant counts
UPDATE adventure SET participant_count = (
    SELECT count() FROM adventure_participant
    WHERE adventure_id = $parent.id AND status IN ["interested", "committed"]
    GROUP ALL
)[0].count ?? 0;

-- Backfill forum reply counts
UPDATE forum_post SET
    reply_count = (SELECT count() FROM forum_post WHERE reply_to_id = $parent.id GROUP ALL)[0].count ?? 0,
    last_reply_on = (SELECT created_on FROM forum_post WHERE reply_to_id = $parent.id ORDER BY created_on DESC LIMIT 1)[0].created_on;
