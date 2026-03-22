package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gobenpark/colign/internal/auth"
)

const testSecret = "test-secret-key-for-testing"

// dummyHandler records whether it was called and captures context values
// set by the middleware.
type dummyHandler struct {
	called bool
	userID int64
	email  string
	orgID  int64
}

func (h *dummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	h.userID = GetUserID(r.Context())
	h.email, _ = r.Context().Value(ContextKeyEmail).(string)
	h.orgID = GetOrgID(r.Context())
	w.WriteHeader(http.StatusOK)
}

// generateExpiredToken creates a JWT token that is already expired,
// signed with the given secret using jwt-go directly.
func generateExpiredToken(secret string) string {
	claims := auth.Claims{
		UserID: 1,
		Email:  "expired@example.com",
		Name:   "Expired User",
		OrgID:  10,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "colign",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

// ---------------------------------------------------------------------------
// JWTAuth middleware tests
// ---------------------------------------------------------------------------

func TestJWTAuth_NoAuthorizationHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockValidator := NewMockTokenValidator(ctrl)

	handler := &dummyHandler{}
	middleware := JWTAuth(mockValidator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handler.called, "next handler should not be called")

	var body map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err, "response body should be valid JSON")
	assert.Equal(t, "authorization header required", body["error"])
}

func TestJWTAuth_InvalidHeaderFormat_NoBearerPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockValidator := NewMockTokenValidator(ctrl)

	tests := []struct {
		name   string
		header string
	}{
		{"plain token without prefix", "some-token-value"},
		{"Basic auth instead of Bearer", "Basic dXNlcjpwYXNz"},
		{"bearer lowercase", "bearer some-token"},
		{"empty Bearer value", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &dummyHandler{}
			middleware := JWTAuth(mockValidator)(handler)

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Authorization", tt.header)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.False(t, handler.called, "next handler should not be called")

			var body map[string]string
			err := json.Unmarshal(rec.Body.Bytes(), &body)
			require.NoError(t, err, "response body should be valid JSON")
			assert.Equal(t, "invalid authorization header format", body["error"])
		})
	}
}

func TestJWTAuth_InvalidMalformedToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockValidator := NewMockTokenValidator(ctrl)

	tests := []struct {
		name  string
		token string
	}{
		{"random garbage", "not-a-jwt-token"},
		{"partial JWT", "eyJhbGciOiJIUzI1NiJ9.garbage"},
		{"empty string token", "some-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator.EXPECT().
				ValidateAccessToken(tt.token).
				Return(nil, errors.New("invalid token")).
				Times(1)

			handler := &dummyHandler{}
			middleware := JWTAuth(mockValidator)(handler)

			req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			rec := httptest.NewRecorder()

			middleware.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.False(t, handler.called, "next handler should not be called")

			var body map[string]string
			err := json.Unmarshal(rec.Body.Bytes(), &body)
			require.NoError(t, err, "response body should be valid JSON")
			assert.Equal(t, "invalid or expired token", body["error"])
		})
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	jwtManager := auth.NewJWTManager(testSecret)
	handler := &dummyHandler{}
	middleware := JWTAuth(jwtManager)(handler)

	expiredToken := generateExpiredToken(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, handler.called, "next handler should not be called")

	var body map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	require.NoError(t, err, "response body should be valid JSON")
	assert.Equal(t, "invalid or expired token", body["error"])
}

func TestJWTAuth_ValidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager(testSecret)
	handler := &dummyHandler{}
	middleware := JWTAuth(jwtManager)(handler)

	var (
		expectedUserID int64 = 42
		expectedEmail        = "user@example.com"
		expectedName         = "Test User"
		expectedOrgID  int64 = 7
	)

	token, err := jwtManager.GenerateAccessToken(expectedUserID, expectedEmail, expectedName, expectedOrgID)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handler.called, "next handler should be called")
	assert.Equal(t, expectedUserID, handler.userID)
	assert.Equal(t, expectedEmail, handler.email)
	assert.Equal(t, expectedOrgID, handler.orgID)
}

func TestJWTAuth_ValidToken_ContextValuesAreCorrect(t *testing.T) {
	jwtManager := auth.NewJWTManager(testSecret)

	var (
		capturedUserID int64
		capturedEmail  string
		capturedOrgID  int64
	)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, _ = r.Context().Value(ContextKeyUserID).(int64)
		capturedEmail, _ = r.Context().Value(ContextKeyEmail).(string)
		capturedOrgID, _ = r.Context().Value(ContextKeyOrgID).(int64)
		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuth(jwtManager)(inner)

	token, err := jwtManager.GenerateAccessToken(99, "ctx@test.com", "Ctx User", 55)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, int64(99), capturedUserID)
	assert.Equal(t, "ctx@test.com", capturedEmail)
	assert.Equal(t, int64(55), capturedOrgID)
}

func TestJWTAuth_ValidToken_MockValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockValidator := NewMockTokenValidator(ctrl)

	expectedClaims := &auth.Claims{
		UserID: 77,
		Email:  "mock@test.com",
		Name:   "Mock User",
		OrgID:  33,
	}

	mockValidator.EXPECT().
		ValidateAccessToken("mock-token-value").
		Return(expectedClaims, nil).
		Times(1)

	handler := &dummyHandler{}
	middleware := JWTAuth(mockValidator)(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer mock-token-value")
	rec := httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handler.called, "next handler should be called")
	assert.Equal(t, int64(77), handler.userID)
	assert.Equal(t, "mock@test.com", handler.email)
	assert.Equal(t, int64(33), handler.orgID)
}

// ---------------------------------------------------------------------------
// GetUserID tests
// ---------------------------------------------------------------------------

func TestGetUserID_ReturnsCorrectValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyUserID, int64(123))
	assert.Equal(t, int64(123), GetUserID(ctx))
}

func TestGetUserID_ReturnsZeroWhenNotSet(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, int64(0), GetUserID(ctx))
}

func TestGetUserID_ReturnsZeroWhenWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyUserID, "not-an-int64")
	assert.Equal(t, int64(0), GetUserID(ctx))
}

func TestGetUserID_ReturnsZeroWhenWrongIntType(t *testing.T) {
	// int (not int64) — type assertion to int64 should fail
	ctx := context.WithValue(context.Background(), ContextKeyUserID, int(42))
	assert.Equal(t, int64(0), GetUserID(ctx))
}

// ---------------------------------------------------------------------------
// GetOrgID tests
// ---------------------------------------------------------------------------

func TestGetOrgID_ReturnsCorrectValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), ContextKeyOrgID, int64(456))
	assert.Equal(t, int64(456), GetOrgID(ctx))
}

func TestGetOrgID_ReturnsZeroWhenNotSet(t *testing.T) {
	ctx := context.Background()
	assert.Equal(t, int64(0), GetOrgID(ctx))
}
