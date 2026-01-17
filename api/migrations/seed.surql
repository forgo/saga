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
CREATE user:demo SET
    email = "demo@forgo.software",
    username = "demo",
    hash = "$2a$10$Uu16ncWz8w9YXs.wQUajr.qSfxTK2xt0RSZqL6F7M/VpQdnKMGp6m",
    firstname = "Demo",
    lastname = "User",
    email_verified = true,
    created_on = time::now(),
    updated_on = time::now();

-- ============================================================================
-- DEMO CIRCLE
-- ============================================================================

CREATE member:demo SET
    name = "Demo User",
    email = "demo@forgo.software",
    user = user:demo,
    created_on = time::now(),
    updated_on = time::now();

CREATE circle:demo SET
    name = "Close Friends",
    description = "My closest friends",
    icon = "person.3.fill",
    color = "#6B46C1",
    created_on = time::now(),
    updated_on = time::now();

RELATE member:demo->responsible_for->circle:demo SET
    pending_approval = false;

-- ============================================================================
-- ACTIVITIES (general-purpose interaction types)
-- ============================================================================

CREATE activity:hungout SET
    name = "Hung out",
    icon = "person.2.fill",
    warn = 1209600,    -- 14 days in seconds
    critical = 2592000, -- 30 days in seconds
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:hungout->managed_by->circle:demo;

CREATE activity:called SET
    name = "Called",
    icon = "phone.fill",
    warn = 604800,     -- 7 days
    critical = 1209600, -- 14 days
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:called->managed_by->circle:demo;

CREATE activity:texted SET
    name = "Texted",
    icon = "message.fill",
    warn = 259200,     -- 3 days
    critical = 604800,  -- 7 days
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:texted->managed_by->circle:demo;

CREATE activity:games SET
    name = "Played games",
    icon = "gamecontroller.fill",
    warn = 604800,     -- 7 days
    critical = 1209600, -- 14 days
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:games->managed_by->circle:demo;

CREATE activity:coffee SET
    name = "Coffee",
    icon = "cup.and.saucer.fill",
    warn = 604800,     -- 7 days
    critical = 1209600, -- 14 days
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:coffee->managed_by->circle:demo;

-- ============================================================================
-- PEOPLE (friends you want to stay in touch with)
-- ============================================================================

CREATE person:alex SET
    name = "Alex Chen",
    nickname = "Alex",
    birthday = d"1990-03-15",
    notes = "Met at college, loves hiking",
    created_on = time::now(),
    updated_on = time::now();
RELATE person:alex->part_of->circle:demo;

CREATE person:sam SET
    name = "Sam Rivera",
    nickname = "Sam",
    birthday = d"1992-08-22",
    notes = "Coworker at previous job, into board games",
    created_on = time::now(),
    updated_on = time::now();
RELATE person:sam->part_of->circle:demo;

CREATE person:jordan SET
    name = "Jordan Lee",
    notes = "Gym buddy",
    created_on = time::now(),
    updated_on = time::now();
RELATE person:jordan->part_of->circle:demo;

-- ============================================================================
-- TIMERS FOR ALEX (varied states for testing)
-- ============================================================================

-- Alex: Hung out - normal state (last met 3 days ago)
CREATE timer:alex_hungout SET
    reset_date = time::now() - 3d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:alex->tracked_by->timer:alex_hungout;
RELATE timer:alex_hungout->adheres_to->activity:hungout;

-- Alex: Texted - warning state (5 days ago)
CREATE timer:alex_texted SET
    reset_date = time::now() - 5d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:alex->tracked_by->timer:alex_texted;
RELATE timer:alex_texted->adheres_to->activity:texted;

-- Alex: Called - critical state (3 weeks ago)
CREATE timer:alex_called SET
    reset_date = time::now() - 21d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:alex->tracked_by->timer:alex_called;
RELATE timer:alex_called->adheres_to->activity:called;

-- ============================================================================
-- TIMERS FOR SAM (all normal states)
-- ============================================================================

-- Sam: Games - recently played
CREATE timer:sam_games SET
    reset_date = time::now() - 2d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:sam->tracked_by->timer:sam_games;
RELATE timer:sam_games->adheres_to->activity:games;

