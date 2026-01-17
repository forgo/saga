-- Saga Schema
-- Version: 2.0.0
-- Created: 2026-01-06
--
-- Comprehensive social coordination platform combining:
-- - Relationship tracking with activity timers
-- - User profiles with location privacy
-- - Interest-based discovery with teaching/learning
-- - Availability & structured hangouts
-- - Questionnaire-based compatibility matching
-- - Enhanced events (Partiful-style)
-- - Event roles and host management
-- - Matching pools (Donut-style)
-- - Adventure planning with ranked voting (multi-day, multi-event)
-- - Rideshare matching (attached to Events/Adventures)
-- - Forums for Adventures and Events
-- - Guild alliances (cross-guild partnerships)
-- - Trust relations & IRL verification
-- - Reviews & reputation (tag-based)
-- - Resonance scoring system
-- - Nudge notifications
-- - Moderation & safety
--
-- Terminology (v2):
-- - Guild (formerly Circle): Community with shared purpose
-- - Adventure (formerly Trip): Multi-day, multi-location coordination
-- - Rideshare (formerly Commute): Transportation attached to Event/Adventure
-- - Alliance (formerly Association): Partnership between Guilds

-- =============================================================================
-- 1. USER & AUTHENTICATION
-- =============================================================================

DEFINE TABLE user SCHEMAFULL;
DEFINE FIELD email ON user TYPE string ASSERT string::is_email($value);
DEFINE FIELD username ON user TYPE option<string>;
DEFINE FIELD hash ON user TYPE option<string>;
DEFINE FIELD firstname ON user TYPE option<string>;
DEFINE FIELD lastname ON user TYPE option<string>;
DEFINE FIELD email_verified ON user TYPE bool DEFAULT false;
DEFINE FIELD role ON user TYPE string DEFAULT "user" ASSERT $value IN ["user", "moderator", "admin"];
DEFINE FIELD created_on ON user TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON user TYPE datetime DEFAULT time::now();
DEFINE FIELD login_on ON user TYPE option<datetime>;

DEFINE INDEX user_email ON user COLUMNS email UNIQUE;
DEFINE INDEX user_username ON user COLUMNS username UNIQUE;

-- Federated identity (OAuth providers)
DEFINE TABLE identity SCHEMAFULL;
DEFINE FIELD user ON identity TYPE record<user>;
DEFINE FIELD provider ON identity TYPE string;
DEFINE FIELD provider_user_id ON identity TYPE string;
DEFINE FIELD provider_email ON identity TYPE option<string>;
DEFINE FIELD email_verified_by_provider ON identity TYPE bool DEFAULT false;
DEFINE FIELD access_token ON identity TYPE option<string>;
DEFINE FIELD refresh_token ON identity TYPE option<string>;
DEFINE FIELD created_on ON identity TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON identity TYPE datetime DEFAULT time::now();

DEFINE INDEX identity_provider ON identity COLUMNS provider, provider_user_id UNIQUE;
DEFINE INDEX identity_user ON identity COLUMNS user;

-- Passkeys (WebAuthn)
DEFINE TABLE passkey SCHEMAFULL;
DEFINE FIELD user ON passkey TYPE record<user>;
DEFINE FIELD credential_id ON passkey TYPE string;
DEFINE FIELD public_key ON passkey TYPE bytes;
DEFINE FIELD sign_count ON passkey TYPE int DEFAULT 0;
DEFINE FIELD name ON passkey TYPE string;
DEFINE FIELD created_on ON passkey TYPE datetime DEFAULT time::now();
DEFINE FIELD last_used_on ON passkey TYPE option<datetime>;

DEFINE INDEX passkey_credential ON passkey COLUMNS credential_id UNIQUE;
DEFINE INDEX passkey_user ON passkey COLUMNS user;

-- Refresh tokens
DEFINE TABLE refresh_token SCHEMAFULL;
DEFINE FIELD user ON refresh_token TYPE record<user>;
DEFINE FIELD token_hash ON refresh_token TYPE string;
DEFINE FIELD expires_at ON refresh_token TYPE datetime;
DEFINE FIELD created_at ON refresh_token TYPE datetime DEFAULT time::now();
DEFINE FIELD revoked ON refresh_token TYPE bool DEFAULT false;

DEFINE INDEX refresh_token_hash ON refresh_token COLUMNS token_hash UNIQUE;
DEFINE INDEX refresh_token_user ON refresh_token COLUMNS user;

-- =============================================================================
-- 2. GUILDS & MEMBERSHIP (formerly Circles)
-- =============================================================================

DEFINE TABLE guild SCHEMAFULL;
DEFINE FIELD name ON guild TYPE string DEFAULT "My Guild" ASSERT string::len($value) <= 100;
DEFINE FIELD description ON guild TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD icon ON guild TYPE option<string> ASSERT string::len($value) <= 50;
DEFINE FIELD color ON guild TYPE option<string> ASSERT string::len($value) <= 20;
DEFINE FIELD visibility ON guild TYPE string DEFAULT "private";
DEFINE FIELD created_on ON guild TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON guild TYPE datetime DEFAULT time::now();

DEFINE TABLE member SCHEMAFULL;
DEFINE FIELD name ON member TYPE string ASSERT string::len($value) <= 100;
DEFINE FIELD email ON member TYPE string;
DEFINE FIELD user ON member TYPE record<user>;
DEFINE FIELD created_on ON member TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON member TYPE datetime DEFAULT time::now();

DEFINE INDEX member_user ON member COLUMNS user;

-- Member authenticates via User
DEFINE TABLE authenticates_via SCHEMAFULL TYPE RELATION FROM member TO user;

-- Member is responsible for Guild
DEFINE TABLE responsible_for SCHEMAFULL TYPE RELATION FROM member TO guild;
DEFINE FIELD pending_approval ON responsible_for TYPE bool DEFAULT false;

-- =============================================================================
-- 3. GUILD ALLIANCES (Cross-Guild Partnerships)
-- =============================================================================

DEFINE TABLE guild_alliance SCHEMAFULL;
DEFINE FIELD guild_a_id ON guild_alliance TYPE record<guild>;
DEFINE FIELD guild_b_id ON guild_alliance TYPE record<guild>;
DEFINE FIELD status ON guild_alliance TYPE string DEFAULT "pending";
DEFINE FIELD initiated_by_id ON guild_alliance TYPE record<user>;
DEFINE FIELD approved_by_id ON guild_alliance TYPE option<record<user>>;
DEFINE FIELD created_on ON guild_alliance TYPE datetime DEFAULT time::now();
DEFINE FIELD approved_on ON guild_alliance TYPE option<datetime>;
DEFINE FIELD revoked_on ON guild_alliance TYPE option<datetime>;

DEFINE INDEX guild_alliance_a ON guild_alliance FIELDS guild_a_id;
DEFINE INDEX guild_alliance_b ON guild_alliance FIELDS guild_b_id;
DEFINE INDEX guild_alliance_status ON guild_alliance FIELDS status;
DEFINE INDEX guild_alliance_pair ON guild_alliance FIELDS guild_a_id, guild_b_id UNIQUE;

-- =============================================================================
-- 4. PEOPLE & ACTIVITIES (Relationship Tracking)
-- =============================================================================

DEFINE TABLE person SCHEMAFULL;
DEFINE FIELD guild_id ON person TYPE record<guild>;
DEFINE FIELD name ON person TYPE string ASSERT string::len($value) <= 100;
DEFINE FIELD nickname ON person TYPE option<string> ASSERT string::len($value) <= 50;
DEFINE FIELD avatar ON person TYPE option<string>;
DEFINE FIELD birthday ON person TYPE option<datetime>;
DEFINE FIELD notes ON person TYPE option<string> ASSERT string::len($value) <= 1000;
DEFINE FIELD created_on ON person TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON person TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_person_guild ON person FIELDS guild_id;

-- Person belongs to Guild (legacy relation, kept for compatibility)
DEFINE TABLE part_of SCHEMAFULL TYPE RELATION FROM person TO guild;

DEFINE TABLE activity SCHEMAFULL;
DEFINE FIELD guild_id ON activity TYPE record<guild>;
DEFINE FIELD name ON activity TYPE string ASSERT string::len($value) <= 50;
DEFINE FIELD icon ON activity TYPE string ASSERT string::len($value) <= 50;
DEFINE FIELD warn ON activity TYPE float ASSERT $value >= 60 AND $value <= 2592000;
DEFINE FIELD critical ON activity TYPE float ASSERT $value >= 60 AND $value <= 7776000;
DEFINE FIELD created_on ON activity TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON activity TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_activity_guild ON activity FIELDS guild_id;
DEFINE INDEX idx_activity_guild_name ON activity FIELDS guild_id, name UNIQUE;

-- Activity is managed by Guild (legacy relation, kept for compatibility)
DEFINE TABLE managed_by SCHEMAFULL TYPE RELATION FROM activity TO guild;

