-- Migration 007: Features v2
-- Version: 2.0.4
-- Date: 2026-01-08
--
-- This migration adds:
-- 1. Trust Ratings - Event-anchored trust system with reviews, endorsements, anti-abuse
-- 2. Role Catalogs - Guild/user-scoped reusable role templates for events/rideshares
-- 3. Adventure Decoupling - Allow adventures organized by guilds OR individual users
-- 4. Voting System - Generic voting with FPTP, ranked choice, approval, multi-select

-- =============================================================================
-- 1. TRUST RATINGS
-- =============================================================================

-- Main trust rating table (event-anchored)
DEFINE TABLE trust_rating SCHEMAFULL;
DEFINE FIELD rater_id ON trust_rating TYPE record<user>;
DEFINE FIELD ratee_id ON trust_rating TYPE record<user>;
DEFINE FIELD anchor_type ON trust_rating TYPE string ASSERT $value IN ["event", "rideshare"];
DEFINE FIELD anchor_id ON trust_rating TYPE string;
DEFINE FIELD trust_level ON trust_rating TYPE string ASSERT $value IN ["trust", "distrust"];
DEFINE FIELD trust_review ON trust_rating TYPE string ASSERT string::len($value) <= 240;
DEFINE FIELD review_visibility ON trust_rating TYPE string DEFAULT "public"
    ASSERT $value IN ["public", "admin_only"];
DEFINE FIELD created_on ON trust_rating TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON trust_rating TYPE datetime DEFAULT time::now();

-- One rating per (rater, ratee, anchor_type, anchor_id) combination
DEFINE INDEX trust_rating_unique ON trust_rating FIELDS rater_id, ratee_id, anchor_type, anchor_id UNIQUE;
DEFINE INDEX idx_trust_rating_ratee ON trust_rating FIELDS ratee_id;
DEFINE INDEX idx_trust_rating_rater ON trust_rating FIELDS rater_id;
DEFINE INDEX idx_trust_rating_anchor ON trust_rating FIELDS anchor_type, anchor_id;

-- Audit ledger (immutable history of trust rating changes)
DEFINE TABLE trust_rating_history SCHEMAFULL;
DEFINE FIELD trust_rating_id ON trust_rating_history TYPE record<trust_rating>;
DEFINE FIELD previous_level ON trust_rating_history TYPE option<string>;
DEFINE FIELD new_level ON trust_rating_history TYPE string;
DEFINE FIELD previous_review ON trust_rating_history TYPE option<string>;
DEFINE FIELD new_review ON trust_rating_history TYPE string;
DEFINE FIELD changed_on ON trust_rating_history TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_trust_history_rating ON trust_rating_history FIELDS trust_rating_id;
DEFINE INDEX idx_trust_history_changed ON trust_rating_history FIELDS changed_on;

-- Endorsements from co-attendees on a trust rating
DEFINE TABLE trust_endorsement SCHEMAFULL;
DEFINE FIELD trust_rating_id ON trust_endorsement TYPE record<trust_rating>;
DEFINE FIELD endorser_id ON trust_endorsement TYPE record<user>;
DEFINE FIELD endorsement_type ON trust_endorsement TYPE string ASSERT $value IN ["agree", "disagree"];
DEFINE FIELD note ON trust_endorsement TYPE option<string> ASSERT string::len($value) <= 240;
DEFINE FIELD created_on ON trust_endorsement TYPE datetime DEFAULT time::now();

-- One endorsement per (endorser, rating) combination
DEFINE INDEX trust_endorsement_unique ON trust_endorsement FIELDS endorser_id, trust_rating_id UNIQUE;
DEFINE INDEX idx_trust_endorsement_rating ON trust_endorsement FIELDS trust_rating_id;

-- Daily throttling for trust ratings
DEFINE TABLE trust_rating_daily_count SCHEMAFULL;
DEFINE FIELD user_id ON trust_rating_daily_count TYPE record<user>;
DEFINE FIELD date ON trust_rating_daily_count TYPE string; -- YYYY-MM-DD format
DEFINE FIELD count ON trust_rating_daily_count TYPE int DEFAULT 0;

DEFINE INDEX trust_rating_daily_unique ON trust_rating_daily_count FIELDS user_id, date UNIQUE;

