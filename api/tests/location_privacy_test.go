package tests

/*
FEATURE: Location Privacy
DOMAIN: Privacy & Data Protection

ACCEPTANCE CRITERIA:
===================

AC-LOC-001: Coordinates Never Exposed
  GIVEN user profile with location
  WHEN profile fetched via API
  THEN lat/lng fields NOT present in response

AC-LOC-002: Distance Buckets
  GIVEN two users 1.5km apart
  WHEN one views other's profile
  THEN distance shown as "~2km"

AC-LOC-003: Coarse Location Only
  GIVEN user profile
  WHEN fetched publicly
  THEN only city, country, timezone visible
*/

import (
	"encoding/json"
	"testing"

	"github.com/forgo/saga/api/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationPrivacy_CoordinatesNotInJSON(t *testing.T) {
	// AC-LOC-001: Coordinates should never appear in JSON output
	location := &model.Location{
		Lat:     37.7749,
		Lng:     -122.4194,
		City:    "San Francisco",
		Country: "United States",
	}

	// Serialize to JSON
	data, err := json.Marshal(location)
	require.NoError(t, err)

	// Verify coordinates are NOT in JSON
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	_, hasLat := jsonMap["lat"]
	_, hasLng := jsonMap["lng"]
	assert.False(t, hasLat, "lat should not be in JSON output")
	assert.False(t, hasLng, "lng should not be in JSON output")

	// Verify city and country ARE present
	assert.Contains(t, string(data), "San Francisco")
	assert.Contains(t, string(data), "United States")
}

func TestLocationPrivacy_InternalLocationHasCoordinates(t *testing.T) {
	// LocationInternal should include coordinates (for internal use/storage)
	locationInternal := &model.LocationInternal{
		Lat:         37.7749,
		Lng:         -122.4194,
		City:        "San Francisco",
		Country:     "United States",
		CountryCode: "US",
	}

	// Serialize to JSON
	data, err := json.Marshal(locationInternal)
	require.NoError(t, err)

	// Verify coordinates ARE in JSON (this is for internal storage)
	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	_, hasLat := jsonMap["lat"]
	_, hasLng := jsonMap["lng"]
	assert.True(t, hasLat, "lat should be in internal JSON")
	assert.True(t, hasLng, "lng should be in internal JSON")
}

func TestLocationPrivacy_InternalToPublicStripsCoordinates(t *testing.T) {
	// AC-LOC-003: ToPublic should strip coordinates
	locationInternal := &model.LocationInternal{
		Lat:         37.7749,
		Lng:         -122.4194,
		City:        "San Francisco",
		Country:     "United States",
		CountryCode: "US",
	}

	// Convert to public representation
	public := locationInternal.ToPublic()

	// Public location should have city and country
	assert.Equal(t, "San Francisco", public.City)
	assert.Equal(t, "United States", public.Country)

	// Verify coordinates don't appear in JSON serialization
	data, err := json.Marshal(public)
	require.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	_, hasLat := jsonMap["lat"]
	_, hasLng := jsonMap["lng"]
	assert.False(t, hasLat, "lat should not be in public JSON")
	assert.False(t, hasLng, "lng should not be in public JSON")
}

func TestLocationPrivacy_DistanceBucketNearby(t *testing.T) {
	// AC-LOC-002: Distance < 1km = "nearby"
	bucket := model.GetDistanceBucket(0.5)
	assert.Equal(t, model.DistanceNearby, bucket)

	bucket = model.GetDistanceBucket(0.9)
	assert.Equal(t, model.DistanceNearby, bucket)
}

func TestLocationPrivacy_DistanceBucket2km(t *testing.T) {
	// AC-LOC-002: Distance 1-2km = "~2km"
	bucket := model.GetDistanceBucket(1.0)
	assert.Equal(t, model.Distance2km, bucket)

	bucket = model.GetDistanceBucket(1.5)
	assert.Equal(t, model.Distance2km, bucket)

	bucket = model.GetDistanceBucket(1.9)
	assert.Equal(t, model.Distance2km, bucket)
}

func TestLocationPrivacy_DistanceBucket5km(t *testing.T) {
	// AC-LOC-002: Distance 2-5km = "~5km"
	bucket := model.GetDistanceBucket(2.0)
	assert.Equal(t, model.Distance5km, bucket)

	bucket = model.GetDistanceBucket(3.5)
	assert.Equal(t, model.Distance5km, bucket)

	bucket = model.GetDistanceBucket(4.9)
	assert.Equal(t, model.Distance5km, bucket)
}

func TestLocationPrivacy_DistanceBucket10km(t *testing.T) {
	// AC-LOC-002: Distance 5-10km = "~10km"
	bucket := model.GetDistanceBucket(5.0)
	assert.Equal(t, model.Distance10km, bucket)

	bucket = model.GetDistanceBucket(7.5)
	assert.Equal(t, model.Distance10km, bucket)

	bucket = model.GetDistanceBucket(9.9)
	assert.Equal(t, model.Distance10km, bucket)
}

func TestLocationPrivacy_DistanceBucket20kmPlus(t *testing.T) {
	// AC-LOC-002: Distance > 20km = ">20km"
	// Note: Based on the model, 10-20km isn't explicitly covered, so check threshold
	bucket := model.GetDistanceBucket(20.0)
	assert.Equal(t, model.Distance20kmPlus, bucket)

	bucket = model.GetDistanceBucket(50.0)
	assert.Equal(t, model.Distance20kmPlus, bucket)

	bucket = model.GetDistanceBucket(100.0)
	assert.Equal(t, model.Distance20kmPlus, bucket)
}

func TestLocationPrivacy_PublicProfileHasCoarseLocation(t *testing.T) {
	// AC-LOC-003: Public profile only shows city, country, not coordinates
	profile := &model.UserProfile{
		ID:         "profile:test123",
		UserID:     "user:test123",
		Visibility: model.VisibilityPublic,
		Location: &model.Location{
			Lat:         37.7749,
			Lng:         -122.4194,
			City:        "San Francisco",
			Country:     "United States",
			CountryCode: "US",
		},
	}

	public := profile.ToPublic()

	// Verify coarse location is present
	assert.NotEmpty(t, public.City)
	assert.NotEmpty(t, public.Country)
	assert.Equal(t, "San Francisco", public.City)
	assert.Equal(t, "United States", public.Country)

	// Verify no coordinates in public profile
	data, err := json.Marshal(public)
	require.NoError(t, err)

	var jsonMap map[string]interface{}
	err = json.Unmarshal(data, &jsonMap)
	require.NoError(t, err)

	_, hasLat := jsonMap["lat"]
	_, hasLng := jsonMap["lng"]
	assert.False(t, hasLat, "lat should not appear in public profile")
	assert.False(t, hasLng, "lng should not appear in public profile")
}

func TestLocationPrivacy_DistanceBucketValues(t *testing.T) {
	// Verify the bucket string values match expected display format
	assert.Equal(t, model.DistanceBucket("nearby"), model.DistanceNearby)
	assert.Equal(t, model.DistanceBucket("~2km"), model.Distance2km)
	assert.Equal(t, model.DistanceBucket("~5km"), model.Distance5km)
	assert.Equal(t, model.DistanceBucket("~10km"), model.Distance10km)
	assert.Equal(t, model.DistanceBucket(">20km"), model.Distance20kmPlus)
}