DEFINE TABLE timer SCHEMAFULL;
DEFINE FIELD person_id ON timer TYPE record<person>;
DEFINE FIELD activity_id ON timer TYPE record<activity>;
DEFINE FIELD reset_date ON timer TYPE datetime;
DEFINE FIELD enabled ON timer TYPE bool DEFAULT true;
DEFINE FIELD push ON timer TYPE bool DEFAULT false;
DEFINE FIELD created_on ON timer TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON timer TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_timer_person ON timer FIELDS person_id;
DEFINE INDEX idx_timer_activity ON timer FIELDS activity_id;
DEFINE INDEX idx_timer_person_activity ON timer FIELDS person_id, activity_id UNIQUE;

-- Person is tracked by Timer (legacy relation, kept for compatibility)
DEFINE TABLE tracked_by SCHEMAFULL TYPE RELATION FROM person TO timer;

-- Timer adheres to Activity (legacy relation, kept for compatibility)
DEFINE TABLE adheres_to SCHEMAFULL TYPE RELATION FROM timer TO activity;

-- Reminders
DEFINE TABLE reminder SCHEMAFULL;
DEFINE FIELD member_id ON reminder TYPE record<member>;
DEFINE FIELD person_id ON reminder TYPE option<record<person>>;
DEFINE FIELD title ON reminder TYPE string ASSERT string::len($value) <= 200;
DEFINE FIELD type ON reminder TYPE string ASSERT $value IN ["birthday", "anniversary", "custom"];
DEFINE FIELD date ON reminder TYPE datetime;
DEFINE FIELD recurring ON reminder TYPE bool DEFAULT true;
DEFINE FIELD notify_days_before ON reminder TYPE int DEFAULT 1 ASSERT $value >= 0 AND $value <= 30;
DEFINE FIELD created_on ON reminder TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON reminder TYPE datetime DEFAULT time::now();

DEFINE INDEX reminder_member ON reminder COLUMNS member_id;
DEFINE INDEX reminder_date ON reminder COLUMNS date;

-- =============================================================================
-- 5. USER PROFILES & INTERESTS
-- =============================================================================

DEFINE TABLE user_profile SCHEMAFULL;
DEFINE FIELD user ON user_profile TYPE record<user>;
DEFINE FIELD bio ON user_profile TYPE option<string>;
DEFINE FIELD tagline ON user_profile TYPE option<string>;
DEFINE FIELD languages ON user_profile TYPE option<array<string>>;
DEFINE FIELD timezone ON user_profile TYPE option<string>;
-- Location: lat/lng internal only, city/neighborhood shown to others
DEFINE FIELD location ON user_profile TYPE option<object>;
-- Visibility: guilds, public, private
DEFINE FIELD visibility ON user_profile TYPE string DEFAULT "guilds";
DEFINE FIELD last_active ON user_profile TYPE option<datetime>;
DEFINE FIELD created_on ON user_profile TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON user_profile TYPE datetime DEFAULT time::now();

DEFINE INDEX user_profile_user ON user_profile FIELDS user UNIQUE;

DEFINE TABLE interest SCHEMAFULL;
DEFINE FIELD name ON interest TYPE string;
DEFINE FIELD category ON interest TYPE string;
DEFINE FIELD icon ON interest TYPE option<string>;
DEFINE FIELD created_on ON interest TYPE datetime DEFAULT time::now();

DEFINE INDEX interest_name ON interest FIELDS name UNIQUE;
DEFINE INDEX interest_category ON interest FIELDS category;

-- User-Interest with teaching/learning intent
DEFINE TABLE has_interest SCHEMAFULL TYPE RELATION FROM user TO interest;
DEFINE FIELD level ON has_interest TYPE string DEFAULT "interested";
DEFINE FIELD wants_to_teach ON has_interest TYPE bool DEFAULT false;
DEFINE FIELD wants_to_learn ON has_interest TYPE bool DEFAULT false;
DEFINE FIELD intent ON has_interest TYPE option<string>;
DEFINE FIELD created_on ON has_interest TYPE datetime DEFAULT time::now();

-- =============================================================================
-- 6. AVAILABILITY & HANGOUTS
-- =============================================================================

DEFINE TABLE availability SCHEMAFULL;
DEFINE FIELD user ON availability TYPE record<user>;
DEFINE FIELD status ON availability TYPE string DEFAULT "available";
DEFINE FIELD start_time ON availability TYPE datetime;
DEFINE FIELD end_time ON availability TYPE datetime;
DEFINE FIELD location ON availability TYPE option<object>;
DEFINE FIELD hangout_type ON availability TYPE string DEFAULT "meet_anyone";
DEFINE FIELD activity_description ON availability TYPE option<string>;
DEFINE FIELD activity_venue ON availability TYPE option<string>;
DEFINE FIELD interest_id ON availability TYPE option<record<interest>>;
DEFINE FIELD max_people ON availability TYPE int DEFAULT 1;
DEFINE FIELD note ON availability TYPE option<string>;
DEFINE FIELD visibility ON availability TYPE string DEFAULT "guilds";
DEFINE FIELD expires_at ON availability TYPE datetime;
DEFINE FIELD created_on ON availability TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON availability TYPE datetime DEFAULT time::now();

DEFINE INDEX availability_user ON availability FIELDS user;
DEFINE INDEX availability_time ON availability FIELDS start_time, end_time;
DEFINE INDEX availability_type ON availability FIELDS hangout_type;

DEFINE TABLE hangout_request SCHEMAFULL;
DEFINE FIELD availability ON hangout_request TYPE record<availability>;
DEFINE FIELD requester ON hangout_request TYPE record<user>;
DEFINE FIELD note ON hangout_request TYPE string;
DEFINE FIELD status ON hangout_request TYPE string DEFAULT "pending";
DEFINE FIELD responded_on ON hangout_request TYPE option<datetime>;
DEFINE FIELD created_on ON hangout_request TYPE datetime DEFAULT time::now();

DEFINE TABLE hangout SCHEMAFULL;
DEFINE FIELD participants ON hangout TYPE array<record<user>>;
DEFINE FIELD availability ON hangout TYPE option<record<availability>>;
DEFINE FIELD hangout_type ON hangout TYPE string;
DEFINE FIELD activity_description ON hangout TYPE option<string>;
DEFINE FIELD scheduled_time ON hangout TYPE datetime;
DEFINE FIELD location ON hangout TYPE option<object>;
DEFINE FIELD is_support_session ON hangout TYPE bool DEFAULT false;
DEFINE FIELD status ON hangout TYPE string DEFAULT "scheduled";
DEFINE FIELD created_on ON hangout TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON hangout TYPE datetime DEFAULT time::now();

-- =============================================================================
-- 7. QUESTIONNAIRES & COMPATIBILITY
-- =============================================================================

DEFINE TABLE question SCHEMAFULL;
DEFINE FIELD text ON question TYPE string;
DEFINE FIELD category ON question TYPE string;
DEFINE FIELD options ON question TYPE array<object> FLEXIBLE;
DEFINE FIELD options.*.value ON question TYPE string;
DEFINE FIELD options.*.label ON question TYPE string;
DEFINE FIELD options.*.implicit_bias ON question TYPE float;
DEFINE FIELD is_dealbreaker_eligible ON question TYPE bool DEFAULT false;
DEFINE FIELD sort_order ON question TYPE int DEFAULT 0;
DEFINE FIELD active ON question TYPE bool DEFAULT true;
DEFINE FIELD guild_id ON question TYPE option<record<guild>>;
DEFINE FIELD created_by ON question TYPE option<record<member>>;
DEFINE FIELD created_on ON question TYPE datetime DEFAULT time::now();

DEFINE INDEX question_category ON question FIELDS category;
DEFINE INDEX question_active ON question FIELDS active;
DEFINE INDEX question_guild ON question FIELDS guild_id;

DEFINE TABLE answer SCHEMAFULL;
DEFINE FIELD user ON answer TYPE record<user>;
DEFINE FIELD question ON answer TYPE record<question>;
DEFINE FIELD selected_option ON answer TYPE string;
DEFINE FIELD acceptable_options ON answer TYPE array<string>;
DEFINE FIELD importance ON answer TYPE string DEFAULT "somewhat";
DEFINE FIELD is_dealbreaker ON answer TYPE bool DEFAULT false;
DEFINE FIELD alignment_weight ON answer TYPE float DEFAULT 0.5;
DEFINE FIELD yikes_options ON answer TYPE option<array<string>>;
DEFINE FIELD created_on ON answer TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON answer TYPE datetime DEFAULT time::now();

DEFINE INDEX answer_user_question ON answer FIELDS user, question UNIQUE;
DEFINE INDEX answer_user ON answer FIELDS user;

-- Guild values questionnaires
DEFINE TABLE guild_values SCHEMAFULL;
DEFINE FIELD guild_id ON guild_values TYPE record<guild>;
DEFINE FIELD name ON guild_values TYPE string;
DEFINE FIELD description ON guild_values TYPE option<string>;
DEFINE FIELD questions ON guild_values TYPE array<record<question>>;
DEFINE FIELD required ON guild_values TYPE bool DEFAULT false;
DEFINE FIELD created_by ON guild_values TYPE record<member>;
DEFINE FIELD created_on ON guild_values TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON guild_values TYPE datetime DEFAULT time::now();

