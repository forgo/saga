package service

import (
	"math"
	"testing"

	"github.com/forgo/saga/api/internal/model"
)

// ============================================================================
// HaversineDistance Tests
// ============================================================================

func TestHaversineDistance_SamePoint_ReturnsZero(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	distance := svc.HaversineDistance(40.7128, -74.0060, 40.7128, -74.0060)

	if distance != 0 {
		t.Errorf("expected 0, got %f", distance)
	}
}

func TestHaversineDistance_NYCtoLA_ReturnsKnownDistance(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// New York City: 40.7128, -74.0060
	// Los Angeles: 34.0522, -118.2437
	// Known distance: ~3944 km
	distance := svc.HaversineDistance(40.7128, -74.0060, 34.0522, -118.2437)

	// Allow 1% tolerance for floating point and Earth model variations
	expectedKm := 3944.0
	tolerance := expectedKm * 0.01
	if math.Abs(distance-expectedKm) > tolerance {
		t.Errorf("expected ~%f km, got %f km", expectedKm, distance)
	}
}

func TestHaversineDistance_LondonToParis_ReturnsKnownDistance(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// London: 51.5074, -0.1278
	// Paris: 48.8566, 2.3522
	// Known distance: ~343 km
	distance := svc.HaversineDistance(51.5074, -0.1278, 48.8566, 2.3522)

	expectedKm := 343.0
	tolerance := expectedKm * 0.02 // 2% tolerance
	if math.Abs(distance-expectedKm) > tolerance {
		t.Errorf("expected ~%f km, got %f km", expectedKm, distance)
	}
}

func TestHaversineDistance_SydneyToTokyo_ReturnsKnownDistance(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Sydney: -33.8688, 151.2093
	// Tokyo: 35.6762, 139.6503
	// Known distance: ~7823 km
	distance := svc.HaversineDistance(-33.8688, 151.2093, 35.6762, 139.6503)

	expectedKm := 7823.0
	tolerance := expectedKm * 0.02
	if math.Abs(distance-expectedKm) > tolerance {
		t.Errorf("expected ~%f km, got %f km", expectedKm, distance)
	}
}

func TestHaversineDistance_EquatorPoints_ReturnsCorrectDistance(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Two points on the equator, 90 degrees apart
	// At equator, 90 degrees = ~10,008 km (quarter of Earth's circumference at equator)
	distance := svc.HaversineDistance(0, 0, 0, 90)

	// Earth circumference at equator ~40,075 km, so quarter is ~10,019 km
	expectedKm := 10008.0
	tolerance := expectedKm * 0.01
	if math.Abs(distance-expectedKm) > tolerance {
		t.Errorf("expected ~%f km, got %f km", expectedKm, distance)
	}
}

func TestHaversineDistance_AntipodalPoints_ReturnsHalfCircumference(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Antipodal points (exact opposite on Earth)
	// From (0, 0) to (0, 180) - half way around the equator
	distance := svc.HaversineDistance(0, 0, 0, 180)

	// Half Earth circumference at equator ~20,037 km
	expectedKm := 20015.0 // Using our Earth radius constant
	tolerance := expectedKm * 0.01
	if math.Abs(distance-expectedKm) > tolerance {
		t.Errorf("expected ~%f km, got %f km", expectedKm, distance)
	}
}

func TestHaversineDistance_NorthSouthPoles_ReturnsHalfCircumference(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// North pole to South pole
	distance := svc.HaversineDistance(90, 0, -90, 0)

	// Half Earth circumference through poles ~20,004 km
	expectedKm := 20015.0
	tolerance := expectedKm * 0.01
	if math.Abs(distance-expectedKm) > tolerance {
		t.Errorf("expected ~%f km, got %f km", expectedKm, distance)
	}
}

func TestHaversineDistance_ShortDistance_Accurate(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Two points ~1km apart (roughly 0.009 degrees at equator)
	distance := svc.HaversineDistance(0, 0, 0, 0.009)

	// Should be approximately 1 km
	if distance < 0.9 || distance > 1.1 {
		t.Errorf("expected ~1 km, got %f km", distance)
	}
}

