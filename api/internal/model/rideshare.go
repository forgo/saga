package model

import "time"

// RideshareStatus constants
const (
	RideshareStatusOpen      = "open"      // Accepting passengers
	RideshareStatusFull      = "full"      // All seats taken
	RideshareStatusDeparted  = "departed"  // Trip in progress
	RideshareStatusCompleted = "completed" // Trip finished
	RideshareStatusCancelled = "cancelled" // Trip cancelled
)

// Rideshare represents transportation coordination attached to an Event or Adventure (formerly Commute)
type Rideshare struct {
	ID             string    `json:"id"`
	EventID        *string   `json:"event_id,omitempty"`     // Attached to event
	AdventureID    *string   `json:"adventure_id,omitempty"` // Attached to adventure (inter-event transport)
	DriverID       string    `json:"driver_id"`              // User ID of driver
	Title          string    `json:"title"`                  // e.g., "Ride to the hike"
	Description    *string   `json:"description,omitempty"`
	Origin         RideshareLocation  `json:"origin"`
	Destination    RideshareLocation  `json:"destination"`
	DepartureTime  time.Time `json:"departure_time"`
	ArrivalTime    *time.Time `json:"arrival_time,omitempty"`
	SeatsTotal     int       `json:"seats_total"`     // Total passenger seats
	SeatsAvailable int       `json:"seats_available"` // Computed from bookings
	Status         string    `json:"status"`          // open, full, departed, completed, cancelled
	TrustRequired  bool      `json:"trust_required"`  // Requires mutual trust
	CreatedOn      time.Time `json:"created_on"`
	UpdatedOn      time.Time `json:"updated_on"`
}

// RideshareLocation represents a location point for rideshares (privacy-safe)
type RideshareLocation struct {
	Name         string  `json:"name"`                   // e.g., "Near Coffee Bean on Main St"
	Description  *string `json:"description,omitempty"`  // Additional details
	Address      *string `json:"address,omitempty"`      // Full address (only for confirmed riders)
	Neighborhood *string `json:"neighborhood,omitempty"` // General area
	City         string  `json:"city"`
	Country      *string `json:"country,omitempty"`
	// Internal coordinates - never exposed directly
	Lat float64 `json:"-"`
	Lng float64 `json:"-"`
}

// RideshareSegment represents a leg of the journey with pick-up/drop-off points
type RideshareSegment struct {
	ID              string            `json:"id"`
	RideshareID     string            `json:"rideshare_id"`
	SequenceOrder   int               `json:"sequence_order"` // Order in the route
	PickupPoint     RideshareLocation `json:"pickup_point"`
	DropoffPoint    RideshareLocation `json:"dropoff_point"`
	EstimatedMinutes *int    `json:"estimated_minutes,omitempty"` // Minutes for this segment
	Notes           *string  `json:"notes,omitempty"`
}

// RideshareSeatStatus constants
const (
	SeatStatusRequested = "requested" // Passenger requested seat
	SeatStatusConfirmed = "confirmed" // Driver confirmed
	SeatStatusCancelled = "cancelled" // Cancelled by either party
)

// RideshareSeat represents a seat/slot in a rideshare
type RideshareSeat struct {
	ID               string     `json:"id"`
	RideshareID      string     `json:"rideshare_id"`
	PassengerID      string     `json:"passenger_id"` // User ID
	Status           string     `json:"status"`       // requested, confirmed, cancelled
	PickupSegmentID  *string    `json:"pickup_segment_id,omitempty"`
	DropoffSegmentID *string    `json:"dropoff_segment_id,omitempty"`
	RequestedOn      time.Time  `json:"requested_on"`
	ConfirmedOn      *time.Time `json:"confirmed_on,omitempty"`
	Notes            *string    `json:"notes,omitempty"` // Passenger's notes
}