-- Trust rating validation triggers
DEFINE EVENT trust_rating_self_check ON TABLE trust_rating WHEN $event = "CREATE" THEN {
    IF $after.rater_id = $after.ratee_id {
        THROW "Cannot rate yourself";
    };
};

-- Auto-update updated_on when trust_level or review changes
-- Note: Only fires when content changes to prevent recursive trigger
DEFINE EVENT trust_rating_updated ON TABLE trust_rating WHEN $event = "UPDATE" AND ($before.trust_level != $after.trust_level OR $before.trust_review != $after.trust_review) THEN {
    UPDATE trust_rating SET updated_on = time::now() WHERE id = $after.id;
};

-- Create history entry on updates
DEFINE EVENT trust_rating_history_on_update ON TABLE trust_rating WHEN $event = "UPDATE" THEN {
    IF $before.trust_level != $after.trust_level OR $before.trust_review != $after.trust_review {
        CREATE trust_rating_history SET
            trust_rating_id = $after.id,
            previous_level = $before.trust_level,
            new_level = $after.trust_level,
            previous_review = $before.trust_review,
            new_review = $after.trust_review,
            changed_on = time::now();
    };
};

-- Endorsement self-check
DEFINE EVENT trust_endorsement_self_check ON TABLE trust_endorsement WHEN $event = "CREATE" THEN {
    LET $rating = SELECT rater_id FROM trust_rating WHERE id = $after.trust_rating_id;
    IF array::len($rating) > 0 AND $rating[0].rater_id = $after.endorser_id {
        THROW "Cannot endorse your own rating";
    };
};

-- Daily limit trigger (10 ratings per day)
DEFINE EVENT trust_rating_daily_limit ON TABLE trust_rating WHEN $event = "CREATE" THEN {
    LET $today = time::format(time::now(), "%Y-%m-%d");
    LET $existing = SELECT count FROM trust_rating_daily_count
        WHERE user_id = $after.rater_id AND date = $today;
    LET $current_count = $existing[0].count ?? 0;
    IF $current_count >= 10 {
        THROW "Daily trust rating limit (10) exceeded";
    };
    -- Upsert daily count
    UPSERT trust_rating_daily_count SET
        user_id = $after.rater_id,
        date = $today,
        count = $current_count + 1
    WHERE user_id = $after.rater_id AND date = $today;
};

-- Function to check if a user can rate another (anchor validation)
DEFINE FUNCTION fn::can_rate_user($rater_id: string, $ratee_id: string, $anchor_type: string, $anchor_id: string) {
    -- Check if anchor is verified completed
    IF $anchor_type = "event" {
        LET $event = SELECT completion_verified FROM event WHERE id = type::record("event", $anchor_id);
        IF array::len($event) = 0 OR !$event[0].completion_verified {
            RETURN false;
        };
        -- Check both users have confirmed attendance
        LET $rater_rsvp = SELECT completion_confirmed FROM unified_rsvp
            WHERE user_id = $rater_id AND target_type = "event" AND target_id = $anchor_id;
        LET $ratee_rsvp = SELECT completion_confirmed FROM unified_rsvp
            WHERE user_id = $ratee_id AND target_type = "event" AND target_id = $anchor_id;
        IF array::len($rater_rsvp) = 0 OR $rater_rsvp[0].completion_confirmed IS NONE {
            RETURN false;
        };
        IF array::len($ratee_rsvp) = 0 OR $ratee_rsvp[0].completion_confirmed IS NONE {
            RETURN false;
        };
        RETURN true;
    };
    IF $anchor_type = "rideshare" {
        LET $rideshare = SELECT status FROM rideshare WHERE id = type::record("rideshare", $anchor_id);
        IF array::len($rideshare) = 0 OR $rideshare[0].status != "completed" {
            RETURN false;
        };
        -- Check rater is driver or passenger
        LET $is_driver = SELECT id FROM rideshare
            WHERE id = type::record("rideshare", $anchor_id) AND driver_id = type::record("user", $rater_id);
        LET $is_passenger = SELECT id FROM rideshare_seat
            WHERE rideshare_id = type::record("rideshare", $anchor_id)
            AND passenger_id = type::record("user", $rater_id)
            AND status = "confirmed";
        IF array::len($is_driver) = 0 AND array::len($is_passenger) = 0 {
            RETURN false;
        };
        -- Check ratee is driver or passenger
        LET $ratee_is_driver = SELECT id FROM rideshare
            WHERE id = type::record("rideshare", $anchor_id) AND driver_id = type::record("user", $ratee_id);
        LET $ratee_is_passenger = SELECT id FROM rideshare_seat
            WHERE rideshare_id = type::record("rideshare", $anchor_id)
            AND passenger_id = type::record("user", $ratee_id)
            AND status = "confirmed";
        IF array::len($ratee_is_driver) = 0 AND array::len($ratee_is_passenger) = 0 {
            RETURN false;
        };
        RETURN true;
    };
    RETURN false;
};

