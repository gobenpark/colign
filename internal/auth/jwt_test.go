package auth_test

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/gobenpark/colign/internal/auth"
	"github.com/gobenpark/colign/internal/auth/mock_auth"
)

//go:generate mockgen -destination=mock_auth/mock_api_token_validator.go -package=mock_auth github.com/gobenpark/colign/internal/auth APITokenValidator

// ---------------------------------------------------------------------------
// Helper: create an expired token manually using jwt-go
// ---------------------------------------------------------------------------

func createExpiredToken(t *testing.T, secret string, userID int64, email, name string, orgID int64) string {
	t.Helper()
	past := time.Now().Add(-1 * time.Hour)
	claims := auth.Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		OrgID:  orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(past),
			IssuedAt:  jwt.NewNumericDate(past.Add(-15 * time.Minute)),
			Issuer:    "colign",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)
	return tokenStr
}

// ---------------------------------------------------------------------------
// 1. NewJWTManager
// ---------------------------------------------------------------------------

func TestJWTManager_NewJWTManager(t *testing.T) {
	t.Run("creates manager with given secret", func(t *testing.T) {
		m := auth.NewJWTManager("test-secret")
		require.NotNil(t, m)
	})

	t.Run("creates manager with empty secret", func(t *testing.T) {
		m := auth.NewJWTManager("")
		require.NotNil(t, m)
	})
}

// ---------------------------------------------------------------------------
// 2. GenerateAccessToken
// ---------------------------------------------------------------------------

