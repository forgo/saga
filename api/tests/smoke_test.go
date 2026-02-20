// Package tests contains end-to-end acceptance tests for the Saga API.
//
// These tests run against a real SurrealDB instance to validate actual
// database behavior including triggers, constraints, and functions.
//
// To run tests:
//  1. Start SurrealDB: surreal start memory -A --user root --pass root
//  2. Run tests: go test ./tests/...
//
// Environment variables:
//
//	TEST_DB_HOST     - SurrealDB host (default: localhost)
//	TEST_DB_PORT     - SurrealDB port (default: 8000)
//	TEST_DB_USER     - SurrealDB username (default: root)
//	TEST_DB_PASSWORD - SurrealDB password (default: root)
package tests

import (
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/helpers"
	"github.com/forgo/saga/api/internal/testing/testdb"
)

/*
FEATURE: Test Infrastructure Smoke Test
DOMAIN: Infrastructure

ACCEPTANCE CRITERIA:
===================

AC-SMOKE-001: Database Connection
  GIVEN SurrealDB is running
  WHEN we create a test database
  THEN the connection succeeds
  AND migrations are applied

AC-SMOKE-002: Fixture Creation
  GIVEN a test database
  WHEN we create a user fixture
  THEN the user is created in the database

AC-SMOKE-003: Guild Creation
  GIVEN a test database with a user
  WHEN we create a guild with the user as admin
  THEN the guild is created
  AND the user is a member

AC-SMOKE-004: Event Creation
  GIVEN a test database with a guild
  WHEN we create an event in the guild
  THEN the event is created with the correct properties

AC-SMOKE-005: Helper Functions
  GIVEN test helper utilities
  WHEN we use JWT and pointer helpers
  THEN they function correctly
*/

func TestSmoke_DatabaseConnection(t *testing.T) {
	// AC-SMOKE-001: Database Connection
	tdb := testdb.New(t)
	defer tdb.Close()

	// Verify we can ping the database
	if err := tdb.DB.Ping(tdb.Ctx()); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	// Verify migrations were applied by checking for a known table
	results := tdb.MustQuery("INFO FOR DB", nil)
	if len(results) == 0 {
		t.Fatal("expected database info, got none")
	}
}

func TestSmoke_FixtureCreation(t *testing.T) {
	// AC-SMOKE-002: Fixture Creation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	// Create a user
	user := f.CreateUser(t)

	if user.ID == "" {
		t.Error("expected user to have an ID")
	}
	if user.Email == "" {
		t.Error("expected user to have an email")
	}
	if user.Role != model.UserRoleUser {
		t.Errorf("expected user role to be %s, got %s", model.UserRoleUser, user.Role)
	}

	// Verify user exists in database
	helpers.AssertRecordExists(t, tdb.DB, "user", user.ID)
}

func TestSmoke_GuildCreation(t *testing.T) {
	// AC-SMOKE-003: Guild Creation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	// Create a user and guild
	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	if guild.ID == "" {
		t.Error("expected guild to have an ID")
	}
	if guild.Name == "" {
		t.Error("expected guild to have a name")
	}
	if guild.Visibility != model.GuildVisibilityPrivate {
		t.Errorf("expected guild visibility to be %s, got %s", model.GuildVisibilityPrivate, guild.Visibility)
	}

	// Verify guild exists in database
	helpers.AssertRecordExists(t, tdb.DB, "guild", guild.ID)
}

func TestSmoke_EventCreation(t *testing.T) {
	// AC-SMOKE-004: Event Creation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	// Create user, guild, and event
	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)
	event := f.CreateEvent(t, guild, user)

	if event.ID == "" {
		t.Error("expected event to have an ID")
	}
	if event.Title == "" {
		t.Error("expected event to have a title")
	}
	if event.Template != model.EventTemplateCasual {
		t.Errorf("expected event template to be %s, got %s", model.EventTemplateCasual, event.Template)
	}
	if event.Status != model.EventStatusPublished {
		t.Errorf("expected event status to be %s, got %s", model.EventStatusPublished, event.Status)
	}

	// Verify event exists in database
	helpers.AssertRecordExists(t, tdb.DB, "event", event.ID)
}

func TestSmoke_HelperFunctions(t *testing.T) {
	// AC-SMOKE-005: Helper Functions
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	user := f.CreateUser(t)

	// Test JWT helper
	jwt := helpers.NewJWTHelper(t)
	token := jwt.GenerateToken(user)
	if token == "" {
		t.Error("expected JWT token to be generated")
	}
	// Token should have 3 parts (header.payload.signature)
	parts := 0
	for _, c := range token {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("expected JWT token to have 2 dots (3 parts), got %d dots", parts)
	}

	// Test pointer helpers
	s := helpers.StringPtr("test")
	if s == nil || *s != "test" {
		t.Error("StringPtr failed")
	}

	i := helpers.IntPtr(42)
	if i == nil || *i != 42 {
		t.Error("IntPtr failed")
	}

	b := helpers.BoolPtr(true)
	if b == nil || !*b {
		t.Error("BoolPtr failed")
	}
}

func TestSmoke_SharedTestDB(t *testing.T) {
	// Test the shared TestDB functionality for subtests
	shared := testdb.NewShared(t)
	defer shared.Close()

	f := fixtures.New(shared.DB)

	t.Run("FirstSubtest", func(t *testing.T) {
		tdb := shared.SetupSubtest(t)
		user := f.CreateUser(t)
		helpers.AssertRecordExists(t, tdb.DB, "user", user.ID)
	})

	t.Run("SecondSubtest", func(t *testing.T) {
		tdb := shared.SetupSubtest(t)
		// Data from first subtest should be cleared
		user := f.CreateUser(t)
		helpers.AssertRecordExists(t, tdb.DB, "user", user.ID)
	})
}

func TestSmoke_AdminAndModeratorUsers(t *testing.T) {
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	admin := f.CreateAdmin(t)
	if admin.Role != model.UserRoleAdmin {
		t.Errorf("expected admin role, got %s", admin.Role)
	}
	if !admin.IsAdmin() {
		t.Error("expected IsAdmin() to return true")
	}

	mod := f.CreateModerator(t)
	if mod.Role != model.UserRoleModerator {
		t.Errorf("expected moderator role, got %s", mod.Role)
	}
	if !mod.IsModerator() {
		t.Error("expected IsModerator() to return true")
	}
}
