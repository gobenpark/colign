package oauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gobenpark/colign/internal/auth"
)

func TestAuthorizeHandlerRefreshesOAuthSessionFromRefreshCookie(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	refreshedAccessToken, err := jwtManager.GenerateAccessToken(1, "user@example.com", "User", 42)
	require.NoError(t, err)

	handler := NewAuthorizeHandler(nil, jwtManager, "http://localhost:8080")
	stub := &stubOAuthSessionService{
		refreshFn: func(_ context.Context, refreshToken string) (*auth.TokenPair, error) {
			assert.Equal(t, "refresh-token", refreshToken)
			return &auth.TokenPair{
				AccessToken:  refreshedAccessToken,
				RefreshToken: "rotated-refresh-token",
				ExpiresAt:    time.Now().Add(auth.AccessTokenDuration).Unix(),
			}, nil
		},
	}
	handler.newAuthService = func() oauthSessionService { return stub }

	req := httptest.NewRequest(http.MethodGet, authorizeURL(), nil)
	req.AddCookie(&http.Cookie{Name: oauthAccessCookieName, Value: expiredToken(t, "test-secret", 1, "user@example.com", "User", 42)})
	req.AddCookie(&http.Cookie{Name: oauthRefreshCookieName, Value: "refresh-token"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() { _ = res.Body.Close() }()

	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Contains(t, rec.Body.String(), "Authorize Access")
	assert.Contains(t, rec.Body.String(), "user@example.com")
	assert.Equal(t, 1, stub.refreshCalls)

	accessCookie := findCookie(t, res.Cookies(), oauthAccessCookieName)
	refreshCookie := findCookie(t, res.Cookies(), oauthRefreshCookieName)
	assert.Equal(t, refreshedAccessToken, accessCookie.Value)
	assert.Equal(t, "rotated-refresh-token", refreshCookie.Value)
	assert.Equal(t, int(auth.AccessTokenDuration/time.Second), accessCookie.MaxAge)
	assert.Equal(t, int(auth.RefreshTokenDuration/time.Second), refreshCookie.MaxAge)
}

func TestAuthorizeHandlerUsesValidAccessCookieWithoutRefresh(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret")
	accessToken, err := jwtManager.GenerateAccessToken(1, "user@example.com", "User", 42)
	require.NoError(t, err)

	handler := NewAuthorizeHandler(nil, jwtManager, "http://localhost:8080")
	stub := &stubOAuthSessionService{
		refreshFn: func(_ context.Context, _ string) (*auth.TokenPair, error) {
			return nil, errors.New("refresh should not be called")
		},
	}
	handler.newAuthService = func() oauthSessionService { return stub }

	req := httptest.NewRequest(http.MethodGet, authorizeURL(), nil)
	req.AddCookie(&http.Cookie{Name: oauthAccessCookieName, Value: accessToken})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer func() { _ = res.Body.Close() }()

	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Contains(t, rec.Body.String(), "user@example.com")
	assert.Equal(t, 0, stub.refreshCalls)
	assert.Empty(t, res.Cookies())
}

func authorizeURL() string {
	values := url.Values{}
	values.Set("client_id", "client-123")
	values.Set("redirect_uri", "http://localhost/callback")
	values.Set("state", "state-123")
	values.Set("code_challenge", "challenge")
	values.Set("code_challenge_method", "S256")
	return "/oauth/authorize?" + values.Encode()
}

func expiredToken(t *testing.T, secret string, userID int64, email, name string, orgID int64) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, auth.Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		OrgID:  orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Minute)),
			Issuer:    "colign",
		},
	})

	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return signed
}

func findCookie(t *testing.T, cookies []*http.Cookie, name string) *http.Cookie {
	t.Helper()

	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}

type stubOAuthSessionService struct {
	refreshCalls int
	refreshFn    func(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
	loginFn      func(ctx context.Context, req auth.LoginRequest) (*auth.TokenPair, error)
}

func (s *stubOAuthSessionService) Login(ctx context.Context, req auth.LoginRequest) (*auth.TokenPair, error) {
	if s.loginFn == nil {
		return nil, errors.New("login not implemented")
	}
	return s.loginFn(ctx, req)
}

func (s *stubOAuthSessionService) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	s.refreshCalls++
	if s.refreshFn == nil {
		return nil, errors.New("refresh not implemented")
	}
	return s.refreshFn(ctx, refreshToken)
}

func TestAuthorizeHandler_MissingClientID(t *testing.T) {
	handler := &AuthorizeHandler{}

	values := url.Values{}
	values.Set("redirect_uri", "http://localhost/callback")
	values.Set("code_challenge", "challenge")
	values.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+values.Encode(), nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing required parameters")
}

func TestAuthorizeHandler_MissingRedirectURI(t *testing.T) {
	handler := &AuthorizeHandler{}

	values := url.Values{}
	values.Set("client_id", "client-123")
	values.Set("code_challenge", "challenge")
	values.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+values.Encode(), nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing required parameters")
}

func TestAuthorizeHandler_MissingCodeChallenge(t *testing.T) {
	handler := &AuthorizeHandler{}

	values := url.Values{}
	values.Set("client_id", "client-123")
	values.Set("redirect_uri", "http://localhost/callback")
	values.Set("code_challenge_method", "S256")

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+values.Encode(), nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "missing required parameters")
}

func TestAuthorizeHandler_WrongCodeChallengeMethod(t *testing.T) {
	handler := &AuthorizeHandler{}

	values := url.Values{}
	values.Set("client_id", "client-123")
	values.Set("redirect_uri", "http://localhost/callback")
	values.Set("code_challenge", "challenge")
	values.Set("code_challenge_method", "plain")

	req := httptest.NewRequest(http.MethodGet, "/oauth/authorize?"+values.Encode(), nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "only S256 code_challenge_method is supported")
}
