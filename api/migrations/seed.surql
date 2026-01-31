-- Saga Development Seed Data
-- WARNING: This is for development only. Never run in production.
--
-- Demo credentials:
--   Email:    demo@forgo.software
--   Password: password123
--
-- Second user:
--   Email:    second@forgo.software
--   Password: password123

-- ============================================================================
-- DEMO USER (Primary test account)
-- ============================================================================

-- Password: password123 (bcrypt hash)
UPSERT user:demo SET
    email = "demo@forgo.software",
    username = "demo",
    hash = "$2a$10$Uu16ncWz8w9YXs.wQUajr.qSfxTK2xt0RSZqL6F7M/VpQdnKMGp6m",
    firstname = "Demo",
    lastname = "User",
    email_verified = true,
    created_on = time::now(),
    updated_on = time::now();

-- ============================================================================
-- DEMO GUILD
-- ============================================================================

UPSERT member:demo SET
    name = "Demo User",
    email = "demo@forgo.software",
    user = user:demo,
    created_on = time::now(),
    updated_on = time::now();

UPSERT guild:demo SET
    name = "Close Friends",
    description = "My closest friends",
    icon = "person.3.fill",
    color = "#6B46C1",
    visibility = "private",
    created_on = time::now(),
    updated_on = time::now();

RELATE member:demo->responsible_for:rf_demo_demo->guild:demo SET
    pending_approval = false,
    role = "admin";

-- ============================================================================
-- ACTIVITIES (general-purpose interaction types)
-- ============================================================================

UPSERT activity:hungout SET
    name = "Hung out",
    icon = "person.2.fill",
    warn = 1209600,    -- 14 days in seconds
    critical = 2592000, -- 30 days in seconds
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:hungout->managed_by:mb_hungout_demo->guild:demo;

UPSERT activity:called SET
    name = "Called",
    icon = "phone.fill",
    warn = 604800,     -- 7 days
    critical = 1209600, -- 14 days
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:called->managed_by:mb_called_demo->guild:demo;

UPSERT activity:texted SET
    name = "Texted",
    icon = "message.fill",
    warn = 259200,     -- 3 days
    critical = 604800,  -- 7 days
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:texted->managed_by:mb_texted_demo->guild:demo;

UPSERT activity:games SET
    name = "Played games",
    icon = "gamecontroller.fill",
    warn = 604800,     -- 7 days
    critical = 1209600, -- 14 days
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:games->managed_by:mb_games_demo->guild:demo;

UPSERT activity:coffee SET
    name = "Coffee",
    icon = "cup.and.saucer.fill",
    warn = 604800,     -- 7 days
    critical = 1209600, -- 14 days
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:coffee->managed_by:mb_coffee_demo->guild:demo;

-- ============================================================================
-- PEOPLE (friends you want to stay in touch with)
-- ============================================================================

UPSERT person:alex SET
    name = "Alex Chen",
    nickname = "Alex",
    birthday = d"1990-03-15",
    notes = "Met at college, loves hiking",
    created_on = time::now(),
    updated_on = time::now();
RELATE person:alex->part_of:po_alex_demo->guild:demo;

UPSERT person:sam SET
    name = "Sam Rivera",
    nickname = "Sam",
    birthday = d"1992-08-22",
    notes = "Coworker at previous job, into board games",
    created_on = time::now(),
    updated_on = time::now();
RELATE person:sam->part_of:po_sam_demo->guild:demo;

UPSERT person:jordan SET
    name = "Jordan Lee",
    notes = "Gym buddy",
    created_on = time::now(),
    updated_on = time::now();
RELATE person:jordan->part_of:po_jordan_demo->guild:demo;

-- ============================================================================
-- TIMERS FOR ALEX (varied states for testing)
-- ============================================================================

-- Alex: Hung out - normal state (last met 3 days ago)
UPSERT timer:alex_hungout SET
    reset_date = time::now() - 3d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:alex->tracked_by:tb_alex_hungout->timer:alex_hungout;
RELATE timer:alex_hungout->adheres_to:at_alex_hungout->activity:hungout;

-- Alex: Texted - warning state (5 days ago)
UPSERT timer:alex_texted SET
    reset_date = time::now() - 5d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:alex->tracked_by:tb_alex_texted->timer:alex_texted;
RELATE timer:alex_texted->adheres_to:at_alex_texted->activity:texted;

-- Alex: Called - critical state (3 weeks ago)
UPSERT timer:alex_called SET
    reset_date = time::now() - 21d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:alex->tracked_by:tb_alex_called->timer:alex_called;
RELATE timer:alex_called->adheres_to:at_alex_called->activity:called;

-- ============================================================================
-- TIMERS FOR SAM (all normal states)
-- ============================================================================

-- Sam: Games - recently played
UPSERT timer:sam_games SET
    reset_date = time::now() - 2d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:sam->tracked_by:tb_sam_games->timer:sam_games;
