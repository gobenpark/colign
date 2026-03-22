package apitoken

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newTestDB creates a bun.DB backed by sqlmock for integration-style tests.
func newTestDB(t *testing.T) (*bun.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	bunDB := bun.NewDB(sqlDB, pgdialect.New())
	t.Cleanup(func() { _ = sqlDB.Close() })
	return bunDB, mock
}

// ---------------------------------------------------------------------------
// generateToken
// ---------------------------------------------------------------------------

func TestGenerateToken_HasColPrefix(t *testing.T) {
	tok, err := generateToken()
	require.NoError(t, err, "generateToken should not return an error")
	assert.True(t, len(tok) >= 4 && tok[:4] == "col_", "token should start with \"col_\" prefix, got %q", tok)
}

func TestGenerateToken_Length(t *testing.T) {
	tok, err := generateToken()
	require.NoError(t, err)
	// 4 byte prefix "col_" + 24 random bytes encoded as 48 hex chars = 52
	assert.Len(t, tok, 52, "token should be exactly 52 characters (4 prefix + 48 hex)")
}

func TestGenerateToken_HexSuffix(t *testing.T) {
	tok, err := generateToken()
	require.NoError(t, err)

	suffix := tok[4:] // strip "col_"
	_, decodeErr := hex.DecodeString(suffix)
	assert.NoError(t, decodeErr, "suffix after \"col_\" should be valid hex, got %q", suffix)
}

func TestGenerateToken_Uniqueness(t *testing.T) {
	const iterations = 100
	seen := make(map[string]struct{}, iterations)

	for i := 0; i < iterations; i++ {
		tok, err := generateToken()
		require.NoError(t, err, "iteration %d", i)
		assert.NotContains(t, seen, tok, "duplicate token on iteration %d: %q", i, tok)
		seen[tok] = struct{}{}
	}
}

// ---------------------------------------------------------------------------
// hashToken
// ---------------------------------------------------------------------------

func TestHashToken_Deterministic(t *testing.T) {
	input := "col_deadbeef1234567890abcdef1234567890abcdef12345678"
	h1 := hashToken(input)
	h2 := hashToken(input)
	assert.Equal(t, h1, h2, "same input must produce the same hash")
}

func TestHashToken_Length(t *testing.T) {
	h := hashToken("anything")
	// SHA-256 produces 32 bytes = 64 hex characters
	assert.Len(t, h, 64, "hash should be 64 hex characters (SHA-256)")
}

func TestHashToken_HexOutput(t *testing.T) {
	h := hashToken("test-token")
	_, err := hex.DecodeString(h)
	assert.NoError(t, err, "hash should be valid hex, got %q", h)
}

func TestHashToken_DifferentInputsDifferentOutputs(t *testing.T) {
	h1 := hashToken("input-a")
	h2 := hashToken("input-b")
	assert.NotEqual(t, h1, h2, "different inputs must produce different hashes")
}

func TestHashToken_EmptyInput(t *testing.T) {
	h := hashToken("")
	assert.Len(t, h, 64, "hash of empty string should still be 64 hex characters")
	// SHA-256 of "" is a well-known constant
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", h)
}

// ---------------------------------------------------------------------------
// CreateOAuth — pure validation (empty / whitespace clientID)
// These error paths return before any DB call, so db can be nil.
// ---------------------------------------------------------------------------

func TestCreateOAuth_EmptyClientID_ReturnsError(t *testing.T) {
	svc := &Service{db: nil}

	tok, raw, err := svc.CreateOAuth(context.Background(), 1, 1, "", "test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oauth client id is required")
	assert.Nil(t, tok)
	assert.Empty(t, raw)
}

func TestCreateOAuth_WhitespaceOnlyClientID_ReturnsError(t *testing.T) {
	svc := &Service{db: nil}

	tok, raw, err := svc.CreateOAuth(context.Background(), 1, 1, "   \t\n  ", "test")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "oauth client id is required")
	assert.Nil(t, tok)
	assert.Empty(t, raw)
}

// ---------------------------------------------------------------------------
// DB-dependent: Create
// ---------------------------------------------------------------------------

