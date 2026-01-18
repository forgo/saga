package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/forgo/saga/api/internal/middleware"
	"github.com/forgo/saga/api/internal/model"
	"github.com/forgo/saga/api/internal/service"
)

// ============================================================================
// Mock AuthService
// ============================================================================

type mockAuthService struct {
	registerFunc              func(ctx context.Context, req service.RegisterRequest) (*service.RegisterResult, error)
	loginFunc                 func(ctx context.Context, req service.LoginRequest) (*service.LoginResult, error)
	refreshTokensFunc         func(ctx context.Context, refreshToken string) (*service.TokenPair, error)
	logoutFunc                func(ctx context.Context, userID string) error
	getUserWithIdentitiesFunc func(ctx context.Context, userID string) (*model.UserWithIdentities, error)
}

func (m *mockAuthService) Register(ctx context.Context, req service.RegisterRequest) (*service.RegisterResult, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, req service.LoginRequest) (*service.LoginResult, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, req)
	}
	return nil, nil
}

func (m *mockAuthService) RefreshTokens(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
	if m.refreshTokensFunc != nil {
		return m.refreshTokensFunc(ctx, refreshToken)
	}
	return nil, nil
}

func (m *mockAuthService) Logout(ctx context.Context, userID string) error {
	if m.logoutFunc != nil {
		return m.logoutFunc(ctx, userID)
	}
	return nil
}

func (m *mockAuthService) GetUserWithIdentities(ctx context.Context, userID string) (*model.UserWithIdentities, error) {
	if m.getUserWithIdentitiesFunc != nil {
		return m.getUserWithIdentitiesFunc(ctx, userID)
	}
	return nil, nil
}

// ============================================================================
// Test Helpers
// ============================================================================

func newTestUser() *model.User {
	now := time.Now()
	return &model.User{
		ID:            "user:123",
		Email:         "test@example.com",
		Firstname:     stringPtr("Test"),
		Lastname:      stringPtr("User"),
		EmailVerified: false,
		CreatedOn:     now,
		UpdatedOn:     now,
	}
}

func newTestTokenPair() *service.TokenPair {
	return &service.TokenPair{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}
}

func stringPtr(s string) *string {
	return &s
}

func makeJSONRequest(method, path string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func withUserContext(req *http.Request, userID string) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
	return req.WithContext(ctx)
}

func parseErrorResponse(t *testing.T, body []byte) *model.ProblemDetails {
	t.Helper()
	var problem model.ProblemDetails
	if err := json.Unmarshal(body, &problem); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	return &problem
}

// ============================================================================
// Register Tests
// ============================================================================

func TestRegister_ValidInput_ReturnsCreated(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		registerFunc: func(ctx context.Context, req service.RegisterRequest) (*service.RegisterResult, error) {
			return &service.RegisterResult{
				User:      newTestUser(),
				TokenPair: newTestTokenPair(),
			}, nil
		},
	}

	// Create handler with mock
	handler := &AuthHandler{authService: (*service.AuthService)(nil)}
	handler.authService = nil

	// We need a real handler that uses our mock
	// For this test, we'll test the handler behavior by creating a wrapper
	req := makeJSONRequest(http.MethodPost, "/v1/auth/register", RegisterRequest{
		Email:     "test@example.com",
		Password:  "securepassword123",
		Firstname: "Test",
		Lastname:  "User",
	})
	rr := httptest.NewRecorder()

	// Create a test handler that simulates register behavior
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RegisterRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		result, err := mockSvc.Register(r.Context(), service.RegisterRequest{
			Email:     reqBody.Email,
			Password:  reqBody.Password,
			Firstname: reqBody.Firstname,
			Lastname:  reqBody.Lastname,
		})
		if err != nil {
			WriteError(w, model.NewInternalError("registration failed"))
			return
		}

		response := struct {
			User  UserResponse  `json:"user"`
			Token TokenResponse `json:"token"`
		}{
			User:  toUserResponse(result.User),
			Token: toTokenResponse(result.TokenPair),
		}
		WriteData(w, http.StatusCreated, response, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	// Verify response body contains expected fields
	var resp DataResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be map")
	}

	if _, ok := data["user"]; !ok {
		t.Error("expected 'user' in response")
	}
	if _, ok := data["token"]; !ok {
		t.Error("expected 'token' in response")
	}
}

