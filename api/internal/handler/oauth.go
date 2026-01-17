package handler

import (
	"errors"
	"net/http"

	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// OAuthHandler handles OAuth authentication endpoints
type OAuthHandler struct {
	oauthService *service.OAuthService
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
	}
}

// OAuthRequest represents an OAuth callback request body
type OAuthCallbackRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
	State        string `json:"state,omitempty"`
}

// OAuthResponse represents a successful OAuth response
type OAuthSuccessResponse struct {
	User      UserResponse  `json:"user"`
	Token     TokenResponse `json:"token"`
	IsNewUser bool          `json:"is_new_user"`
}

// LinkRequiredResponse indicates account linking is needed
type LinkRequiredResponse struct {
	LinkRequired bool   `json:"link_required"`
	LinkToken    string `json:"link_token"`
	Email        string `json:"email"`
	Message      string `json:"message"`
}

// Google handles POST /v1/auth/oauth/google
func (h *OAuthHandler) Google(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req OAuthCallbackRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if req.Code == "" {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "code", Message: "authorization code is required"},
		}))
		return
	}

	if req.CodeVerifier == "" {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "code_verifier", Message: "PKCE code verifier is required"},
		}))
		return
	}

	result, err := h.oauthService.AuthenticateGoogle(r.Context(), service.OAuthRequest{
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		State:        req.State,
	})

	if err != nil {
		h.handleOAuthError(w, err)
		return
	}

	h.writeOAuthResult(w, result)
}

// Apple handles POST /v1/auth/oauth/apple
func (h *OAuthHandler) Apple(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, model.NewMethodNotAllowedError("POST"))
		return
	}

	var req OAuthCallbackRequest
	if err := DecodeJSON(r, &req); err != nil {
		WriteError(w, model.NewBadRequestError("invalid request body"))
		return
	}

	if req.Code == "" {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "code", Message: "authorization code is required"},
		}))
		return
	}

	if req.CodeVerifier == "" {
		WriteError(w, model.NewValidationError([]model.FieldError{
			{Field: "code_verifier", Message: "PKCE code verifier is required"},
		}))
		return
	}

	result, err := h.oauthService.AuthenticateApple(r.Context(), service.OAuthRequest{
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		State:        req.State,
	})

	if err != nil {
		h.handleOAuthError(w, err)
		return
	}

	h.writeOAuthResult(w, result)
}

func (h *OAuthHandler) writeOAuthResult(w http.ResponseWriter, result *service.OAuthResult) {
	if result.LinkRequired {
		// Account linking required
		response := LinkRequiredResponse{
			LinkRequired: true,
			LinkToken:    result.LinkToken,
			Email:        result.ExistingUser.Email,
			Message:      "An account with this email already exists. Please authenticate with your existing account to link.",
		}
		WriteData(w, http.StatusOK, response, nil)
		return
	}

	// Successful authentication
	response := OAuthSuccessResponse{
		User:      toUserResponse(result.User),
		Token:     toTokenResponse(result.TokenPair),
		IsNewUser: result.IsNewUser,
	}

	status := http.StatusOK
	if result.IsNewUser {
		status = http.StatusCreated
	}

	WriteData(w, status, response, map[string]string{
		"self": "/v1/auth/me",
	})
}

func (h *OAuthHandler) handleOAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidAuthCode):
		WriteError(w, model.NewBadRequestError("invalid authorization code"))
	case errors.Is(err, service.ErrPKCEVerifyFailed):
		WriteError(w, model.NewBadRequestError("PKCE verification failed"))
	case errors.Is(err, service.ErrProviderError):
		WriteError(w, model.NewBadRequestError("OAuth provider error: "+err.Error()))
	case errors.Is(err, service.ErrInvalidIDToken):
		WriteError(w, model.NewBadRequestError("invalid ID token from provider"))
	case errors.Is(err, service.ErrEmailNotVerified):
		WriteError(w, model.NewBadRequestError("email not verified by OAuth provider"))
	case errors.Is(err, service.ErrUserNotFound):
		WriteError(w, model.NewNotFoundError("user"))
	default:
		WriteError(w, model.NewInternalError("OAuth authentication failed"))
	}
}
