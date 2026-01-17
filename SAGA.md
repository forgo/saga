Saga - Comprehensive Social Coordination Platform

Vision

Transform into a full social coordination platform combining:

- Couchsurfing: Availability pools, host/guest reviews
- Meetup: Interest-based events, geographic discovery
- Internations: Cross-cultural communities, learning/teaching
- OkCupid: Questionnaire-based compatibility matching
- Donut: Automated random matching for forced socialization
- Partiful: Beautiful themed events, minimal friction

Plus new ideas: Travel coordination, itinerary voting, rideshare matching, map integration.

---

Hierarchical Entity Architecture (NEW)

Core Insight

The platform has three rigors of social engagement, each with increasing commitment:

```
 ┌─────────────────────────────────────────────────────────────────────────────┐
 │                           ENGAGEMENT HIERARCHY                              │
 ├─────────────────────────────────────────────────────────────────────────────┤
 │                                                                             │
 │  SPONTANEOUS         EVENT                   ADVENTURE                      │
 │  ─────────────       ─────                   ─────────                      │
 │  Lowest barrier      Medium barrier          Highest barrier                │
 │  Matching-based      Host-controlled         Multi-component                │
 │  Ephemeral           Single location         Multiple locations             │
 │  "Right now"         Concrete time           Multi-day span                 │
 │                                                                             │
 │  ┌───────────┐       ┌───────────┐           ┌────────────────────────────┐ │
 │  │ Hangout   │       │   Event   │           │        Adventure           │ │
 │  │ Request   │       │           │           │  ┌───────┐  ┌───────┐      │ │
 │  └───────────┘       │ ┌───────┐ │           │  │Event 1│  │Event 2│ ...  │ │
 │                      │ │Rideshr│ │           │  │┌─────┐│  │┌─────┐│      │ │
 │                      │ │ ...   │ │           │  ││Rides││  ││Rides││      │ │
 │                      │ └───────┘ │           │  │└─────┘│  │└─────┘│      │ │
 │                      │ ┌──────┐  │           │  └───────┘  └───────┘      │ │
 │                      │ │Forum │  │           │  ┌──────────────────┐      │ │
 │                      │ └──────┘  │           │  │ Inter-event      │      │ │
 │                      └───────────┘           │  │ Rideshares       │      │ │
 │                                              │  └──────────────────┘      │ │
 │                                              │  ┌──────┐                  │ │
 │                                              │  │Forum │ (master)         │ │
 │                                              │  └──────┘                  │ │
 │                                              └────────────────────────────┘ │
 │                                                                             │
 └─────────────────────────────────────────────────────────────────────────────┘
```

Terminology Changes

| Old Term    | New Term  | Reason                                                |
| ----------- | --------- | ----------------------------------------------------- |
| Trip        | Adventure | More evocative, implies multi-component journey       |
| Commute     | Rideshare | Clearer purpose, attached to Events/Adventures        |
| Circle      | Guild     | More evocative, implies community with shared purpose |
| Association | Alliance  | Stronger partnership connotation between Guilds       |

Entity Composition

Guild (formerly Circle):

- Community with shared purpose
- Contains members, events, adventures
- Has moderation settings and values
- Can form Alliances with other Guilds

Adventure (formerly Trip):

- Multi-location, multi-day coordination
- Contains zero or more Events
- Contains zero or more Rideshares (inter-event transport)
- Has a master Forum for discussion
- Users RSVP to Adventure first, then opt into individual components
- Host sets RSVP requirements (values matching, questions)

Event (enhanced):

- Single location, concrete time, focused activity
- Can exist standalone OR within an Adventure
- Contains zero or more Rideshares (event-specific transport)
- Has its own Forum for discussion
- Host controls RSVP requirements
- If part of Adventure: inherits visibility from parent

Rideshare (formerly Commute):

- Transportation coordination between locations
- Must be attached to either an Event OR an Adventure (never standalone)
- Inherits visibility from parent
- Driver/passenger coordination with seat booking

Spontaneous (existing Availability/Hangout system):

- No hierarchy, no forums, no rideshares
- Matching-based discovery
- Ephemeral, low commitment

Alliance (formerly Association):

- Partnership between two Guilds
- Enables cross-guild adventures and discovery
- Bidirectional approval required

Security Model (Cascading Visibility)

Adventure.visibility → Event.visibility → Rideshare.visibility

Rules:

1.  Child cannot be MORE visible than parent
2.  Non-participants cannot see child entities
3.  RSVP to Adventure grants access to see (not join) its Events
4.  RSVP to Event grants access to see its Rideshares

Visibility Levels:
public - Anyone can discover
guilds - Only members of associated guilds
invite_only - Only explicitly invited users
private - Only organizers (draft mode)

Access Enforcement:
// Pseudo-code for visibility checking
func CanViewEvent(user, event) bool {
if event.AdventureID != nil {
adventure := GetAdventure(event.AdventureID)
if !IsAdventureParticipant(user, adventure) {
return false
}
}
return CheckVisibility(user, event)
}

func CanViewRideshare(user, rideshare) bool {
if rideshare.EventID != nil {
if !CanViewEvent(user, GetEvent(rideshare.EventID)) {
return false
}
} else if rideshare.AdventureID != nil {
if !IsAdventureParticipant(user, GetAdventure(rideshare.AdventureID)) {
return false
}
}
return true
}

Guild Alliances

Guilds can form alliances (partnerships) to:

- Share events across guilds
- Coordinate cross-guild adventures
- Enable member discovery across allied guilds

```
 ┌─────────────┐      alliance      ┌─────────────┐
 │   Guild A   │◄──────────────────►│   Guild B   │
 │ (SF Hikers) │                    │ (LA Hikers) │
 └─────────────┘                    └─────────────┘
        │                                   │
        └───────── Joint Adventure ─────────┘
```

Forums

Forums provide threaded discussion for:

- Adventures: Master forum for all participants
- Events: Focused discussion for event attendees

Forum features:

- Threaded replies
- Pinned posts (by organizers)
- @mentions
- Reactions
- Visibility inherits from parent entity

Discovery Patterns

Users can discover engagement opportunities by:

1.  Location: "What's happening in Amsterdam?"
2.  Time: "What's happening this weekend?"
3.  Type: "Show me Adventures" / "Show me Events" / "Spontaneous now"
4.  Interest: "Hiking events near me"
5.  Guilds: "Public events from my guilds and alliances"

API Views

The API should support three primary view modes:

Timeline View:

- Chronological feed of all engagement types
- Filters: type, guild, time range
- Groups by: today, tomorrow, this week, later

Calendar View:

- Month/week/day layouts
- Events and Adventures plotted by date
- Color-coded by guild or type
- iCal export

Map View:

- Geographic pins for events/adventures
- Cluster nearby pins
- Filter by date range, type, guild
- "Near me" mode

Dashboard Organization

```
 ┌───────────────────────────────────────────────────────────┐
 │  HOME DASHBOARD                                           │
 ├───────────────────────────────────────────────────────────┤
 │                                                           │
 │  [🔥 Spontaneous]  [📅 Events]  [🗺️ Adventures]           │
 │                                                           │
 │  ┌──────────────────────────────────────────────────────┐ │
 │  │ Happening Now                                        │ │
 │  │ • 3 people available nearby                          │ │
 │  │ • Board Game Night starting in 30 min                │ │
 │  └──────────────────────────────────────────────────────┘ │
 │                                                           │
 │  ┌──────────────────────────────────────────────────────┐ │
 │  │ This Week                                            │ │
 │  │ • Friday: Hiking Meetup (Event)                      │ │
 │  │ • Saturday-Sunday: Beach Trip (Adventure)            │ │
 │  └──────────────────────────────────────────────────────┘ │
 │                                                           │
 │  ┌──────────────────────────────────────────────────────┐ │
 │  │ Coming Up                                            │ │
 │  │ • Jan 15-20: Amsterdam Trip (Adventure) - 12 going   │ │
 │  │ • Jan 22: Wine Tasting (Event) - 8 spots left        │ │
 │  └──────────────────────────────────────────────────────┘ │
 │                                                           │
 └───────────────────────────────────────────────────────────┘
```

---