-- =============================================================================
-- 2. ROLE CATALOGS
-- =============================================================================

-- Role catalog for reusable role templates
DEFINE TABLE role_catalog SCHEMAFULL;
DEFINE FIELD scope_type ON role_catalog TYPE string ASSERT $value IN ["guild", "user"];
DEFINE FIELD scope_id ON role_catalog TYPE string; -- guild:<id> or user:<id>
DEFINE FIELD role_type ON role_catalog TYPE string ASSERT $value IN ["event", "rideshare"];
DEFINE FIELD name ON role_catalog TYPE string ASSERT string::len($value) <= 50;
DEFINE FIELD description ON role_catalog TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD icon ON role_catalog TYPE option<string> ASSERT string::len($value) <= 50;
DEFINE FIELD is_active ON role_catalog TYPE bool DEFAULT true;
DEFINE FIELD created_by ON role_catalog TYPE record<user>;
DEFINE FIELD created_on ON role_catalog TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON role_catalog TYPE datetime DEFAULT time::now();

-- Unique role name per scope and type
DEFINE INDEX role_catalog_unique ON role_catalog FIELDS scope_type, scope_id, role_type, name UNIQUE;
DEFINE INDEX idx_role_catalog_scope ON role_catalog FIELDS scope_type, scope_id;
DEFINE INDEX idx_role_catalog_active ON role_catalog FIELDS is_active;

-- Add catalog reference to existing event_role table
DEFINE FIELD catalog_role_id ON event_role TYPE option<record<role_catalog>>;

-- Rideshare roles (mirrors event_role structure)
DEFINE TABLE rideshare_role SCHEMAFULL;
DEFINE FIELD rideshare_id ON rideshare_role TYPE record<rideshare>;
DEFINE FIELD catalog_role_id ON rideshare_role TYPE option<record<role_catalog>>;
DEFINE FIELD name ON rideshare_role TYPE string ASSERT string::len($value) <= 50;
DEFINE FIELD description ON rideshare_role TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD max_slots ON rideshare_role TYPE int DEFAULT 1;
DEFINE FIELD filled_slots ON rideshare_role TYPE int DEFAULT 0;
DEFINE FIELD sort_order ON rideshare_role TYPE int DEFAULT 0;
DEFINE FIELD created_by ON rideshare_role TYPE record<user>;
DEFINE FIELD created_on ON rideshare_role TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON rideshare_role TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_rideshare_role_rideshare ON rideshare_role FIELDS rideshare_id;

-- Rideshare role assignments
DEFINE TABLE rideshare_role_assignment SCHEMAFULL;
DEFINE FIELD rideshare_id ON rideshare_role_assignment TYPE record<rideshare>;
DEFINE FIELD role_id ON rideshare_role_assignment TYPE record<rideshare_role>;
DEFINE FIELD user_id ON rideshare_role_assignment TYPE record<user>;
DEFINE FIELD note ON rideshare_role_assignment TYPE option<string> ASSERT string::len($value) <= 200;
DEFINE FIELD status ON rideshare_role_assignment TYPE string DEFAULT "confirmed"
    ASSERT $value IN ["pending", "confirmed", "cancelled"];
DEFINE FIELD assigned_on ON rideshare_role_assignment TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON rideshare_role_assignment TYPE datetime DEFAULT time::now();

DEFINE INDEX rideshare_role_assignment_unique ON rideshare_role_assignment FIELDS role_id, user_id UNIQUE;
DEFINE INDEX idx_rideshare_role_assignment_rideshare ON rideshare_role_assignment FIELDS rideshare_id;
DEFINE INDEX idx_rideshare_role_assignment_user ON rideshare_role_assignment FIELDS user_id;

