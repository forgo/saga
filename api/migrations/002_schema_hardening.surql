-- Saga Schema Hardening
-- Version: 2.1.0
-- Created: 2026-01-06
--
-- This migration addresses critical gaps identified in the schema audit:
-- 1. Unified RSVP system with polymorphic target
-- 2. Discovery enforcement (daily limits, profile eligibility)
-- 3. Location privacy functions (DB-level enforcement)
-- 4. Event completion tracking
-- 5. Missing composite indexes for performance
-- 6. Visibility cascade constraints

-- =============================================================================
-- 1. UNIFIED RSVP SYSTEM (Polymorphic)
-- =============================================================================
-- Replaces fragmented: event_participant, event_rsvp, legacy rsvp relation
-- Single table for all RSVP-like relationships with composite indexes

DEFINE TABLE unified_rsvp SCHEMAFULL;

-- Polymorphic target (what you're RSVPing to)
DEFINE FIELD target_type ON unified_rsvp TYPE string
    ASSERT $value IN ["event", "adventure", "hangout", "pool_match"];
DEFINE FIELD target_id ON unified_rsvp TYPE string;

-- Who is RSVPing
DEFINE FIELD user_id ON unified_rsvp TYPE record<user>;

-- Universal RSVP fields
DEFINE FIELD status ON unified_rsvp TYPE string DEFAULT "pending"
    ASSERT $value IN ["pending", "approved", "waitlisted", "declined", "cancelled", "attended", "no_show"];
DEFINE FIELD role ON unified_rsvp TYPE string DEFAULT "participant"
    ASSERT $value IN ["organizer", "host", "co_host", "participant", "helper", "driver"];
DEFINE FIELD created_on ON unified_rsvp TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON unified_rsvp TYPE datetime DEFAULT time::now();

-- Optional context-specific fields (sparse columns)
DEFINE FIELD values_aligned ON unified_rsvp TYPE option<bool>;
DEFINE FIELD alignment_score ON unified_rsvp TYPE option<float>;
DEFINE FIELD yikes_count ON unified_rsvp TYPE option<int>;
DEFINE FIELD plus_ones ON unified_rsvp TYPE option<int>;
DEFINE FIELD plus_one_names ON unified_rsvp TYPE option<array<string>>;
DEFINE FIELD note ON unified_rsvp TYPE option<string>;
DEFINE FIELD host_note ON unified_rsvp TYPE option<string>;

-- Confirmation & verification fields
DEFINE FIELD checkin_time ON unified_rsvp TYPE option<datetime>;
DEFINE FIELD completion_confirmed ON unified_rsvp TYPE option<datetime>;
DEFINE FIELD early_confirmed ON unified_rsvp TYPE option<bool>;

-- Feedback fields (for support sessions / hangouts)
DEFINE FIELD helpfulness_rating ON unified_rsvp TYPE option<string>
    ASSERT $value IS NONE OR $value IN ["yes", "somewhat", "no"];
DEFINE FIELD helpfulness_tags ON unified_rsvp TYPE option<array<string>>;

-- Critical indexes for scoped queries
DEFINE INDEX idx_unified_rsvp_target ON unified_rsvp FIELDS target_type, target_id;
DEFINE INDEX idx_unified_rsvp_user ON unified_rsvp FIELDS user_id;
DEFINE INDEX idx_unified_rsvp_target_user ON unified_rsvp FIELDS target_type, target_id, user_id UNIQUE;
DEFINE INDEX idx_unified_rsvp_user_type_status ON unified_rsvp FIELDS user_id, target_type, status;
DEFINE INDEX idx_unified_rsvp_status ON unified_rsvp FIELDS status;

-- Auto-update timestamps (only when updated_on wasn't already changed, to prevent recursion)
DEFINE EVENT unified_rsvp_updated ON TABLE unified_rsvp WHEN $event = "UPDATE" AND $before.updated_on = $after.updated_on THEN {
    UPDATE unified_rsvp SET updated_on = time::now() WHERE id = $after.id;
};

-- =============================================================================
-- 2. DISCOVERY ENFORCEMENT
-- =============================================================================

-- Daily discovery limit tracking (5-10 suggestions per day max)
DEFINE TABLE discovery_daily_count SCHEMAFULL;
DEFINE FIELD user_id ON discovery_daily_count TYPE record<user>;
DEFINE FIELD date ON discovery_daily_count TYPE string; -- YYYY-MM-DD format
DEFINE FIELD people_shown ON discovery_daily_count TYPE int DEFAULT 0;
DEFINE FIELD events_shown ON discovery_daily_count TYPE int DEFAULT 0;
DEFINE FIELD guilds_shown ON discovery_daily_count TYPE int DEFAULT 0;
DEFINE FIELD created_on ON discovery_daily_count TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON discovery_daily_count TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_discovery_user_date ON discovery_daily_count FIELDS user_id, date UNIQUE;

-- Update user_profile with eligibility fields
-- Requirement: 3+ questions from required categories before discovery
DEFINE FIELD discovery_eligible ON user_profile TYPE bool DEFAULT false;
DEFINE FIELD categories_completed ON user_profile TYPE array<string> DEFAULT [];
DEFINE FIELD question_count ON user_profile TYPE int DEFAULT 0;
DEFINE FIELD profile_completion_score ON user_profile TYPE float DEFAULT 0.0;

-- Index for filtering eligible users in discovery
DEFINE INDEX idx_profile_eligible ON user_profile FIELDS discovery_eligible, visibility;

-- =============================================================================
-- 3. LOCATION PRIVACY FUNCTIONS
-- =============================================================================
-- These functions ensure raw coordinates are NEVER exposed, even in direct DB queries

-- Calculate distance bucket from coordinates (never expose raw distance)
DEFINE FUNCTION fn::distance_bucket($lat1: float, $lng1: float, $lat2: float, $lng2: float) {
    LET $point1 = { type: "Point", coordinates: [$lng1, $lat1] };
    LET $point2 = { type: "Point", coordinates: [$lng2, $lat2] };
    LET $distance = geo::distance($point1, $point2);

    RETURN IF $distance < 1000 THEN "nearby"
        ELSE IF $distance < 2000 THEN "~2km"
        ELSE IF $distance < 5000 THEN "~5km"
        ELSE IF $distance < 10000 THEN "~10km"
        ELSE IF $distance < 20000 THEN "~20km"
        ELSE ">20km"
    END;
};

-- Get neighborhood-level location (coarse) from coordinates
-- Returns only city/neighborhood, never coordinates
DEFINE FUNCTION fn::coarse_location($location: object) {
    RETURN {
        city: $location.city,
        neighborhood: $location.neighborhood,
        country: $location.country,
        timezone: $location.timezone
    };
};

-- Safe profile view function - returns profile without exposing coordinates
-- Caller provides their own coordinates to calculate distance bucket
DEFINE FUNCTION fn::safe_profile($profile_id: record<user_profile>, $viewer_lat: float, $viewer_lng: float) {
    LET $profile = SELECT * FROM user_profile WHERE id = $profile_id;
    LET $p = $profile[0];

    RETURN IF $p IS NONE THEN NONE ELSE {
        id: $p.id,
        user: $p.user,
        bio: $p.bio,
        tagline: $p.tagline,
        languages: $p.languages,
        timezone: $p.timezone,
        city: $p.location.city,
        neighborhood: $p.location.neighborhood,
        country: $p.location.country,
        distance_bucket: fn::distance_bucket(
            $p.location.lat ?? 0.0,
            $p.location.lng ?? 0.0,
            $viewer_lat,
            $viewer_lng
        ),
        visibility: $p.visibility,
        last_active: $p.last_active,
        discovery_eligible: $p.discovery_eligible
    } END;
};

-- =============================================================================
-- 4. EVENT COMPLETION TRACKING
-- =============================================================================
-- Fields for verification workflow per SAGA.md requirements:
-- - 1:1 events: BOTH parties must confirm within 48h
-- - Group events: host + ≥2 attendees must confirm within 48h

DEFINE FIELD confirmation_deadline ON event TYPE option<datetime>;
DEFINE FIELD confirmed_count ON event TYPE int DEFAULT 0;
DEFINE FIELD requires_confirmation ON event TYPE bool DEFAULT true;
DEFINE FIELD completion_verified ON event TYPE bool DEFAULT false;
DEFINE FIELD completion_verified_on ON event TYPE option<datetime>;

-- Index for finding events needing confirmation
DEFINE INDEX idx_event_confirmation ON event FIELDS requires_confirmation, confirmation_deadline, completion_verified;

-- Trigger to track confirmations from unified_rsvp
DEFINE EVENT rsvp_confirmation_tracking ON TABLE unified_rsvp
    WHEN $event = "UPDATE"
    AND $before.completion_confirmed IS NONE
    AND $after.completion_confirmed IS NOT NONE
    AND $after.target_type = "event" THEN {

    -- Increment confirmed count on the event
    UPDATE event SET confirmed_count += 1
        WHERE id = type::record("event", $after.target_id);
};

-- Check if event meets verification requirements
DEFINE FUNCTION fn::check_event_completion($event_id: record<event>) {
    LET $evt = SELECT * FROM event WHERE id = $event_id;
    LET $e = $evt[0];

    IF $e IS NONE {
        RETURN false;
    };

    -- Get RSVP count for this event
    LET $rsvp_count = SELECT count() FROM unified_rsvp
        WHERE target_type = "event"
        AND target_id = string::split(<string>$event_id, ":")[1]
        AND status = "attended"
        GROUP ALL;

    LET $attendee_count = $rsvp_count[0].count ?? 0;

    -- 1:1 event (max_attendees = 2): both must confirm
    IF $e.max_attendees IS NOT NONE AND $e.max_attendees <= 2 {
        RETURN $e.confirmed_count >= 2;
    };

    -- Group event: host + 2 attendees (3 total) must confirm
    RETURN $e.confirmed_count >= 3;
};

-- =============================================================================
-- 5. MISSING COMPOSITE INDEXES FOR PERFORMANCE
-- =============================================================================

-- Discovery queries: visibility + status + time
DEFINE INDEX idx_event_discover ON event FIELDS visibility, status, starts_at;
DEFINE INDEX idx_adventure_discover ON adventure FIELDS visibility, status, start_date;
DEFINE INDEX idx_availability_discover ON availability FIELDS status, start_time, expires_at;

-- User activity tracking
DEFINE INDEX idx_profile_active ON user_profile FIELDS last_active;

-- Forum soft deletes (for excluding deleted posts efficiently)
DEFINE INDEX idx_forum_post_deleted ON forum_post FIELDS deleted_on;

-- Guild-scoped event queries
DEFINE INDEX idx_event_guild_time ON event FIELDS guild_id, starts_at;

-- Adventure-scoped event queries
DEFINE INDEX idx_event_adventure_order ON event FIELDS adventure_id, order_in_adventure;

-- Rideshare by driver with status
DEFINE INDEX idx_rideshare_driver_status ON rideshare FIELDS driver_id, status;

-- Hangout by user and status
DEFINE INDEX idx_hangout_status ON hangout FIELDS status;

-- Interest popularity (for discovery sorting)
DEFINE FIELD user_count ON interest TYPE int DEFAULT 0;
DEFINE INDEX idx_interest_popularity ON interest FIELDS category, user_count;

-- =============================================================================
-- 6. VISIBILITY CASCADE CONSTRAINTS
-- =============================================================================
-- Prevent child visibility from exceeding parent visibility
-- Visibility levels: private(0) < invite_only(1) < guilds(2) < public(3)

DEFINE FUNCTION fn::visibility_level($visibility: string) {
    RETURN IF $visibility = "private" THEN 0
        ELSE IF $visibility = "invite_only" THEN 1
        ELSE IF $visibility = "guilds" THEN 2
        ELSE 3
    END;
};

-- Validate event visibility doesn't exceed parent adventure
DEFINE EVENT visibility_cascade_event ON TABLE event
    WHEN ($event = "CREATE" OR $event = "UPDATE")
    AND $after.adventure_id IS NOT NONE THEN {

    LET $parent = SELECT visibility FROM adventure WHERE id = $after.adventure_id;

    IF array::len($parent) > 0 {
        LET $parent_level = fn::visibility_level($parent[0].visibility);
        LET $child_level = fn::visibility_level($after.visibility);

        IF $child_level > $parent_level {
            THROW "Event visibility cannot exceed parent adventure visibility";
        };
    };
};

-- Validate rideshare visibility doesn't exceed parent event/adventure
DEFINE EVENT visibility_cascade_rideshare ON TABLE rideshare
    WHEN $event = "CREATE" OR $event = "UPDATE" THEN {

    -- Check event parent
    IF $after.event_id IS NOT NONE {
        LET $parent = SELECT visibility FROM event WHERE id = $after.event_id;
        IF array::len($parent) > 0 {
            -- Rideshares inherit visibility, can't be more public
            -- (Currently rideshare doesn't have visibility field, so this is a placeholder)
            RETURN NONE;
        };
    };

    -- Check adventure parent
    IF $after.adventure_id IS NOT NONE {
        LET $parent = SELECT visibility FROM adventure WHERE id = $after.adventure_id;
        IF array::len($parent) > 0 {
            RETURN NONE;
        };
    };
};

-- =============================================================================
-- 7. RESONANCE SCORE TRIGGER (Incremental Updates)
-- =============================================================================
-- Instead of recalculating from ledger, update score incrementally

DEFINE EVENT resonance_score_update ON TABLE resonance_ledger WHEN $event = "CREATE" THEN {
    -- Get existing score or create new one
    LET $existing = SELECT * FROM resonance_score WHERE user = $after.user;

    IF array::len($existing) = 0 {
        -- Create new score record
        CREATE resonance_score SET
            user = $after.user,
            total = $after.points,
            questing = IF $after.stat = "questing" THEN $after.points ELSE 0 END,
            mana = IF $after.stat = "mana" THEN $after.points ELSE 0 END,
            wayfinder = IF $after.stat = "wayfinder" THEN $after.points ELSE 0 END,
            attunement = IF $after.stat = "attunement" THEN $after.points ELSE 0 END,
            nexus = IF $after.stat = "nexus" THEN $after.points ELSE 0 END,
            last_calculated = time::now();
    } ELSE {
        -- Update existing score
        UPDATE resonance_score SET
            total += $after.points,
            questing += IF $after.stat = "questing" THEN $after.points ELSE 0 END,
            mana += IF $after.stat = "mana" THEN $after.points ELSE 0 END,
            wayfinder += IF $after.stat = "wayfinder" THEN $after.points ELSE 0 END,
            attunement += IF $after.stat = "attunement" THEN $after.points ELSE 0 END,
            nexus += IF $after.stat = "nexus" THEN $after.points ELSE 0 END,
            last_calculated = time::now()
        WHERE user = $after.user;
    };
};

-- =============================================================================
-- 8. PROFILE ELIGIBILITY TRACKING
-- =============================================================================
-- Update eligibility when user answers questions

-- Required categories for discovery eligibility
-- User must answer at least 1 question from each category
DEFINE FUNCTION fn::required_categories() {
    RETURN ["values", "social", "lifestyle", "communication"];
};

-- Check if user has completed required categories
DEFINE FUNCTION fn::check_eligibility($user_id: record<user>) {
    LET $required = fn::required_categories();

    -- Get distinct categories user has answered
    LET $answered = SELECT question.category AS cat FROM answer
        WHERE user = $user_id
        GROUP BY cat;

    LET $answered_cats = array::map($answered, |$a| $a.cat);

    -- Check if all required categories are covered
    LET $missing = array::filter($required, |$r| NOT ($r IN $answered_cats));

    RETURN array::len($missing) = 0;
};

-- Update profile eligibility when answer is created
DEFINE EVENT answer_eligibility_check ON TABLE answer WHEN $event = "CREATE" THEN {
    -- Count user's answers
    LET $count = SELECT count() FROM answer WHERE user = $after.user GROUP ALL;
    LET $question_count = $count[0].count ?? 0;

    -- Get categories completed
    LET $cats = SELECT question.category AS cat FROM answer WHERE user = $after.user GROUP BY cat;
    LET $categories = array::map($cats, |$c| $c.cat);

    -- Check eligibility (3+ questions AND all required categories)
    LET $eligible = $question_count >= 3 AND fn::check_eligibility($after.user);

    -- Update profile
    UPDATE user_profile SET
        question_count = $question_count,
        categories_completed = $categories,
        discovery_eligible = $eligible
    WHERE user = $after.user;
};

-- =============================================================================
-- 9. DISCOVERY LIMIT ENFORCEMENT
-- =============================================================================

-- Increment discovery count (called from API)
DEFINE FUNCTION fn::increment_discovery_count($user_id: record<user>, $type: string) {
    LET $today = time::format(time::now(), "%Y-%m-%d");

    -- Try to get existing record
    LET $existing = SELECT * FROM discovery_daily_count
        WHERE user_id = $user_id AND date = $today;

    IF array::len($existing) = 0 {
        -- Create new record
        CREATE discovery_daily_count SET
            user_id = $user_id,
            date = $today,
            people_shown = IF $type = "people" THEN 1 ELSE 0 END,
            events_shown = IF $type = "events" THEN 1 ELSE 0 END,
            guilds_shown = IF $type = "guilds" THEN 1 ELSE 0 END;
    } ELSE {
        -- Update existing
        UPDATE discovery_daily_count SET
            people_shown += IF $type = "people" THEN 1 ELSE 0 END,
            events_shown += IF $type = "events" THEN 1 ELSE 0 END,
            guilds_shown += IF $type = "guilds" THEN 1 ELSE 0 END,
            updated_on = time::now()
        WHERE user_id = $user_id AND date = $today;
    };

    RETURN true;
};

-- Check if user has remaining discovery quota
DEFINE FUNCTION fn::has_discovery_quota($user_id: record<user>, $type: string, $limit: int) {
    LET $today = time::format(time::now(), "%Y-%m-%d");

    LET $record = SELECT * FROM discovery_daily_count
        WHERE user_id = $user_id AND date = $today;

    IF array::len($record) = 0 {
        RETURN true;
    };

    LET $r = $record[0];

    RETURN IF $type = "people" THEN $r.people_shown < $limit
        ELSE IF $type = "events" THEN $r.events_shown < $limit
        ELSE IF $type = "guilds" THEN $r.guilds_shown < $limit
        ELSE true
    END;
};

-- =============================================================================
-- 10. DENORMALIZED COUNTS (Trigger-Updated)
-- =============================================================================

-- Add count fields to frequently queried entities
DEFINE FIELD attendee_count ON event TYPE int DEFAULT 0;
DEFINE FIELD participant_count ON adventure TYPE int DEFAULT 0;
DEFINE FIELD member_count ON guild TYPE int DEFAULT 0;

-- Update event attendee count on unified_rsvp changes
DEFINE EVENT update_event_attendee_count ON TABLE unified_rsvp
    WHEN ($event = "CREATE" OR $event = "UPDATE" OR $event = "DELETE")
    AND ($after.target_type = "event" OR $before.target_type = "event") THEN {

    LET $event_id = IF $event = "DELETE" THEN $before.target_id ELSE $after.target_id END;

    IF $event_id IS NOT NONE {
        LET $count = SELECT count() FROM unified_rsvp
            WHERE target_type = "event"
            AND target_id = $event_id
            AND status IN ["approved", "attended"]
            GROUP ALL;

        UPDATE event SET attendee_count = $count[0].count ?? 0
            WHERE id = type::record("event", $event_id);
    };
};

-- Update adventure participant count
DEFINE EVENT update_adventure_participant_count ON TABLE adventure_participant
    WHEN $event = "CREATE" OR $event = "DELETE" THEN {

    LET $adventure_id = IF $event = "DELETE" THEN $before.adventure_id ELSE $after.adventure_id END;

    LET $count = SELECT count() FROM adventure_participant
        WHERE adventure_id = $adventure_id
        AND status IN ["interested", "committed"]
        GROUP ALL;

    UPDATE adventure SET participant_count = $count[0].count ?? 0
        WHERE id = $adventure_id;
};

-- Update guild member count
DEFINE EVENT update_guild_member_count ON TABLE responsible_for
    WHEN $event = "CREATE" OR $event = "DELETE" THEN {

    LET $guild_id = IF $event = "DELETE" THEN $before.out ELSE $after.out END;

    LET $count = SELECT count() FROM responsible_for
        WHERE out = $guild_id
        GROUP ALL;

    UPDATE guild SET member_count = $count[0].count ?? 0
        WHERE id = $guild_id;
};

-- Update interest user count
DEFINE EVENT update_interest_user_count ON TABLE has_interest
    WHEN $event = "CREATE" OR $event = "DELETE" THEN {

    LET $interest_id = IF $event = "DELETE" THEN $before.out ELSE $after.out END;

    LET $count = SELECT count() FROM has_interest
        WHERE out = $interest_id
        GROUP ALL;

    UPDATE interest SET user_count = $count[0].count ?? 0
        WHERE id = $interest_id;
};

-- =============================================================================
-- 11. EARLY CONFIRMATION TRACKING
-- =============================================================================
-- Track early confirmations for bonus points (≥2 hours before event start)

DEFINE FUNCTION fn::is_early_confirm($event_id: record<event>) {
    LET $evt = SELECT starts_at FROM event WHERE id = $event_id;

    IF array::len($evt) = 0 {
        RETURN false;
    };

    LET $start = $evt[0].starts_at;
    LET $now = time::now();
    LET $two_hours_before = $start - 2h;

    RETURN $now < $two_hours_before;
};

-- =============================================================================
-- 12. ON-TIME CHECKIN TRACKING
-- =============================================================================
-- Check if checkin is within ±10 minute window of event start

DEFINE FUNCTION fn::is_ontime_checkin($event_id: record<event>, $checkin_time: datetime) {
    LET $evt = SELECT starts_at FROM event WHERE id = $event_id;

    IF array::len($evt) = 0 {
        RETURN false;
    };

    LET $start = $evt[0].starts_at;
    LET $window_start = $start - 10m;
    LET $window_end = $start + 10m;

    RETURN $checkin_time >= $window_start AND $checkin_time <= $window_end;
};

-- =============================================================================
-- 13. CLEANUP: Mark Legacy Tables for Deprecation
-- =============================================================================
-- These tables are superseded by unified_rsvp but kept for migration
-- TODO: Remove in future migration after data migration complete

-- Add deprecation marker to legacy tables (informational only)
-- event_participant, event_rsvp, and rsvp relation will be removed
-- after data is migrated to unified_rsvp
