package service

import (
	"math"

	"github.com/forgo/saga/api/internal/model"
)

// GeoService handles geographic calculations
type GeoService struct{}

// NewGeoService creates a new geo service
func NewGeoService() *GeoService {
	return &GeoService{}
}

// EarthRadiusKm is the Earth's radius in kilometers
const EarthRadiusKm = 6371.0

// HaversineDistance calculates the distance between two points in kilometers
// using the Haversine formula (accounts for Earth's curvature)
func (s *GeoService) HaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

// GetDistanceBucket returns a privacy-preserving distance bucket
func (s *GeoService) GetDistanceBucket(distanceKm float64) model.DistanceBucket {
	return model.GetDistanceBucket(distanceKm)
}

// DistanceBetweenLocations calculates distance between two LocationInternal objects
func (s *GeoService) DistanceBetweenLocations(loc1, loc2 *model.LocationInternal) float64 {
	if loc1 == nil || loc2 == nil {
		return -1 // Unknown distance
	}
	return s.HaversineDistance(loc1.Lat, loc1.Lng, loc2.Lat, loc2.Lng)
}

// IsWithinRadius checks if a point is within a given radius of another point
func (s *GeoService) IsWithinRadius(centerLat, centerLng, pointLat, pointLng, radiusKm float64) bool {
	distance := s.HaversineDistance(centerLat, centerLng, pointLat, pointLng)
	return distance <= radiusKm
}

// BoundingBox calculates a rough bounding box for initial filtering
// before applying Haversine for accuracy (optimization for database queries)
type BoundingBox struct {
	MinLat float64 `json:"min_lat"`
	MaxLat float64 `json:"max_lat"`
	MinLng float64 `json:"min_lng"`
	MaxLng float64 `json:"max_lng"`
}

// GetBoundingBox returns a bounding box around a center point with given radius
// This is an approximation used for database query optimization
func (s *GeoService) GetBoundingBox(lat, lng, radiusKm float64) BoundingBox {
	// Approximate degrees per km
	// At equator: 1 degree latitude ≈ 111 km
	// Longitude varies by latitude
	latDelta := radiusKm / 111.0
	lngDelta := radiusKm / (111.0 * math.Cos(lat*math.Pi/180))

	return BoundingBox{
		MinLat: lat - latDelta,
		MaxLat: lat + latDelta,
		MinLng: lng - lngDelta,
		MaxLng: lng + lngDelta,
	}
}

// NearbySearchConfig holds configuration for nearby searches
type NearbySearchConfig struct {
	CenterLat  float64
	CenterLng  float64
	RadiusKm   float64
	MaxResults int
}

// LocationWithDistance pairs a location with its computed distance for sorting.
type LocationWithDistance struct {
	Location *model.LocationInternal
	Distance float64
	Bucket   model.DistanceBucket
	UserID   string // or other identifier
}

// CalculateDistances calculates distances from a center point to multiple locations
func (s *GeoService) CalculateDistances(center *model.LocationInternal, locations []LocationWithDistance) []LocationWithDistance {
	if center == nil {
		return locations
	}

	for i := range locations {
		if locations[i].Location != nil {
			locations[i].Distance = s.DistanceBetweenLocations(center, locations[i].Location)
			locations[i].Bucket = s.GetDistanceBucket(locations[i].Distance)
		}
	}

	return locations
}

// Default search radii
const (
	DefaultSearchRadiusKm = 25.0
	MaxSearchRadiusKm     = 100.0
	NearbyRadiusKm        = 1.0
)