// RideshareMatch represents a potential rideshare match
type RideshareMatch struct {
	Rideshare      Rideshare    `json:"rideshare"`
	DriverID       string       `json:"driver_id"`
	MatchScore     float64      `json:"match_score"`
	DistanceKm     float64      `json:"distance_km,omitempty"` // Approximate route overlap
	TimeOverlap    bool         `json:"time_overlap"`
	TrustStatus    TrustSummary `json:"trust_status"`
	AvailableSeats int          `json:"available_seats"`
}

// RideshareWithSeats includes full seat information
type RideshareWithSeats struct {
	Rideshare Rideshare        `json:"rideshare"`
	Segments  []RideshareSegment `json:"segments"`
	Seats     []RideshareSeat  `json:"seats"`
}

// Constraints
const (
	MaxSeatsPerRideshare    = 8
	MaxSegmentsPerRideshare = 10
	MaxRidesharesPerEvent   = 10
	MaxRideshareDescLength  = 500
	MaxLocationNameLength   = 100
)

// CreateRideshareRequest represents a request to create a rideshare
type CreateRideshareRequest struct {
	EventID       *string           `json:"event_id,omitempty"`
	AdventureID   *string           `json:"adventure_id,omitempty"`
	Title         string            `json:"title"`
	Description   *string           `json:"description,omitempty"`
	Origin        RideshareLocation `json:"origin"`
	Destination   RideshareLocation `json:"destination"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   *time.Time `json:"arrival_time,omitempty"`
	SeatsTotal    int       `json:"seats_total"`
	TrustRequired bool      `json:"trust_required"`
}

// UpdateRideshareRequest represents a request to update a rideshare
type UpdateRideshareRequest struct {
	Title         *string            `json:"title,omitempty"`
	Description   *string            `json:"description,omitempty"`
	Origin        *RideshareLocation `json:"origin,omitempty"`
	Destination   *RideshareLocation `json:"destination,omitempty"`
	DepartureTime *time.Time         `json:"departure_time,omitempty"`
	ArrivalTime   *time.Time         `json:"arrival_time,omitempty"`
	SeatsTotal    *int               `json:"seats_total,omitempty"`
	Status        *string            `json:"status,omitempty"`
	TrustRequired *bool              `json:"trust_required,omitempty"`
}

// AddSegmentRequest represents a request to add a route segment
type AddRideshareSegmentRequest struct {
	PickupPoint      RideshareLocation `json:"pickup_point"`
	DropoffPoint     RideshareLocation `json:"dropoff_point"`
	EstimatedMinutes *int              `json:"estimated_minutes,omitempty"`
	Notes            *string           `json:"notes,omitempty"`
}

// RequestSeatRequest represents a request to book a seat
type RequestSeatRequest struct {
	PickupSegmentID  *string `json:"pickup_segment_id,omitempty"`
	DropoffSegmentID *string `json:"dropoff_segment_id,omitempty"`
	Notes            *string `json:"notes,omitempty"`
}

// RespondToSeatRequest represents driver's response to seat request
type RespondToSeatRequest struct {
	Confirmed bool    `json:"confirmed"`
	Notes     *string `json:"notes,omitempty"`
}

// RideshareSearchFilters for finding rideshares
type RideshareSearchFilters struct {
	EventID         *string    `json:"event_id,omitempty"`
	AdventureID     *string    `json:"adventure_id,omitempty"`
	DepartureAfter  *time.Time `json:"departure_after,omitempty"`
	DepartureBefore *time.Time `json:"departure_before,omitempty"`
	City            *string    `json:"city,omitempty"`
	TrustedOnly     bool       `json:"trusted_only"`
	AvailableOnly   bool       `json:"available_only"` // Only show rideshares with open seats
}

// Backward compatibility type aliases (deprecated, will be removed)
type Commute = Rideshare
type CommuteLocation = RideshareLocation
type CommuteSegment = RideshareSegment
type CommuteSeat = RideshareSeat
type CommuteMatch = RideshareMatch
type CommuteWithSeats = RideshareWithSeats