RELATE timer:sam_games->adheres_to:at_sam_games->activity:games;

-- Sam: Texted - recent
UPSERT timer:sam_texted SET
    reset_date = time::now() - 1d,
    enabled = true,
    push = false,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:sam->tracked_by:tb_sam_texted->timer:sam_texted;
RELATE timer:sam_texted->adheres_to:at_sam_texted->activity:texted;

-- ============================================================================
-- TIMERS FOR JORDAN
-- ============================================================================

-- Jordan: Hung out - warning state (18 days ago)
UPSERT timer:jordan_hungout SET
    reset_date = time::now() - 18d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:jordan->tracked_by:tb_jordan_hungout->timer:jordan_hungout;
RELATE timer:jordan_hungout->adheres_to:at_jordan_hungout->activity:hungout;

-- ============================================================================
-- SECOND USER (for multi-user/guild testing)
-- ============================================================================

UPSERT user:second SET
    email = "second@forgo.software",
    username = "second",
    hash = "$2a$10$Uu16ncWz8w9YXs.wQUajr.qSfxTK2xt0RSZqL6F7M/VpQdnKMGp6m",
    firstname = "Second",
    lastname = "User",
    email_verified = true,
    created_on = time::now(),
    updated_on = time::now();

UPSERT member:second SET
    name = "Second User",
    email = "second@forgo.software",
    user = user:second,
    created_on = time::now(),
    updated_on = time::now();

UPSERT guild:second SET
    name = "Work Buddies",
    description = "Former colleagues",
    icon = "briefcase.fill",
    color = "#38B2AC",
    created_on = time::now(),
    updated_on = time::now();

RELATE member:second->responsible_for:rf_second_second->guild:second SET
    pending_approval = false,
    role = "admin";

-- Activities for second guild
UPSERT activity:hungout2 SET
    name = "Hung out",
    icon = "person.2.fill",
    warn = 1209600,
    critical = 2592000,
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:hungout2->managed_by:mb_hungout2_second->guild:second;

UPSERT activity:called2 SET
    name = "Called",
    icon = "phone.fill",
    warn = 604800,
    critical = 1209600,
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:called2->managed_by:mb_called2_second->guild:second;

UPSERT activity:lunch2 SET
    name = "Lunch",
    icon = "fork.knife",
    warn = 1209600,
    critical = 2592000,
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:lunch2->managed_by:mb_lunch2_second->guild:second;

-- Person for second guild
UPSERT person:taylor SET
    name = "Taylor Martinez",
    nickname = "Tay",
    notes = "Former manager, great mentor",
    created_on = time::now(),
    updated_on = time::now();
RELATE person:taylor->part_of:po_taylor_second->guild:second;

-- Timers for Taylor
UPSERT timer:taylor_hungout SET
    reset_date = time::now() - 10d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:taylor->tracked_by:tb_taylor_hungout->timer:taylor_hungout;
RELATE timer:taylor_hungout->adheres_to:at_taylor_hungout->activity:hungout2;

UPSERT timer:taylor_called SET
    reset_date = time::now() - 5d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:taylor->tracked_by:tb_taylor_called->timer:taylor_called;
RELATE timer:taylor_called->adheres_to:at_taylor_called->activity:called2;

UPSERT timer:taylor_lunch SET
    reset_date = time::now() - 25d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:taylor->tracked_by:tb_taylor_lunch->timer:taylor_lunch;
RELATE timer:taylor_lunch->adheres_to:at_taylor_lunch->activity:lunch2;

-- ============================================================================
-- PENDING INVITE (for testing guild join flow)
-- ============================================================================

UPSERT user:pending SET
    email = "pending@forgo.software",
    username = "pending",
    hash = "$2a$10$Uu16ncWz8w9YXs.wQUajr.qSfxTK2xt0RSZqL6F7M/VpQdnKMGp6m",
    firstname = "Pending",
    lastname = "User",
    email_verified = true,
    created_on = time::now(),
    updated_on = time::now();

UPSERT member:pending SET
    name = "Pending User",
    email = "pending@forgo.software",
    user = user:pending,
    created_on = time::now(),
    updated_on = time::now();

-- Pending invite to Demo Circle (for testing approval flow)
RELATE member:pending->responsible_for:rf_pending_demo->guild:demo SET
    pending_approval = true;

-- ============================================================================
-- SAMPLE EVENT (for testing events feature)
-- ============================================================================

UPSERT event:gamenight SET
    guild_id = guild:demo,
    title = "Board Game Night",
    description = "Monthly board game meetup at Alex's place. Bring snacks!",
    location = "123 Main St, Apt 4B",
    starts_at = time::now() + 7d,
    ends_at = time::now() + 7d + 4h,
    all_day = false,
    created_by = member:demo,
    created_on = time::now(),
    updated_on = time::now();

-- Demo user's RSVP
RELATE member:demo->rsvp:rsvp_demo_gamenight->event:gamenight SET
    status = "yes",
    responded_on = time::now();
