package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents the register endpoint request body
type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
}

// LoginRequest represents the login endpoint request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents the refresh endpoint request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Username      *string `json:"username,omitempty"`
	Firstname     *string `json:"firstname,omitempty"`
	Lastname      *string `json:"lastname,omitempty"`
	EmailVerified bool    `json:"email_verified"`
	CreatedOn     string  `json:"created_on"`
	UpdatedOn     string  `json:"updated_on"`
}

// Register handles POST /v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req RegisterRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	result, err := h.authService.Register(r.Context(), service.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		Firstname: req.Firstname,
		Lastname:  req.Lastname,
	})

	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	// Build response
	response := struct {
		User  UserResponse  `json:"user"`
		Token TokenResponse `json:"token"`
	}{
		User:  toUserResponse(result.User),
		Token: toTokenResponse(result.TokenPair),
	}

	WriteData(w, http.StatusCreated, response, map[string]string{
		"self": "/v1/auth/me",
	})
}

// Login handles POST /v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req LoginRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	result, err := h.authService.Login(r.Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		h.handleAuthError(w, err)
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

// Refresh handles POST /v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req RefreshRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if req.RefreshToken == "" {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "refresh_token", Message: "refresh_token is required"},
		}))
		return
	}

	tokenPair, err := h.authService.RefreshTokens(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleAuthError(w, err)
		return
	}

	WriteData(w, http.StatusOK, toTokenResponse(tokenPair), nil)
}

// Logout handles POST /v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	if err := h.authService.Logout(r.Context(), userID); err != nil {
		WriteError(w, model.NewInternalError("logout failed"))
		return
	}

	WriteNoContent(w)
}

// Me handles GET /v1/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, model.NewMethodNotAllowedError("GET"))
		return
	}

	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		WriteError(w, model.NewUnauthorizedError("authentication required"))
		return
	}

	userWithIdentities, err := h.authService.GetUserWithIdentities(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			WriteError(w, model.NewNotFoundError("user"))
			return
		}
		WriteError(w, model.NewInternalError("failed to get user"))
		return
	}

	response := struct {
		User       UserResponse       `json:"user"`
		Identities []IdentityResponse `json:"identities"`
		Passkeys   []PasskeyResponse  `json:"passkeys"`
	}{
		User:       toUserResponse(userWithIdentities.User),
		Identities: toIdentitiesResponse(userWithIdentities.Identities),
		Passkeys:   toPasskeysResponse(userWithIdentities.Passkeys),
	}

	WriteData(w, http.StatusOK, response, map[string]string{
		"self": "/v1/auth/me",
	})
}

// IdentityResponse represents an identity in API responses
type IdentityResponse struct {
	ID            string `json:"id"`
	Provider      string `json:"provider"`
	ProviderEmail string `json:"provider_email,omitempty"`
	CreatedOn     string `json:"created_on"`
}

// PasskeyResponse represents a passkey in API responses
type PasskeyResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	CreatedOn  string  `json:"created_on"`
	LastUsedOn *string `json:"last_used_on,omitempty"`
}

func (h *AuthHandler) handleAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidCredentials):
		WriteError(w, model.NewUnauthorizedError("invalid email or password"))
	case errors.Is(err, service.ErrEmailAlreadyExists):
		WriteError(w, model.NewConflictError("email already registered"))
	case errors.Is(err, service.ErrUserNotFound):
		WriteError(w, model.NewNotFoundError("user"))
	case errors.Is(err, service.ErrPasswordRequired):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "password", Message: "password is required"},
		}))
	case errors.Is(err, service.ErrPasswordTooShort):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "password", Message: "password must be at least 8 characters"},
		}))
	case errors.Is(err, service.ErrPasswordTooLong):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "password", Message: "password must be at most 128 characters"},
		}))
	case errors.Is(err, service.ErrInvalidEmail):
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "email", Message: "invalid email format"},
		}))
	case errors.Is(err, service.ErrInvalidRefreshToken),
		errors.Is(err, service.ErrRefreshTokenExpired),
		errors.Is(err, service.ErrRefreshTokenRevoked):
		WriteError(w, model.NewUnauthorizedError("invalid or expired refresh token"))
	default:
		slog.Error("unhandled auth error", "error", err)
		WriteError(w, model.NewInternalError("authentication error"))
	}
}

// Helper functions

func toUserResponse(user *model.User) UserResponse {
	return UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		Firstname:     user.Firstname,
		Lastname:      user.Lastname,
		EmailVerified: user.EmailVerified,
		CreatedOn:     user.CreatedOn.Format("2006-01-02T15:04:05Z"),
		UpdatedOn:     user.UpdatedOn.Format("2006-01-02T15:04:05Z"),
	}
}

func toTokenResponse(tokenPair *service.TokenPair) TokenResponse {
	return TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
	}
}

func toIdentitiesResponse(identities []*model.Identity) []IdentityResponse {
	result := make([]IdentityResponse, 0, len(identities))
	for _, identity := range identities {
		resp := IdentityResponse{
			ID:        identity.ID,
			Provider:  identity.Provider,
			CreatedOn: identity.CreatedOn.Format("2006-01-02T15:04:05Z"),
		}
		if identity.ProviderEmail != nil {
			resp.ProviderEmail = *identity.ProviderEmail
		}
		result = append(result, resp)
	}
	return result
}

func toPasskeysResponse(passkeys []*model.Passkey) []PasskeyResponse {
	result := make([]PasskeyResponse, 0, len(passkeys))
	for _, passkey := range passkeys {
		resp := PasskeyResponse{
			ID:        passkey.ID,
			Name:      passkey.Name,
			CreatedOn: passkey.CreatedOn.Format("2006-01-02T15:04:05Z"),
		}
		if passkey.LastUsedOn != nil {
			formatted := passkey.LastUsedOn.Format("2006-01-02T15:04:05Z")
			resp.LastUsedOn = &formatted
		}
		result = append(result, resp)
	}
	return result
}