DEFINE INDEX guild_values_guild ON guild_values FIELDS guild_id;

DEFINE TABLE guild_values_response SCHEMAFULL;
DEFINE FIELD guild_id ON guild_values_response TYPE record<guild>;
DEFINE FIELD user ON guild_values_response TYPE record<user>;
DEFINE FIELD values_id ON guild_values_response TYPE record<guild_values>;
DEFINE FIELD status ON guild_values_response TYPE string DEFAULT "pending";
DEFINE FIELD completed_on ON guild_values_response TYPE datetime;
DEFINE FIELD bias_score ON guild_values_response TYPE float DEFAULT 0.0;
DEFINE FIELD approved_by ON guild_values_response TYPE option<record<member>>;
DEFINE FIELD approved_on ON guild_values_response TYPE option<datetime>;
DEFINE FIELD rejected_by ON guild_values_response TYPE option<record<member>>;
DEFINE FIELD rejection_note ON guild_values_response TYPE option<string>;
DEFINE FIELD created_on ON guild_values_response TYPE datetime DEFAULT time::now();

DEFINE INDEX guild_values_response_user ON guild_values_response FIELDS user, values_id UNIQUE;
DEFINE INDEX guild_values_response_status ON guild_values_response FIELDS status;

-- User bias tracking (internal only)
DEFINE TABLE user_bias_profile SCHEMAFULL;
DEFINE FIELD user ON user_bias_profile TYPE record<user>;
DEFINE FIELD accumulated_bias ON user_bias_profile TYPE float DEFAULT 0.0;
DEFINE FIELD answer_count ON user_bias_profile TYPE int DEFAULT 0;
DEFINE FIELD status ON user_bias_profile TYPE string DEFAULT "normal";
DEFINE FIELD last_updated ON user_bias_profile TYPE datetime DEFAULT time::now();

DEFINE INDEX user_bias_profile_user ON user_bias_profile FIELDS user UNIQUE;
DEFINE INDEX user_bias_profile_status ON user_bias_profile FIELDS status;

-- =============================================================================
-- 8. ADVENTURES (Multi-Day, Multi-Event Coordination)
-- =============================================================================

DEFINE TABLE adventure SCHEMAFULL;
DEFINE FIELD guild_id ON adventure TYPE option<record<guild>>;
DEFINE FIELD title ON adventure TYPE string ASSERT string::len($value) <= 100;
DEFINE FIELD description ON adventure TYPE option<string> ASSERT string::len($value) <= 1000;
DEFINE FIELD start_date ON adventure TYPE datetime;
DEFINE FIELD end_date ON adventure TYPE datetime;
DEFINE FIELD status ON adventure TYPE string DEFAULT "idea";
DEFINE FIELD visibility ON adventure TYPE string DEFAULT "guilds";
DEFINE FIELD created_by_id ON adventure TYPE record<user>;
DEFINE FIELD values_required ON adventure TYPE bool DEFAULT false;
DEFINE FIELD values_questions ON adventure TYPE option<array<string>>;
DEFINE FIELD budget_min ON adventure TYPE option<float>;
DEFINE FIELD budget_max ON adventure TYPE option<float>;
DEFINE FIELD currency ON adventure TYPE option<string>;
DEFINE FIELD cover_image ON adventure TYPE option<string>;
DEFINE FIELD voting_open ON adventure TYPE bool DEFAULT true;
DEFINE FIELD voting_deadline ON adventure TYPE option<datetime>;
DEFINE FIELD created_on ON adventure TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON adventure TYPE datetime DEFAULT time::now();

DEFINE INDEX adventure_guild ON adventure FIELDS guild_id;
DEFINE INDEX adventure_status ON adventure FIELDS status;
DEFINE INDEX adventure_visibility ON adventure FIELDS visibility;
DEFINE INDEX adventure_dates ON adventure FIELDS start_date, end_date;
DEFINE INDEX adventure_created ON adventure FIELDS created_on;

DEFINE TABLE adventure_participant SCHEMAFULL;
DEFINE FIELD adventure_id ON adventure_participant TYPE record<adventure>;
DEFINE FIELD user_id ON adventure_participant TYPE record<user>;
DEFINE FIELD member_id ON adventure_participant TYPE option<record<member>>;
DEFINE FIELD role ON adventure_participant TYPE string DEFAULT "participant";
DEFINE FIELD status ON adventure_participant TYPE string DEFAULT "interested";
DEFINE FIELD joined_on ON adventure_participant TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON adventure_participant TYPE datetime DEFAULT time::now();

DEFINE INDEX adventure_participant_adventure ON adventure_participant FIELDS adventure_id;
DEFINE INDEX adventure_participant_user ON adventure_participant FIELDS user_id;
DEFINE INDEX adventure_participant_unique ON adventure_participant FIELDS adventure_id, user_id UNIQUE;

DEFINE TABLE destination SCHEMAFULL;
DEFINE FIELD adventure_id ON destination TYPE record<adventure>;
DEFINE FIELD name ON destination TYPE string ASSERT string::len($value) <= 100;
DEFINE FIELD description ON destination TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD proposed_by_id ON destination TYPE record<user>;
DEFINE FIELD city ON destination TYPE option<string>;
DEFINE FIELD country ON destination TYPE option<string>;
DEFINE FIELD region ON destination TYPE option<string>;
DEFINE FIELD lat ON destination TYPE option<float>;
DEFINE FIELD lng ON destination TYPE option<float>;
DEFINE FIELD order_index ON destination TYPE int DEFAULT 0;
DEFINE FIELD estimated_cost_min ON destination TYPE option<float>;
DEFINE FIELD estimated_cost_max ON destination TYPE option<float>;
DEFINE FIELD info_url ON destination TYPE option<string>;
DEFINE FIELD image_url ON destination TYPE option<string>;
DEFINE FIELD created_on ON destination TYPE datetime DEFAULT time::now();

DEFINE INDEX destination_adventure ON destination FIELDS adventure_id;
DEFINE INDEX destination_proposed_by ON destination FIELDS proposed_by_id;

DEFINE TABLE destination_vote SCHEMAFULL;
DEFINE FIELD destination_id ON destination_vote TYPE record<destination>;
DEFINE FIELD user_id ON destination_vote TYPE record<user>;
DEFINE FIELD rank ON destination_vote TYPE int ASSERT $value >= 1;
DEFINE FIELD veto ON destination_vote TYPE bool DEFAULT false;
DEFINE FIELD reason ON destination_vote TYPE option<string>;
DEFINE FIELD created_on ON destination_vote TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON destination_vote TYPE datetime DEFAULT time::now();

DEFINE INDEX destination_vote_dest ON destination_vote FIELDS destination_id;
DEFINE INDEX destination_vote_user ON destination_vote FIELDS user_id;
DEFINE INDEX destination_vote_unique ON destination_vote FIELDS destination_id, user_id UNIQUE;

DEFINE TABLE adventure_activity SCHEMAFULL;
DEFINE FIELD adventure_id ON adventure_activity TYPE record<adventure>;
DEFINE FIELD destination_id ON adventure_activity TYPE option<record<destination>>;
DEFINE FIELD title ON adventure_activity TYPE string ASSERT string::len($value) <= 100;
DEFINE FIELD description ON adventure_activity TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD scheduled_date ON adventure_activity TYPE option<datetime>;
DEFINE FIELD scheduled_time ON adventure_activity TYPE option<string>;
DEFINE FIELD duration_minutes ON adventure_activity TYPE option<int>;
DEFINE FIELD location ON adventure_activity TYPE option<string>;
DEFINE FIELD estimated_cost ON adventure_activity TYPE option<float>;
DEFINE FIELD booking_url ON adventure_activity TYPE option<string>;
DEFINE FIELD proposed_by_id ON adventure_activity TYPE record<user>;
DEFINE FIELD is_confirmed ON adventure_activity TYPE bool DEFAULT false;
DEFINE FIELD created_on ON adventure_activity TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON adventure_activity TYPE datetime DEFAULT time::now();

DEFINE INDEX adventure_activity_adventure ON adventure_activity FIELDS adventure_id;
DEFINE INDEX adventure_activity_dest ON adventure_activity FIELDS destination_id;
DEFINE INDEX adventure_activity_date ON adventure_activity FIELDS scheduled_date;
DEFINE INDEX adventure_activity_confirmed ON adventure_activity FIELDS is_confirmed;

DEFINE TABLE activity_vote SCHEMAFULL;
DEFINE FIELD activity_id ON activity_vote TYPE record<adventure_activity>;
DEFINE FIELD user_id ON activity_vote TYPE record<user>;
DEFINE FIELD interested ON activity_vote TYPE bool;
DEFINE FIELD created_on ON activity_vote TYPE datetime DEFAULT time::now();