-- Role catalog limit trigger (50 per scope)
DEFINE EVENT check_role_catalog_limit ON TABLE role_catalog WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM role_catalog
        WHERE scope_type = $after.scope_type AND scope_id = $after.scope_id GROUP ALL);
    IF $count[0].count > 50 {
        THROW "Maximum 50 role templates per scope"
    };
};

-- Rideshare role limit (20 per rideshare)
DEFINE EVENT check_rideshare_role_limit ON TABLE rideshare_role WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM rideshare_role WHERE rideshare_id = $after.rideshare_id GROUP ALL);
    IF $count[0].count > 20 {
        THROW "Maximum 20 roles per rideshare"
    };
};

-- Track filled slots on rideshare role assignment
DEFINE EVENT increment_rideshare_filled_slots ON TABLE rideshare_role_assignment WHEN $event = "CREATE" THEN {
    UPDATE rideshare_role SET filled_slots += 1 WHERE id = $after.role_id;
};

DEFINE EVENT decrement_rideshare_filled_slots ON TABLE rideshare_role_assignment WHEN $event = "DELETE" THEN {
    UPDATE rideshare_role SET filled_slots -= 1 WHERE id = $before.role_id;
};

-- Auto-update timestamps
-- NOTE: These events were removed because they cause infinite recursion when the
-- repository's UPDATE query already sets updated_on = time::now(). The UPDATE
-- triggers the event, which does another UPDATE, triggering the event again.
-- Timestamps are now handled directly in repository code.

-- =============================================================================
-- 3. ADVENTURE DECOUPLING
-- =============================================================================

-- Add organizer fields to adventure table
DEFINE FIELD organizer_type ON adventure TYPE string DEFAULT "guild"
    ASSERT $value IN ["guild", "user"];
DEFINE FIELD organizer_id ON adventure TYPE string; -- "guild:<id>" or "user:<id>"
DEFINE FIELD organizer_user_id ON adventure TYPE record<user>; -- The person organizing

-- Freeze fields (when organizer loses guild membership)
DEFINE FIELD freeze_reason ON adventure TYPE option<string>;
DEFINE FIELD frozen_on ON adventure TYPE option<datetime>;

-- Add "frozen" status to adventure status validation
-- Update the existing check_adventure_status trigger
REMOVE EVENT IF EXISTS check_adventure_status ON adventure;
DEFINE EVENT check_adventure_status ON TABLE adventure WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    LET $valid_statuses = ["idea", "planning", "confirmed", "active", "completed", "cancelled", "frozen"];
    IF NOT ($after.status IN $valid_statuses) {
        THROW "Invalid adventure status"
    };
};

-- Adventure admission table (permission to view/RSVP)
DEFINE TABLE adventure_admission SCHEMAFULL;
DEFINE FIELD adventure_id ON adventure_admission TYPE record<adventure>;
DEFINE FIELD user_id ON adventure_admission TYPE record<user>;
DEFINE FIELD status ON adventure_admission TYPE string DEFAULT "requested"
    ASSERT $value IN ["requested", "admitted", "rejected"];
DEFINE FIELD requested_by ON adventure_admission TYPE string DEFAULT "self"
    ASSERT $value IN ["self", "invited"];
DEFINE FIELD invited_by_id ON adventure_admission TYPE option<record<user>>;
DEFINE FIELD rejection_reason ON adventure_admission TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD requested_on ON adventure_admission TYPE datetime DEFAULT time::now();
DEFINE FIELD decided_on ON adventure_admission TYPE option<datetime>;
DEFINE FIELD updated_on ON adventure_admission TYPE datetime DEFAULT time::now();

DEFINE INDEX adventure_admission_unique ON adventure_admission FIELDS adventure_id, user_id UNIQUE;
DEFINE INDEX idx_adventure_admission_adventure ON adventure_admission FIELDS adventure_id;
DEFINE INDEX idx_adventure_admission_user ON adventure_admission FIELDS user_id;
DEFINE INDEX idx_adventure_admission_status ON adventure_admission FIELDS status;

