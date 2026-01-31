-- Saga Bug Fixes Migration
-- Version: 2.2.0
-- Created: 2026-01-06
--
-- This migration fixes critical bugs discovered during schema audit:
-- 1. Duplicate block tables (user_block vs block) - consolidate to single table
-- 2. Missing rideshare visibility field and cascade trigger
-- 3. Resonance score can go negative - add floor at 0
-- 4. Add self-block check to user_block for safety

-- =============================================================================
-- 1. BLOCK TABLE CONSOLIDATION
-- =============================================================================
-- Issue: Two block tables exist with different field names:
--   - user_block: blocker, blocked (NO self-block trigger)
--   - block: blocker_user_id, blocked_user_id (HAS self-block trigger)
--
-- Solution: Add self-block trigger to user_block for safety.
-- We keep both tables since existing code may reference either.
-- The user_block table uses simpler field names (blocker/blocked).

DEFINE EVENT user_block_self_check ON TABLE user_block WHEN $event = "CREATE" THEN {
    IF $after.blocker = $after.blocked {
        THROW "Cannot block yourself";
    };
};

-- =============================================================================
-- 2. RIDESHARE VISIBILITY
-- =============================================================================
-- Issue: Rideshare table has no visibility field, allowing potential data leaks.
-- Rideshares should inherit visibility from parent event/adventure.
--
-- Add visibility field with default matching typical use case.

DEFINE FIELD visibility ON rideshare TYPE string DEFAULT "guilds"
    ASSERT $value IN ["private", "invite_only", "guilds", "public"];

-- Index for visibility filtering
DEFINE INDEX idx_rideshare_visibility ON rideshare FIELDS visibility;
DEFINE INDEX OVERWRITE idx_rideshare_driver_status ON rideshare FIELDS driver_id, status;

-- Visibility cascade enforcement: rideshare visibility cannot exceed parent
DEFINE EVENT OVERWRITE visibility_cascade_rideshare ON TABLE rideshare
    WHEN $event = "CREATE" OR $event = "UPDATE" THEN {

    LET $child_level = IF $after.visibility = "private" THEN 0
        ELSE IF $after.visibility = "invite_only" THEN 1
        ELSE IF $after.visibility = "guilds" THEN 2
        ELSE 3 END;

    -- Check event parent first
    IF $after.event_id != NONE {
        LET $parent_event = (SELECT visibility FROM event WHERE id = $after.event_id)[0];
        LET $parent_level = IF $parent_event.visibility = "private" THEN 0
            ELSE IF $parent_event.visibility = "invite_only" THEN 1
            ELSE IF $parent_event.visibility = "guilds" THEN 2
            ELSE 3 END;

        IF $child_level > $parent_level {
            THROW "Rideshare visibility cannot exceed parent event visibility";
        };
    };

    -- Check adventure parent if no event
    IF $after.event_id = NONE AND $after.adventure_id != NONE {
        LET $parent_adv = (SELECT visibility FROM adventure WHERE id = $after.adventure_id)[0];
        LET $parent_level = IF $parent_adv.visibility = "private" THEN 0
            ELSE IF $parent_adv.visibility = "invite_only" THEN 1
            ELSE IF $parent_adv.visibility = "guilds" THEN 2
            ELSE 3 END;

        IF $child_level > $parent_level {
            THROW "Rideshare visibility cannot exceed parent adventure visibility";
        };
    };
};

-- =============================================================================
-- 3. RESONANCE SCORE FLOOR
-- =============================================================================
-- Issue: Resonance scores can go negative if penalties exceed rewards.
-- This looks bad in UI and violates expected scoring behavior.
--
-- Solution: Recreate the trigger with MAX(0, value) floor protection.
-- We remove and redefine to ensure clean trigger definition.

REMOVE EVENT resonance_score_update ON TABLE resonance_ledger;

DEFINE EVENT resonance_score_update ON TABLE resonance_ledger WHEN $event = "CREATE" THEN {
    -- Get existing score or create new one
    LET $existing = SELECT * FROM resonance_score WHERE user = $after.user;
    -- Helper: floor at 0 using math::max with array
    LET $pts = IF $after.points < 0 THEN 0 ELSE $after.points END;

    IF array::len($existing) = 0 {
        -- Create new score record (new scores start at 0 floor)
        CREATE resonance_score SET
            user = $after.user,
            total = $pts,
            questing = IF $after.stat = "questing" THEN $pts ELSE 0 END,
            mana = IF $after.stat = "mana" THEN $pts ELSE 0 END,
            wayfinder = IF $after.stat = "wayfinder" THEN $pts ELSE 0 END,
            attunement = IF $after.stat = "attunement" THEN $pts ELSE 0 END,
            nexus = IF $after.stat = "nexus" THEN $pts ELSE 0 END,
            last_calculated = time::now();
    } ELSE {
        -- Update existing score with floor protection
        LET $current = $existing[0];

        UPDATE resonance_score SET
            total = IF (total + $after.points) < 0 THEN 0 ELSE (total + $after.points) END,
            questing = IF $after.stat = "questing" THEN (IF (questing + $after.points) < 0 THEN 0 ELSE (questing + $after.points) END) ELSE questing END,
            mana = IF $after.stat = "mana" THEN (IF (mana + $after.points) < 0 THEN 0 ELSE (mana + $after.points) END) ELSE mana END,
            wayfinder = IF $after.stat = "wayfinder" THEN (IF (wayfinder + $after.points) < 0 THEN 0 ELSE (wayfinder + $after.points) END) ELSE wayfinder END,
            attunement = IF $after.stat = "attunement" THEN (IF (attunement + $after.points) < 0 THEN 0 ELSE (attunement + $after.points) END) ELSE attunement END,
            nexus = IF $after.stat = "nexus" THEN (IF (nexus + $after.points) < 0 THEN 0 ELSE (nexus + $after.points) END) ELSE nexus END,
            last_calculated = time::now()
        WHERE user = $after.user;
    };
};

-- =============================================================================
-- 4. DENORMALIZED COUNTS FOR PERFORMANCE
-- =============================================================================
-- Add member_count to guild for efficient list queries

DEFINE FIELD OVERWRITE member_count ON guild TYPE int DEFAULT 0;

-- Trigger to maintain guild member count
DEFINE EVENT OVERWRITE update_guild_member_count ON TABLE responsible_for
    WHEN $event = "CREATE" OR $event = "DELETE" THEN {

    LET $guild_id = IF $event = "CREATE" THEN $after.out ELSE $before.out END;

    LET $count = (SELECT count() FROM responsible_for WHERE out = $guild_id GROUP ALL)[0].count;

    UPDATE $guild_id SET member_count = $count ?? 0;
};

-- =============================================================================
-- NOTES FOR FUTURE MIGRATIONS
-- =============================================================================
--
-- The following issues are documented but NOT fixed in this migration
-- as they require data migration verification first:
--
-- 1. Legacy RSVP tables (rsvp, event_participant, event_rsvp)
--    - Should be dropped after verifying all data migrated to unified_rsvp
--
-- 2. Cascade delete rules
--    - Relations don't specify ON DELETE behavior
--    - Adventure/event deletion could orphan child records
--    - Requires careful analysis of deletion semantics
--
-- 3. Limit-checking trigger optimization
--    - 30+ triggers use SELECT count() on every CREATE
--    - Could use denormalized counts instead
--    - Lower priority, only affects bulk operations
--
-- 4. Forum post limit race condition
--    - Trigger runs after insert, throws if >1000
--    - Should be checked at application layer before insert
