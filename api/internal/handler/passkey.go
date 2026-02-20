package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// PasskeyHandler handles passkey/WebAuthn endpoints
type PasskeyHandler struct {
	passkeyService *service.PasskeyService
}

// NewPasskeyHandler creates a new passkey handler
func NewPasskeyHandler(passkeyService *service.PasskeyService) *PasskeyHandler {
	return &PasskeyHandler{
		passkeyService: passkeyService,
	}
}

// PasskeyRegisterStartResponse represents the register start response.
type PasskeyRegisterStartResponse struct {
	Challenge              string                          `json:"challenge"`
	RP                     service.RelyingPartyInfo        `json:"rp"`
	User                   service.WebAuthnUserInfo        `json:"user"`
	PubKeyCredParams       []service.PubKeyCredParam       `json:"pubKeyCredParams"`
	Timeout                int                             `json:"timeout"`
	Attestation            string                          `json:"attestation"`
	AuthenticatorSelection *service.AuthenticatorSelection `json:"authenticatorSelection,omitempty"`
	ExcludeCredentials     []service.CredentialDescriptor  `json:"excludeCredentials,omitempty"`
}

// RegisterStart handles POST /v1/auth/passkey/register/start
func (h *PasskeyHandler) RegisterStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	result, err := h.passkeyService.StartRegistration(r.Context(), service.RegistrationStartRequest{
		UserID: userID,
	})
	if err != nil {
		h.handlePasskeyError(w, err)
		return
	}

	response := PasskeyRegisterStartResponse{
		Challenge:              result.Challenge,
		RP:                     result.RP,
		User:                   result.User,
		PubKeyCredParams:       result.PubKeyCredParams,
		Timeout:                result.Timeout,
		Attestation:            result.Attestation,
		AuthenticatorSelection: result.AuthenticatorSelection,
		ExcludeCredentials:     result.ExcludeCredentials,
	}

	WriteData(w, http.StatusOK, response, nil)
}

// PasskeyRegisterFinishRequest represents the register finish request body.
type PasskeyRegisterFinishRequest struct {
	Credential *service.CredentialResponse `json:"credential"`
	Name       string                      `json:"name"`
}

// RegisterFinish handles POST /v1/auth/passkey/register/finish
func (h *PasskeyHandler) RegisterFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	var req PasskeyRegisterFinishRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if req.Credential == nil {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "credential", Message: "credential is required"},
		}))
		return
	}

	if req.Name == "" {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "name", Message: "passkey name is required"},
		}))
		return
	}

	result, err := h.passkeyService.FinishRegistration(r.Context(), service.RegistrationFinishRequest{
		UserID:     userID,
		Name:       req.Name,
		Credential: req.Credential,
	})
	if err != nil {
		h.handlePasskeyError(w, err)
		return
	}

	response := PasskeyResponse{
		ID:        result.Passkey.ID,
		Name:      result.Passkey.Name,
		CreatedOn: result.Passkey.CreatedOn.Format("2006-01-02T15:04:05Z"),
	}

	WriteData(w, http.StatusCreated, response, nil)
}

// PasskeyLoginStartRequest represents the login start request body.
type PasskeyLoginStartRequest struct {
	Email string `json:"email,omitempty"`
}

// LoginStart handles POST /v1/auth/passkey/login/start
func (h *PasskeyHandler) LoginStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req PasskeyLoginStartRequest
	// Allow empty body for discoverable credentials
	_ = DecodeJSON(r, &req)

	result, err := h.passkeyService.StartLogin(r.Context(), service.LoginStartRequest{
		Email: req.Email,
	})
	if err != nil {
		h.handlePasskeyError(w, err)
		return
	}

	WriteData(w, http.StatusOK, result, nil)
}

// PasskeyLoginFinishRequest represents the login finish request body.
type PasskeyLoginFinishRequest struct {
	Credential *service.AssertionResponse `json:"credential"`
}

// LoginFinish handles POST /v1/auth/passkey/login/finish
func (h *PasskeyHandler) LoginFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req PasskeyLoginFinishRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if req.Credential == nil {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "credential", Message: "credential is required"},
		}))
		return
	}

	result, err := h.passkeyService.FinishLogin(r.Context(), service.LoginFinishRequest{
		Credential: req.Credential,
	})
	if err != nil {
		h.handlePasskeyError(w, err)
		return
	}

	response := struct {
		User  UserResponse  `json:"user"`
		Token TokenResponse `json:"token"`
	}{
		User:  toUserResponse(result.User),
		Token: toTokenResponse(result.TokenPair),
	}

	WriteData(w, http.StatusOK, response, map[string]string{
		"self": "/v1/auth/me",
	})
}

// Delete handles DELETE /v1/auth/passkey/{id}
func (h *PasskeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		WriteError(w, model.NewMethodNotAllowedError("DELETE"))
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	// Extract passkey ID from URL path
	// Expected format: /v1/auth/passkey/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		WriteError(w, model.NewBadRequestError("invalid passkey ID"))
		return
	}
	passkeyID := parts[len(parts)-1]

	if err := h.passkeyService.DeletePasskey(r.Context(), userID, passkeyID); err != nil {
		h.handlePasskeyError(w, err)
		return
	}

	WriteNoContent(w)
}

func (h *PasskeyHandler) handlePasskeyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		WriteError(w, model.NewNotFoundError("user"))
	case errors.Is(err, service.ErrPasskeyNotFound):
		WriteError(w, model.NewNotFoundError("passkey"))
	case errors.Is(err, service.ErrInvalidChallenge):
		WriteError(w, model.NewBadRequestError("invalid or expired challenge"))
	case errors.Is(err, service.ErrInvalidCredential):
		WriteError(w, model.NewBadRequestError("invalid credential"))
	case errors.Is(err, service.ErrCredentialNotAllowed):
		WriteError(w, model.NewForbiddenError("credential not allowed"))
	case errors.Is(err, service.ErrSignCountMismatch):
		WriteError(w, model.NewBadRequestError("sign count mismatch - potential cloned authenticator"))
	case errors.Is(err, service.ErrPasskeyLimitReached):
		WriteError(w, model.NewLimitExceededError("passkeys per user", 10, 10))
	default:
		WriteError(w, model.NewInternalError("passkey operation failed"))
	}
}