func TestHaversineDistance_Symmetric(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Distance A to B should equal B to A
	distAB := svc.HaversineDistance(40.7128, -74.0060, 34.0522, -118.2437)
	distBA := svc.HaversineDistance(34.0522, -118.2437, 40.7128, -74.0060)

	if math.Abs(distAB-distBA) > 0.001 {
		t.Errorf("distance should be symmetric: A->B=%f, B->A=%f", distAB, distBA)
	}
}

// ============================================================================
// GetDistanceBucket Tests
// ============================================================================

func TestGetDistanceBucket_Nearby(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	tests := []struct {
		distance float64
		expected model.DistanceBucket
	}{
		{0, model.DistanceNearby},
		{0.5, model.DistanceNearby},
		{0.99, model.DistanceNearby},
	}

	for _, tt := range tests {
		result := svc.GetDistanceBucket(tt.distance)
		if result != tt.expected {
			t.Errorf("GetDistanceBucket(%f): expected %q, got %q", tt.distance, tt.expected, result)
		}
	}
}

func TestGetDistanceBucket_AllBuckets(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	tests := []struct {
		distance float64
		expected model.DistanceBucket
	}{
		{0.5, model.DistanceNearby},    // < 1 km
		{1.5, model.Distance2km},       // 1-2 km
		{3.0, model.Distance5km},       // 2-5 km
		{7.0, model.Distance10km},      // 5-10 km
		{25.0, model.Distance20kmPlus}, // > 20 km
	}

	for _, tt := range tests {
		result := svc.GetDistanceBucket(tt.distance)
		if result != tt.expected {
			t.Errorf("GetDistanceBucket(%f): expected %q, got %q", tt.distance, tt.expected, result)
		}
	}
}

func TestGetDistanceBucket_Boundaries(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Test boundary values
	tests := []struct {
		distance float64
		expected model.DistanceBucket
	}{
		{1.0, model.Distance2km},       // Exactly 1 km
		{2.0, model.Distance5km},       // Exactly 2 km
		{5.0, model.Distance10km},      // Exactly 5 km
		{10.0, model.Distance20kmPlus}, // Exactly 10 km
		{20.0, model.Distance20kmPlus}, // Exactly 20 km
	}

	for _, tt := range tests {
		result := svc.GetDistanceBucket(tt.distance)
		if result != tt.expected {
			t.Errorf("GetDistanceBucket(%f): expected %q, got %q", tt.distance, tt.expected, result)
		}
	}
}

// ============================================================================
// DistanceBetweenLocations Tests
// ============================================================================

func TestDistanceBetweenLocations_ValidLocations_ReturnsDistance(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	loc1 := &model.LocationInternal{Lat: 40.7128, Lng: -74.0060}
	loc2 := &model.LocationInternal{Lat: 34.0522, Lng: -118.2437}

	distance := svc.DistanceBetweenLocations(loc1, loc2)

	// NYC to LA ~3944 km
	if distance < 3900 || distance > 4000 {
		t.Errorf("expected ~3944 km, got %f km", distance)
	}
}

func TestDistanceBetweenLocations_NilFirst_ReturnsNegative(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	loc2 := &model.LocationInternal{Lat: 34.0522, Lng: -118.2437}

	distance := svc.DistanceBetweenLocations(nil, loc2)

	if distance != -1 {
		t.Errorf("expected -1 for nil first location, got %f", distance)
	}
}

func TestDistanceBetweenLocations_NilSecond_ReturnsNegative(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	loc1 := &model.LocationInternal{Lat: 40.7128, Lng: -74.0060}

	distance := svc.DistanceBetweenLocations(loc1, nil)

	if distance != -1 {
		t.Errorf("expected -1 for nil second location, got %f", distance)
	}
}

func TestDistanceBetweenLocations_BothNil_ReturnsNegative(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	distance := svc.DistanceBetweenLocations(nil, nil)

	if distance != -1 {
		t.Errorf("expected -1 for both nil locations, got %f", distance)
	}
}

// ============================================================================
// IsWithinRadius Tests
// ============================================================================

func TestIsWithinRadius_InsideRadius_ReturnsTrue(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Point 0.5km from center, radius 1km
	result := svc.IsWithinRadius(0, 0, 0, 0.0045, 1.0)

	if !result {
		t.Error("point should be within radius")
	}
}