func TestCreate_InsertsTokenAndReturnsRaw(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	// bun uses RETURNING clause, so the INSERT is issued as a Query, not Exec.
	mock.ExpectQuery(`INSERT INTO "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_used_at", "created_at"}).
			AddRow(1, nil, time.Now()))

	tok, raw, err := svc.Create(context.Background(), 10, 20, "my-token")

	require.NoError(t, err)
	require.NotNil(t, tok)
	assert.NotEmpty(t, raw)
	assert.Equal(t, int64(10), tok.UserID)
	assert.Equal(t, int64(20), tok.OrgID)
	assert.Equal(t, "my-token", tok.Name)
	assert.Equal(t, "personal", tok.TokenType)
	assert.Empty(t, tok.OAuthClientID)
	assert.Len(t, raw, 52, "raw token should be 52 chars")
	assert.Equal(t, raw[:8], tok.Prefix, "prefix should be first 8 chars of raw token")
	assert.Equal(t, hashToken(raw), tok.TokenHash, "stored hash should match hash of raw token")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreate_DBError_ReturnsError(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.ExpectQuery(`INSERT INTO "api_tokens"`).
		WillReturnError(assert.AnError)

	tok, raw, err := svc.Create(context.Background(), 10, 20, "my-token")

	require.Error(t, err)
	assert.Nil(t, tok)
	assert.Empty(t, raw)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// DB-dependent: CreateOAuth with valid clientID
// ---------------------------------------------------------------------------

func TestCreateOAuth_ValidClientID_InsertsToken(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	// First DELETE: remove existing token for same client
	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	// Second DELETE: prune stale oauth tokens
	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	// INSERT new token (bun uses RETURNING, so this is a Query)
	mock.ExpectQuery(`INSERT INTO "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_used_at", "created_at"}).
			AddRow(1, nil, time.Now()))

	tok, raw, err := svc.CreateOAuth(context.Background(), 10, 20, "client-xyz", "oauth-token")

	require.NoError(t, err)
	require.NotNil(t, tok)
	assert.NotEmpty(t, raw)
	assert.Equal(t, "oauth", tok.TokenType)
	assert.Equal(t, "client-xyz", tok.OAuthClientID)
	assert.Equal(t, int64(10), tok.UserID)
	assert.Equal(t, int64(20), tok.OrgID)
	assert.Len(t, raw, 52)
	assert.Equal(t, raw[:8], tok.Prefix)
	assert.Equal(t, hashToken(raw), tok.TokenHash)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateOAuth_TrimsClientID(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`INSERT INTO "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_used_at", "created_at"}).
			AddRow(1, nil, time.Now()))

	tok, _, err := svc.CreateOAuth(context.Background(), 1, 1, "  padded-client  ", "name")

	require.NoError(t, err)
	assert.Equal(t, "padded-client", tok.OAuthClientID, "clientID should be trimmed")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// DB-dependent: List
// ---------------------------------------------------------------------------

func TestList_ReturnsTokens(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	rows := sqlmock.NewRows([]string{"id", "user_id", "org_id", "name", "token_type", "token_hash", "prefix"}).
		AddRow(1, 10, 20, "tok-1", "personal", "hash1", "col_abcd").
		AddRow(2, 10, 20, "tok-2", "personal", "hash2", "col_efgh")

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(rows)

	tokens, err := svc.List(context.Background(), 10, 20)

	require.NoError(t, err)
	assert.Len(t, tokens, 2)
	assert.Equal(t, "tok-1", tokens[0].Name)
	assert.Equal(t, "tok-2", tokens[1].Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_EmptyResult(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	rows := sqlmock.NewRows([]string{"id", "user_id", "org_id", "name", "token_type", "token_hash", "prefix"})
	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(rows)

	tokens, err := svc.List(context.Background(), 10, 20)

	require.NoError(t, err)
	assert.Empty(t, tokens)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestList_DBError(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnError(assert.AnError)

	tokens, err := svc.List(context.Background(), 10, 20)

	require.Error(t, err)
	assert.Nil(t, tokens)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// DB-dependent: Delete
// ---------------------------------------------------------------------------

func TestDelete_Success(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 10, 99)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDelete_DBError(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnError(assert.AnError)

	err := svc.Delete(context.Background(), 10, 99)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// DB-dependent: ValidateToken
// ---------------------------------------------------------------------------

func TestValidateToken_Found(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	rawToken := "col_aabbccddee112233445566778899aabbccddee11223344"

	rows := sqlmock.NewRows([]string{"id", "user_id", "org_id", "name", "token_type", "token_hash", "prefix"}).
		AddRow(42, 10, 20, "found-token", "personal", hashToken(rawToken), rawToken[:8])

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(rows)

	// Background goroutine will fire an UPDATE for last_used_at; allow it.
	mock.ExpectExec(`UPDATE "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tok, err := svc.ValidateToken(context.Background(), rawToken)

	require.NoError(t, err)
	require.NotNil(t, tok)
	assert.Equal(t, int64(42), tok.ID)
	assert.Equal(t, int64(10), tok.UserID)
	assert.Equal(t, int64(20), tok.OrgID)
}