-- Prevent organizer_type changes after creation
DEFINE EVENT prevent_organizer_type_change ON TABLE adventure WHEN $event = "UPDATE" THEN {
    IF $before.organizer_type IS NOT NONE AND $before.organizer_type != $after.organizer_type {
        THROW "Cannot change adventure organizer_type after creation";
    };
};

-- Limit admissions per adventure (500)
DEFINE EVENT check_adventure_admission_limit ON TABLE adventure_admission WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM adventure_admission
        WHERE adventure_id = $after.adventure_id AND status = "admitted" GROUP ALL);
    IF $count[0].count > 500 {
        THROW "Maximum 500 admitted participants per adventure"
    };
};

-- Remove auto-update event as it causes recursion
-- Updated_on and decided_on are managed by the repository directly
REMOVE EVENT IF EXISTS adventure_admission_updated ON adventure_admission;

-- Data migration: Populate organizer fields from existing guild_id
-- Note: Run this once after migration
-- UPDATE adventure SET
--     organizer_type = IF guild_id IS NOT NONE THEN "guild" ELSE "user" END,
--     organizer_id = IF guild_id IS NOT NONE
--         THEN string::concat("guild:", string::replace(<string> guild_id, "guild:", ""))
--         ELSE string::concat("user:", <string> created_by_id) END,
--     organizer_user_id = created_by_id
-- WHERE organizer_type IS NONE;

-- =============================================================================
-- 4. VOTING SYSTEM
-- =============================================================================

-- Main vote table
DEFINE TABLE vote SCHEMAFULL;
DEFINE FIELD scope_type ON vote TYPE string ASSERT $value IN ["guild", "global"];
DEFINE FIELD scope_id ON vote TYPE option<record<guild>>; -- null for global votes
DEFINE FIELD created_by ON vote TYPE record<user>;
DEFINE FIELD title ON vote TYPE string ASSERT string::len($value) <= 200;
DEFINE FIELD description ON vote TYPE option<string> ASSERT string::len($value) <= 2000;
DEFINE FIELD vote_type ON vote TYPE string ASSERT $value IN ["fptp", "ranked_choice", "approval", "multi_select"];
DEFINE FIELD opens_at ON vote TYPE datetime;
DEFINE FIELD closes_at ON vote TYPE datetime;
DEFINE FIELD status ON vote TYPE string DEFAULT "draft"
    ASSERT $value IN ["draft", "open", "closed", "cancelled"];
DEFINE FIELD results_visibility ON vote TYPE string DEFAULT "after_close"
    ASSERT $value IN ["live", "after_close", "admin_only"];
DEFINE FIELD max_options_selectable ON vote TYPE option<int>; -- For multi_select only
DEFINE FIELD allow_abstain ON vote TYPE bool DEFAULT false;
DEFINE FIELD created_on ON vote TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON vote TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_vote_scope ON vote FIELDS scope_type, scope_id;
DEFINE INDEX idx_vote_status ON vote FIELDS status;
DEFINE INDEX idx_vote_opens ON vote FIELDS opens_at;
DEFINE INDEX idx_vote_closes ON vote FIELDS closes_at;

-- Vote options
DEFINE TABLE vote_option SCHEMAFULL;
DEFINE FIELD vote_id ON vote_option TYPE record<vote>;
DEFINE FIELD option_text ON vote_option TYPE string ASSERT string::len($value) <= 200;
DEFINE FIELD option_description ON vote_option TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD sort_order ON vote_option TYPE int DEFAULT 0;
DEFINE FIELD created_by ON vote_option TYPE record<user>;
DEFINE FIELD created_on ON vote_option TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_vote_option_vote ON vote_option FIELDS vote_id;
DEFINE INDEX idx_vote_option_order ON vote_option FIELDS vote_id, sort_order;

-- Vote ballots (immutable after creation)
DEFINE TABLE vote_ballot SCHEMAFULL;
DEFINE FIELD vote_id ON vote_ballot TYPE record<vote>;
DEFINE FIELD voter_user_id ON vote_ballot TYPE record<user>;
-- Snapshot of voter identity at time of voting (for audit transparency)
DEFINE FIELD voter_snapshot ON vote_ballot TYPE object FLEXIBLE; -- {username, display_name}
-- Ballot data varies by vote type:
-- FPTP: {option_id: "vote_option:xxx"}
-- Ranked Choice: {rankings: ["vote_option:a", "vote_option:b", "vote_option:c"]}
-- Approval: {approved_options: ["vote_option:a", "vote_option:b"]}
-- Multi-select: {selected_options: ["vote_option:a", "vote_option:b"]}
DEFINE FIELD ballot_data ON vote_ballot TYPE object FLEXIBLE;
DEFINE FIELD is_abstain ON vote_ballot TYPE bool DEFAULT false;
DEFINE FIELD created_on ON vote_ballot TYPE datetime DEFAULT time::now();