func TestIsWithinRadius_OutsideRadius_ReturnsFalse(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Point ~111km from center (1 degree at equator), radius 10km
	result := svc.IsWithinRadius(0, 0, 1, 0, 10.0)

	if result {
		t.Error("point should be outside radius")
	}
}

func TestIsWithinRadius_ExactlyOnRadius_ReturnsTrue(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// Calculate a point exactly at 10km
	// At equator, 10km ≈ 0.0899 degrees
	centerLat := 0.0
	centerLng := 0.0
	pointLat := 0.0
	pointLng := 0.0899

	distance := svc.HaversineDistance(centerLat, centerLng, pointLat, pointLng)
	radius := distance // Use the exact distance as radius

	result := svc.IsWithinRadius(centerLat, centerLng, pointLat, pointLng, radius)

	if !result {
		t.Error("point exactly on radius should be within radius (<=)")
	}
}

func TestIsWithinRadius_SamePoint_ReturnsTrue(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	result := svc.IsWithinRadius(40.7128, -74.0060, 40.7128, -74.0060, 0.1)

	if !result {
		t.Error("same point should always be within any positive radius")
	}
}

// ============================================================================
// GetBoundingBox Tests
// ============================================================================

func TestGetBoundingBox_AtEquator_SymmetricBox(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	box := svc.GetBoundingBox(0, 0, 10.0)

	// At equator, lat and lng deltas should be similar
	latDelta := box.MaxLat - box.MinLat
	lngDelta := box.MaxLng - box.MinLng

	// 10km radius should give roughly 0.18 degrees (10/111 * 2)
	expectedDelta := (10.0 / 111.0) * 2
	tolerance := expectedDelta * 0.1

	if math.Abs(latDelta-expectedDelta) > tolerance {
		t.Errorf("lat delta: expected ~%f, got %f", expectedDelta, latDelta)
	}
	if math.Abs(lngDelta-expectedDelta) > tolerance {
		t.Errorf("lng delta: expected ~%f at equator, got %f", expectedDelta, lngDelta)
	}
}

func TestGetBoundingBox_AtHighLatitude_CompressedLongitude(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	// At 60 degrees latitude, longitude degrees are compressed (cos(60°) = 0.5)
	box := svc.GetBoundingBox(60, 0, 10.0)

	latDelta := box.MaxLat - box.MinLat
	lngDelta := box.MaxLng - box.MinLng

	// Longitude delta should be roughly 2x latitude delta at 60°
	// Because cos(60°) ≈ 0.5, so need 2x the degrees for same distance
	if lngDelta < latDelta*1.8 {
		t.Errorf("at 60° latitude, lng delta (%f) should be ~2x lat delta (%f)", lngDelta, latDelta)
	}
}

func TestGetBoundingBox_CenteredOnInput(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	centerLat := 45.0
	centerLng := -90.0
	box := svc.GetBoundingBox(centerLat, centerLng, 10.0)

	// Box should be centered on input
	midLat := (box.MinLat + box.MaxLat) / 2
	midLng := (box.MinLng + box.MaxLng) / 2

	if math.Abs(midLat-centerLat) > 0.0001 {
		t.Errorf("box not centered: expected lat %f, mid is %f", centerLat, midLat)
	}
	if math.Abs(midLng-centerLng) > 0.0001 {
		t.Errorf("box not centered: expected lng %f, mid is %f", centerLng, midLng)
	}
}

func TestGetBoundingBox_SmallRadius(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	box := svc.GetBoundingBox(0, 0, 1.0)

	// 1km radius should give roughly 0.018 degrees (1/111 * 2)
	expectedDelta := (1.0 / 111.0) * 2
	latDelta := box.MaxLat - box.MinLat

	if math.Abs(latDelta-expectedDelta) > 0.002 {
		t.Errorf("lat delta for 1km: expected ~%f, got %f", expectedDelta, latDelta)
	}
}

// ============================================================================
// CalculateDistances Tests
// ============================================================================

