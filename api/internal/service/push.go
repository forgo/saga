package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// Push service errors
var (
	ErrPushDisabled       = errors.New("push notifications are disabled")
	ErrNoDeviceTokens     = errors.New("no device tokens found for user")
	ErrInvalidDeviceToken = errors.New("invalid device token")
)

// PushNotification represents a push notification to send
type PushNotification struct {
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	ImageURL string            `json:"image_url,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
	Badge    *int              `json:"badge,omitempty"`
	Sound    string            `json:"sound,omitempty"`
}

// PushResult represents the result of sending a push notification
type PushResult struct {
	Success      bool   `json:"success"`
	DeviceToken  string `json:"device_token"`
	MessageID    string `json:"message_id,omitempty"`
	Error        string `json:"error,omitempty"`
	ShouldRetry  bool   `json:"should_retry"`
	TokenInvalid bool   `json:"token_invalid"`
}

// DeviceTokenRepository interface for push service
type DeviceTokenRepository interface {
	GetByUserID(ctx context.Context, userID string) ([]*model.DeviceToken, error)
	GetByToken(ctx context.Context, token string) (*model.DeviceToken, error)
	MarkInactive(ctx context.Context, token string) error
	UpdateLastUsed(ctx context.Context, id string) error
}

// PushService handles sending push notifications
type PushService struct {
	deviceRepo DeviceTokenRepository
	enabled    bool
}

// PushServiceConfig holds configuration for the push service
type PushServiceConfig struct {
	DeviceRepo         DeviceTokenRepository
	Enabled            bool
	FCMCredentialsPath string
}

// NewPushService creates a new push service
func NewPushService(cfg PushServiceConfig) (*PushService, error) {
	svc := &PushService{
		deviceRepo: cfg.DeviceRepo,
		enabled:    cfg.Enabled,
	}

	if cfg.Enabled && cfg.FCMCredentialsPath != "" {
		// TODO: Initialize Firebase Admin SDK when adding real FCM support
		// This would look something like:
		//
		// opt := option.WithCredentialsFile(cfg.FCMCredentialsPath)
		// app, err := firebase.NewApp(context.Background(), nil, opt)
		// if err != nil {
		//     return nil, fmt.Errorf("initializing firebase: %w", err)
		// }
		// client, err := app.Messaging(context.Background())
		// if err != nil {
		//     return nil, fmt.Errorf("initializing firebase messaging: %w", err)
		// }
		// svc.fcmClient = client
		//
		log.Printf("[PushService] FCM credentials path configured but Firebase SDK not yet integrated")
	}

	return svc, nil
}

// IsEnabled returns whether push notifications are enabled
func (s *PushService) IsEnabled() bool {
	return s.enabled
}

// SendToUser sends a push notification to all of a user's devices
func (s *PushService) SendToUser(ctx context.Context, userID string, notification *PushNotification) ([]PushResult, error) {
	if !s.enabled {
		return nil, ErrPushDisabled
	}

	if s.deviceRepo == nil {
		return nil, errors.New("device repository not configured")
	}

	devices, err := s.deviceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting devices: %w", err)
	}

	if len(devices) == 0 {
		return nil, ErrNoDeviceTokens
	}

	results := make([]PushResult, 0, len(devices))
	for _, device := range devices {
		result := s.sendToDevice(ctx, device, notification)
		results = append(results, result)

		// Handle invalid tokens
		if result.TokenInvalid {
			if err := s.deviceRepo.MarkInactive(ctx, device.Token); err != nil {
				log.Printf("[PushService] Failed to mark token inactive: %v", err)
			}
		}

		// Update last used on success
		if result.Success {
			if err := s.deviceRepo.UpdateLastUsed(ctx, device.ID); err != nil {
				log.Printf("[PushService] Failed to update last used: %v", err)
			}
		}
	}

	return results, nil
}

// sendToDevice sends a push notification to a specific device
func (s *PushService) sendToDevice(ctx context.Context, device *model.DeviceToken, notification *PushNotification) PushResult {
	result := PushResult{
		DeviceToken: device.Token,
	}

	// TODO: Replace with actual FCM implementation
	// When Firebase is integrated, this would look like:
	//
	// message := &messaging.Message{
	//     Token: device.Token,
	//     Notification: &messaging.Notification{
	//         Title:    notification.Title,
	//         Body:     notification.Body,
	//         ImageURL: notification.ImageURL,
	//     },
	//     Data: notification.Data,
	// }
	//
	// // Platform-specific config
	// switch device.Platform {
	// case model.PlatformIOS:
	//     message.APNS = &messaging.APNSConfig{
	//         Payload: &messaging.APNSPayload{
	//             Aps: &messaging.Aps{
	//                 Sound: notification.Sound,
	//                 Badge: notification.Badge,
	//             },
	//         },
	//     }
	// case model.PlatformAndroid:
	//     message.Android = &messaging.AndroidConfig{
	//         Priority: "high",
	//     }
	// }
	//
	// messageID, err := s.fcmClient.Send(ctx, message)
	// if err != nil {
	//     // Check for specific error types
	//     if messaging.IsUnregistered(err) || messaging.IsInvalidArgument(err) {
	//         result.TokenInvalid = true
	//     } else if messaging.IsUnavailable(err) {
	//         result.ShouldRetry = true
	//     }
	//     result.Error = err.Error()
	//     return result
	// }
	//
	// result.Success = true
	// result.MessageID = messageID
	// return result

	// Stub implementation - log and succeed
	log.Printf("[PushService] Would send push to %s (%s): %s - %s",
		device.Platform, maskToken(device.Token), notification.Title, notification.Body)

	result.Success = true
	result.MessageID = fmt.Sprintf("stub_%d", time.Now().UnixNano())
	return result
}

// SendMulticast sends a notification to multiple users
func (s *PushService) SendMulticast(ctx context.Context, userIDs []string, notification *PushNotification) (map[string][]PushResult, error) {
	if !s.enabled {
		return nil, ErrPushDisabled
	}

	results := make(map[string][]PushResult)
	for _, userID := range userIDs {
		userResults, err := s.SendToUser(ctx, userID, notification)
		if err != nil {
			if !errors.Is(err, ErrNoDeviceTokens) {
				log.Printf("[PushService] Failed to send to user %s: %v", userID, err)
			}
			continue
		}
		results[userID] = userResults
	}

	return results, nil
}

// maskToken masks a device token for logging
func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
