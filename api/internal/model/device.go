package model

import "time"

// DevicePlatform represents a device's platform type
type DevicePlatform string

const (
	PlatformIOS     DevicePlatform = "ios"
	PlatformAndroid DevicePlatform = "android"
	PlatformWeb     DevicePlatform = "web"
)

// IsValid returns true if the platform is valid
func (p DevicePlatform) IsValid() bool {
	switch p {
	case PlatformIOS, PlatformAndroid, PlatformWeb:
		return true
	default:
		return false
	}
}

// DeviceToken represents a device push notification token
type DeviceToken struct {
	ID        string         `json:"id"`
	UserID    string         `json:"user_id"`
	Platform  DevicePlatform `json:"platform"`
	Token     string         `json:"token"`
	Name      string         `json:"name,omitempty"` // e.g., "iPhone 15"
	Active    bool           `json:"active"`
	CreatedOn time.Time      `json:"created_on"`
	UpdatedOn time.Time      `json:"updated_on"`
	LastUsed  *time.Time     `json:"last_used,omitempty"`
}

// RegisterDeviceRequest represents a request to register a device token
type RegisterDeviceRequest struct {
	Platform DevicePlatform `json:"platform"`
	Token    string         `json:"token"`
	Name     string         `json:"name,omitempty"`
}

// Validate validates the register device request
func (r *RegisterDeviceRequest) Validate() []FieldError {
	var errors []FieldError

	if !r.Platform.IsValid() {
		errors = append(errors, FieldError{
			Field:   "platform",
			Message: "platform must be ios, android, or web",
		})
	}

	if r.Token == "" {
		errors = append(errors, FieldError{
			Field:   "token",
			Message: "token is required",
		})
	}

	if len(r.Token) > 4096 {
		errors = append(errors, FieldError{
			Field:   "token",
			Message: "token exceeds maximum length",
		})
	}

	if len(r.Name) > 100 {
		errors = append(errors, FieldError{
			Field:   "name",
			Message: "name exceeds maximum length",
		})
	}

	return errors
}

// Business constraints for devices
const (
	MaxDevicesPerUser = 10
)