DEFINE INDEX activity_vote_activity ON activity_vote FIELDS activity_id;
DEFINE INDEX activity_vote_user ON activity_vote FIELDS user_id;
DEFINE INDEX activity_vote_unique ON activity_vote FIELDS activity_id, user_id UNIQUE;

-- =============================================================================
-- 9. EVENTS & RSVPs (Can be standalone or nested in Adventure)
-- =============================================================================

DEFINE TABLE event SCHEMAFULL;
DEFINE FIELD guild_id ON event TYPE option<record<guild>>;
DEFINE FIELD adventure_id ON event TYPE option<record<adventure>>;
DEFINE FIELD order_in_adventure ON event TYPE option<int>;
DEFINE FIELD title ON event TYPE string ASSERT string::len($value) <= 200;
DEFINE FIELD description ON event TYPE option<string> ASSERT string::len($value) <= 2000;
DEFINE FIELD location ON event TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD location_data ON event TYPE option<object>;
DEFINE FIELD starts_at ON event TYPE datetime;
DEFINE FIELD ends_at ON event TYPE option<datetime>;
DEFINE FIELD all_day ON event TYPE bool DEFAULT false;
DEFINE FIELD created_by ON event TYPE record<member>;
DEFINE FIELD template ON event TYPE string DEFAULT "casual";
DEFINE FIELD visibility ON event TYPE string DEFAULT "guilds";
DEFINE FIELD max_attendees ON event TYPE option<int>;
DEFINE FIELD waitlist_enabled ON event TYPE bool DEFAULT false;
DEFINE FIELD requires_approval ON event TYPE bool DEFAULT false;
DEFINE FIELD allow_plus_ones ON event TYPE bool DEFAULT false;
DEFINE FIELD max_plus_ones ON event TYPE int DEFAULT 1;
DEFINE FIELD cover_image ON event TYPE option<string>;
DEFINE FIELD theme_color ON event TYPE option<string>;
DEFINE FIELD values_required ON event TYPE bool DEFAULT false;
DEFINE FIELD values_questions ON event TYPE option<array<string>>;
DEFINE FIELD auto_approve_aligned ON event TYPE bool DEFAULT false;
DEFINE FIELD yikes_threshold ON event TYPE int DEFAULT 2;
DEFINE FIELD is_support_event ON event TYPE bool DEFAULT false;
DEFINE FIELD status ON event TYPE string DEFAULT "published";
DEFINE FIELD created_on ON event TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON event TYPE datetime DEFAULT time::now();

DEFINE INDEX event_guild ON event COLUMNS guild_id;
DEFINE INDEX event_adventure ON event COLUMNS adventure_id;
DEFINE INDEX event_starts ON event COLUMNS starts_at;
DEFINE INDEX event_visibility ON event FIELDS visibility;
DEFINE INDEX event_status ON event FIELDS status;
DEFINE INDEX event_template ON event FIELDS template;

-- Event participant with confirmation tracking
DEFINE TABLE event_participant SCHEMAFULL TYPE RELATION FROM member TO event;
DEFINE FIELD rsvp ON event_participant TYPE string DEFAULT "yes";
DEFINE FIELD rsvp_note ON event_participant TYPE option<string>;
DEFINE FIELD responded_on ON event_participant TYPE option<datetime>;
DEFINE FIELD completion_confirmed ON event_participant TYPE option<datetime>;
DEFINE FIELD checkin_time ON event_participant TYPE option<datetime>;
DEFINE FIELD helpfulness_rating ON event_participant TYPE option<string>;
DEFINE FIELD helpfulness_tags ON event_participant TYPE option<array<string>>;
DEFINE FIELD created_on ON event_participant TYPE datetime DEFAULT time::now();

-- Legacy RSVP relation (for backwards compatibility)
DEFINE TABLE rsvp SCHEMAFULL TYPE RELATION FROM member TO event;
DEFINE FIELD status ON rsvp TYPE string ASSERT $value IN ["yes", "no", "maybe", "pending"];
DEFINE FIELD responded_on ON rsvp TYPE option<datetime>;

-- =============================================================================
-- 10. EVENT HOSTS & ROLES
-- =============================================================================

DEFINE TABLE event_host SCHEMAFULL;
DEFINE FIELD event_id ON event_host TYPE record<event>;
DEFINE FIELD user_id ON event_host TYPE record<user>;
DEFINE FIELD role ON event_host TYPE string DEFAULT "primary";
DEFINE FIELD added_on ON event_host TYPE datetime DEFAULT time::now();
DEFINE FIELD added_by ON event_host TYPE record<user>;

DEFINE INDEX event_host_event ON event_host FIELDS event_id;
DEFINE INDEX event_host_user ON event_host FIELDS user_id;
DEFINE INDEX event_host_unique ON event_host FIELDS event_id, user_id UNIQUE;

DEFINE TABLE event_rsvp SCHEMAFULL;
DEFINE FIELD event_id ON event_rsvp TYPE record<event>;
DEFINE FIELD user_id ON event_rsvp TYPE record<user>;
DEFINE FIELD status ON event_rsvp TYPE string DEFAULT "pending";
DEFINE FIELD rsvp_type ON event_rsvp TYPE string DEFAULT "going";
DEFINE FIELD values_aligned ON event_rsvp TYPE bool DEFAULT true;
DEFINE FIELD alignment_score ON event_rsvp TYPE float DEFAULT 0.0;
DEFINE FIELD yikes_count ON event_rsvp TYPE int DEFAULT 0;
DEFINE FIELD waiting_reason ON event_rsvp TYPE option<string>;
DEFINE FIELD host_note ON event_rsvp TYPE option<string>;
DEFINE FIELD responded_by ON event_rsvp TYPE option<record<user>>;
DEFINE FIELD responded_on ON event_rsvp TYPE option<datetime>;
DEFINE FIELD requested_on ON event_rsvp TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON event_rsvp TYPE datetime DEFAULT time::now();
DEFINE FIELD plus_ones ON event_rsvp TYPE int DEFAULT 0;
DEFINE FIELD plus_one_names ON event_rsvp TYPE option<array<string>>;
DEFINE FIELD completion_confirmed ON event_rsvp TYPE option<datetime>;
DEFINE FIELD checkin_time ON event_rsvp TYPE option<datetime>;
DEFINE FIELD helpfulness_rating ON event_rsvp TYPE option<string>;
DEFINE FIELD helpfulness_tags ON event_rsvp TYPE option<array<string>>;

DEFINE INDEX event_rsvp_event ON event_rsvp FIELDS event_id;
DEFINE INDEX event_rsvp_user ON event_rsvp FIELDS user_id;
DEFINE INDEX event_rsvp_status ON event_rsvp FIELDS status;
DEFINE INDEX event_rsvp_unique ON event_rsvp FIELDS event_id, user_id UNIQUE;

DEFINE TABLE event_role SCHEMAFULL;
DEFINE FIELD event_id ON event_role TYPE record<event>;
DEFINE FIELD name ON event_role TYPE string ASSERT string::len($value) <= 50;
DEFINE FIELD description ON event_role TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD max_slots ON event_role TYPE int DEFAULT 1;
DEFINE FIELD filled_slots ON event_role TYPE int DEFAULT 0;
DEFINE FIELD is_default ON event_role TYPE bool DEFAULT false;
DEFINE FIELD sort_order ON event_role TYPE int DEFAULT 0;
DEFINE FIELD suggested_interests ON event_role TYPE option<array<string>>;
DEFINE FIELD created_by ON event_role TYPE record<user>;
DEFINE FIELD created_on ON event_role TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON event_role TYPE datetime DEFAULT time::now();

DEFINE INDEX event_role_event ON event_role FIELDS event_id;
DEFINE INDEX event_role_default ON event_role FIELDS event_id, is_default;

DEFINE TABLE event_role_assignment SCHEMAFULL;
DEFINE FIELD event_id ON event_role_assignment TYPE record<event>;
DEFINE FIELD role_id ON event_role_assignment TYPE record<event_role>;
DEFINE FIELD user_id ON event_role_assignment TYPE record<user>;
DEFINE FIELD note ON event_role_assignment TYPE option<string> ASSERT string::len($value) <= 200;
DEFINE FIELD status ON event_role_assignment TYPE string DEFAULT "confirmed";
DEFINE FIELD assigned_on ON event_role_assignment TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON event_role_assignment TYPE datetime DEFAULT time::now();

DEFINE INDEX role_assignment_event ON event_role_assignment FIELDS event_id;
DEFINE INDEX role_assignment_role ON event_role_assignment FIELDS role_id;
DEFINE INDEX role_assignment_user ON event_role_assignment FIELDS user_id;
DEFINE INDEX role_assignment_unique ON event_role_assignment FIELDS role_id, user_id UNIQUE;

-- =============================================================================
-- 11. FORUMS (for Adventures and Events)
-- =============================================================================

DEFINE TABLE forum SCHEMAFULL;
DEFINE FIELD adventure_id ON forum TYPE option<record<adventure>>;
DEFINE FIELD event_id ON forum TYPE option<record<event>>;
DEFINE FIELD created_on ON forum TYPE datetime DEFAULT time::now();