Anti-Patterns (HARD REQUIREMENTS)

- ❌ NO subscriptions, tiers, microtransactions
- ❌ NO ads, data selling, enshittification
- ✅ One-time purchase ($9.99) + optional donation
- ✅ Sparse UI with powerful backend
- ✅ Privacy-first, user controls everything

---

Data Model Extensions

Core New Tables (SurrealDB)

-- 1. USER PROFILE EXTENSION
DEFINE TABLE user_profile SCHEMAFULL;
DEFINE FIELD user ON user_profile TYPE record<user>;
DEFINE FIELD bio ON user_profile TYPE option<string>;
DEFINE FIELD tagline ON user_profile TYPE option<string>;
DEFINE FIELD languages ON user_profile TYPE option<array<string>>;
DEFINE FIELD timezone ON user_profile TYPE option<string>;
DEFINE FIELD location ON user_profile TYPE option<object>; -- {lat, lng, city, country}
DEFINE FIELD visibility ON user_profile TYPE string DEFAULT "circles";

-- 2. INTEREST SYSTEM
DEFINE TABLE interest SCHEMAFULL;
DEFINE FIELD name ON interest TYPE string;
DEFINE FIELD category ON interest TYPE string; -- hobby, skill, language, cuisine, etc.
DEFINE FIELD icon ON interest TYPE option<string>;

DEFINE TABLE has_interest SCHEMAFULL TYPE RELATION FROM user TO interest;
DEFINE FIELD level ON has_interest TYPE string; -- curious, interested, experienced, expert
DEFINE FIELD wants_to_teach ON has_interest TYPE bool DEFAULT false;
DEFINE FIELD wants_to_learn ON has_interest TYPE bool DEFAULT false;

-- 3. AVAILABILITY SYSTEM
DEFINE TABLE availability SCHEMAFULL;
DEFINE FIELD user ON availability TYPE record<user>;
DEFINE FIELD status ON availability TYPE string; -- available, maybe, busy
DEFINE FIELD start_time ON availability TYPE datetime;
DEFINE FIELD end_time ON availability TYPE datetime;
DEFINE FIELD location ON availability TYPE option<object>; -- {lat, lng, radius_km}
DEFINE FIELD activity_types ON availability TYPE option<array<string>>;
DEFINE FIELD max_people ON availability TYPE int DEFAULT 1;
DEFINE FIELD note ON availability TYPE option<string>;

-- 4. QUESTIONNAIRE SYSTEM
DEFINE TABLE question SCHEMAFULL;
DEFINE FIELD text ON question TYPE string;
DEFINE FIELD category ON question TYPE string; -- values, lifestyle, social, etc.
DEFINE FIELD options ON question TYPE array<object>;
DEFINE FIELD is_dealbreaker_eligible ON question TYPE bool DEFAULT false;

DEFINE TABLE answer SCHEMAFULL;
DEFINE FIELD user ON answer TYPE record<user>;
DEFINE FIELD question ON answer TYPE record<question>;
DEFINE FIELD selected_option ON answer TYPE string;
DEFINE FIELD acceptable_options ON answer TYPE array<string>;
DEFINE FIELD importance ON answer TYPE string; -- irrelevant to mandatory
DEFINE FIELD is_dealbreaker ON answer TYPE bool DEFAULT false;

-- 5. ENHANCED EVENTS (Partiful-style)
-- Extends existing event table
DEFINE FIELD template ON event TYPE option<string>; -- casual, dinner_party, birthday, etc.
DEFINE FIELD visibility ON event TYPE string DEFAULT "guild"; -- public, guild, invite_only
DEFINE FIELD capacity ON event TYPE option<int>;
DEFINE FIELD waitlist_enabled ON event TYPE bool DEFAULT false;
DEFINE FIELD cost ON event TYPE option<object>;
DEFINE FIELD cover_image ON event TYPE option<string>;
DEFINE FIELD theme_color ON event TYPE option<string>;

-- 6. MATCHING POOLS (Donut-style)
DEFINE TABLE matching_pool SCHEMAFULL;
DEFINE FIELD guild_id ON matching_pool TYPE option<record<guild>>;
DEFINE FIELD name ON matching_pool TYPE string;
DEFINE FIELD frequency ON matching_pool TYPE string; -- weekly, biweekly, monthly
DEFINE FIELD match_size ON matching_pool TYPE int DEFAULT 2;
DEFINE FIELD activity_suggestion ON matching_pool TYPE option<string>;
DEFINE FIELD next_match_on ON matching_pool TYPE datetime;

DEFINE TABLE pool_member SCHEMAFULL TYPE RELATION FROM member TO matching_pool;
DEFINE FIELD excluded_members ON pool_member TYPE option<array<record<member>>>;

DEFINE TABLE match_result SCHEMAFULL;
DEFINE FIELD pool_id ON match_result TYPE record<matching_pool>;
DEFINE FIELD members ON match_result TYPE array<record<member>>;
DEFINE FIELD status ON match_result TYPE string; -- pending, scheduled, completed, skipped

-- 7. ADVENTURES (formerly Trip)
DEFINE TABLE adventure SCHEMAFULL;
DEFINE FIELD guild_id ON adventure TYPE option<record<guild>>;
DEFINE FIELD title ON adventure TYPE string;
DEFINE FIELD description ON adventure TYPE option<string>;
DEFINE FIELD start_date ON adventure TYPE datetime;
DEFINE FIELD end_date ON adventure TYPE datetime;
DEFINE FIELD status ON adventure TYPE string DEFAULT "idea"; -- idea, planning, confirmed, active, completed, cancelled
DEFINE FIELD visibility ON adventure TYPE string DEFAULT "guilds"; -- public, guilds, invite_only, private
DEFINE FIELD created_by_id ON adventure TYPE record<user>;
DEFINE FIELD values_required ON adventure TYPE bool DEFAULT false;
DEFINE FIELD values_questions ON adventure TYPE option<array<string>>;
DEFINE FIELD budget_min ON adventure TYPE option<float>;
DEFINE FIELD budget_max ON adventure TYPE option<float>;
DEFINE FIELD currency ON adventure TYPE option<string>;
DEFINE FIELD cover_image ON adventure TYPE option<string>;
DEFINE FIELD created_on ON adventure TYPE datetime;
DEFINE FIELD updated_on ON adventure TYPE datetime;

DEFINE TABLE adventure_participant SCHEMAFULL;
DEFINE FIELD adventure_id ON adventure_participant TYPE record<adventure>;
DEFINE FIELD user_id ON adventure_participant TYPE record<user>;
DEFINE FIELD role ON adventure_participant TYPE string DEFAULT "participant"; -- organizer, participant
DEFINE FIELD status ON adventure_participant TYPE string DEFAULT "interested"; -- interested, committed, maybe, out
DEFINE FIELD joined_on ON adventure_participant TYPE datetime;

DEFINE TABLE destination SCHEMAFULL;
DEFINE FIELD adventure_id ON destination TYPE record<adventure>;
DEFINE FIELD name ON destination TYPE string;
DEFINE FIELD description ON destination TYPE option<string>;
DEFINE FIELD location ON destination TYPE object; -- {lat, lng, city, country}
DEFINE FIELD proposed_by ON destination TYPE record<user>;
DEFINE FIELD order_index ON destination TYPE int DEFAULT 0;
DEFINE FIELD created_on ON destination TYPE datetime;

DEFINE TABLE destination_vote SCHEMAFULL;
DEFINE FIELD destination_id ON destination_vote TYPE record<destination>;
DEFINE FIELD user_id ON destination_vote TYPE record<user>;
DEFINE FIELD rank ON destination_vote TYPE int;
DEFINE FIELD veto ON destination_vote TYPE bool DEFAULT false;
DEFINE FIELD reason ON destination_vote TYPE option<string>;

-- 8. EVENT HIERARCHY (Event links to Adventure)
-- Extend event table with adventure_id
DEFINE FIELD adventure_id ON event TYPE option<record<adventure>>;
DEFINE FIELD order_in_adventure ON event TYPE option<int>; -- sequence within adventure

