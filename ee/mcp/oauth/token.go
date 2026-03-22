package oauth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/apitoken"
)

type TokenHandler struct {
	db              *bun.DB
	apiTokenService *apitoken.Service
}

func NewTokenHandler(db *bun.DB, apiTokenService *apitoken.Service) *TokenHandler {
	return &TokenHandler{db: db, apiTokenService: apiTokenService}
}

// ServeHTTP handles POST /oauth/token — exchanges authorization code for access token.
func (h *TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	grantType := r.FormValue("grant_type")
	if grantType != "authorization_code" {
		writeTokenError(w, "unsupported_grant_type", "only authorization_code is supported")
		return
	}

	code := r.FormValue("code")
	codeVerifier := r.FormValue("code_verifier")
	if code == "" || codeVerifier == "" {
		writeTokenError(w, "invalid_request", "code and code_verifier are required")
		return
	}

	// Look up authorization code
	authCode := new(OAuthAuthorizationCode)
	err := h.db.NewSelect().Model(authCode).
		Where("oac.code = ?", code).
		Where("oac.used = ?", false).
		Where("oac.expires_at > ?", time.Now()).
		Scan(r.Context())
	if err != nil {
		writeTokenError(w, "invalid_grant", "invalid or expired authorization code")
		return
	}

	// Verify PKCE
	if !verifyPKCE(codeVerifier, authCode.CodeChallenge) {
		writeTokenError(w, "invalid_grant", "code_verifier does not match code_challenge")
		return
	}

	// Mark code as used
	authCode.Used = true
	if _, err := h.db.NewUpdate().Model(authCode).WherePK().Column("used").Exec(r.Context()); err != nil {
		writeTokenError(w, "server_error", "internal error")
		return
	}

	// Create an OAuth token scoped to this MCP client.
	token, rawToken, err := h.apiTokenService.CreateOAuth(r.Context(), authCode.UserID, authCode.OrgID, authCode.ClientID, "MCP OAuth")
	if err != nil {
		writeTokenError(w, "server_error", "failed to create access token")
		return
	}
	_ = token

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token": rawToken,
		"token_type":   "bearer",
	})
}

// verifyPKCE checks that SHA256(code_verifier) == code_challenge.
func verifyPKCE(verifier, challenge string) bool {
	h := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(h[:])
	return computed == challenge
}

func writeTokenError(w http.ResponseWriter, errCode, description string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":             errCode,
		"error_description": description,
	})
}