DEFINE INDEX forum_adventure ON forum FIELDS adventure_id;
DEFINE INDEX forum_event ON forum FIELDS event_id;

DEFINE TABLE forum_post SCHEMAFULL;
DEFINE FIELD forum_id ON forum_post TYPE record<forum>;
DEFINE FIELD author_id ON forum_post TYPE record<user>;
DEFINE FIELD content ON forum_post TYPE string ASSERT string::len($value) <= 5000;
DEFINE FIELD reply_to_id ON forum_post TYPE option<record<forum_post>>;
DEFINE FIELD is_pinned ON forum_post TYPE bool DEFAULT false;
DEFINE FIELD reactions ON forum_post TYPE option<object>;
DEFINE FIELD mentions ON forum_post TYPE option<array<record<user>>>;
DEFINE FIELD created_on ON forum_post TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON forum_post TYPE datetime DEFAULT time::now();
DEFINE FIELD deleted_on ON forum_post TYPE option<datetime>;

DEFINE INDEX forum_post_forum ON forum_post FIELDS forum_id;
DEFINE INDEX forum_post_author ON forum_post FIELDS author_id;
DEFINE INDEX forum_post_reply_to ON forum_post FIELDS reply_to_id;
DEFINE INDEX forum_post_pinned ON forum_post FIELDS forum_id, is_pinned;
DEFINE INDEX forum_post_created ON forum_post FIELDS created_on;

-- =============================================================================
-- 12. MATCHING POOLS (Donut-style)
-- =============================================================================

DEFINE TABLE matching_pool SCHEMAFULL;
DEFINE FIELD guild_id ON matching_pool TYPE record<guild>;
DEFINE FIELD name ON matching_pool TYPE string ASSERT string::len($value) <= 100;
DEFINE FIELD description ON matching_pool TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD frequency ON matching_pool TYPE string DEFAULT "weekly";
DEFINE FIELD match_size ON matching_pool TYPE int DEFAULT 2 ASSERT $value >= 2 AND $value <= 6;
DEFINE FIELD activity_suggestion ON matching_pool TYPE option<string> ASSERT string::len($value) <= 200;
DEFINE FIELD next_match_on ON matching_pool TYPE datetime;
DEFINE FIELD last_match_on ON matching_pool TYPE option<datetime>;
DEFINE FIELD active ON matching_pool TYPE bool DEFAULT true;
DEFINE FIELD created_by ON matching_pool TYPE record<user>;
DEFINE FIELD created_on ON matching_pool TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON matching_pool TYPE datetime DEFAULT time::now();

DEFINE INDEX matching_pool_guild ON matching_pool FIELDS guild_id;
DEFINE INDEX matching_pool_active ON matching_pool FIELDS active;
DEFINE INDEX matching_pool_next_match ON matching_pool FIELDS next_match_on;

DEFINE TABLE pool_member SCHEMAFULL;
DEFINE FIELD pool_id ON pool_member TYPE record<matching_pool>;
DEFINE FIELD member_id ON pool_member TYPE record<member>;
DEFINE FIELD user_id ON pool_member TYPE record<user>;
DEFINE FIELD active ON pool_member TYPE bool DEFAULT true;
DEFINE FIELD excluded_members ON pool_member TYPE option<array<record<member>>>;
DEFINE FIELD joined_on ON pool_member TYPE datetime DEFAULT time::now();

DEFINE INDEX pool_member_pool ON pool_member FIELDS pool_id;
DEFINE INDEX pool_member_user ON pool_member FIELDS user_id;
DEFINE INDEX pool_member_active ON pool_member FIELDS pool_id, active;
DEFINE INDEX pool_member_unique ON pool_member FIELDS pool_id, member_id UNIQUE;

DEFINE TABLE match_result SCHEMAFULL;
DEFINE FIELD pool_id ON match_result TYPE record<matching_pool>;
DEFINE FIELD members ON match_result TYPE array<record<member>>;
DEFINE FIELD member_user_ids ON match_result TYPE array<record<user>>;
DEFINE FIELD status ON match_result TYPE string DEFAULT "pending";
DEFINE FIELD match_round ON match_result TYPE string;
DEFINE FIELD scheduled_event ON match_result TYPE option<record<event>>;
DEFINE FIELD scheduled_time ON match_result TYPE option<datetime>;
DEFINE FIELD created_on ON match_result TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON match_result TYPE datetime DEFAULT time::now();

DEFINE INDEX match_result_pool ON match_result FIELDS pool_id;
DEFINE INDEX match_result_round ON match_result FIELDS pool_id, match_round;
DEFINE INDEX match_result_status ON match_result FIELDS status;
DEFINE INDEX match_result_created ON match_result FIELDS created_on;

-- =============================================================================
-- 13. RIDESHARES (Attached to Events or Adventures)
-- =============================================================================

DEFINE TABLE rideshare SCHEMAFULL;
DEFINE FIELD event_id ON rideshare TYPE option<record<event>>;
DEFINE FIELD adventure_id ON rideshare TYPE option<record<adventure>>;
DEFINE FIELD driver_id ON rideshare TYPE record<user>;
DEFINE FIELD title ON rideshare TYPE string ASSERT string::len($value) <= 100;
DEFINE FIELD description ON rideshare TYPE option<string> ASSERT string::len($value) <= 500;
DEFINE FIELD origin ON rideshare TYPE object;
DEFINE FIELD destination ON rideshare TYPE object;
DEFINE FIELD departure_time ON rideshare TYPE datetime;
DEFINE FIELD arrival_time ON rideshare TYPE option<datetime>;
DEFINE FIELD seats_total ON rideshare TYPE int DEFAULT 4;
DEFINE FIELD seats_available ON rideshare TYPE int DEFAULT 4;
DEFINE FIELD status ON rideshare TYPE string DEFAULT "open";
DEFINE FIELD trust_required ON rideshare TYPE bool DEFAULT true;
DEFINE FIELD created_on ON rideshare TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON rideshare TYPE datetime DEFAULT time::now();

DEFINE INDEX rideshare_event ON rideshare FIELDS event_id;
DEFINE INDEX rideshare_adventure ON rideshare FIELDS adventure_id;
DEFINE INDEX rideshare_driver ON rideshare FIELDS driver_id;
DEFINE INDEX rideshare_status ON rideshare FIELDS status;
DEFINE INDEX rideshare_departure ON rideshare FIELDS departure_time;

DEFINE TABLE rideshare_segment SCHEMAFULL;
DEFINE FIELD rideshare_id ON rideshare_segment TYPE record<rideshare>;
DEFINE FIELD sequence_order ON rideshare_segment TYPE int DEFAULT 0;
DEFINE FIELD pickup_point ON rideshare_segment TYPE object;
DEFINE FIELD dropoff_point ON rideshare_segment TYPE object;
DEFINE FIELD estimated_minutes ON rideshare_segment TYPE option<int>;
DEFINE FIELD notes ON rideshare_segment TYPE option<string>;

DEFINE INDEX segment_rideshare ON rideshare_segment FIELDS rideshare_id;
DEFINE INDEX segment_order ON rideshare_segment FIELDS rideshare_id, sequence_order;

DEFINE TABLE rideshare_seat SCHEMAFULL;
DEFINE FIELD rideshare_id ON rideshare_seat TYPE record<rideshare>;
DEFINE FIELD passenger_id ON rideshare_seat TYPE record<user>;
DEFINE FIELD status ON rideshare_seat TYPE string DEFAULT "requested";
DEFINE FIELD pickup_segment_id ON rideshare_seat TYPE option<record<rideshare_segment>>;
DEFINE FIELD dropoff_segment_id ON rideshare_seat TYPE option<record<rideshare_segment>>;
DEFINE FIELD requested_on ON rideshare_seat TYPE datetime DEFAULT time::now();
DEFINE FIELD confirmed_on ON rideshare_seat TYPE option<datetime>;
DEFINE FIELD notes ON rideshare_seat TYPE option<string>;

DEFINE INDEX seat_rideshare ON rideshare_seat FIELDS rideshare_id;
DEFINE INDEX seat_passenger ON rideshare_seat FIELDS passenger_id;
DEFINE INDEX seat_status ON rideshare_seat FIELDS status;
DEFINE INDEX seat_unique ON rideshare_seat FIELDS rideshare_id, passenger_id UNIQUE;

-- =============================================================================
-- 14. TRUST & IRL VERIFICATION
-- =============================================================================

DEFINE TABLE trust_relation SCHEMAFULL;
DEFINE FIELD user_a_id ON trust_relation TYPE record<user>;
DEFINE FIELD user_b_id ON trust_relation TYPE record<user>;
DEFINE FIELD status ON trust_relation TYPE string DEFAULT "active";
DEFINE FIELD created_on ON trust_relation TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON trust_relation TYPE datetime DEFAULT time::now();