-- 9. RIDESHARES (formerly Commute)
-- Must be attached to Event OR Adventure, never standalone
DEFINE TABLE rideshare SCHEMAFULL;
DEFINE FIELD event_id ON rideshare TYPE option<record<event>>;
DEFINE FIELD adventure_id ON rideshare TYPE option<record<adventure>>;
DEFINE FIELD driver_id ON rideshare TYPE record<user>;
DEFINE FIELD title ON rideshare TYPE string;
DEFINE FIELD description ON rideshare TYPE option<string>;
DEFINE FIELD origin ON rideshare TYPE object; -- {lat, lng, address, city}
DEFINE FIELD destination ON rideshare TYPE object;
DEFINE FIELD departure_time ON rideshare TYPE datetime;
DEFINE FIELD arrival_time ON rideshare TYPE option<datetime>;
DEFINE FIELD seats_total ON rideshare TYPE int DEFAULT 4;
DEFINE FIELD seats_available ON rideshare TYPE int DEFAULT 4;
DEFINE FIELD status ON rideshare TYPE string DEFAULT "open"; -- open, full, departed, completed, cancelled
DEFINE FIELD created_on ON rideshare TYPE datetime;
DEFINE FIELD updated_on ON rideshare TYPE datetime;
-- Constraint: event_id OR adventure_id must be set (enforced in business logic)

DEFINE TABLE rideshare_segment SCHEMAFULL;
DEFINE FIELD rideshare_id ON rideshare_segment TYPE record<rideshare>;
DEFINE FIELD sequence_order ON rideshare_segment TYPE int DEFAULT 0;
DEFINE FIELD pickup_point ON rideshare_segment TYPE object;
DEFINE FIELD dropoff_point ON rideshare_segment TYPE object;
DEFINE FIELD estimated_minutes ON rideshare_segment TYPE option<int>;
DEFINE FIELD notes ON rideshare_segment TYPE option<string>;

DEFINE TABLE rideshare_seat SCHEMAFULL;
DEFINE FIELD rideshare_id ON rideshare_seat TYPE record<rideshare>;
DEFINE FIELD passenger_id ON rideshare_seat TYPE record<user>;
DEFINE FIELD status ON rideshare_seat TYPE string DEFAULT "confirmed"; -- requested, confirmed, cancelled
DEFINE FIELD pickup_segment_id ON rideshare_seat TYPE option<record<rideshare_segment>>;
DEFINE FIELD dropoff_segment_id ON rideshare_seat TYPE option<record<rideshare_segment>>;
DEFINE FIELD requested_on ON rideshare_seat TYPE datetime;
DEFINE FIELD confirmed_on ON rideshare_seat TYPE option<datetime>;
DEFINE FIELD notes ON rideshare_seat TYPE option<string>;

-- 10. FORUMS (for Adventures and Events)
DEFINE TABLE forum SCHEMAFULL;
DEFINE FIELD adventure_id ON forum TYPE option<record<adventure>>;
DEFINE FIELD event_id ON forum TYPE option<record<event>>;
DEFINE FIELD created_on ON forum TYPE datetime;
-- Constraint: exactly one of adventure_id OR event_id must be set

DEFINE TABLE forum_post SCHEMAFULL;
DEFINE FIELD forum_id ON forum_post TYPE record<forum>;
DEFINE FIELD author_id ON forum_post TYPE record<user>;
DEFINE FIELD content ON forum_post TYPE string;
DEFINE FIELD reply_to_id ON forum_post TYPE option<record<forum_post>>; -- for threading
DEFINE FIELD is_pinned ON forum_post TYPE bool DEFAULT false;
DEFINE FIELD reactions ON forum_post TYPE option<object>; -- {emoji: [user_ids]}
DEFINE FIELD mentions ON forum_post TYPE option<array<record<user>>>;
DEFINE FIELD created_on ON forum_post TYPE datetime;
DEFINE FIELD updated_on ON forum_post TYPE datetime;
DEFINE FIELD deleted_on ON forum_post TYPE option<datetime>; -- soft delete

-- 11. GUILD ALLIANCES (formerly Circle Associations)
DEFINE TABLE guild_alliance SCHEMAFULL;
DEFINE FIELD guild_a_id ON guild_alliance TYPE record<guild>;
DEFINE FIELD guild_b_id ON guild_alliance TYPE record<guild>;
DEFINE FIELD status ON guild_alliance TYPE string DEFAULT "pending"; -- pending, active, revoked
DEFINE FIELD initiated_by_id ON guild_alliance TYPE record<user>;
DEFINE FIELD approved_by_id ON guild_alliance TYPE option<record<user>>;
DEFINE FIELD created_on ON guild_alliance TYPE datetime;
DEFINE FIELD approved_on ON guild_alliance TYPE option<datetime>;
DEFINE FIELD revoked_on ON guild_alliance TYPE option<datetime>;

-- 9. REVIEWS & REPUTATION
DEFINE TABLE review SCHEMAFULL TYPE RELATION FROM member TO member;
DEFINE FIELD context ON review TYPE string; -- hosted, was_guest, event, matched
DEFINE FIELD rating ON review TYPE int; -- 1-5
DEFINE FIELD text ON review TYPE option<string>;
DEFINE FIELD tags ON review TYPE option<array<string>>;
DEFINE FIELD would_meet_again ON review TYPE bool DEFAULT true;

---

Feature Modules (Prioritized)

Phase 1: Consolidated Schema (Design Phase)

All terminology changes consolidated into initial migration

| Feature             | Description                          | Files                               |
| ------------------- | ------------------------------------ | ----------------------------------- |
| Consolidated Schema | All naming changes in initial schema | migrations/001_initial_schema.surql |
| Guild Model         | Replace Circle with Guild            | internal/model/guild.go             |
| Adventure Model     | Replace Trip with Adventure          | internal/model/adventure.go         |
| Rideshare Model     | Replace Commute with Rideshare       | internal/model/rideshare.go         |
| Event Hierarchy     | Add adventure_id to Event            | internal/model/event.go             |
| Forum System        | Forums for Adventures/Events         | internal/model/forum.go             |
| Guild Alliances     | Link guilds together                 | internal/model/guild_alliance.go    |

Phase 2: Foundation

Extends existing infrastructure

| Feature          | Description                        | Files                                                       |
| ---------------- | ---------------------------------- | ----------------------------------------------------------- |
| User Profiles    | Bio, languages, timezone, location | internal/model/profile.go, internal/service/profile.go      |
| Interest System  | Tag interests, teaching/learning   | internal/model/interest.go, internal/repository/interest.go |
| Enhanced Events  | Templates, themes, capacity        | Extend internal/model/event.go, internal/handler/event.go   |
| Geographic Layer | Location storage, distance calcs   | internal/service/geo.go                                     |

Phase 3: Discovery & Spontaneous

Enable spontaneous coordination

| Feature               | Description                 | Files                                                            |
| --------------------- | --------------------------- | ---------------------------------------------------------------- |
| Availability Windows  | "I'm free now" signaling    | internal/model/availability.go, internal/service/availability.go |
| Timeline/Calendar/Map | Three view modes            | internal/handler/discover.go                                     |
| Location Discovery    | "What's in Amsterdam?"      | Extend internal/handler/discover.go                              |
| Interest Matching     | Connect by shared interests | internal/service/matching.go                                     |

Phase 4: Compatibility Matching

OkCupid-style questionnaires

| Feature            | Description                            | Files                                                              |
| ------------------ | -------------------------------------- | ------------------------------------------------------------------ |
| Question Engine    | Create/answer questions                | internal/model/questionnaire.go, internal/service/questionnaire.go |
| Compatibility Calc | Weighted scoring algorithm             | internal/service/compatibility.go                                  |
| Dealbreakers       | Hard incompatibility filters           | Part of compatibility service                                      |
| RSVP Gating        | Values questions for events/adventures | Extend event/adventure handlers                                    |

Phase 5: Automated Social

Donut-style forced socialization

| Feature         | Description             | Files                                            |
| --------------- | ----------------------- | ------------------------------------------------ |
| Matching Pools  | Create pools with rules | internal/model/pool.go, internal/service/pool.go |
| Auto-Matcher    | Scheduled matching job  | internal/service/pool_matcher.go                 |
| Exclusion Lists | Never match with X      | Part of pool model                               |
| Nudge System    | Timer-based reminders   | internal/service/nudge.go                        |

