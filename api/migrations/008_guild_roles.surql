-- ============================================================================
-- Migration 008: Guild Roles
-- Adds role-based permissions to guild membership
-- ============================================================================

-- Add role field to responsible_for relation
DEFINE FIELD role ON responsible_for TYPE string DEFAULT "member"
    ASSERT $value IN ["member", "moderator", "admin"];

-- Existing members without a role will get "member" as default
-- Guild creators should be manually promoted to admin or handled by application logic
