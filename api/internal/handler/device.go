package handler

import (
	"context"
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
)

// DeviceTokenRepository interface for the handler
type DeviceTokenRepository interface {
	Create(ctx context.Context, token *model.DeviceToken) error
	GetByUserID(ctx context.Context, userID string) ([]*model.DeviceToken, error)
	GetByID(ctx context.Context, id string) (*model.DeviceToken, error)
	Delete(ctx context.Context, id string) error
	CountByUserID(ctx context.Context, userID string) (int, error)
	UpsertByToken(ctx context.Context, token *model.DeviceToken) error
}

// DeviceHandler handles device token HTTP requests
type DeviceHandler struct {
	deviceRepo DeviceTokenRepository
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(deviceRepo DeviceTokenRepository) *DeviceHandler {
	return &DeviceHandler{deviceRepo: deviceRepo}
}

// Register handles POST /v1/devices - register a device token
func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req model.RegisterDeviceRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	// Validate request
	if errors := req.Validate(); len(errors) > 0 {
		WriteError(w, model.NewValidationError(errors))
		return
	}

	// Check device limit
	count, err := h.deviceRepo.CountByUserID(ctx, userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to count devices"))
		return
	}
	if count >= model.MaxDevicesPerUser {
		WriteError(w, model.NewLimitExceededError(
			"maximum devices per user reached",
			model.MaxDevicesPerUser,
			model.MaxDevicesPerUser,
		))
		return
	}

	// Create device token
	device := &model.DeviceToken{
		UserID:   userID,
		Platform: req.Platform,
		Token:    req.Token,
		Name:     req.Name,
		Active:   true,
	}

	// Use upsert to handle re-registration of same token
	if err := h.deviceRepo.UpsertByToken(ctx, device); err != nil {
		WriteError(w, model.NewInternalError("failed to register device"))
		return
	}

	WriteData(w, http.StatusCreated, device, nil)
}

// List handles GET /v1/devices - list user's registered devices
func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	devices, err := h.deviceRepo.GetByUserID(ctx, userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to list devices"))
		return
	}

	// Don't expose full tokens in the response
	sanitized := make([]map[string]interface{}, len(devices))
	for i, d := range devices {
		sanitized[i] = map[string]interface{}{
			"id":         d.ID,
			"platform":   d.Platform,
			"name":       d.Name,
			"active":     d.Active,
			"created_on": d.CreatedOn,
			"last_used":  d.LastUsed,
		}
	}

	WriteData(w, http.StatusOK, sanitized, nil)
}

// Delete handles DELETE /v1/devices/{deviceId} - unregister a device
func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := middleware.GetUserID(ctx)
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	deviceID := r.PathValue("deviceId")
	if deviceID == "" {
		WriteError(w, model.NewBadRequestError("device ID required"))
		return
	}

	// Verify ownership
	device, err := h.deviceRepo.GetByID(ctx, deviceID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get device"))
		return
	}
	if device == nil {
		WriteError(w, model.NewNotFoundError("device not found"))
		return
	}
	if device.UserID != userID {
		WriteError(w, model.NewNotFoundError("device not found"))
		return
	}

	// Delete
	if err := h.deviceRepo.Delete(ctx, deviceID); err != nil {
		WriteError(w, model.NewInternalError("failed to delete device"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
