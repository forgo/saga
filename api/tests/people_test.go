package tests

/*
FEATURE: People/Contacts
DOMAIN: Guild People Management

ACCEPTANCE CRITERIA:
===================

AC-PEOPLE-001: Create Person
  GIVEN user is guild member
  WHEN user creates person with name
  THEN person is created in guild

AC-PEOPLE-002: List Guild People
  GIVEN guild has people A, B, C
  WHEN user lists people
  THEN all people returned

AC-PEOPLE-003: Update Person
  GIVEN existing person in guild
  WHEN user updates person name
  THEN changes persisted

AC-PEOPLE-004: Delete Person
  GIVEN person with no active timers
  WHEN user deletes person
  THEN person removed

AC-PEOPLE-005: Cross-Guild Isolation
  GIVEN person X in guild A
  WHEN user queries guild B people
  THEN person X not returned
*/

import (
	"context"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/repository"
	"github.com/forgo/saga/api/internal/testing/fixtures"
	"github.com/forgo/saga/api/internal/testing/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPerson_Create(t *testing.T) {
	// AC-PEOPLE-001: Create Person
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	person := &model.Person{
		GuildID:  guild.ID,
		Name:     "John Doe",
		Nickname: "Johnny",
	}

	err := personRepo.Create(ctx, person)
	require.NoError(t, err)
	assert.NotEmpty(t, person.ID)
	assert.NotZero(t, person.CreatedOn)

	// Verify person can be retrieved
	fetched, err := personRepo.GetByID(ctx, person.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "John Doe", fetched.Name)
	assert.Equal(t, "Johnny", fetched.Nickname)
}

func TestPerson_Create_WithAllFields(t *testing.T) {
	// Create person with all optional fields
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	birthday := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
	person := &model.Person{
		GuildID:  guild.ID,
		Name:     "Jane Doe",
		Nickname: "Janey",
		Birthday: &birthday,
		Notes:    "Best friend from college",
		Avatar:   "avatar.jpg",
	}

	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	// Verify all fields
	fetched, err := personRepo.GetByID(ctx, person.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "Jane Doe", fetched.Name)
	assert.Equal(t, "Janey", fetched.Nickname)
	assert.Equal(t, "Best friend from college", fetched.Notes)
	assert.Equal(t, "avatar.jpg", fetched.Avatar)
	require.NotNil(t, fetched.Birthday)
	assert.Equal(t, 1990, fetched.Birthday.Year())
}

func TestPerson_ListByGuild(t *testing.T) {
	// AC-PEOPLE-002: List Guild People
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create multiple people
	names := []string{"Alice", "Bob", "Charlie"}
	for _, name := range names {
		person := &model.Person{
			GuildID: guild.ID,
			Name:    name,
		}
		err := personRepo.Create(ctx, person)
		require.NoError(t, err)
	}

	// List people
	fetched, err := personRepo.GetByGuildID(ctx, guild.ID)
	require.NoError(t, err)
	assert.Len(t, fetched, 3, "Should return all 3 people")

	// Verify people are returned (sorted by name)
	fetchedNames := make([]string, len(fetched))
	for i, p := range fetched {
		fetchedNames[i] = p.Name
	}
	assert.Contains(t, fetchedNames, "Alice")
	assert.Contains(t, fetchedNames, "Bob")
	assert.Contains(t, fetchedNames, "Charlie")
}

func TestPerson_ListByGuild_Empty(t *testing.T) {
	// List people for guild with no people
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// List people - should return empty list
	fetched, err := personRepo.GetByGuildID(ctx, guild.ID)
	require.NoError(t, err)
	assert.Len(t, fetched, 0, "Should return empty list for guild with no people")
}

func TestPerson_Update(t *testing.T) {
	// AC-PEOPLE-003: Update Person
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	// Update name and add nickname
	person.Name = "John Smith"
	person.Nickname = "Johnny"
	person.Notes = "Updated contact info"

	err = personRepo.Update(ctx, person)
	require.NoError(t, err)

	// Verify update persisted
	fetched, err := personRepo.GetByID(ctx, person.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, "John Smith", fetched.Name)
	assert.Equal(t, "Johnny", fetched.Nickname)
	assert.Equal(t, "Updated contact info", fetched.Notes)
}

func TestPerson_Delete(t *testing.T) {
	// AC-PEOPLE-004: Delete Person
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Create person
	person := &model.Person{
		GuildID: guild.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	// Verify it exists
	fetched, err := personRepo.GetByID(ctx, person.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	// Delete person
	err = personRepo.Delete(ctx, person.ID)
	require.NoError(t, err)

	// Verify deleted
	fetched, err = personRepo.GetByID(ctx, person.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched, "Person should be deleted")
}

func TestPerson_CrossGuildIsolation(t *testing.T) {
	// AC-PEOPLE-005: Cross-Guild Isolation
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild1 := f.CreateGuild(t, user)
	guild2 := f.CreateGuild(t, user)

	// Create person in guild1
	person := &model.Person{
		GuildID: guild1.ID,
		Name:    "John Doe",
	}
	err := personRepo.Create(ctx, person)
	require.NoError(t, err)

	// List people for guild2 - should be empty
	guild2People, err := personRepo.GetByGuildID(ctx, guild2.ID)
	require.NoError(t, err)
	assert.Len(t, guild2People, 0, "Guild2 should have no people")

	// List people for guild1 - should have 1
	guild1People, err := personRepo.GetByGuildID(ctx, guild1.ID)
	require.NoError(t, err)
	assert.Len(t, guild1People, 1, "Guild1 should have 1 person")
}

func TestPerson_Count(t *testing.T) {
	// Test person count per guild
	tdb := testdb.New(t)
	defer tdb.Close()

	f := fixtures.New(tdb.DB)
	personRepo := repository.NewPersonRepository(tdb.DB)
	ctx := context.Background()

	user := f.CreateUser(t)
	guild := f.CreateGuild(t, user)

	// Initially zero
	count, err := personRepo.CountByGuildID(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Create people
	for i := 0; i < 5; i++ {
		person := &model.Person{
			GuildID: guild.ID,
			Name:    "Person" + string(rune('A'+i)),
		}
		err := personRepo.Create(ctx, person)
		require.NoError(t, err)
	}

	// Count should be 5
	count, err = personRepo.CountByGuildID(ctx, guild.ID)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}