-- Sam: Texted - recent
CREATE timer:sam_texted SET
    reset_date = time::now() - 1d,
    enabled = true,
    push = false,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:sam->tracked_by->timer:sam_texted;
RELATE timer:sam_texted->adheres_to->activity:texted;

-- ============================================================================
-- TIMERS FOR JORDAN
-- ============================================================================

-- Jordan: Hung out - warning state (18 days ago)
CREATE timer:jordan_hungout SET
    reset_date = time::now() - 18d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:jordan->tracked_by->timer:jordan_hungout;
RELATE timer:jordan_hungout->adheres_to->activity:hungout;

-- ============================================================================
-- SECOND USER (for multi-user/circle testing)
-- ============================================================================

CREATE user:second SET
    email = "second@forgo.software",
    username = "second",
    hash = "$2a$10$Uu16ncWz8w9YXs.wQUajr.qSfxTK2xt0RSZqL6F7M/VpQdnKMGp6m",
    firstname = "Second",
    lastname = "User",
    email_verified = true,
    created_on = time::now(),
    updated_on = time::now();

CREATE member:second SET
    name = "Second User",
    email = "second@forgo.software",
    user = user:second,
    created_on = time::now(),
    updated_on = time::now();

CREATE circle:second SET
    name = "Work Buddies",
    description = "Former colleagues",
    icon = "briefcase.fill",
    color = "#38B2AC",
    created_on = time::now(),
    updated_on = time::now();

RELATE member:second->responsible_for->circle:second SET
    pending_approval = false;

-- Activities for second circle
CREATE activity:hungout2 SET
    name = "Hung out",
    icon = "person.2.fill",
    warn = 1209600,
    critical = 2592000,
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:hungout2->managed_by->circle:second;

CREATE activity:called2 SET
    name = "Called",
    icon = "phone.fill",
    warn = 604800,
    critical = 1209600,
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:called2->managed_by->circle:second;

CREATE activity:lunch2 SET
    name = "Lunch",
    icon = "fork.knife",
    warn = 1209600,
    critical = 2592000,
    created_on = time::now(),
    updated_on = time::now();
RELATE activity:lunch2->managed_by->circle:second;

-- Person for second circle
CREATE person:taylor SET
    name = "Taylor Martinez",
    nickname = "Tay",
    notes = "Former manager, great mentor",
    created_on = time::now(),
    updated_on = time::now();
RELATE person:taylor->part_of->circle:second;

-- Timers for Taylor
CREATE timer:taylor_hungout SET
    reset_date = time::now() - 10d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:taylor->tracked_by->timer:taylor_hungout;
RELATE timer:taylor_hungout->adheres_to->activity:hungout2;

CREATE timer:taylor_called SET
    reset_date = time::now() - 5d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:taylor->tracked_by->timer:taylor_called;
RELATE timer:taylor_called->adheres_to->activity:called2;

CREATE timer:taylor_lunch SET
    reset_date = time::now() - 25d,
    enabled = true,
    push = true,
    created_on = time::now(),
    updated_on = time::now();
RELATE person:taylor->tracked_by->timer:taylor_lunch;
RELATE timer:taylor_lunch->adheres_to->activity:lunch2;

-- ============================================================================
-- PENDING INVITE (for testing circle join flow)
-- ============================================================================

CREATE user:pending SET
    email = "pending@forgo.software",
    username = "pending",
    hash = "$2a$10$Uu16ncWz8w9YXs.wQUajr.qSfxTK2xt0RSZqL6F7M/VpQdnKMGp6m",
    firstname = "Pending",
    lastname = "User",
    email_verified = true,
    created_on = time::now(),
    updated_on = time::now();

CREATE member:pending SET
    name = "Pending User",
    email = "pending@forgo.software",
    user = user:pending,
    created_on = time::now(),
    updated_on = time::now();

-- Pending invite to Demo Circle (for testing approval flow)
RELATE member:pending->responsible_for->circle:demo SET
    pending_approval = true;

-- ============================================================================
-- SAMPLE EVENT (for testing events feature)
-- ============================================================================

CREATE event:gamenight SET
    circle_id = circle:demo,
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
RELATE member:demo->rsvp->event:gamenight SET
    status = "yes",
    responded_on = time::now();