Phase 6: Adventures & Rideshares

Multi-day, multi-event coordination

| Feature                | Description                         | Files                                                        |
| ---------------------- | ----------------------------------- | ------------------------------------------------------------ |
| Adventure Planning     | Create adventures, add destinations | internal/service/adventure.go, internal/handler/adventure.go |
| Event Nesting          | Events within adventures            | Extend event service                                         |
| Ranked Voting          | Vote on destinations                | internal/handler/adventure.go                                |
| Rideshare Coordination | Ride sharing for events/adventures  | internal/service/rideshare.go, internal/handler/rideshare.go |
| Forum Integration      | Discussion threads                  | internal/service/forum.go, internal/handler/forum.go         |

Phase 7: Trust & Reputation

Community safety

| Feature           | Description                  | Files                                                |
| ----------------- | ---------------------------- | ---------------------------------------------------- |
| Review System     | Rate hosts/guests/organizers | internal/model/review.go, internal/service/review.go |
| Reputation Scores | Aggregate trust score        | Part of review service                               |
| Tag Feedback      | Quick feedback tags          | Part of review model                                 |
| Moderation Tools  | Report/block users           | internal/service/moderation.go                       |

---

Matching Algorithms

Compatibility Score (OkCupid-style)

1.  Find shared answered questions between users A and B
2.  For each question:
    - Weight by importance (irrelevant=0, mandatory=250)
    - Score 1 if B's answer is acceptable to A, else 0
3.  Calculate A→B score = earned_points / total_weight
4.  Calculate B→A score similarly
5.  Final = sqrt(A→B × B→A) × 100 (geometric mean)
6.  If dealbreaker violated → 0%

Availability Matching

1.  Query availabilities within radius + time overlap
2.  Score by: distance, activity overlap, compatibility
3.  Rank and return top matches
4.  Real-time SSE when new matches appear

Pool Matching (Donut-style)

1.  Build compatibility matrix for pool members
2.  Apply exclusions (never match pairs)
3.  Reduce score for recent matches (avoid repeats)
4.  Hungarian algorithm for optimal group formation
5.  Notify matched members via push/SSE

---

UI/UX Philosophy

Principles

1.  Progressive Disclosure: Simple by default, reveal on demand
2.  Information Density: Every pixel earns its place
3.  Calm Technology: Rare, meaningful notifications
4.  Honest Interface: No dark patterns, clear privacy

Key Screens

- Home: Guild list with activity indicators + availability toggle
- Discover: Map view with nearby events/people
- Events: Template picker → minimal form
- Profile: Bio + interests + availability settings
- Match: Swipe-free compatibility view

---

Healthy Community Design (iOS App Patterns)

Location Privacy (Hard Requirement)

Never expose exact locations. Even if users want to opt-in, it's a safety risk.

What we store:
location: {
lat: 34.0522, // Internal only, never exposed to other users
lng: -118.2437, // Internal only, never exposed to other users
city: "Los Angeles", // Shown to others
neighborhood: "Silver Lake", // Optional, user-controlled
country: "US" // Shown to others
}

What other users see:
DISTANCE DISPLAY:

- "Nearby" (< 1 km)
- "~2 km away"
- "~5 km away"
- "~10 km away"
- "> 20 km away"

Never: Exact coordinates, street addresses, or precise distances

Activity Timestamps (Freshness Indicators):
ACTIVITY TIERS (shown on profiles/discovery):
🟢 "Active now" - within last 10 minutes
🟢 "Active recently" - within last 30 minutes
🟡 "Active this hour" - within last hour
🟡 "Active today" - within last 24 hours
⚪ "Active yesterday" - 24-48 hours ago
⚪ "Active this week" - 2-7 days ago
⚫ "Away" - > 7 days (or no indicator)

Sorting Priority:

1.  Distance (closest first)
2.  Activity recency (most recent first within same distance tier)
3.  Compatibility score (within same distance + recency tier)

Example result order:

1.  Alex (Nearby, Active now, 87% match)
2.  Sam (Nearby, Active today, 92% match)
3.  Jordan (~2km, Active now, 78% match)
4.  Taylor (~5km, Active recently, 85% match)

---

Anti-Pattern: Shallow Binary Interactions

Problem: Tinder-style swipe left/right creates:

- Snap judgments based on photos
- Gamification of human connection
- Poorly behaved users who treat others as disposable
- Alienates "normal" people seeking genuine connection

Solution: Friction-Based Discovery

1.  NO swiping interfaces anywhere in the app
2.  Profiles expand via tap → read → express interest
3.  Interest requires writing a short note (why connect?)
4.  Compatibility % shown but never as a sorting filter
5.  Discovery limited to 5-10 suggestions per day

Implementation Patterns:
| Screen | Anti-Pattern | Saga Pattern |
|---------------------|----------------------|--------------------------------------------------|
| Discovery | Swipe cards | Grid with expand-on-tap |
| Expressing Interest | Single tap "like" | Tap + compose note (min 20 chars) |
| Matching | Instant match pop-up | "Pending connections" queue reviewed 1x/day |
| Profiles | Photo carousel | Photo + values summary + compatibility breakdown |

Anti-Pattern: Toxicity & Hate Groups

Problem: Platforms become havens for negativity when:

- Anonymity enables bad behavior
- Algorithmic engagement prioritizes outrage
- Moderation is reactive only
- Groups form around exclusion/hate

Solution: Values-First Architecture

1.  IDENTITY

    - No anonymous accounts
    - Guild membership requires invite/approval
    - Profile must include 3+ answered questions before discovery

2.  QUESTION CATEGORIES
    Required categories (must answer 1+ from each):

    - Values & Ethics
    - Social Style
    - Lifestyle Preferences
    - Communication

    Flagged categories (answers trigger review):

    - Political extremism indicators
    - Exclusionary language patterns

3.  GUILD GOVERNANCE

    - Guilds are private by default
    - Public events require organizer reputation score ≥ 4.0
    - Any member can flag content → immediate review queue

4.  PROACTIVE MODERATION
    - NLP scan on: bio, event descriptions, messages
    - Automatic flagging: slurs, hate speech, harassment patterns
    - Graduated response: warning → temp ban → permanent ban

Question Design for Values Alignment:
-- Example questions that surface values without politics
CREATE question SET
text = "When a friend is struggling, I prefer to...",
category = "values",
options = [
{value: "listen", label: "Listen without trying to fix"},
{value: "advise", label: "Offer practical advice"},
{value: "distract", label: "Help them take their mind off it"},
{value: "space", label: "Give them space until they reach out"}
],
is_dealbreaker_eligible = false;

CREATE question SET
text = "In group settings, I value...",
category = "social",
options = [
{value: "inclusion", label: "Making sure everyone feels included"},
{value: "depth", label: "Deep conversations with a few people"},
{value: "energy", label: "High energy and lots of activity"},
{value: "structure", label: "Clear plans and organization"}
],
is_dealbreaker_eligible = false;

-- Dealbreaker-eligible questions (user can mark as hard requirement)
CREATE question SET
text = "How important is respecting others' boundaries to you?",
category = "values",
options = [
{value: "essential", label: "Non-negotiable - always respect boundaries"},
{value: "important", label: "Very important with rare exceptions"},
{value: "contextual", label: "Depends on the situation"},
{value: "flexible", label: "People should be more flexible"}
],
is_dealbreaker_eligible = true;

Positive Categorization System

Instead of negative filters, use positive attraction:

INTERESTS WITH INTENT

- Not just "I like hiking" but "I want to organize hikes"
- Not just "speaks Spanish" but "wants to practice Spanish with others"
- Not just "likes cooking" but "wants to host dinner parties"

TEACHING/LEARNING PAIRS

- Every interest can be: learning | practicing | experienced | teaching
- System matches teachers with learners automatically
- Creates natural mentorship dynamics

VALUES TAGS (positive framing)

- "Welcomes newcomers"
- "Good listener"
- "Reliable"
- "Creates inclusive spaces"
- "Brings snacks"
- "Great host"