DEFINE INDEX trust_user_a ON trust_relation FIELDS user_a_id;
DEFINE INDEX trust_user_b ON trust_relation FIELDS user_b_id;
DEFINE INDEX trust_pair ON trust_relation FIELDS user_a_id, user_b_id UNIQUE;
DEFINE INDEX trust_status ON trust_relation FIELDS status;

DEFINE TABLE irl_verification SCHEMAFULL;
DEFINE FIELD user_a_id ON irl_verification TYPE record<user>;
DEFINE FIELD user_b_id ON irl_verification TYPE record<user>;
DEFINE FIELD verified_on ON irl_verification TYPE datetime DEFAULT time::now();
DEFINE FIELD context ON irl_verification TYPE string DEFAULT "other";
DEFINE FIELD reference_id ON irl_verification TYPE option<string>;
DEFINE FIELD user_a_confirmed ON irl_verification TYPE bool DEFAULT false;
DEFINE FIELD user_b_confirmed ON irl_verification TYPE bool DEFAULT false;
DEFINE FIELD user_a_confirmed_on ON irl_verification TYPE option<datetime>;
DEFINE FIELD user_b_confirmed_on ON irl_verification TYPE option<datetime>;

DEFINE INDEX irl_user_a ON irl_verification FIELDS user_a_id;
DEFINE INDEX irl_user_b ON irl_verification FIELDS user_b_id;
DEFINE INDEX irl_pair ON irl_verification FIELDS user_a_id, user_b_id UNIQUE;
DEFINE INDEX irl_context ON irl_verification FIELDS context;

-- =============================================================================
-- 15. REVIEWS & REPUTATION
-- =============================================================================

DEFINE TABLE review SCHEMAFULL TYPE RELATION FROM member TO member;
DEFINE FIELD context ON review TYPE string;
DEFINE FIELD reference_id ON review TYPE option<string>;
DEFINE FIELD would_meet_again ON review TYPE bool DEFAULT true;
DEFINE FIELD positive_tags ON review TYPE option<array<string>>;
DEFINE FIELD improvement_tags ON review TYPE option<array<string>>;
DEFINE FIELD private_note ON review TYPE option<string>;
DEFINE FIELD created_on ON review TYPE datetime DEFAULT time::now();

DEFINE INDEX review_reviewee ON review FIELDS out;
DEFINE INDEX review_context ON review FIELDS context;

-- =============================================================================
-- 16. RESONANCE SCORING
-- =============================================================================

DEFINE TABLE resonance_ledger SCHEMAFULL;
DEFINE FIELD user ON resonance_ledger TYPE record<user>;
DEFINE FIELD stat ON resonance_ledger TYPE string;
DEFINE FIELD points ON resonance_ledger TYPE int;
DEFINE FIELD source_object_id ON resonance_ledger TYPE string;
DEFINE FIELD reason_code ON resonance_ledger TYPE string;
DEFINE FIELD created_on ON resonance_ledger TYPE datetime DEFAULT time::now();

DEFINE INDEX resonance_user ON resonance_ledger FIELDS user;
DEFINE INDEX resonance_stat ON resonance_ledger FIELDS stat;
DEFINE INDEX resonance_source ON resonance_ledger FIELDS source_object_id;
DEFINE INDEX resonance_unique ON resonance_ledger FIELDS user, stat, source_object_id UNIQUE;

DEFINE TABLE resonance_score SCHEMAFULL;
DEFINE FIELD user ON resonance_score TYPE record<user>;
DEFINE FIELD total ON resonance_score TYPE int DEFAULT 0;
DEFINE FIELD questing ON resonance_score TYPE int DEFAULT 0;
DEFINE FIELD mana ON resonance_score TYPE int DEFAULT 0;
DEFINE FIELD wayfinder ON resonance_score TYPE int DEFAULT 0;
DEFINE FIELD attunement ON resonance_score TYPE int DEFAULT 0;
DEFINE FIELD nexus ON resonance_score TYPE int DEFAULT 0;
DEFINE FIELD last_calculated ON resonance_score TYPE datetime DEFAULT time::now();

DEFINE INDEX resonance_score_user ON resonance_score FIELDS user UNIQUE;

DEFINE TABLE resonance_daily_cap SCHEMAFULL;
DEFINE FIELD user ON resonance_daily_cap TYPE record<user>;
DEFINE FIELD date ON resonance_daily_cap TYPE string;
DEFINE FIELD questing_earned ON resonance_daily_cap TYPE int DEFAULT 0;
DEFINE FIELD mana_earned ON resonance_daily_cap TYPE int DEFAULT 0;
DEFINE FIELD wayfinder_earned ON resonance_daily_cap TYPE int DEFAULT 0;
DEFINE FIELD attunement_earned ON resonance_daily_cap TYPE int DEFAULT 0;

DEFINE INDEX daily_cap_user_date ON resonance_daily_cap FIELDS user, date UNIQUE;

DEFINE TABLE support_pair_count SCHEMAFULL;
DEFINE FIELD helper ON support_pair_count TYPE record<user>;
DEFINE FIELD receiver ON support_pair_count TYPE record<user>;
DEFINE FIELD count ON support_pair_count TYPE int DEFAULT 0;
DEFINE FIELD last_session ON support_pair_count TYPE datetime;

DEFINE INDEX support_pair ON support_pair_count FIELDS helper, receiver UNIQUE;

-- =============================================================================
-- 17. NUDGE SYSTEM
-- =============================================================================

DEFINE TABLE nudge SCHEMAFULL;
DEFINE FIELD user_id ON nudge TYPE record<user>;
DEFINE FIELD type ON nudge TYPE string;
DEFINE FIELD channel ON nudge TYPE string DEFAULT "sse";
DEFINE FIELD title ON nudge TYPE string ASSERT string::len($value) <= 100;
DEFINE FIELD message ON nudge TYPE string ASSERT string::len($value) <= 500;
DEFINE FIELD data ON nudge TYPE option<object>;
DEFINE FIELD sent_at ON nudge TYPE datetime DEFAULT time::now();
DEFINE FIELD read_at ON nudge TYPE option<datetime>;
DEFINE FIELD acted_at ON nudge TYPE option<datetime>;
DEFINE FIELD expires_at ON nudge TYPE option<datetime>;

DEFINE INDEX nudge_user ON nudge FIELDS user_id;
DEFINE INDEX nudge_type ON nudge FIELDS type;
DEFINE INDEX nudge_user_type ON nudge FIELDS user_id, type;
DEFINE INDEX nudge_unread ON nudge FIELDS user_id, read_at;
DEFINE INDEX nudge_sent ON nudge FIELDS sent_at;
DEFINE INDEX nudge_expires ON nudge FIELDS expires_at;

DEFINE TABLE nudge_history SCHEMAFULL;
DEFINE FIELD user_id ON nudge_history TYPE record<user>;
DEFINE FIELD type ON nudge_history TYPE string;
DEFINE FIELD target_id ON nudge_history TYPE string;
DEFINE FIELD sent_count ON nudge_history TYPE int DEFAULT 1;
DEFINE FIELD last_sent ON nudge_history TYPE datetime DEFAULT time::now();
DEFINE FIELD first_sent ON nudge_history TYPE datetime DEFAULT time::now();

DEFINE INDEX nudge_history_user ON nudge_history FIELDS user_id;
DEFINE INDEX nudge_history_lookup ON nudge_history FIELDS user_id, type, target_id UNIQUE;
DEFINE INDEX nudge_history_last_sent ON nudge_history FIELDS last_sent;

DEFINE TABLE nudge_preference SCHEMAFULL;
DEFINE FIELD user_id ON nudge_preference TYPE record<user>;
DEFINE FIELD type ON nudge_preference TYPE string;
DEFINE FIELD enabled ON nudge_preference TYPE bool DEFAULT true;
DEFINE FIELD channel ON nudge_preference TYPE option<string>;
DEFINE FIELD cooldown_seconds ON nudge_preference TYPE option<int>;
DEFINE FIELD created_on ON nudge_preference TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON nudge_preference TYPE datetime DEFAULT time::now();

DEFINE INDEX nudge_preference_user ON nudge_preference FIELDS user_id;
DEFINE INDEX nudge_preference_lookup ON nudge_preference FIELDS user_id, type UNIQUE;

DEFINE TABLE nudge_settings SCHEMAFULL;
DEFINE FIELD user_id ON nudge_settings TYPE record<user>;
DEFINE FIELD enabled ON nudge_settings TYPE bool DEFAULT true;
DEFINE FIELD quiet_start_hour ON nudge_settings TYPE option<int>;
DEFINE FIELD quiet_end_hour ON nudge_settings TYPE option<int>;
DEFINE FIELD timezone ON nudge_settings TYPE option<string>;
DEFINE FIELD global_cooldown_seconds ON nudge_settings TYPE int DEFAULT 3600;
DEFINE FIELD created_on ON nudge_settings TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON nudge_settings TYPE datetime DEFAULT time::now();

DEFINE INDEX nudge_settings_user ON nudge_settings FIELDS user_id UNIQUE;

-- =============================================================================
-- 18. MODERATION & SAFETY
-- =============================================================================