func TestCalculateDistances_WithValidCenter_CalculatesAll(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	center := &model.LocationInternal{Lat: 0, Lng: 0}
	locations := []LocationWithDistance{
		{Location: &model.LocationInternal{Lat: 0, Lng: 0.009}, UserID: "user1"}, // ~1km
		{Location: &model.LocationInternal{Lat: 0, Lng: 0.045}, UserID: "user2"}, // ~5km
		{Location: &model.LocationInternal{Lat: 0, Lng: 0.09}, UserID: "user3"},  // ~10km
	}

	result := svc.CalculateDistances(center, locations)

	// Check distances are populated
	if result[0].Distance < 0.9 || result[0].Distance > 1.1 {
		t.Errorf("user1 distance: expected ~1km, got %f", result[0].Distance)
	}
	if result[1].Distance < 4.5 || result[1].Distance > 5.5 {
		t.Errorf("user2 distance: expected ~5km, got %f", result[1].Distance)
	}

	// Check buckets are set
	if result[0].Bucket != model.DistanceNearby && result[0].Bucket != model.Distance2km {
		t.Errorf("user1 bucket: expected Nearby or ~2km, got %s", result[0].Bucket)
	}
}

func TestCalculateDistances_NilCenter_ReturnsUnchanged(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	locations := []LocationWithDistance{
		{Location: &model.LocationInternal{Lat: 0, Lng: 0.009}, UserID: "user1", Distance: 0},
	}

	result := svc.CalculateDistances(nil, locations)

	// Distance should remain 0 (unchanged)
	if result[0].Distance != 0 {
		t.Errorf("distance should be unchanged with nil center, got %f", result[0].Distance)
	}
}

func TestCalculateDistances_SomeNilLocations_HandlesGracefully(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	center := &model.LocationInternal{Lat: 0, Lng: 0}
	locations := []LocationWithDistance{
		{Location: &model.LocationInternal{Lat: 0, Lng: 0.009}, UserID: "user1"},
		{Location: nil, UserID: "user2"}, // Nil location
		{Location: &model.LocationInternal{Lat: 0, Lng: 0.045}, UserID: "user3"},
	}

	result := svc.CalculateDistances(center, locations)

	// Valid locations should have distances
	if result[0].Distance < 0.9 || result[0].Distance > 1.1 {
		t.Errorf("user1 distance: expected ~1km, got %f", result[0].Distance)
	}
	// Nil location should have distance 0 (unchanged from initial)
	if result[1].Distance != 0 {
		t.Errorf("user2 (nil loc) distance should be 0, got %f", result[1].Distance)
	}
	if result[2].Distance < 4.5 || result[2].Distance > 5.5 {
		t.Errorf("user3 distance: expected ~5km, got %f", result[2].Distance)
	}
}

func TestCalculateDistances_EmptySlice_ReturnsEmpty(t *testing.T) {
	t.Parallel()
	svc := NewGeoService()

	center := &model.LocationInternal{Lat: 0, Lng: 0}
	var locations []LocationWithDistance

	result := svc.CalculateDistances(center, locations)

	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d items", len(result))
	}
}

// ============================================================================
// Constants Tests
// ============================================================================

func TestConstants_ReasonableValues(t *testing.T) {
	t.Parallel()

	if EarthRadiusKm < 6370 || EarthRadiusKm > 6372 {
		t.Errorf("EarthRadiusKm should be ~6371, got %f", EarthRadiusKm)
	}

	if DefaultSearchRadiusKm != 25.0 {
		t.Errorf("DefaultSearchRadiusKm: expected 25, got %f", DefaultSearchRadiusKm)
	}

	if MaxSearchRadiusKm != 100.0 {
		t.Errorf("MaxSearchRadiusKm: expected 100, got %f", MaxSearchRadiusKm)
	}

	if NearbyRadiusKm != 1.0 {
		t.Errorf("NearbyRadiusKm: expected 1, got %f", NearbyRadiusKm)
	}
}

// ============================================================================
// NewGeoService Tests
// ============================================================================

func TestNewGeoService_ReturnsInstance(t *testing.T) {
	t.Parallel()

	svc := NewGeoService()

	if svc == nil {
		t.Error("NewGeoService() should return non-nil")
	}
}