Constructive Feedback System (Not Just Ratings)

Problem with 1-5 stars:

- Binary in practice (5 = good, anything else = bad)
- No actionable information
- Creates anxiety around ratings
- Doesn't help anyone improve

Solution: Structured Positive Feedback

POST-EVENT FEEDBACK FLOW:

1.  "Did you attend?" (yes/no/partial)

2.  "What went well?" (multi-select tags)
    □ Good venue/location
    □ Right group size
    □ Interesting conversation
    □ Well organized
    □ Inclusive atmosphere
    □ Started/ended on time
    □ Good food/drinks
    □ Made new connections

3.  "What would make it even better?" (multi-select)
    □ More structured activities
    □ Better venue
    □ Different time
    □ Smaller group
    □ Larger group
    □ More lead time for planning
    □ Clearer expectations
    □ Dietary options

4.  "Would you attend again?" (yes/maybe/no)

5.  OPTIONAL: Private note to organizer
    - Not public, not part of rating
    - Constructive suggestions only

Reputation Display:
Instead of: ⭐ 4.2 (23 reviews)

Show:
┌─────────────────────────────────┐
│ Alex has hosted 12 events │
│ │
│ 🎯 Well organized (10) │
│ 🤝 Inclusive (9) │
│ 💬 Great conversation (8) │
│ 📍 Good venues (7) │
│ │
│ 11 of 12 attendees would return │
└─────────────────────────────────┘

Community Moderation Tools

Graduated Response System:
Level 0: NUDGE

- Trigger: First minor flag (e.g., slightly off-topic event)
- Action: Private message with guidelines
- Record: Not visible to others

Level 1: WARNING