func TestValidateToken_NotFound(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	rows := sqlmock.NewRows([]string{"id", "user_id", "org_id", "name", "token_type", "token_hash", "prefix"})
	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(rows)

	tok, err := svc.ValidateToken(context.Background(), "col_nonexistent000000000000000000000000000000000000")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API token")
	assert.Nil(t, tok)
}

// ---------------------------------------------------------------------------
// DB-dependent: ValidateTokenForAuth
// ---------------------------------------------------------------------------

func TestValidateTokenForAuth_Success(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	// The background goroutine UPDATE may fire before or after the user SELECT,
	// so we must allow out-of-order matching.
	mock.MatchExpectationsInOrder(false)

	rawToken := "col_aabbccddee112233445566778899aabbccddee11223344"

	tokenRows := sqlmock.NewRows([]string{"id", "user_id", "org_id", "name", "token_type", "token_hash", "prefix"}).
		AddRow(42, 10, 20, "auth-token", "personal", hashToken(rawToken), rawToken[:8])

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(tokenRows)

	// Background goroutine: UPDATE last_used_at
	mock.ExpectExec(`UPDATE "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	userRows := sqlmock.NewRows([]string{"id", "email", "name"}).
		AddRow(10, "user@example.com", "Test User")
	mock.ExpectQuery(`SELECT .+ FROM "users"`).
		WillReturnRows(userRows)

	userID, email, orgID, err := svc.ValidateTokenForAuth(context.Background(), rawToken)

	require.NoError(t, err)
	assert.Equal(t, int64(10), userID)
	assert.Equal(t, "user@example.com", email)
	assert.Equal(t, int64(20), orgID)

	// Give background goroutine time to fire.
	time.Sleep(50 * time.Millisecond)
}

func TestValidateTokenForAuth_InvalidToken(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	rows := sqlmock.NewRows([]string{"id", "user_id", "org_id", "name", "token_type", "token_hash", "prefix"})
	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(rows)

	userID, email, orgID, err := svc.ValidateTokenForAuth(context.Background(), "col_bad00000000000000000000000000000000000000000000")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API token")
	assert.Zero(t, userID)
	assert.Empty(t, email)
	assert.Zero(t, orgID)
}

func TestValidateTokenForAuth_UserNotFound(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	// Background goroutine UPDATE may race with user SELECT.
	mock.MatchExpectationsInOrder(false)

	rawToken := "col_aabbccddee112233445566778899aabbccddee11223344"

	tokenRows := sqlmock.NewRows([]string{"id", "user_id", "org_id", "name", "token_type", "token_hash", "prefix"}).
		AddRow(42, 10, 20, "auth-token", "personal", hashToken(rawToken), rawToken[:8])

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(tokenRows)

	// Background goroutine: UPDATE last_used_at
	mock.ExpectExec(`UPDATE "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// User lookup returns no rows
	userRows := sqlmock.NewRows([]string{"id", "email", "name"})
	mock.ExpectQuery(`SELECT .+ FROM "users"`).
		WillReturnRows(userRows)

	userID, email, orgID, err := svc.ValidateTokenForAuth(context.Background(), rawToken)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found for API token")
	assert.Zero(t, userID)
	assert.Empty(t, email)
	assert.Zero(t, orgID)

	// Give background goroutine time to fire.
	time.Sleep(50 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// NewService
// ---------------------------------------------------------------------------

func TestNewService(t *testing.T) {
	bunDB, _ := newTestDB(t)
	svc := NewService(bunDB)

	require.NotNil(t, svc)
	assert.Equal(t, bunDB, svc.db)
}