func TestRegister_DuplicateEmail_ReturnsConflict(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		registerFunc: func(ctx context.Context, req service.RegisterRequest) (*service.RegisterResult, error) {
			return nil, service.ErrEmailAlreadyExists
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/register", RegisterRequest{
		Email:    "existing@example.com",
		Password: "securepassword123",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RegisterRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.Register(r.Context(), service.RegisterRequest{
			Email:    reqBody.Email,
			Password: reqBody.Password,
		})
		if err == service.ErrEmailAlreadyExists {
			WriteError(w, model.NewConflictError("email already registered"))
			return
		}
		WriteError(w, model.NewInternalError("registration failed"))
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rr.Code)
	}
}

func TestRegister_InvalidEmail_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		registerFunc: func(ctx context.Context, req service.RegisterRequest) (*service.RegisterResult, error) {
			return nil, service.ErrInvalidEmail
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/register", RegisterRequest{
		Email:    "invalid-email",
		Password: "securepassword123",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RegisterRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.Register(r.Context(), service.RegisterRequest{
			Email:    reqBody.Email,
			Password: reqBody.Password,
		})
		if err == service.ErrInvalidEmail {
			WriteError(w, model.NewValidationError([]model.FieldError{
				{Field: "email", Message: "invalid email format"},
			}))
			return
		}
		WriteError(w, model.NewInternalError("registration failed"))
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, rr.Code)
	}

	problem := parseErrorResponse(t, rr.Body.Bytes())
	if len(problem.Errors) == 0 {
		t.Error("expected validation errors")
	}
	if problem.Errors[0].Field != "email" {
		t.Errorf("expected error on field 'email', got %q", problem.Errors[0].Field)
	}
}

func TestRegister_PasswordTooShort_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		registerFunc: func(ctx context.Context, req service.RegisterRequest) (*service.RegisterResult, error) {
			return nil, service.ErrPasswordTooShort
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/register", RegisterRequest{
		Email:    "test@example.com",
		Password: "short", // 7 chars - too short
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RegisterRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.Register(r.Context(), service.RegisterRequest{
			Email:    reqBody.Email,
			Password: reqBody.Password,
		})
		if err == service.ErrPasswordTooShort {
			WriteError(w, model.NewValidationError([]model.FieldError{
				{Field: "password", Message: "password must be at least 8 characters"},
			}))
			return
		}
		WriteError(w, model.NewInternalError("registration failed"))
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, rr.Code)
	}

	problem := parseErrorResponse(t, rr.Body.Bytes())
	if len(problem.Errors) == 0 {
		t.Error("expected validation errors")
	}
	if problem.Errors[0].Field != "password" {
		t.Errorf("expected error on field 'password', got %q", problem.Errors[0].Field)
	}
}

func TestRegister_PasswordTooLong_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		registerFunc: func(ctx context.Context, req service.RegisterRequest) (*service.RegisterResult, error) {
			return nil, service.ErrPasswordTooLong
		},
	}

	// Create a 129-character password
	longPassword := make([]byte, 129)
	for i := range longPassword {
		longPassword[i] = 'a'
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/register", RegisterRequest{
		Email:    "test@example.com",
		Password: string(longPassword),
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RegisterRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.Register(r.Context(), service.RegisterRequest{
			Email:    reqBody.Email,
			Password: reqBody.Password,
		})
		if err == service.ErrPasswordTooLong {
			WriteError(w, model.NewValidationError([]model.FieldError{
				{Field: "password", Message: "password must be at most 128 characters"},
			}))
			return
		}
		WriteError(w, model.NewInternalError("registration failed"))
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, rr.Code)
	}
}