DEFINE TABLE user_flag SCHEMAFULL;
DEFINE FIELD reporter ON user_flag TYPE record<user>;
DEFINE FIELD reported_user ON user_flag TYPE record<user>;
DEFINE FIELD category ON user_flag TYPE string;
DEFINE FIELD description ON user_flag TYPE option<string>;
DEFINE FIELD status ON user_flag TYPE string DEFAULT "pending";
DEFINE FIELD reviewed_by ON user_flag TYPE option<record<user>>;
DEFINE FIELD reviewed_on ON user_flag TYPE option<datetime>;
DEFINE FIELD action_taken ON user_flag TYPE option<string>;
DEFINE FIELD created_on ON user_flag TYPE datetime DEFAULT time::now();

DEFINE INDEX flag_reported ON user_flag FIELDS reported_user;
DEFINE INDEX flag_status ON user_flag FIELDS status;

DEFINE TABLE user_block SCHEMAFULL;
DEFINE FIELD blocker ON user_block TYPE record<user>;
DEFINE FIELD blocked ON user_block TYPE record<user>;
DEFINE FIELD created_on ON user_block TYPE datetime DEFAULT time::now();

DEFINE INDEX block_blocker ON user_block FIELDS blocker;
DEFINE INDEX block_pair ON user_block FIELDS blocker, blocked UNIQUE;

DEFINE TABLE user_moderation SCHEMAFULL;
DEFINE FIELD user ON user_moderation TYPE record<user>;
DEFINE FIELD level ON user_moderation TYPE int DEFAULT 0;
DEFINE FIELD reason ON user_moderation TYPE option<string>;
DEFINE FIELD expires_on ON user_moderation TYPE option<datetime>;
DEFINE FIELD created_on ON user_moderation TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON user_moderation TYPE datetime DEFAULT time::now();

DEFINE INDEX moderation_user ON user_moderation FIELDS user UNIQUE;
DEFINE INDEX moderation_level ON user_moderation FIELDS level;

DEFINE TABLE report SCHEMAFULL;
DEFINE FIELD reporter_user_id ON report TYPE record<user>;
DEFINE FIELD reported_user_id ON report TYPE record<user>;
DEFINE FIELD guild_id ON report TYPE option<record<guild>>;
DEFINE FIELD category ON report TYPE string
    ASSERT $value IN ["spam", "harassment", "hate_speech", "inappropriate_content", "made_uncomfortable", "other"];
DEFINE FIELD description ON report TYPE option<string>;
DEFINE FIELD content_type ON report TYPE option<string>;
DEFINE FIELD content_id ON report TYPE option<string>;
DEFINE FIELD status ON report TYPE string DEFAULT "pending"
    ASSERT $value IN ["pending", "reviewed", "resolved", "dismissed"];
DEFINE FIELD reviewed_by_id ON report TYPE option<record<user>>;
DEFINE FIELD review_notes ON report TYPE option<string>;
DEFINE FIELD action_taken ON report TYPE option<string>;
DEFINE FIELD created_on ON report TYPE datetime DEFAULT time::now();
DEFINE FIELD reviewed_on ON report TYPE option<datetime>;
DEFINE FIELD resolved_on ON report TYPE option<datetime>;

DEFINE INDEX idx_report_reporter ON report FIELDS reporter_user_id;
DEFINE INDEX idx_report_reported ON report FIELDS reported_user_id;
DEFINE INDEX idx_report_status ON report FIELDS status;
DEFINE INDEX idx_report_created ON report FIELDS created_on;
DEFINE INDEX idx_report_category ON report FIELDS category;

DEFINE TABLE moderation_action SCHEMAFULL;
DEFINE FIELD user_id ON moderation_action TYPE record<user>;
DEFINE FIELD level ON moderation_action TYPE string
    ASSERT $value IN ["nudge", "warning", "suspension", "ban"];
DEFINE FIELD reason ON moderation_action TYPE string;
DEFINE FIELD report_id ON moderation_action TYPE option<record<report>>;
DEFINE FIELD admin_user_id ON moderation_action TYPE option<record<user>>;
DEFINE FIELD duration_days ON moderation_action TYPE option<int>;
DEFINE FIELD expires_on ON moderation_action TYPE option<datetime>;
DEFINE FIELD is_active ON moderation_action TYPE bool DEFAULT true;
DEFINE FIELD restrictions ON moderation_action TYPE option<array<string>>;
DEFINE FIELD created_on ON moderation_action TYPE datetime DEFAULT time::now();
DEFINE FIELD lifted_on ON moderation_action TYPE option<datetime>;
DEFINE FIELD lifted_by_id ON moderation_action TYPE option<record<user>>;
DEFINE FIELD lift_reason ON moderation_action TYPE option<string>;

DEFINE INDEX idx_action_user ON moderation_action FIELDS user_id;
DEFINE INDEX idx_action_active ON moderation_action FIELDS is_active;
DEFINE INDEX idx_action_level ON moderation_action FIELDS level;
DEFINE INDEX idx_action_expires ON moderation_action FIELDS expires_on;

DEFINE TABLE block SCHEMAFULL;
DEFINE FIELD blocker_user_id ON block TYPE record<user>;
DEFINE FIELD blocked_user_id ON block TYPE record<user>;
DEFINE FIELD reason ON block TYPE option<string>;
DEFINE FIELD created_on ON block TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_block_blocker ON block FIELDS blocker_user_id;
DEFINE INDEX idx_block_blocked ON block FIELDS blocked_user_id;
DEFINE INDEX idx_block_pair ON block FIELDS blocker_user_id, blocked_user_id UNIQUE;

DEFINE TABLE guild_moderation_settings SCHEMAFULL;
DEFINE FIELD guild_id ON guild_moderation_settings TYPE record<guild>;
DEFINE FIELD require_approval_to_join ON guild_moderation_settings TYPE bool DEFAULT true;
DEFINE FIELD min_reputation_to_join ON guild_moderation_settings TYPE option<int>;
DEFINE FIELD allow_public_events ON guild_moderation_settings TYPE bool DEFAULT false;
DEFINE FIELD public_event_min_reputation ON guild_moderation_settings TYPE option<int>;
DEFINE FIELD updated_on ON guild_moderation_settings TYPE datetime DEFAULT time::now();

DEFINE INDEX idx_guild_mod_settings ON guild_moderation_settings FIELDS guild_id UNIQUE;

-- =============================================================================
-- 19. BUSINESS CONSTRAINTS & EVENTS
-- =============================================================================

-- Limit people per guild (50)
DEFINE EVENT check_person_limit ON TABLE part_of WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM part_of WHERE out = $after.out GROUP ALL);
    IF $count[0].count > 50 {
        THROW "Maximum 50 people per guild"
    };
};

-- Limit members per guild (20)
DEFINE EVENT check_member_limit ON TABLE responsible_for WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM responsible_for WHERE out = $after.out GROUP ALL);
    IF $count[0].count > 20 {
        THROW "Maximum 20 members per guild"
    };
};

-- Limit activities per guild (50)
DEFINE EVENT check_activity_limit ON TABLE managed_by WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM managed_by WHERE out = $after.out GROUP ALL);
    IF $count[0].count > 50 {
        THROW "Maximum 50 activities per guild"
    };
};

-- Limit timers per person (50)
DEFINE EVENT check_timer_limit ON TABLE tracked_by WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM tracked_by WHERE in = $after.in GROUP ALL);
    IF $count[0].count > 50 {
        THROW "Maximum 50 timers per person"
    };
};

-- Limit roles per event (20)
DEFINE EVENT check_role_limit ON TABLE event_role WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM event_role WHERE event_id = $after.event_id GROUP ALL);
    IF $count[0].count > 20 {
        THROW "Maximum 20 roles per event"
    };
};

-- Limit role assignments per user per event (10)
DEFINE EVENT check_user_role_limit ON TABLE event_role_assignment WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM event_role_assignment
        WHERE event_id = $after.event_id AND user_id = $after.user_id GROUP ALL);
    IF $count[0].count > 10 {
        THROW "Maximum 10 roles per user per event"
    };
};

-- Limit hosts per event (5)
DEFINE EVENT check_host_limit ON TABLE event_host WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM event_host WHERE event_id = $after.event_id GROUP ALL);
    IF $count[0].count > 5 {
        THROW "Maximum 5 hosts per event"
    };
};

-- Limit seats per rideshare (8)
DEFINE EVENT check_seat_limit ON TABLE rideshare_seat WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM rideshare_seat WHERE rideshare_id = $after.rideshare_id GROUP ALL);
    IF $count[0].count > 8 {
        THROW "Maximum 8 seats per rideshare"
    };
};

-- Limit segments per rideshare (10)
DEFINE EVENT check_segment_limit ON TABLE rideshare_segment WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM rideshare_segment WHERE rideshare_id = $after.rideshare_id GROUP ALL);
    IF $count[0].count > 10 {
        THROW "Maximum 10 segments per rideshare"
    };
};