func TestJWTManager_GenerateAccessToken(t *testing.T) {
	secret := "my-super-secret-key"
	m := auth.NewJWTManager(secret)

	t.Run("returns a non-empty token string", func(t *testing.T) {
		token, err := m.GenerateAccessToken(1, "user@example.com", "Alice", 10)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("token contains correct claims", func(t *testing.T) {
		tokenStr, err := m.GenerateAccessToken(42, "bob@example.com", "Bob", 99)
		require.NoError(t, err)

		claims, err := m.ValidateAccessToken(tokenStr)
		require.NoError(t, err)

		assert.Equal(t, int64(42), claims.UserID)
		assert.Equal(t, "bob@example.com", claims.Email)
		assert.Equal(t, "Bob", claims.Name)
		assert.Equal(t, int64(99), claims.OrgID)
		assert.Equal(t, "colign", claims.Issuer)
	})

	t.Run("token expiry is approximately AccessTokenDuration from now", func(t *testing.T) {
		before := time.Now()
		tokenStr, err := m.GenerateAccessToken(1, "a@b.com", "A", 1)
		require.NoError(t, err)
		after := time.Now()

		claims, err := m.ValidateAccessToken(tokenStr)
		require.NoError(t, err)

		expectedEarliest := before.Add(auth.AccessTokenDuration)
		expectedLatest := after.Add(auth.AccessTokenDuration)

		exp := claims.ExpiresAt.Time
		assert.False(t, exp.Before(expectedEarliest.Add(-time.Second)), "expiry too early")
		assert.False(t, exp.After(expectedLatest.Add(time.Second)), "expiry too late")
	})

	t.Run("issued at is approximately now", func(t *testing.T) {
		before := time.Now()
		tokenStr, err := m.GenerateAccessToken(1, "a@b.com", "A", 1)
		require.NoError(t, err)

		claims, err := m.ValidateAccessToken(tokenStr)
		require.NoError(t, err)

		iat := claims.IssuedAt.Time
		assert.WithinDuration(t, before, iat, 2*time.Second)
	})
}

// ---------------------------------------------------------------------------
// 3. ValidateAccessToken
// ---------------------------------------------------------------------------

func TestJWTManager_ValidateAccessToken(t *testing.T) {
	secret := "validate-test-secret"
	m := auth.NewJWTManager(secret)

	t.Run("validates a valid token", func(t *testing.T) {
		tokenStr, err := m.GenerateAccessToken(7, "valid@test.com", "Valid", 3)
		require.NoError(t, err)

		claims, err := m.ValidateAccessToken(tokenStr)
		require.NoError(t, err)
		assert.Equal(t, int64(7), claims.UserID)
		assert.Equal(t, "valid@test.com", claims.Email)
		assert.Equal(t, "Valid", claims.Name)
		assert.Equal(t, int64(3), claims.OrgID)
	})

	t.Run("rejects expired token", func(t *testing.T) {
		tokenStr := createExpiredToken(t, secret, 1, "expired@test.com", "Exp", 1)

		claims, err := m.ValidateAccessToken(tokenStr)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.ErrorIs(t, err, jwt.ErrTokenExpired)
	})

	t.Run("rejects token signed with wrong secret", func(t *testing.T) {
		other := auth.NewJWTManager("different-secret")
		tokenStr, err := other.GenerateAccessToken(1, "wrong@test.com", "Wrong", 1)
		require.NoError(t, err)

		claims, err := m.ValidateAccessToken(tokenStr)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("rejects token with wrong signing method", func(t *testing.T) {
		claims := auth.Claims{
			UserID: 1,
			Email:  "a@b.com",
			Name:   "A",
			OrgID:  1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "colign",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenStr, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		result, err := m.ValidateAccessToken(tokenStr)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("rejects completely malformed token", func(t *testing.T) {
		claims, err := m.ValidateAccessToken("this-is-not-a-jwt")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("rejects empty token", func(t *testing.T) {
		claims, err := m.ValidateAccessToken("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

// ---------------------------------------------------------------------------
// 4. ExtractClaims
// ---------------------------------------------------------------------------

func TestExtractClaims(t *testing.T) {
	secret := "extract-claims-secret"
	m := auth.NewJWTManager(secret)
	validToken, err := m.GenerateAccessToken(5, "extract@test.com", "Extract", 20)
	require.NoError(t, err)

	t.Run("valid Bearer header", func(t *testing.T) {
		claims, err := auth.ExtractClaims(m, "Bearer "+validToken)
		require.NoError(t, err)
		assert.Equal(t, int64(5), claims.UserID)
		assert.Equal(t, "extract@test.com", claims.Email)
		assert.Equal(t, "Extract", claims.Name)
		assert.Equal(t, int64(20), claims.OrgID)
	})

	t.Run("missing Bearer prefix", func(t *testing.T) {
		claims, err := auth.ExtractClaims(m, "Token "+validToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "invalid authorization header")
	})

	t.Run("empty header", func(t *testing.T) {
		claims, err := auth.ExtractClaims(m, "")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "invalid authorization header")
	})

	t.Run("malformed header with no space", func(t *testing.T) {
		claims, err := auth.ExtractClaims(m, "BearerTOKEN")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "invalid authorization header")
	})

	t.Run("header with only Bearer and no token", func(t *testing.T) {
		// "Bearer " split by space gives ["Bearer", ""] — part[1] is empty string
		claims, err := auth.ExtractClaims(m, "Bearer ")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

// ---------------------------------------------------------------------------
// 5. ResolveFromHeader
// ---------------------------------------------------------------------------

func TestResolveFromHeader(t *testing.T) {
	secret := "resolve-secret"
	m := auth.NewJWTManager(secret)
	ctx := context.Background()

	t.Run("JWT token path - valid token", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		tokenStr, err := m.GenerateAccessToken(11, "jwt@test.com", "JWT", 55)
		require.NoError(t, err)

		mockValidator := mock_auth.NewMockAPITokenValidator(ctrl)
		// No EXPECT — the mock should NOT be called for a non-col_ token.

		claims, err := auth.ResolveFromHeader(m, mockValidator, ctx, "Bearer "+tokenStr)
		require.NoError(t, err)
		assert.Equal(t, int64(11), claims.UserID)
		assert.Equal(t, "jwt@test.com", claims.Email)
		assert.Equal(t, "JWT", claims.Name)
		assert.Equal(t, int64(55), claims.OrgID)
	})

	t.Run("API token (col_*) path - valid", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockValidator := mock_auth.NewMockAPITokenValidator(ctrl)
		mockValidator.EXPECT().
			ValidateTokenForAuth(gomock.Any(), "col_abc123").
			Return(int64(100), "api@test.com", int64(200), nil)

		claims, err := auth.ResolveFromHeader(m, mockValidator, ctx, "Bearer col_abc123")
		require.NoError(t, err)
		assert.Equal(t, int64(100), claims.UserID)
		assert.Equal(t, "api@test.com", claims.Email)
		assert.Equal(t, int64(200), claims.OrgID)
	})

	t.Run("API token (col_*) path - validator returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		mockValidator := mock_auth.NewMockAPITokenValidator(ctrl)
		mockValidator.EXPECT().
			ValidateTokenForAuth(gomock.Any(), "col_bad").
			Return(int64(0), "", int64(0), errors.New("invalid api token"))

		claims, err := auth.ResolveFromHeader(m, mockValidator, ctx, "Bearer col_bad")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "invalid api token")
	})

	t.Run("nil apiTokenValidator with col_ token", func(t *testing.T) {
		claims, err := auth.ResolveFromHeader(m, nil, ctx, "Bearer col_abc123")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "API token authentication not available")
	})

	t.Run("invalid header - missing Bearer prefix", func(t *testing.T) {
		claims, err := auth.ResolveFromHeader(m, nil, ctx, "Token something")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "invalid authorization header")
	})

	t.Run("invalid header - empty string", func(t *testing.T) {
		claims, err := auth.ResolveFromHeader(m, nil, ctx, "")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "invalid authorization header")
	})

	t.Run("invalid header - no space", func(t *testing.T) {
		claims, err := auth.ResolveFromHeader(m, nil, ctx, "BearerXYZ")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.EqualError(t, err, "invalid authorization header")
	})

	t.Run("JWT token path - invalid JWT after Bearer", func(t *testing.T) {
		claims, err := auth.ResolveFromHeader(m, nil, ctx, "Bearer not-a-valid-jwt")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

// ---------------------------------------------------------------------------
// 6. GenerateRefreshToken
// ---------------------------------------------------------------------------

func TestGenerateRefreshToken(t *testing.T) {
	t.Run("returns hex string of correct length", func(t *testing.T) {
		token, err := auth.GenerateRefreshToken()
		require.NoError(t, err)
		// 32 bytes -> 64 hex characters
		assert.Len(t, token, 64)
	})

	t.Run("returns valid hex characters only", func(t *testing.T) {
		token, err := auth.GenerateRefreshToken()
		require.NoError(t, err)

		decoded, err := hex.DecodeString(token)
		require.NoError(t, err)
		assert.Len(t, decoded, 32)
	})

	t.Run("uniqueness - successive calls produce different tokens", func(t *testing.T) {
		token1, err := auth.GenerateRefreshToken()
		require.NoError(t, err)

		token2, err := auth.GenerateRefreshToken()
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)
	})

	t.Run("uniqueness - many tokens are all distinct", func(t *testing.T) {
		seen := make(map[string]bool)
		for i := 0; i < 100; i++ {
			token, err := auth.GenerateRefreshToken()
			require.NoError(t, err)
			assert.False(t, seen[token], "duplicate refresh token detected")
			seen[token] = true
		}
	})
}

// ---------------------------------------------------------------------------
// 7. GenerateTokenPair
// ---------------------------------------------------------------------------

func TestGenerateTokenPair(t *testing.T) {
	secret := "token-pair-secret"
	m := auth.NewJWTManager(secret)

	t.Run("returns both tokens with correct fields", func(t *testing.T) {
		before := time.Now()
		pair, err := m.GenerateTokenPair(33, "pair@test.com", "Pair", 77)
		require.NoError(t, err)
		after := time.Now()

		require.NotNil(t, pair)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)

		// Refresh token should be 64 hex chars
		assert.Len(t, pair.RefreshToken, 64)

		// ExpiresAt should be approximately AccessTokenDuration from now
		expiresAt := time.Unix(pair.ExpiresAt, 0)
		expectedEarliest := before.Add(auth.AccessTokenDuration)
		expectedLatest := after.Add(auth.AccessTokenDuration)
		assert.False(t, expiresAt.Before(expectedEarliest.Add(-time.Second)), "expires_at too early")
		assert.False(t, expiresAt.After(expectedLatest.Add(time.Second)), "expires_at too late")
	})

	t.Run("access token contains correct claims", func(t *testing.T) {
		pair, err := m.GenerateTokenPair(33, "pair@test.com", "Pair", 77)
		require.NoError(t, err)

		claims, err := m.ValidateAccessToken(pair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, int64(33), claims.UserID)
		assert.Equal(t, "pair@test.com", claims.Email)
		assert.Equal(t, "Pair", claims.Name)
		assert.Equal(t, int64(77), claims.OrgID)
	})

	t.Run("successive calls produce different token pairs", func(t *testing.T) {
		pair1, err := m.GenerateTokenPair(1, "a@b.com", "A", 1)
		require.NoError(t, err)

		pair2, err := m.GenerateTokenPair(1, "a@b.com", "A", 1)
		require.NoError(t, err)

		// Refresh tokens must differ (crypto random)
		assert.NotEqual(t, pair1.RefreshToken, pair2.RefreshToken)
	})
}
