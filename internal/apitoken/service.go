package apitoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/models"
)

type Service struct {
	db *bun.DB
}

const oauthTokenRetention = 30 * 24 * time.Hour

func NewService(db *bun.DB) *Service {
	return &Service{db: db}
}

func generateToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "col_" + hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

func (s *Service) Create(ctx context.Context, userID, orgID int64, name string) (*models.APIToken, string, error) {
	return s.createWithType(ctx, userID, orgID, "", name, "personal")
}

func (s *Service) CreateOAuth(ctx context.Context, userID, orgID int64, clientID, name string) (*models.APIToken, string, error) {
	clientID = strings.TrimSpace(clientID)
	if clientID == "" {
		return nil, "", fmt.Errorf("oauth client id is required")
	}

	// Replace the token for the same MCP client without invalidating other active clients.
	_, _ = s.db.NewDelete().Model((*models.APIToken)(nil)).
		Where("user_id = ?", userID).
		Where("org_id = ?", orgID).
		Where("token_type = ?", "oauth").
		Where("oauth_client_id = ?", clientID).
		Exec(ctx)

	// Opportunistically prune stale OAuth tokens so they do not accumulate forever.
	cutoff := time.Now().Add(-oauthTokenRetention)
	_, _ = s.db.NewDelete().Model((*models.APIToken)(nil)).
		Where("user_id = ?", userID).
		Where("org_id = ?", orgID).
		Where("token_type = ?", "oauth").
		Where("COALESCE(last_used_at, created_at) < ?", cutoff).
		Exec(ctx)

	return s.createWithType(ctx, userID, orgID, clientID, name, "oauth")
}

func (s *Service) createWithType(ctx context.Context, userID, orgID int64, clientID, name, tokenType string) (*models.APIToken, string, error) {
	raw, err := generateToken()
	if err != nil {
		return nil, "", err
	}

	token := &models.APIToken{
		UserID:        userID,
		OrgID:         orgID,
		Name:          name,
		TokenType:     tokenType,
		OAuthClientID: clientID,
		TokenHash:     hashToken(raw),
		Prefix:        raw[:8],
	}

	if _, err := s.db.NewInsert().Model(token).Exec(ctx); err != nil {
		return nil, "", err
	}

	return token, raw, nil
}

func (s *Service) List(ctx context.Context, userID, orgID int64) ([]models.APIToken, error) {
	var tokens []models.APIToken
	err := s.db.NewSelect().Model(&tokens).
		Where("at.user_id = ?", userID).
		Where("at.org_id = ?", orgID).
		Where("at.token_type = ?", "personal").
		OrderExpr("at.created_at DESC").
		Scan(ctx)
	return tokens, err
}

func (s *Service) Delete(ctx context.Context, userID, tokenID int64) error {
	_, err := s.db.NewDelete().Model((*models.APIToken)(nil)).
		Where("id = ?", tokenID).
		Where("user_id = ?", userID).
		Exec(ctx)
	return err
}

func (s *Service) ValidateToken(ctx context.Context, rawToken string) (*models.APIToken, error) {
	h := hashToken(rawToken)
	token := new(models.APIToken)
	err := s.db.NewSelect().Model(token).
		Where("at.token_hash = ?", h).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("invalid API token")
	}

	go func() {
		now := time.Now()
		_, _ = s.db.NewUpdate().Model((*models.APIToken)(nil)).
			Set("last_used_at = ?", now).
			Where("id = ?", token.ID).
			Exec(context.Background())
	}()

	return token, nil
}

// ValidateTokenForAuth implements auth.APITokenValidator.
func (s *Service) ValidateTokenForAuth(ctx context.Context, rawToken string) (int64, string, int64, error) {
	token, err := s.ValidateToken(ctx, rawToken)
	if err != nil {
		return 0, "", 0, err
	}

	user := new(models.User)
	if err := s.db.NewSelect().Model(user).Where("id = ?", token.UserID).Scan(ctx); err != nil {
		return 0, "", 0, fmt.Errorf("user not found for API token")
	}

	return token.UserID, user.Email, token.OrgID, nil
}
