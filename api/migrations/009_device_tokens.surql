-- ============================================================================
-- Migration 009: Device Tokens
-- Stores device push notification tokens for users
-- ============================================================================

DEFINE TABLE device_token SCHEMAFULL;

DEFINE FIELD user_id ON device_token TYPE string;
DEFINE FIELD platform ON device_token TYPE string
    ASSERT $value IN ["ios", "android", "web"];
DEFINE FIELD token ON device_token TYPE string;
DEFINE FIELD name ON device_token TYPE option<string>;
DEFINE FIELD active ON device_token TYPE bool DEFAULT true;
DEFINE FIELD created_on ON device_token TYPE datetime DEFAULT time::now();
DEFINE FIELD updated_on ON device_token TYPE datetime DEFAULT time::now();
DEFINE FIELD last_used ON device_token TYPE option<datetime>;

-- Index for finding devices by user
DEFINE INDEX device_token_user ON device_token FIELDS user_id;

-- Unique constraint: one token per user (prevents duplicate registrations)
DEFINE INDEX device_token_unique ON device_token FIELDS user_id, token UNIQUE;

-- Index for finding by token (for deactivation on failure)
DEFINE INDEX device_token_by_token ON device_token FIELDS token;
