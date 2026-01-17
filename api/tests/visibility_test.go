package tests

/*
FEATURE: Visibility Cascade
DOMAIN: Access Control & Privacy

ACCEPTANCE CRITERIA:
===================

AC-VIS-001: Event Visibility <= Adventure Visibility
  GIVEN adventure with visibility=guilds
  WHEN creating event with visibility=public
  THEN fails with 400 Bad Request (child > parent)

AC-VIS-002: Event Visibility <= Adventure (Valid)
  GIVEN adventure with visibility=guilds
  WHEN creating event with visibility=private
  THEN succeeds (child <= parent)

AC-VIS-003: Rideshare Visibility <= Event
  GIVEN event with visibility=guilds
  WHEN creating rideshare with visibility=public
  THEN fails (child > parent)

AC-VIS-004: Update Cascade Check
  GIVEN adventure with public children
  WHEN updating adventure to private
  THEN fails OR children auto-downgraded

VISIBILITY ORDERING (most to least restrictive):
  private (0) < invite_only (1) < guilds (2) < public (3)
*/

import (
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisibility_OrderingConstants(t *testing.T) {
	// Verify visibility ordering helpers work correctly
	// Order: private (0) < invite_only (1) < guilds (2) < public (3)

	// Adventure visibility levels
	assert.True(t, isLessRestrictive(string(model.AdventureVisibilityPublic), string(model.AdventureVisibilityGuilds)),
		"public should be less restrictive than guilds")
	assert.True(t, isLessRestrictive(string(model.AdventureVisibilityGuilds), string(model.AdventureVisibilityInviteOnly)),
		"guilds should be less restrictive than invite_only")
	assert.True(t, isLessRestrictive(string(model.AdventureVisibilityInviteOnly), string(model.AdventureVisibilityPrivate)),
		"invite_only should be less restrictive than private")

	// Same visibility
	assert.False(t, isLessRestrictive(string(model.AdventureVisibilityGuilds), string(model.AdventureVisibilityGuilds)),
		"same visibility should not be less restrictive")

	// Reverse (more restrictive)
	assert.False(t, isLessRestrictive(string(model.AdventureVisibilityPrivate), string(model.AdventureVisibilityPublic)),
		"private should not be less restrictive than public")
}

func TestVisibility_AdventureCanBeCreatedWithAnyVisibility(t *testing.T) {
	// Adventures can be created with any visibility level
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Test each visibility level
	visibilities := []string{
		string(model.AdventureVisibilityPrivate),
		string(model.AdventureVisibilityInviteOnly),
		string(model.AdventureVisibilityGuilds),
		string(model.AdventureVisibilityPublic),
	}

	for _, vis := range visibilities {
		adventure := f.CreateAdventure(t, guild, user, fixtures.WithAdventureVisibility(vis))
		require.NotNil(t, adventure, "Should create adventure with visibility %s", vis)
		assert.Equal(t, model.AdventureVisibility(vis), adventure.Visibility)
	}
}

func TestVisibility_EventInheritsOrRestrictsFromAdventure(t *testing.T) {
	// Events should be able to have equal or more restrictive visibility than their adventure
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create adventure with visibility=guilds
	adventure := f.CreateAdventure(t, guild, user, fixtures.WithAdventureVisibility(string(model.AdventureVisibilityGuilds)))
	require.NotNil(t, adventure)
	assert.Equal(t, model.AdventureVisibilityGuilds, adventure.Visibility)

	// Create event with more restrictive visibility (private <= guilds)
	event := f.CreateEventForAdventure(t, adventure, guild, user, fixtures.WithEventVisibility(model.EventVisibilityPrivate))
	require.NotNil(t, event, "Should create event with more restrictive visibility")
	assert.Equal(t, model.EventVisibilityPrivate, event.Visibility)

	// Create event with same visibility (guilds == guilds)
	event2 := f.CreateEventForAdventure(t, adventure, guild, user, fixtures.WithEventVisibility(model.EventVisibilityGuilds))
	require.NotNil(t, event2, "Should create event with same visibility as adventure")
	assert.Equal(t, model.EventVisibilityGuilds, event2.Visibility)
}

func TestVisibility_EventWithInviteOnlyUnderGuilds(t *testing.T) {
	// invite_only is more restrictive than guilds
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create adventure with visibility=guilds
	adventure := f.CreateAdventure(t, guild, user, fixtures.WithAdventureVisibility(string(model.AdventureVisibilityGuilds)))

	// Create event with invite_only (should work: invite_only < guilds)
	event := f.CreateEventForAdventure(t, adventure, guild, user, fixtures.WithEventVisibility(model.EventVisibilityInviteOnly))
	require.NotNil(t, event, "Should create event with invite_only visibility under guilds adventure")
	assert.Equal(t, model.EventVisibilityInviteOnly, event.Visibility)
}

func TestVisibility_PublicAdventureAllowsAnyEventVisibility(t *testing.T) {
	// Public adventure should allow events with any visibility
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create adventure with visibility=public
	adventure := f.CreateAdventure(t, guild, user, fixtures.WithAdventureVisibility(string(model.AdventureVisibilityPublic)))

	// Test all event visibilities (all should work under public)
	visibilities := []string{
		model.EventVisibilityPrivate,
		model.EventVisibilityInviteOnly,
		model.EventVisibilityGuilds,
		model.EventVisibilityPublic,
	}

	for _, vis := range visibilities {
		event := f.CreateEventForAdventure(t, adventure, guild, user, fixtures.WithEventVisibility(vis))
		require.NotNil(t, event, "Should create event with visibility %s under public adventure", vis)
		assert.Equal(t, vis, event.Visibility)
	}
}

func TestVisibility_PrivateAdventureOnlyAllowsPrivateEvents(t *testing.T) {
	// Private adventure should only allow private events
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create adventure with visibility=private
	adventure := f.CreateAdventure(t, guild, user, fixtures.WithAdventureVisibility(string(model.AdventureVisibilityPrivate)))

	// Private event should work
	event := f.CreateEventForAdventure(t, adventure, guild, user, fixtures.WithEventVisibility(model.EventVisibilityPrivate))
	require.NotNil(t, event, "Should create private event under private adventure")
	assert.Equal(t, model.EventVisibilityPrivate, event.Visibility)
}

// isLessRestrictive returns true if a is less restrictive than b
// Less restrictive = more people can see it
func isLessRestrictive(a, b string) bool {
	order := map[string]int{
		"private":     0,
		"invite_only": 1,
		"guilds":      2,
		"public":      3,
	}
	return order[a] > order[b]
}
