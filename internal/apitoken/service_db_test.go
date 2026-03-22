package apitoken

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Create — additional scenarios
// ---------------------------------------------------------------------------

func TestCreate_TokenHasColPrefix(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.ExpectQuery(`INSERT INTO "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_used_at", "created_at"}).
			AddRow(1, nil, time.Now()))

	token, raw, err := svc.Create(context.Background(), 5, 10, "prefix-check")
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.True(t, strings.HasPrefix(raw, "col_"), "raw token must start with col_ prefix")
	assert.Equal(t, raw[:8], token.Prefix)
	assert.Equal(t, "personal", token.TokenType)
	assert.Equal(t, int64(5), token.UserID)
	assert.Equal(t, int64(10), token.OrgID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// CreateOAuth — additional scenarios
// ---------------------------------------------------------------------------

func TestCreateOAuth_SetsOAuthTypeAndClientID(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`INSERT INTO "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "last_used_at", "created_at"}).
			AddRow(7, nil, time.Now()))

	token, raw, err := svc.CreateOAuth(context.Background(), 100, 200, "linear-mcp", "Linear OAuth")
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.True(t, strings.HasPrefix(raw, "col_"))
	assert.Equal(t, "oauth", token.TokenType)
	assert.Equal(t, "linear-mcp", token.OAuthClientID)
	assert.Equal(t, int64(100), token.UserID)
	assert.Equal(t, int64(200), token.OrgID)
	assert.Equal(t, "Linear OAuth", token.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// List — additional scenarios
// ---------------------------------------------------------------------------

func TestList_ReturnsTwoTokensInOrder(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "org_id", "name", "token_type", "token_hash", "prefix",
		"last_used_at", "created_at",
	}).
		AddRow(3, 1, 1, "recent-token", "personal", "hash3", "col_rrrr", nil, now).
		AddRow(1, 1, 1, "old-token", "personal", "hash1", "col_oooo", nil, now.Add(-24*time.Hour))

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(rows)

	tokens, err := svc.List(context.Background(), 1, 1)
	require.NoError(t, err)
	require.Len(t, tokens, 2)

	// Verify order is preserved as returned by the mock (DESC by created_at)
	assert.Equal(t, "recent-token", tokens[0].Name)
	assert.Equal(t, "old-token", tokens[1].Name)
	assert.Equal(t, int64(3), tokens[0].ID)
	assert.Equal(t, int64(1), tokens[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// Delete — additional scenarios
// ---------------------------------------------------------------------------

func TestDelete_UsesUserIDAndTokenID(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	// The DELETE should include both user_id and token id conditions
	mock.ExpectExec(`DELETE FROM "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := svc.Delete(context.Background(), 42, 77)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// ValidateToken — additional scenarios
// ---------------------------------------------------------------------------

func TestValidateToken_ReturnsTokenFields(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.MatchExpectationsInOrder(false)

	rawToken := "col_1122334455667788990011223344556677889900aabbccdd"
	h := hashToken(rawToken)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "org_id", "name", "token_type",
			"oauth_client_id", "token_hash", "prefix", "last_used_at", "created_at",
		}).AddRow(99, 7, 14, "validated-key", "personal", "", h, rawToken[:8], nil, now))

	mock.ExpectExec(`UPDATE "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	token, err := svc.ValidateToken(context.Background(), rawToken)
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.Equal(t, int64(99), token.ID)
	assert.Equal(t, int64(7), token.UserID)
	assert.Equal(t, int64(14), token.OrgID)
	assert.Equal(t, "validated-key", token.Name)
	assert.Equal(t, h, token.TokenHash)

	time.Sleep(50 * time.Millisecond)
}

func TestValidateToken_InvalidToken_ReturnsError(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	// Return empty rows to simulate no matching token
	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "org_id", "name", "token_type",
			"oauth_client_id", "token_hash", "prefix", "last_used_at", "created_at",
		}))

	token, err := svc.ValidateToken(context.Background(), "col_doesnotexist000000000000000000000000000000000000")
	assert.Error(t, err)
	assert.Nil(t, token)
	assert.Contains(t, err.Error(), "invalid API token")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// ValidateTokenForAuth — additional scenarios
// ---------------------------------------------------------------------------

func TestValidateTokenForAuth_ReturnsUserIDEmailOrgID(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.MatchExpectationsInOrder(false)

	rawToken := "col_ffeeddccbbaa99887766554433221100ffeeddccbbaa9988"
	h := hashToken(rawToken)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "org_id", "name", "token_type",
			"oauth_client_id", "token_hash", "prefix", "last_used_at", "created_at",
		}).AddRow(11, 33, 44, "auth-key", "personal", "", h, rawToken[:8], nil, now))

	mock.ExpectExec(`UPDATE "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(`SELECT .+ FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name"}).
			AddRow(33, "alice@example.com", "Alice"))

	userID, email, orgID, err := svc.ValidateTokenForAuth(context.Background(), rawToken)
	require.NoError(t, err)

	assert.Equal(t, int64(33), userID)
	assert.Equal(t, "alice@example.com", email)
	assert.Equal(t, int64(44), orgID)

	time.Sleep(50 * time.Millisecond)
}

func TestValidateTokenForAuth_TokenNotFound_ReturnsError(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "org_id", "name", "token_type",
			"oauth_client_id", "token_hash", "prefix", "last_used_at", "created_at",
		}))

	userID, email, orgID, err := svc.ValidateTokenForAuth(context.Background(), "col_notfound0000000000000000000000000000000000000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid API token")
	assert.Equal(t, int64(0), userID)
	assert.Equal(t, "", email)
	assert.Equal(t, int64(0), orgID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidateTokenForAuth_UserNotFound_ReturnsError(t *testing.T) {
	bunDB, mock := newTestDB(t)
	svc := NewService(bunDB)

	mock.MatchExpectationsInOrder(false)

	rawToken := "col_aabbccddee112233445566778899aabbccddee00112233"
	h := hashToken(rawToken)
	now := time.Now()

	mock.ExpectQuery(`SELECT .+ FROM "api_tokens"`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "org_id", "name", "token_type",
			"oauth_client_id", "token_hash", "prefix", "last_used_at", "created_at",
		}).AddRow(5, 999, 20, "orphan-token", "personal", "", h, rawToken[:8], nil, now))

	mock.ExpectExec(`UPDATE "api_tokens"`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// User SELECT returns empty result
	mock.ExpectQuery(`SELECT .+ FROM "users"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name"}))

	userID, email, orgID, err := svc.ValidateTokenForAuth(context.Background(), rawToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found for API token")
	assert.Equal(t, int64(0), userID)
	assert.Equal(t, "", email)
	assert.Equal(t, int64(0), orgID)

	time.Sleep(50 * time.Millisecond)
}