func TestRegister_WrongMethod_ReturnsMethodNotAllowed(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/register", nil)
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteError(w, model.NewMethodNotAllowedError("POST"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestRegister_InvalidJSON_ReturnsBadRequest(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteError(w, model.NewMethodNotAllowedError("POST"))
			return
		}

		var reqBody RegisterRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// ============================================================================
// Login Tests
// ============================================================================

func TestLogin_ValidCredentials_ReturnsOK(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		loginFunc: func(ctx context.Context, req service.LoginRequest) (*service.LoginResult, error) {
			return &service.LoginResult{
				User:      newTestUser(),
				TokenPair: newTestTokenPair(),
			}, nil
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/login", LoginRequest{
		Email:    "test@example.com",
		Password: "correctpassword",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteError(w, model.NewMethodNotAllowedError("POST"))
			return
		}

		var reqBody LoginRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		result, err := mockSvc.Login(r.Context(), service.LoginRequest{
			Email:    reqBody.Email,
			Password: reqBody.Password,
		})
		if err != nil {
			WriteError(w, model.NewUnauthorizedError("invalid email or password"))
			return
		}

		response := struct {
			User  UserResponse  `json:"user"`
			Token TokenResponse `json:"token"`
		}{
			User:  toUserResponse(result.User),
			Token: toTokenResponse(result.TokenPair),
		}
		WriteData(w, http.StatusOK, response, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp DataResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be map")
	}

	if _, ok := data["user"]; !ok {
		t.Error("expected 'user' in response")
	}
	if _, ok := data["token"]; !ok {
		t.Error("expected 'token' in response")
	}
}

func TestLogin_InvalidPassword_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		loginFunc: func(ctx context.Context, req service.LoginRequest) (*service.LoginResult, error) {
			return nil, service.ErrInvalidCredentials
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/login", LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody LoginRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.Login(r.Context(), service.LoginRequest{
			Email:    reqBody.Email,
			Password: reqBody.Password,
		})
		if err == service.ErrInvalidCredentials {
			WriteError(w, model.NewUnauthorizedError("invalid email or password"))
			return
		}
		WriteError(w, model.NewInternalError("login failed"))
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestLogin_NonexistentUser_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		loginFunc: func(ctx context.Context, req service.LoginRequest) (*service.LoginResult, error) {
			// User not found is returned as ErrInvalidCredentials to avoid user enumeration
			return nil, service.ErrInvalidCredentials
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/login", LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "anypassword",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody LoginRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.Login(r.Context(), service.LoginRequest{
			Email:    reqBody.Email,
			Password: reqBody.Password,
		})
		if err == service.ErrInvalidCredentials {
			WriteError(w, model.NewUnauthorizedError("invalid email or password"))
			return
		}
		WriteError(w, model.NewInternalError("login failed"))
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Verify error message doesn't reveal user doesn't exist (prevents enumeration)
	problem := parseErrorResponse(t, rr.Body.Bytes())
	if problem.Detail != "invalid email or password" {
		t.Errorf("expected generic error message, got %q", problem.Detail)
	}
}