-- One ballot per (vote, voter) combination
DEFINE INDEX vote_ballot_unique ON vote_ballot FIELDS vote_id, voter_user_id UNIQUE;
DEFINE INDEX idx_vote_ballot_vote ON vote_ballot FIELDS vote_id;
DEFINE INDEX idx_vote_ballot_voter ON vote_ballot FIELDS voter_user_id;

-- Prevent ballot updates (immutability)
DEFINE EVENT prevent_ballot_update ON TABLE vote_ballot WHEN $event = "UPDATE" THEN {
    THROW "Ballots are immutable - delete and recreate to change vote";
};

-- Validate vote dates
DEFINE EVENT validate_vote_dates ON TABLE vote WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    IF $after.closes_at <= $after.opens_at {
        THROW "Vote close time must be after open time"
    };
};

-- Limit options per vote (20)
DEFINE EVENT check_vote_option_limit ON TABLE vote_option WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM vote_option WHERE vote_id = $after.vote_id GROUP ALL);
    IF $count[0].count > 20 {
        THROW "Maximum 20 options per vote"
    };
};

-- Limit votes per guild (50 active)
DEFINE EVENT check_guild_vote_limit ON TABLE vote WHEN $event = "CREATE" THEN {
    IF $after.scope_type = "guild" AND $after.scope_id IS NOT NONE {
        LET $count = (SELECT count() FROM vote
            WHERE scope_type = "guild" AND scope_id = $after.scope_id
            AND status IN ["draft", "open"] GROUP ALL);
        IF $count[0].count > 50 {
            THROW "Maximum 50 active votes per guild"
        };
    };
};

-- Prevent options modification after voting starts
DEFINE EVENT prevent_option_change_after_voting ON TABLE vote_option
    WHEN $event = "UPDATE" OR $event = "DELETE"
    THEN {
    LET $vote = SELECT status FROM vote WHERE id = $before.vote_id;
    IF array::len($vote) > 0 AND $vote[0].status IN ["open", "closed"] {
        THROW "Cannot modify options after voting has started";
    };
};

-- Auto-update timestamps (only when status or other fields change, not just updated_on)
DEFINE EVENT vote_updated ON TABLE vote WHEN $event = "UPDATE" AND ($before.status != $after.status OR $before.title != $after.title OR $before.description != $after.description) THEN {
    UPDATE vote SET updated_on = time::now() WHERE id = $after.id;
};

-- =============================================================================
-- 5. HELPER FUNCTIONS
-- =============================================================================

-- Check if user can create guild vote
DEFINE FUNCTION fn::can_create_guild_vote($user_id: string, $guild_id: string) {
    -- Check if user is a guild admin (has member record with admin role)
    LET $is_member = SELECT id FROM responsible_for
        WHERE in IN (SELECT id FROM member WHERE user = type::record("user", $user_id))
        AND out = type::record("guild", $guild_id);
    RETURN array::len($is_member) > 0;
};

-- Check if user can vote in a guild vote
DEFINE FUNCTION fn::can_vote_in_guild($user_id: string, $guild_id: string) {
    LET $is_member = SELECT id FROM responsible_for
        WHERE in IN (SELECT id FROM member WHERE user = type::record("user", $user_id))
        AND out = type::record("guild", $guild_id);
    RETURN array::len($is_member) > 0;
};