-- Limit pools per guild (10)
DEFINE EVENT check_pool_limit ON TABLE matching_pool WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM matching_pool WHERE guild_id = $after.guild_id GROUP ALL);
    IF $count[0].count > 10 {
        THROW "Maximum 10 pools per guild"
    };
};

-- Limit members per pool (100)
DEFINE EVENT check_pool_member_limit ON TABLE pool_member WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM pool_member WHERE pool_id = $after.pool_id AND active = true GROUP ALL);
    IF $count[0].count > 100 {
        THROW "Maximum 100 members per pool"
    };
};

-- Limit exclusions per pool member (20)
DEFINE EVENT check_exclusion_limit ON TABLE pool_member WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    IF array::len($after.excluded_members ?? []) > 20 {
        THROW "Maximum 20 exclusions per member"
    };
};

-- Validate nudge types
DEFINE EVENT check_nudge_type ON TABLE nudge WHEN $event = "CREATE" THEN {
    LET $valid_types = ["pending_match", "stale_hangout", "upcoming_hangout", "hangout_followup", "pending_request", "unresponded_request", "pool_match_created", "pool_match_stale"];
    IF NOT ($after.type IN $valid_types) {
        THROW "Invalid nudge type"
    };
};

-- Validate nudge channels
DEFINE EVENT check_nudge_channel ON TABLE nudge WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    LET $valid_channels = ["sse", "push"];
    IF NOT ($after.channel IN $valid_channels) {
        THROW "Invalid nudge channel"
    };
};

-- Validate quiet hours
DEFINE EVENT check_quiet_hours ON TABLE nudge_settings WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    IF $after.quiet_start_hour IS NOT NONE AND ($after.quiet_start_hour < 0 OR $after.quiet_start_hour > 23) {
        THROW "quiet_start_hour must be between 0 and 23"
    };
    IF $after.quiet_end_hour IS NOT NONE AND ($after.quiet_end_hour < 0 OR $after.quiet_end_hour > 23) {
        THROW "quiet_end_hour must be between 0 and 23"
    };
};

-- Validate adventure status
DEFINE EVENT check_adventure_status ON TABLE adventure WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    LET $valid_statuses = ["idea", "planning", "confirmed", "active", "completed", "cancelled"];
    IF NOT ($after.status IN $valid_statuses) {
        THROW "Invalid adventure status"
    };
};

-- Validate adventure participant role
DEFINE EVENT check_adventure_participant_role ON TABLE adventure_participant WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    LET $valid_roles = ["organizer", "participant"];
    IF NOT ($after.role IN $valid_roles) {
        THROW "Invalid participant role"
    };
};

-- Validate adventure participant status
DEFINE EVENT check_adventure_participant_status ON TABLE adventure_participant WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    LET $valid_statuses = ["interested", "committed", "maybe", "out"];
    IF NOT ($after.status IN $valid_statuses) {
        THROW "Invalid participant status"
    };
};

-- Validate adventure dates
DEFINE EVENT check_adventure_dates ON TABLE adventure WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    IF $after.end_date < $after.start_date {
        THROW "End date must be after start date"
    };
};

-- Limit destinations per adventure (20)
DEFINE EVENT check_destination_limit ON TABLE destination WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM destination WHERE adventure_id = $after.adventure_id GROUP ALL);
    IF $count[0].count > 20 {
        THROW "Maximum 20 destinations per adventure"
    };
};

-- Limit activities per adventure (50)
DEFINE EVENT check_adventure_activity_limit ON TABLE adventure_activity WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM adventure_activity WHERE adventure_id = $after.adventure_id GROUP ALL);
    IF $count[0].count > 50 {
        THROW "Maximum 50 activities per adventure"
    };
};

-- Limit events per adventure (20)
DEFINE EVENT check_adventure_event_limit ON TABLE event WHEN $event = "CREATE" THEN {
    IF $after.adventure_id IS NOT NONE {
        LET $count = (SELECT count() FROM event WHERE adventure_id = $after.adventure_id GROUP ALL);
        IF $count[0].count > 20 {
            THROW "Maximum 20 events per adventure"
        };
    };
};

-- Limit rideshares per event (10)
DEFINE EVENT check_event_rideshare_limit ON TABLE rideshare WHEN $event = "CREATE" THEN {
    IF $after.event_id IS NOT NONE {
        LET $count = (SELECT count() FROM rideshare WHERE event_id = $after.event_id GROUP ALL);
        IF $count[0].count > 10 {
            THROW "Maximum 10 rideshares per event"
        };
    };
};

-- Limit rideshares per adventure (20)
DEFINE EVENT check_adventure_rideshare_limit ON TABLE rideshare WHEN $event = "CREATE" THEN {
    IF $after.adventure_id IS NOT NONE {
        LET $count = (SELECT count() FROM rideshare WHERE adventure_id = $after.adventure_id GROUP ALL);
        IF $count[0].count > 20 {
            THROW "Maximum 20 rideshares per adventure"
        };
    };
};

-- Validate rideshare has parent (event OR adventure, not standalone)
DEFINE EVENT check_rideshare_parent ON TABLE rideshare WHEN $event = "CREATE" THEN {
    IF $after.event_id IS NONE AND $after.adventure_id IS NONE {
        THROW "Rideshare must be attached to an event or adventure"
    };
};

-- Validate guild alliance status
DEFINE EVENT check_alliance_status ON TABLE guild_alliance WHEN $event = "CREATE" OR $event = "UPDATE" THEN {
    LET $valid_statuses = ["pending", "active", "revoked"];
    IF NOT ($after.status IN $valid_statuses) {
        THROW "Invalid alliance status"
    };
};

-- Prevent self-alliance
DEFINE EVENT check_self_alliance ON TABLE guild_alliance WHEN $event = "CREATE" THEN {
    IF $after.guild_a_id = $after.guild_b_id {
        THROW "Cannot create alliance with self"
    };
};

-- Limit forum posts per forum (1000)
DEFINE EVENT check_forum_post_limit ON TABLE forum_post WHEN $event = "CREATE" THEN {
    LET $count = (SELECT count() FROM forum_post WHERE forum_id = $after.forum_id GROUP ALL);
    IF $count[0].count > 1000 {
        THROW "Maximum 1000 posts per forum"
    };
};

-- Validate forum has exactly one parent
DEFINE EVENT check_forum_parent ON TABLE forum WHEN $event = "CREATE" THEN {
    IF $after.adventure_id IS NONE AND $after.event_id IS NONE {
        THROW "Forum must be attached to an adventure or event"
    };
    IF $after.adventure_id IS NOT NONE AND $after.event_id IS NOT NONE {
        THROW "Forum cannot be attached to both adventure and event"
    };
};

-- Track filled slots on role assignment
DEFINE EVENT increment_filled_slots ON TABLE event_role_assignment WHEN $event = "CREATE" THEN {
    UPDATE event_role SET filled_slots += 1 WHERE id = $after.role_id;
};

DEFINE EVENT decrement_filled_slots ON TABLE event_role_assignment WHEN $event = "DELETE" THEN {
    UPDATE event_role SET filled_slots -= 1 WHERE id = $before.role_id;
};

-- Prevent self-report
DEFINE EVENT report_self_check ON TABLE report WHEN $event = "CREATE" THEN {
    IF $after.reporter_user_id = $after.reported_user_id {
        THROW "Cannot report yourself";
    };
};

-- Prevent self-block
DEFINE EVENT block_self_check ON TABLE block WHEN $event = "CREATE" THEN {
    IF $after.blocker_user_id = $after.blocked_user_id {
        THROW "Cannot block yourself";
    };
};

-- Auto-update timestamps
DEFINE EVENT pool_updated ON TABLE matching_pool WHEN $event = "UPDATE" THEN {
    UPDATE matching_pool SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT match_updated ON TABLE match_result WHEN $event = "UPDATE" THEN {
    UPDATE match_result SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT nudge_preference_updated ON TABLE nudge_preference WHEN $event = "UPDATE" THEN {
    UPDATE nudge_preference SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT nudge_settings_updated ON TABLE nudge_settings WHEN $event = "UPDATE" THEN {
    UPDATE nudge_settings SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT adventure_updated ON TABLE adventure WHEN $event = "UPDATE" THEN {
    UPDATE adventure SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT adventure_participant_updated ON TABLE adventure_participant WHEN $event = "UPDATE" THEN {
    UPDATE adventure_participant SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT destination_vote_updated ON TABLE destination_vote WHEN $event = "UPDATE" THEN {
    UPDATE destination_vote SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT adventure_activity_updated ON TABLE adventure_activity WHEN $event = "UPDATE" THEN {
    UPDATE adventure_activity SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT rideshare_updated ON TABLE rideshare WHEN $event = "UPDATE" THEN {
    UPDATE rideshare SET updated_on = time::now() WHERE id = $after.id;
};

DEFINE EVENT forum_post_updated ON TABLE forum_post WHEN $event = "UPDATE" THEN {
    UPDATE forum_post SET updated_on = time::now() WHERE id = $after.id;
};
