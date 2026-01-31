package handler

import (
	"net/http"
	"strconv"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// ResonanceHandler handles resonance scoring endpoints
type ResonanceHandler struct {
	resonanceService *service.ResonanceService
}

// NewResonanceHandler creates a new resonance handler
func NewResonanceHandler(resonanceService *service.ResonanceService) *ResonanceHandler {
	return &ResonanceHandler{
		resonanceService: resonanceService,
	}
}

// GetMyResonance handles GET /v1/resonance - get own resonance score
func (h *ResonanceHandler) GetMyResonance(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	score, err := h.resonanceService.GetUserScore(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get resonance score"))
		return
	}

	// Convert to display format
	display := model.ResonanceDisplay{
		Total:      score.Total,
		Questing:   score.Questing,
		Mana:       score.Mana,
		Wayfinder:  score.Wayfinder,
		Attunement: score.Attunement,
		Nexus:      score.Nexus,
	}

	WriteData(w, http.StatusOK, display, map[string]string{
		"self":   "/v1/resonance",
		"ledger": "/v1/resonance/ledger",
	})
}

// GetUserResonance handles GET /v1/users/{userId}/resonance - get another user's resonance
func (h *ResonanceHandler) GetUserResonance(w http.ResponseWriter, r *http.Request) {
	_ = middleware.GetUserID(r.Context()) // Auth check
	if middleware.GetUserID(r.Context()) == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	targetUserID := r.PathValue("userId")
	if targetUserID == "" {
		WriteError(w, model.NewBadRequestError("user ID required"))
		return
	}

	score, err := h.resonanceService.GetUserScore(r.Context(), targetUserID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get resonance score"))
		return
	}

	// Only show total for other users (privacy)
	display := model.ResonanceDisplay{
		Total: score.Total,
	}

	WriteData(w, http.StatusOK, display, map[string]string{
		"self": "/v1/users/" + targetUserID + "/resonance",
	})
}

// GetLedger handles GET /v1/resonance/ledger - get point history
func (h *ResonanceHandler) GetLedger(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	limit := 50
	if r.URL.Query().Get("limit") != "" {
		if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if r.URL.Query().Get("offset") != "" {
		if o, err := strconv.Atoi(r.URL.Query().Get("offset")); err == nil && o >= 0 {
			offset = o
		}
	}

	entries, err := h.resonanceService.GetUserLedger(r.Context(), userID, limit, offset)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to get resonance ledger"))
		return
	}

	WriteCollection(w, http.StatusOK, entries, nil, map[string]string{
		"self": "/v1/resonance/ledger",
	})
}

// RecalculateScore handles POST /v1/resonance/recalculate - force recalculation (admin only)
func (h *ResonanceHandler) RecalculateScore(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	score, err := h.resonanceService.RecalculateScore(r.Context(), userID)
	if err != nil {
		WriteError(w, model.NewInternalError("failed to recalculate score"))
		return
	}

	display := model.ResonanceDisplay{
		Total:      score.Total,
		Questing:   score.Questing,
		Mana:       score.Mana,
		Wayfinder:  score.Wayfinder,
		Attunement: score.Attunement,
		Nexus:      score.Nexus,
	}

	WriteData(w, http.StatusOK, display, map[string]string{
		"self": "/v1/resonance",
	})
}

// GetResonanceExplainer handles GET /v1/resonance/explain - explain scoring system
func (h *ResonanceHandler) GetResonanceExplainer(w http.ResponseWriter, r *http.Request) {
	explainer := map[string]interface{}{
		"stats": []map[string]interface{}{
			{
				"name":        "Questing",
				"icon":        "compass.fill",
				"description": "Points for showing up to events you committed to",
				"daily_cap":   model.DailyCapQuesting,
				"earning": map[string]int{
					"verified_completion": model.PointsQuestingBase,
					"early_confirm_bonus": model.PointsQuestingEarlyConfirm,
					"on_time_checkin":     model.PointsQuestingCheckin,
				},
			},
			{
				"name":        "Mana",
				"icon":        "sparkles",
				"description": "Points for support sessions where the other person said it helped",
				"daily_cap":   model.DailyCapMana,
				"earning": map[string]int{
					"helpful_session": model.PointsManaBase,
					"early_confirm":   model.PointsManaEarlyConfirm,
					"feedback_tag":    model.PointsManaFeedbackTag,
				},
				"notes": "Diminishing returns: sessions 1-3 with same person = 100%, 4-6 = 50%, 7+ = 25%",
			},
			{
				"name":        "Wayfinder",
				"icon":        "map.fill",
				"description": "Points for hosting events that people attended",
				"daily_cap":   model.DailyCapWayfinder,
				"earning": map[string]interface{}{
					"verified_hosting": model.PointsWayfinderBase,
					"per_attendee":     model.PointsWayfinderPerAttendee,
					"max_attendees":    model.PointsWayfinderMaxAttendees,
					"early_confirm":    model.PointsWayfinderEarlyConfirm,
				},
			},
			{
				"name":        "Attunement",
				"icon":        "slider.horizontal.3",
				"description": "Points for answering matching questions",
				"daily_cap":   model.DailyCapAttunement,
				"earning": map[string]int{
					"answer_question":   model.PointsAttunementQuestion,
					"monthly_refresh":   model.PointsAttunementProfileRefresh,
				},
			},
			{
				"name":        "Nexus",
				"icon":        "network",
				"description": "Points for being active in healthy circles",
				"monthly_cap": model.MonthlyCapNexus,
				"notes":       "Calculated monthly based on circle activity and cross-circle bridging",
			},
		},
		"principles": []string{
			"No punishments - you can fail to earn points, never lose them",
			"All awards require verification from other users",
			"Pairwise diminishing returns prevent collusion",
			"Every point award is auditable in your ledger",
		},
	}

	WriteData(w, http.StatusOK, explainer, map[string]string{
		"self": "/v1/resonance/explain",
	})
}