- Trigger: Repeated minor flags or single moderate flag
- Action: Visible warning, feature restrictions (e.g., can't create public events)
- Duration: 7 days
- Record: Visible to guild admins only

Level 2: SUSPENSION

- Trigger: Serious flag or Level 1 violation during warning period
- Action: Account suspended, no access
- Duration: 30 days
- Record: Visible to all guild admins
- Appeal: Can request review after 14 days

Level 3: BAN

- Trigger: Hate speech, harassment, illegal activity, repeated Level 2
- Action: Permanent account termination
- Record: Email/device fingerprint blocklisted
- Appeal: None for hate/harassment; 6-month review for others

User-Initiated Actions:
BLOCK

- Instant, no questions asked
- Blocked user cannot see you anywhere
- Your profile invisible to them
- Mutual events show "1 attendee hidden"

FLAG

- Requires selecting category:
  - Spam/fake account
  - Harassment
  - Hate speech
  - Inappropriate content
  - Made me uncomfortable (private)
  - Other (requires description)
- Optional: Add context
- Creates moderation queue item

LEAVE GUILD

- Instant removal
- Option to report guild itself
- 30-day cooldown before rejoining

Diversity Without Tolerance of Hate

Design Principle: Diversity means welcoming people of all backgrounds, identities, and perspectives that don't cause harm. It does NOT mean tolerating:

- Hate based on identity (race, gender, sexuality, religion, nationality, disability)
- Harassment or bullying
- Dangerous misinformation
- Exclusionary group formation

Implementation:

1.  QUESTION FILTERS

    - Questions test for openness, not politics
    - "I enjoy meeting people from different backgrounds" → high match with diverse community
    - Users who answer toward closed-mindedness see fewer discovery results

2.  GUILD FORMATION

    - Guilds cannot have exclusionary names/descriptions
    - Auto-flag: "whites only", "no [group]", etc.
    - Public guilds require diversity statement

3.  ALGORITHMIC DIVERSITY

    - Discovery intentionally surfaces variety
    - If you only connect with one demographic, system prompts broader exploration
    - "You might also enjoy" includes outside-comfort-zone suggestions

4.  PROTECTED CATEGORIES
    The app explicitly protects and affirms:

    - All races and ethnicities
    - All gender identities and expressions
    - All sexual orientations
    - All religions (and non-religion)
    - All abilities and disabilities
    - All nationalities and immigration statuses
    - All body types

    Community standards explicitly state: discrimination against protected categories = ban

Structured Hangouts (Improving on Couchsurfing)

Problems with Couchsurfing Hangouts:

- Too open-ended ("exploring the area") → nobody commits
- No defined roles → awkward expectations
- "Positive vibes only" pressure → excludes people having hard days
- Excessive filtering → stagnation, no spontaneous connections
- Vague suggestions → decision paralysis

Solution: Role-Based Hangout Types

HANGOUT INTENTS (user selects when posting availability):

🗣️ TALK IT OU
"I have something on my mind I'd like to talk through" - Seeking: A listening ear - Vibe: Supportive, no judgment - Examples: Career crossroads, relationship question, life decision

👂 HERE TO LISTEN
"I'm in a good headspace and happy to listen" - Offering: Attention, maybe advice if wanted - Vibe: Patient, present - Matches with: Talk It Out posts

🎯 CONCRETE ACTIVITY
"I want to do [specific thing] with someone" - Required: Select from curated list OR add custom - Examples:
□ Comedy show tonight at [venue]
□ Try [specific restaurant]
□ Visit [specific museum/exhibit]
□ Play [specific sport/game]
□ Walk [specific trail/neighborhood]
□ Work from [specific cafe] together - NOT allowed: "hang out", "explore", "whatever"

🤝 MUTUAL INTEREST
"I want to connect with someone who shares [interest]" - Links to: User's interest tags - Examples:
□ Photography walk (both have photography interest)
□ Language exchange (complementary languages)
□ Board games (both have gaming interest)

🆕 MEET ANYONE
"I'm open to meeting new people, surprise me" - System picks based on: compatibility, variety, mutual availability - Lower barrier but still structured: - Suggests a venue (coffee shop, park, etc.) - Suggests a timeframe - User confirms or suggests alternative

Matching Flow:

1.  USER A posts availability:
    Type: 🎯 CONCRETE ACTIVITY
    What: "Comedy at Laugh Factory, 8pm show"
    When: Tonight 7pm-11pm
    Open to: 1-3 people

2.  USER B sees in "Available Now" feed:

    - Shown because: B is available, B has comedy interest, B is nearby
    - B taps to see: A's profile summary + compatibility %

3.  USER B expresses interest:

    - Required: Short note (20+ chars)
    - Example: "Love standup! Haven't been to Laugh Factory yet"

4.  USER A receives request:

    - Can: Accept / Politely Decline / Propose Alternative
    - "Politely Decline" sends: "Thanks for reaching out! Can't make it work this time."
    - No explanation required, no negative mark

5.  If accepted:
    - Both get each other's contact preference (in-app chat, phone, etc.)
    - System creates a "hangout" record for later feedback
    - Option to add to calendar

Anti-Stagnation Rules:

1.  DECLINING LIMITS

    - Can decline anyone, but after 5 consecutive declines:
      - Prompt: "Noticed you've passed on a few connections.
        Would you like to adjust your hangout preferences?"
      - No punishment, just gentle nudge

2.  FILTERING LIMITS

    - Cannot filter by: age, gender, appearance
    - CAN filter by: activity type, distance, availability window
    - System occasionally shows "outside your usual" matches

3.  VARIETY ENCOURAGEMENT

    - If only connecting with same people:
      - "You and [friend] hang out a lot!
        Here's someone new you might enjoy meeting."
    - Pool matching (Donut-style) ensures variety

4.  FRESH OVER STALE
    - Recently active users surfaced first
    - Hangout posts expire after 24 hours
    - No "saved searches" that ossify preferences

Concrete Suggestion Engine:

Instead of letting users post vague hangouts, guide them:

Step 1: "What kind of hangout?"
[Talk] [Listen] [Activity] [Interest] [Surprise Me]

Step 2 (if Activity): "What specifically?"

         NEARBY & HAPPENING NOW:
         🎭 Comedy at Laugh Factory (8pm)
         🍜 Try Daikokuya Ramen (open till 11)
         🌳 Griffith Observatory sunset (clear tonight)

         POPULAR IN YOUR AREA:
         ☕ Work from Verve Coffee
         🎳 Bowling at Highland Park Bowl
         🎮 Board games at Game Haus

         YOUR INTERESTS:
         📷 Golden hour at Venice Beach
         🎸 Open mic at Hotel Cafe

         [+ Add your own idea]

Step 3: "When?"
[Now-ish] [This evening] [Tomorrow] [Pick time]

Step 4: "How many people?"
[Just 1 person] [2-3 people] [Small group 4+]

RESULT: "Going to Daikokuya Ramen tonight around 7pm, looking for 1-2 people to join!" - Specific - Time-bound - Clear expectations

Psychological Safety Features:

1.  NO "POSITIVE VIBES ONLY" MESSAGING

    - App never says "be positive" or "good energy only"
    - Acknowledges: people have hard days, that's okay
    - "Talk It Out" option explicitly welcomes processing

2.  LOW-STAKES FRAMING

    - "Grab coffee" not "make a new best friend"
    - One-time hangout, no commitment implied
    - "If you click, great. If not, no big deal."

3.  GRACEFUL EXITS

    - Decline without explanation
    - "Something came up" cancellation (no penalty unless pattern)
    - Post-hangout feedback is optional

4.  MANAGING EXPECTATIONS
    - Compatibility % shown but framed as "shared interests"
    - No "perfect match" language
    - "You both enjoy: hiking, cooking, board games"

iOS App Structure (SwiftUI)

View Hierarchy (Three Rigors):
App
├── AuthenticationView (login/register)
├── MainTabView
│ ├── HomeTab (Dashboard)
│ │ ├── EngagementSelector [Spontaneous | Events | Adventures]
│ │ ├── TimelineView (chronological feed)
│ │ ├── CalendarView (date-based)
│ │ ├── MapView (geographic)
│ │ ├── GuildListView
│ │ └── AvailabilityToggle (floating)
│ │
│ ├── DiscoverTab
│ │ ├── DiscoverModeSelector [Timeline | Calendar | Map]
│ │ ├── FilterBar (location, date, type, interests)
│ │ ├── SpontaneousNowView (available people)
│ │ ├── EventDiscoveryView (grid, not swipe!)
│ │ ├── AdventureDiscoveryView
│ │ └── PeopleDiscoveryView
│ │
│ ├── CreateTab (+ button)
│ │ ├── CreateTypeSelector [Event | Adventure]
│ │ ├── CreateEventFlow
│ │ │ ├── TemplatePickerView
│ │ │ ├── EventDetailsView
│ │ │ ├── RideshareSetupView (optional)
│ │ │ └── EventPreviewView
│ │ └── CreateAdventureFlow
│ │ ├── AdventureDetailsView
│ │ ├── DestinationListView
│ │ ├── EventsWithinAdventureView
│ │ ├── RidesharesSetupView
│ │ └── AdventurePreviewView
│ │
│ ├── ConnectionsTab
│ │ ├── PendingConnectionsView
│ │ ├── MatchSuggestionsView (limited daily)
│ │ ├── CompatibilityDetailView
│ │ └── ConnectionProfileView
│ │
│ └── ProfileTab
│ ├── ProfileEditView
│ ├── InterestsView
│ ├── QuestionsView
│ ├── AvailabilitySettingsView
│ ├── ReputationView
│ ├── GuildAlliancesView
│ └── SettingsView
│
├── Detail Views
│ ├── EventDetailView
│ │ ├── EventInfoSection
│ │ ├── AttendeesSection
│ │ ├── RidesharesSection
│ │ └── ForumSection
│ │
│ ├── AdventureDetailView
│ │ ├── AdventureInfoSection
│ │ ├── DestinationsSection (with voting)
│ │ ├── EventsSection (nested events)
│ │ ├── RidesharesSection (inter-event)
│ │ ├── ParticipantsSection
│ │ └── ForumSection
│ │
│ ├── RideshareDetailView
│ │ ├── RouteMapSection
│ │ ├── SeatsSection
│ │ └── BookingSeatFlow
│ │
│ └── ForumView
│ ├── ThreadedPostsView
│ ├── ComposePostSheet
│ └── ReactionsView
│
└── Sheets/Modals
├── ExpressInterestSheet (with note composer)
├── RSVPSheet (values questions if required)
├── FeedbackFlowSheet
├── ReportSheet
├── QuestionAnswerSheet
├── DestinationVoteSheet
└── RideshareRequestSheet

Key Design Decisions:
| Element | Decision | Rationale |
|--------------|-------------------------------|-------------------------------------------|
| Three Rigors | Spontaneous, Event, Adventure | Graduated commitment levels |
| Discovery | Grid not cards | Prevents swipe muscle memory |
| Interest | Requires note | Adds friction = genuine intent |
| Daily limit | 5-10 suggestions | Quality over quantity |
| Ratings | Tags not stars | Actionable, positive framing |
| Profiles | Values first | Character over appearance |
| Moderation | Graduated | Fair, recoverable, documented |
| Forums | Per Adventure/Event | Threaded discussion, inherited visibility |
| Rideshares | Nested under parent | Security inherits from Event/Adventure |

---

Resonance Scoring System

Design Constraints (Hard Requirements)

1.  No punishments: No negative points, no score reductions. Users may fail to earn points; they are never penalized.
2.  Additive only: All awarded points are appended to an immutable ledger. Totals are sums of ledger entries.
3.  Anti-farming: Points require counterparty verification and/or feedback. Pairwise diminishing returns reduce collusion.
4.  Auditability: Every award creates a ledger entry: (user_id, stat, points, source_object_id, reason_code, timestamp).

Data Model

-- Resonance Ledger (immutable, append-only)
DEFINE TABLE resonance_ledger SCHEMAFULL;
DEFINE FIELD user ON resonance_ledger TYPE record<user>;
DEFINE FIELD stat ON resonance_ledger TYPE string; -- questing, mana, wayfinder, attunement, nexus
DEFINE FIELD points ON resonance_ledger TYPE int;
DEFINE FIELD source_object_id ON resonance_ledger TYPE string; -- event:xyz, question:abc, month:2026-01
DEFINE FIELD reason_code ON resonance_ledger TYPE string;
DEFINE FIELD created_on ON resonance_ledger TYPE datetime;

-- Event completion confirmations
DEFINE FIELD completion_confirmed ON event_participant TYPE option<datetime>;
DEFINE FIELD checkin_time ON event_participant TYPE option<datetime>;

-- Support session feedback
DEFINE FIELD helpfulness_rating ON event_participant TYPE option<string>; -- YES, SOMEWHAT, NOT_REALLY, SKIP
DEFINE FIELD helpfulness_tags ON event_participant TYPE option<array<string>>;

Configurable Windows

const (
ConfirmWindow = 48 _ time.Hour // After event ends
CheckinWindow = 10 _ time.Minute // ± around start time
EarlyConfirmWindow = 2 _ time.Hour // Before start time
NexusWindow = 30 _ 24 \* time.Hour // Rolling 30 days
)

Resonance Formula

Resonance(u) = Questing(u) + Mana(u) + Wayfinder(u) + Attunement(u) + Nexus(u)

---

🧭 Questing (Verified Follow-Through)

Intent: Reward showing up and being punctual.

Eligibility: User earns Questing from event e if completion is verified.

Verification Rules:

- 1:1 event: Both participants must confirm completion within 48 hours
- Group event: Host + ≥2 attendees must confirm within 48 hours

Formula per event:
Questing_award = 10 (verified completion) + 2 (early confirm bonus: confirmed ≥2h before start) + 2 (on-time checkin bonus: ±10 min of start)

Range: 10-14 points per event

Daily Cap: 40 points/day

---

✨ Mana (Support That Landed)

Intent: Reward support sessions where the receiver confirms it was helpful.

Eligibility: Helper earns Mana from support session if:

1.  Event has support_role = true
2.  Both parties confirm completion
3.  Receiver rates helpfulness as YES or SOMEWHAT

Formula per session:
Mana_base = 12 (verified + receiver confirmed helpful) + 2 (early confirm bonus) + 2 (receiver selected "what helped" tag)

Range: 12-16 points before diminishing returns

Pairwise Diminishing Returns (Anti-Collusion):
Sessions 1-3 between same pair (30 days): × 1.0
Sessions 4-6: × 0.5
Sessions 7+: × 0.25

Daily Cap: 32 points/day

Edge Cases:

- Receiver rates NOT_REALLY → Helper gets 0 (no punishment)
- Receiver skips rating → Helper gets 0 (no punishment)
- No negative consequences, just no reward

---

🗺️ Wayfinder (Hosting That Happened

Intent: Reward hosting events that actually occur with verified attendees.

Eligibility: Host earns Wayfinder when event is verified complete.

Formula per event:
A = min(verified_attendees, 4) // Cap at 4 to prevent mega-event farming

Wayfinder_award = 8 (verified hosting) + 2\*A (per verified attendee, max 8) + 2 (early confirm bonus)

1:1 event (A=1): 12-14 points
Group (A=4): 18-20 points

Daily Cap: 30 points/day

---

🎛️ Attunement (Profile Clarity

Intent: Reward filling in matching info that improves recommendations.

Sources:

A) Matching Questions:
+2 points per first-time answer
Cap: 20 points/day (10 questions)

B) Monthly Profile Refresh:
+10 points for updating availability/preferences
Limit: Once per calendar month
Requires: Substantive change (not trivial edit)

Anti-Spam:

- Only first answer to each question earns points
- Edits rate-limited to max +2/day with 6-hour cooldown
- Blank/default answers earn nothing

---

🕸️ Nexus (Active Guilds + Bridging

Intent: Reward meaningful involvement in living guilds and bridging communities.

Computed monthly and appended as a single ledger entry.

Guild Eligibility:
GuildActive = 1 if:

- ≥2 verified events in last 30 days, AND
- ≥3 active members (participated in guild events)
  Else: 0 (prevents shell/staged guilds)

Per-Guild Contribution:
ActivityFactor = min(1, my_verified_completions / 3) // 0→1 by 3 events/month

GuildNexus = round(5 × log₂(1 + ActiveMembers) × ActivityFactor × GuildActive)

Bridge Bonus (Cross-Guild Activity):
For each pair of guilds (g, h) where user is active in both:

Overlap = users active in both guilds
BridgeNexus = round(2 × log₂(1 + Overlap) × min(ActivityFactor_g, ActivityFactor_h))

Monthly Total:
Nexus_monthly = Σ GuildNexus(g) + Σ BridgeNexus(g,h)

Monthly Cap: 200 points (safety cap)

---

Display (Minimal, Warm)

┌─────────────────────────────────┐
│ ✨ Resonance: 847 │
├─────────────────────────────────┤
│ 🧭 Questing 286 │
│ ✨ Mana 194 │
│ 🗺️ Wayfinder 168
│ 🎛️ Attunement 82
│ 🕸️ Nexus 117
└─────────────────────────────────┘

[Why did I earn points?] → Links to ledger view

Tooltips (plain language):

- Questing: "Points for showing up to events you committed to"
- Mana: "Points for support sessions where the other person said it helped"
- Wayfinder: "Points for hosting events that people attended"
- Attunement: "Points for answering matching questions"
- Nexus: "Points for being active in healthy guilds"

---

Anti-Farming Summary

| Measure                | Implementation                                           |
| ---------------------- | -------------------------------------------------------- |
| Verification           | No points without mutual completion confirmation         |
| Quality Gate           | Mana requires receiver YES/SOMEWHAT                      |
| Diminishing Returns    | Same-pair Mana: 1.0 → 0.5 → 0.25 over 30 days            |
| Daily Caps             | Questing 40, Mana 32, Wayfinder 30, Attunement 20        |
| Guild Validity         | Nexus only for guilds with ≥2 events + ≥3 active members |
| Participation Required | Membership alone = 0 Nexus                               |
| Idempotent Ledger      | Each (user, stat, source_object) can only award once     |
| Silent Mitigation      | Detect cliques, tighten multipliers quietly (no bans)    |

---

API Endpoints (New)

# Profile & Discovery

/v1/profile GET, PATCH
/v1/profile/interests GET, POST, DELETE
/v1/profile/availability GET, POST, PATCH, DELETE

/v1/questions GET
/v1/questions/{id}/answer POST, PATCH

/v1/compatibility/{userId} GET

# Discovery (Three Views)

/v1/discover/timeline GET (chronological feed)
/v1/discover/calendar GET (date-based view)
/v1/discover/map GET (geographic pins)
/v1/discover/events GET (geo query)
/v1/discover/adventures GET (geo + date query)
/v1/discover/people GET (interest filters)
/v1/discover/spontaneous GET (nearby available now)

# Guilds & Alliances

/v1/guilds/{id}/pools GET, POST
/v1/guilds/{id}/pools/{id}/join POST
/v1/guilds/{id}/pools/{id}/leave POST
/v1/guilds/{id}/alliances GET, POST
/v1/guilds/{id}/alliances/{id} GET, DELETE (revoke)
/v1/guilds/{id}/alliances/{id}/approve POST

# Adventures (formerly Trips)

/v1/adventures GET, POST
/v1/adventures/{id} GET, PATCH, DELETE
/v1/adventures/{id}/join POST (RSVP)
/v1/adventures/{id}/leave POST
/v1/adventures/{id}/participants GET
/v1/adventures/{id}/destinations GET, POST
/v1/adventures/{id}/destinations/{id}/vote POST
/v1/adventures/{id}/events GET, POST (nested events)
/v1/adventures/{id}/rideshares GET, POST (inter-event transport)
/v1/adventures/{id}/forum GET (get forum)
/v1/adventures/{id}/forum/posts GET, POST

# Events (can be standalone or nested in Adventure)

/v1/events GET, POST
/v1/events/{id} GET, PATCH, DELETE
/v1/events/{id}/rsvp POST
/v1/events/{id}/rideshares GET, POST
/v1/events/{id}/forum GET
/v1/events/{id}/forum/posts GET, POST
/v1/events/{id}/confirm POST (mark complete)
/v1/events/{id}/checkin POST (on-time checkin)
/v1/events/{id}/feedback POST (helpfulness rating + tags)

# Rideshares (formerly Commutes)

/v1/rideshares/{id} GET, PATCH, DELETE
/v1/rideshares/{id}/seats GET
/v1/rideshares/{id}/seats/request POST
/v1/rideshares/{id}/seats/{id} PATCH (confirm/cancel)

# Forums

/v1/forums/{id} GET
/v1/forums/{id}/posts GET, POST
/v1/forums/{id}/posts/{id} GET, PATCH, DELETE
/v1/forums/{id}/posts/{id}/react POST
/v1/forums/{id}/posts/{id}/pin POST (organizers only)

# Reviews & Reputation

/v1/reviews POST
/v1/reviews/received GET

# Resonance

/v1/resonance GET (my total + breakdown)
/v1/resonance/ledger GET (paginated history)
/v1/resonance/users/{userId} GET (view another's public score)

---

New SSE Event Types

// Availability (Spontaneous)
EventAvailabilityCreated = "availability.created"
EventAvailabilityMatch = "availability.match"

// Matching Pools
EventPoolMatchCreated = "pool.match_created"
EventNudge = "nudge"

// Adventures (formerly Trips)
EventAdventureUpdated = "adventure.updated"
EventAdventureJoined = "adventure.joined"
EventDestinationVoted = "adventure.destination_voted"
EventAdventureEventAdded = "adventure.event_added"

// Events
EventEventUpdated = "event.updated"
EventEventRsvp = "event.rsvp"

// Rideshares (formerly Commutes)
EventRideshareCreated = "rideshare.created"
EventRideshareSeatRequested = "rideshare.seat_requested"
EventRideshareSeatConfirmed = "rideshare.seat_confirmed"

// Forums
EventForumPostCreated = "forum.post_created"
EventForumPostReaction = "forum.post_reaction"

// Reviews
EventReviewReceived = "review.received"

---

Monetization

ONE-TIME: $9.99

- All features forever
- All future updates
- No limits

OPTIONAL DONATIONS:

- Coffee ($3): Thank you
- Dinner ($15): Supporters list
- Champion ($50): Beta access

ZERO: Subscriptions, ads, data selling, feature gating

---

Files to Create/Modify

New Files (Phase 1: Consolidated Schema)

| File                             | Purpose                                 |
| -------------------------------- | --------------------------------------- |
| internal/model/guild.go          | Guild (formerly Circle)                 |
| internal/model/adventure.go      | Adventure (multi-day coordination)      |
| internal/model/rideshare.go      | Rideshare (attached to Event/Adventure) |
| internal/model/forum.go          | Forum, ForumPost                        |
| internal/model/guild_alliance.go | Guild alliances                         |
| internal/service/guild.go        | Guild management                        |
| internal/service/adventure.go    | Adventure coordination                  |
| internal/service/rideshare.go    | Rideshare management                    |
| internal/service/forum.go        | Forum management                        |
| internal/handler/guild.go        | Guild endpoints                         |
| internal/handler/adventure.go    | Adventure endpoints                     |
| internal/handler/rideshare.go    | Rideshare endpoints                     |
| internal/handler/forum.go        | Forum endpoints                         |

New Files (Later Phases)

| File                              | Purpose                         |
| --------------------------------- | ------------------------------- |
| internal/model/profile.go         | User profile, interests         |
| internal/model/availability.go    | Availability windows            |
| internal/model/questionnaire.go   | Questions, answers              |
| internal/model/pool.go            | Matching pools                  |
| internal/model/review.go          | Reviews, reputation             |
| internal/model/resonance.go       | Resonance ledger, stats         |
| internal/service/profile.go       | Profile management              |
| internal/service/availability.go  | Availability matching           |
| internal/service/compatibility.go | Match scoring                   |
| internal/service/pool_matcher.go  | Auto-matching job               |
| internal/service/review.go        | Review management               |
| internal/service/geo.go           | Geographic utilities            |
| internal/service/resonance.go     | Scoring pipeline                |
| internal/jobs/nexus_monthly.go    | Monthly Nexus calculation job   |
| internal/handler/discover.go      | Timeline/Calendar/Map endpoints |
| internal/handler/profile.go       | Profile endpoints               |
| internal/handler/pool.go          | Pool endpoints                  |
| internal/handler/review.go        | Review endpoints                |
| internal/handler/resonance.go     | Resonance endpoints             |

Modify Files

| File                                | Changes                                                            |
| ----------------------------------- | ------------------------------------------------------------------ |
| migrations/001_initial_schema.surql | Complete rewrite with new terminology                              |
| internal/model/event.go             | Add adventure_id, order_in_adventure, change circle_id to guild_id |
| internal/service/event.go           | Add adventure hierarchy support                                    |
| internal/handler/event.go           | Add nested event creation, forum access                            |
| cmd/server/main.go                  | Wire new services/handlers                                         |

Delete/Replace Files

| File                        | Action                    |
| --------------------------- | ------------------------- |
| internal/model/circle.go    | Replace with guild.go     |
| internal/model/trip.go      | Replace with adventure.go |
| internal/model/commute.go   | Replace with rideshare.go |
| internal/service/circle.go  | Replace with guild.go     |
| internal/service/trip.go    | Replace with adventure.go |
| internal/service/commute.go | Replace with rideshare.go |
| internal/handler/circle.go  | Replace with guild.go     |
| internal/handler/trip.go    | Replace with adventure.go |
| internal/handler/commute.go | Replace with rideshare.go |

---

Decisions Made

| Decision               | Choice                                  | Rationale                                                    |
| ---------------------- | --------------------------------------- | ------------------------------------------------------------ |
| Circle Rename          | Circle → Guild                          | More evocative, implies community with shared purpose        |
| Entity Hierarchy       | Adventure → Event → Rideshare           | Prevents feature duplication, enables composition            |
| Trip Rename            | Trip → Adventure                        | More evocative, implies multi-component journey              |
| Commute Rename         | Commute → Rideshare                     | Clearer purpose, always attached to Event/Adventure          |
| Association Rename     | Association → Alliance                  | Stronger partnership connotation between Guilds              |
| Rideshare Attachment   | Required (Event OR Adventure)           | No standalone rideshares, security inherits from parent      |
| Three Rigors           | Spontaneous, Event, Adventure           | Graduated commitment levels for dashboard organization       |
| Forums                 | Per Adventure/Event                     | Threaded discussion with inherited visibility                |
| Guild Alliances        | Bidirectional partnerships              | Enable cross-guild adventures and discovery                  |
| Consolidated Migration | All changes in 001_initial_schema.surql | Design phase, no legacy data to migrate                      |
| MVP Scope              | All 7 phases                            | Full vision - comprehensive social coordination              |
| Platform Priority      | iOS first                               | Native SwiftUI, then expand to web/Android                   |
| Default Privacy        | Guilds-only                             | Profile visible only to guild members by default             |
| Visibility Cascade     | Parent → Child                          | Events inherit from Adventure, Rideshares inherit from Event |

---

Implementation Order

Given the consolidated schema approach (still in design phase):

1.  Phase 1: Consolidated Schema - Rewrite 001_initial_schema.surql with all new terminology (Guild, Adventure, Rideshare, Alliance)
2.  Phase 2: Foundation - Profiles, interests, enhanced events, geo layer
3.  Phase 3: Discovery - Availability, Timeline/Calendar/Map views, location-based search
4.  Phase 4: Compatibility - Questionnaires, match scoring, RSVP gating
5.  Phase 5: Automation - Matching pools, nudge system
6.  Phase 6: Adventures & Rideshares - Full adventure coordination, rideshare features
7.  Phase 7: Trust - Reviews, reputation, moderation

Each phase builds on the previous. iOS app evolves alongside API.

---

Next Steps

1.  Rewrite migration migrations/001_initial_schema.surql:

- Rename circle → guild (all references)
- Rename trip → adventure, trip_participant → adventure_participant
- Rename commute → rideshare, update related tables
- Rename circle_association → guild_alliance
- Add adventure_id to event table
- Add forum, forum_post tables
- Update all visibility defaults from "circles" to "guilds"
- Update business constraint events

2.  Implement Go models for new/renamed entities:

- internal/model/guild.go (replace circle.go)
- internal/model/adventure.go
- internal/model/rideshare.go
- internal/model/forum.go
- internal/model/guild_alliance.go

3.  Update existing models:

- Add AdventureID to Event model
- Change CircleID to GuildID everywhere
- Add visibility cascade logic

4.  Create handlers and services for new entities
5.  Update iOS app with new entity names and hierarchy

---

Schema Migration Summary

-- Key terminology changes in migrations/001_initial_schema.surql
-- (Consolidated since still in design phase - no legacy data)

-- 1. circle → guild (throughout)
DEFINE TABLE guild SCHEMAFULL;
-- All circle_id fields become guild_id

-- 2. trip → adventure
DEFINE TABLE adventure SCHEMAFULL;
DEFINE FIELD guild_id ON adventure TYPE option<record<guild>>;
DEFINE FIELD visibility ON adventure TYPE string DEFAULT "guilds";

-- 3. Add event.adventure_id
DEFINE FIELD adventure_id ON event TYPE option<record<adventure>>;
DEFINE FIELD order_in_adventure ON event TYPE option<int>;
DEFINE FIELD guild_id ON event TYPE option<record<guild>>; -- renamed from circle_id

-- 4. commute → rideshare with required parent
DEFINE TABLE rideshare SCHEMAFULL;
DEFINE FIELD event_id ON rideshare TYPE option<record<event>>;
DEFINE FIELD adventure_id ON rideshare TYPE option<record<adventure>>;
-- Constraint: exactly one must be set

-- 5. Add forums
DEFINE TABLE forum SCHEMAFULL;
DEFINE FIELD adventure_id ON forum TYPE option<record<adventure>>;
DEFINE FIELD event_id ON forum TYPE option<record<event>>;

DEFINE TABLE forum_post SCHEMAFULL;
-- threading, reactions, mentions

-- 6. circle_association → guild_alliance
DEFINE TABLE guild_alliance SCHEMAFULL;
DEFINE FIELD guild_a_id ON guild_alliance TYPE record<guild>;
DEFINE FIELD guild_b_id ON guild_alliance TYPE record<guild>;
DEFINE FIELD status ON guild_alliance TYPE string; -- pending, active, revoked