-- Get trust aggregate for a user
DEFINE FUNCTION fn::get_trust_aggregate($user_id: string) {
    LET $received = SELECT count() as cnt, trust_level FROM trust_rating
        WHERE ratee_id = type::record("user", $user_id)
        GROUP BY trust_level;

    LET $trust_count = array::filter($received, |$r| $r.trust_level = "trust")[0].cnt ?? 0;
    LET $distrust_count = array::filter($received, |$r| $r.trust_level = "distrust")[0].cnt ?? 0;

    LET $endorsements = SELECT count() as cnt FROM trust_endorsement
        WHERE trust_rating_id IN (SELECT id FROM trust_rating WHERE ratee_id = type::record("user", $user_id))
        GROUP ALL;

    RETURN {
        trust_count: $trust_count,
        distrust_count: $distrust_count,
        endorsement_count: $endorsements[0].cnt ?? 0,
        net_trust: $trust_count - $distrust_count
    };
};

-- Check if user is admitted to adventure
DEFINE FUNCTION fn::is_admitted_to_adventure($user_id: string, $adventure_id: string) {
    LET $admission = SELECT status FROM adventure_admission
        WHERE user_id = type::record("user", $user_id)
        AND adventure_id = type::record("adventure", $adventure_id)
        AND status = "admitted";
    RETURN array::len($admission) > 0;
};

-- =============================================================================
-- 6. CASCADE DELETE UPDATES
-- =============================================================================

-- Cascade delete for trust ratings when user is deleted
-- (Extends existing cascade_user_delete trigger from 006)
DEFINE EVENT cascade_user_trust_ratings ON TABLE user WHEN $event = "DELETE" THEN {
    -- Delete trust ratings given by this user
    DELETE trust_rating WHERE rater_id = $before.id;
    -- Delete trust ratings received by this user
    DELETE trust_rating WHERE ratee_id = $before.id;
    -- Delete endorsements by this user
    DELETE trust_endorsement WHERE endorser_id = $before.id;
    -- Delete daily counts
    DELETE trust_rating_daily_count WHERE user_id = $before.id;
    -- Delete role catalogs owned by user
    DELETE role_catalog WHERE scope_type = "user" AND scope_id = string::concat("user:", <string> $before.id);
    -- Delete vote ballots
    DELETE vote_ballot WHERE voter_user_id = $before.id;
    -- Delete adventure admissions
    DELETE adventure_admission WHERE user_id = $before.id;
};

-- Cascade delete for trust ratings when trust_rating is deleted
DEFINE EVENT cascade_trust_rating_delete ON TABLE trust_rating WHEN $event = "DELETE" THEN {
    DELETE trust_endorsement WHERE trust_rating_id = $before.id;
    DELETE trust_rating_history WHERE trust_rating_id = $before.id;
};

-- Cascade delete for votes
DEFINE EVENT cascade_vote_delete ON TABLE vote WHEN $event = "DELETE" THEN {
    DELETE vote_option WHERE vote_id = $before.id;
    DELETE vote_ballot WHERE vote_id = $before.id;
};

-- Cascade delete for adventures (extends existing)
DEFINE EVENT cascade_adventure_admissions ON TABLE adventure WHEN $event = "DELETE" THEN {
    DELETE adventure_admission WHERE adventure_id = $before.id;
};

-- Cascade delete for rideshares (extends existing)
DEFINE EVENT cascade_rideshare_roles ON TABLE rideshare WHEN $event = "DELETE" THEN {
    LET $roles = SELECT id FROM rideshare_role WHERE rideshare_id = $before.id;
    DELETE rideshare_role_assignment WHERE role_id IN $roles.id;
    DELETE rideshare_role WHERE rideshare_id = $before.id;
};

-- Cascade delete for guilds (extends existing)
DEFINE EVENT cascade_guild_role_catalogs ON TABLE guild WHEN $event = "DELETE" THEN {
    DELETE role_catalog WHERE scope_type = "guild" AND scope_id = string::concat("guild:", <string> $before.id);
    DELETE vote WHERE scope_type = "guild" AND scope_id = $before.id;
};

-- =============================================================================
-- 7. FIX RECURSIVE EVENTS
-- =============================================================================

-- Fix adventure_updated event to prevent recursion
-- Only fires when content changes, not when updated_on changes
REMOVE EVENT IF EXISTS adventure_updated ON adventure;
DEFINE EVENT adventure_updated ON TABLE adventure WHEN $event = "UPDATE" AND ($before.title != $after.title OR $before.description != $after.description OR $before.status != $after.status OR $before.organizer_user_id != $after.organizer_user_id) THEN {
    UPDATE adventure SET updated_on = time::now() WHERE id = $after.id;
};