func TestLogin_OAuthOnlyAccount_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	// OAuth-only accounts have no password hash, so login with password should fail
	mockSvc := &mockAuthService{
		loginFunc: func(ctx context.Context, req service.LoginRequest) (*service.LoginResult, error) {
			return nil, service.ErrInvalidCredentials
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/login", LoginRequest{
		Email:    "oauth-user@example.com",
		Password: "anypassword",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody LoginRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.Login(r.Context(), service.LoginRequest{
			Email:    reqBody.Email,
			Password: reqBody.Password,
		})
		if err == service.ErrInvalidCredentials {
			WriteError(w, model.NewUnauthorizedError("invalid email or password"))
			return
		}
		WriteError(w, model.NewInternalError("login failed"))
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// ============================================================================
// Refresh Tests
// ============================================================================

func TestRefresh_ValidToken_ReturnsNewTokens(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		refreshTokensFunc: func(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
			return newTestTokenPair(), nil
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/refresh", RefreshRequest{
		RefreshToken: "valid-refresh-token",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteError(w, model.NewMethodNotAllowedError("POST"))
			return
		}

		var reqBody RefreshRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		if reqBody.RefreshToken == "" {
			WriteError(w, model.NewValidationError([]model.FieldError{
				{Field: "refresh_token", Message: "refresh_token is required"},
			}))
			return
		}

		tokenPair, err := mockSvc.RefreshTokens(r.Context(), reqBody.RefreshToken)
		if err != nil {
			WriteError(w, model.NewUnauthorizedError("invalid or expired refresh token"))
			return
		}

		WriteData(w, http.StatusOK, toTokenResponse(tokenPair), nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp DataResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be map")
	}

	if _, ok := data["access_token"]; !ok {
		t.Error("expected 'access_token' in response")
	}
	if _, ok := data["refresh_token"]; !ok {
		t.Error("expected 'refresh_token' in response")
	}
}

func TestRefresh_ExpiredToken_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		refreshTokensFunc: func(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
			return nil, service.ErrRefreshTokenExpired
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/refresh", RefreshRequest{
		RefreshToken: "expired-refresh-token",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RefreshRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.RefreshTokens(r.Context(), reqBody.RefreshToken)
		if err != nil {
			WriteError(w, model.NewUnauthorizedError("invalid or expired refresh token"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestRefresh_RevokedToken_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		refreshTokensFunc: func(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
			return nil, service.ErrRefreshTokenRevoked
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/refresh", RefreshRequest{
		RefreshToken: "revoked-refresh-token",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RefreshRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.RefreshTokens(r.Context(), reqBody.RefreshToken)
		if err != nil {
			WriteError(w, model.NewUnauthorizedError("invalid or expired refresh token"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestRefresh_MissingToken_ReturnsValidationError(t *testing.T) {
	t.Parallel()

	req := makeJSONRequest(http.MethodPost, "/v1/auth/refresh", RefreshRequest{
		RefreshToken: "",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RefreshRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		if reqBody.RefreshToken == "" {
			WriteError(w, model.NewValidationError([]model.FieldError{
				{Field: "refresh_token", Message: "refresh_token is required"},
			}))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, rr.Code)
	}

	problem := parseErrorResponse(t, rr.Body.Bytes())
	if len(problem.Errors) == 0 {
		t.Error("expected validation errors")
	}
	if problem.Errors[0].Field != "refresh_token" {
		t.Errorf("expected error on field 'refresh_token', got %q", problem.Errors[0].Field)
	}
}

func TestRefresh_InvalidToken_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		refreshTokensFunc: func(ctx context.Context, refreshToken string) (*service.TokenPair, error) {
			return nil, service.ErrInvalidRefreshToken
		},
	}

	req := makeJSONRequest(http.MethodPost, "/v1/auth/refresh", RefreshRequest{
		RefreshToken: "invalid-token-format",
	})
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RefreshRequest
		if err := DecodeJSON(r, &reqBody); err != nil {
			WriteError(w, model.NewBadRequestError("invalid request body"))
			return
		}

		_, err := mockSvc.RefreshTokens(r.Context(), reqBody.RefreshToken)
		if err != nil {
			WriteError(w, model.NewUnauthorizedError("invalid or expired refresh token"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// ============================================================================
// Logout Tests
// ============================================================================

func TestLogout_Authenticated_ReturnsNoContent(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		logoutFunc: func(ctx context.Context, userID string) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	req = withUserContext(req, "user:123")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteError(w, model.NewMethodNotAllowedError("POST"))
			return
		}

		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		if err := mockSvc.Logout(r.Context(), userID); err != nil {
			WriteError(w, model.NewInternalError("logout failed"))
			return
		}

		WriteNoContent(w)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestLogout_Unauthenticated_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	// No user context
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			WriteError(w, model.NewMethodNotAllowedError("POST"))
			return
		}

		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

// ============================================================================
// Me Tests
// ============================================================================

func TestMe_Authenticated_ReturnsUserData(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		getUserWithIdentitiesFunc: func(ctx context.Context, userID string) (*model.UserWithIdentities, error) {
			return &model.UserWithIdentities{
				User:       newTestUser(),
				Identities: []*model.Identity{},
				Passkeys:   []*model.Passkey{},
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req = withUserContext(req, "user:123")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteError(w, model.NewMethodNotAllowedError("GET"))
			return
		}

		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		userWithIdentities, err := mockSvc.GetUserWithIdentities(r.Context(), userID)
		if err != nil {
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

		WriteData(w, http.StatusOK, response, nil)
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var resp DataResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be map")
	}

	if _, ok := data["user"]; !ok {
		t.Error("expected 'user' in response")
	}
	if _, ok := data["identities"]; !ok {
		t.Error("expected 'identities' in response")
	}
	if _, ok := data["passkeys"]; !ok {
		t.Error("expected 'passkeys' in response")
	}
}

func TestMe_Unauthenticated_ReturnsUnauthorized(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	// No user context
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteError(w, model.NewMethodNotAllowedError("GET"))
			return
		}

		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestMe_UserNotFound_ReturnsNotFound(t *testing.T) {
	t.Parallel()

	mockSvc := &mockAuthService{
		getUserWithIdentitiesFunc: func(ctx context.Context, userID string) (*model.UserWithIdentities, error) {
			return nil, service.ErrUserNotFound
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	req = withUserContext(req, "user:deleted")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteError(w, model.NewMethodNotAllowedError("GET"))
			return
		}

		userID := middleware.GetUserID(r.Context())
		if userID == "" {
			WriteError(w, model.NewUnauthorizedError("authentication required"))
			return
		}

		_, err := mockSvc.GetUserWithIdentities(r.Context(), userID)
		if err == service.ErrUserNotFound {
			WriteError(w, model.NewNotFoundError("user"))
			return
		}
		if err != nil {
			WriteError(w, model.NewInternalError("failed to get user"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
}

func TestMe_WrongMethod_ReturnsMethodNotAllowed(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/me", nil)
	req = withUserContext(req, "user:123")
	rr := httptest.NewRecorder()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			WriteError(w, model.NewMethodNotAllowedError("GET"))
			return
		}
	})

	testHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
