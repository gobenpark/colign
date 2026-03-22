package oauth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/auth"
)

type AuthorizeHandler struct {
	db             *bun.DB
	jwtManager     *auth.JWTManager
	baseURL        string
	newAuthService func() oauthSessionService
}

const (
	oauthAccessCookieName  = "colign_token"
	oauthRefreshCookieName = "colign_refresh_token"
)

type oauthSessionService interface {
	Login(ctx context.Context, req auth.LoginRequest) (*auth.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error)
}

func NewAuthorizeHandler(db *bun.DB, jwtManager *auth.JWTManager, baseURL string) *AuthorizeHandler {
	h := &AuthorizeHandler{
		db:         db,
		jwtManager: jwtManager,
		baseURL:    baseURL,
	}
	h.newAuthService = func() oauthSessionService {
		return auth.NewService(h.db, h.jwtManager)
	}
	return h
}

// ServeHTTP handles GET /oauth/authorize.
// If the user has a valid JWT cookie, show the consent screen.
// Otherwise, show a login form.
func (h *AuthorizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	state := r.URL.Query().Get("state")
	codeChallenge := r.URL.Query().Get("code_challenge")
	codeChallengeMethod := r.URL.Query().Get("code_challenge_method")

	if clientID == "" || redirectURI == "" || codeChallenge == "" {
		http.Error(w, "missing required parameters", http.StatusBadRequest)
		return
	}
	if codeChallengeMethod != "S256" {
		http.Error(w, "only S256 code_challenge_method is supported", http.StatusBadRequest)
		return
	}

	if claims := h.authenticateOAuthSession(w, r); claims != nil {
		// User is authenticated — handle consent
		if r.Method == http.MethodPost {
			h.handleConsent(w, r, claims, clientID, redirectURI, state, codeChallenge)
			return
		}
		h.showConsentPage(w, claims, clientID, redirectURI, state, codeChallenge)
		return
	}

	// Handle login form submission
	if r.Method == http.MethodPost && r.FormValue("action") == "login" {
		h.handleLogin(w, r, clientID, redirectURI, state, codeChallenge)
		return
	}

	// Show login page
	h.showLoginPage(w, clientID, redirectURI, state, codeChallenge)
}

func (h *AuthorizeHandler) handleLogin(w http.ResponseWriter, r *http.Request, clientID, redirectURI, state, codeChallenge string) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	tokenPair, err := h.newAuthService().Login(r.Context(), auth.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		h.showLoginPage(w, clientID, redirectURI, state, codeChallenge)
		return
	}

	h.setOAuthSessionCookies(w, tokenPair)

	claims, _ := h.jwtManager.ValidateAccessToken(tokenPair.AccessToken)
	h.showConsentPage(w, claims, clientID, redirectURI, state, codeChallenge)
}

func (h *AuthorizeHandler) authenticateOAuthSession(w http.ResponseWriter, r *http.Request) *auth.Claims {
	accessCookie, err := r.Cookie(oauthAccessCookieName)
	if err == nil && accessCookie.Value != "" {
		claims, err := h.jwtManager.ValidateAccessToken(accessCookie.Value)
		if err == nil {
			return claims
		}
	}

	refreshCookie, err := r.Cookie(oauthRefreshCookieName)
	if err != nil || refreshCookie.Value == "" {
		return nil
	}

	tokenPair, err := h.newAuthService().RefreshToken(r.Context(), refreshCookie.Value)
	if err != nil {
		h.clearOAuthSessionCookies(w)
		return nil
	}

	h.setOAuthSessionCookies(w, tokenPair)
	claims, err := h.jwtManager.ValidateAccessToken(tokenPair.AccessToken)
	if err != nil {
		h.clearOAuthSessionCookies(w)
		return nil
	}
	return claims
}

func (h *AuthorizeHandler) setOAuthSessionCookies(w http.ResponseWriter, tokenPair *auth.TokenPair) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthAccessCookieName,
		Value:    tokenPair.AccessToken,
		Path:     "/oauth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.AccessTokenDuration / time.Second),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     oauthRefreshCookieName,
		Value:    tokenPair.RefreshToken,
		Path:     "/oauth",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(auth.RefreshTokenDuration / time.Second),
	})
}

func (h *AuthorizeHandler) clearOAuthSessionCookies(w http.ResponseWriter) {
	for _, name := range []string{oauthAccessCookieName, oauthRefreshCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/oauth",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
	}
}

func (h *AuthorizeHandler) handleConsent(w http.ResponseWriter, r *http.Request, claims *auth.Claims, clientID, redirectURI, state, codeChallenge string) {
	code, err := generateAuthCode()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	authCode := &OAuthAuthorizationCode{
		UserID:        claims.UserID,
		OrgID:         claims.OrgID,
		ClientID:      clientID,
		Code:          code,
		CodeChallenge: codeChallenge,
		RedirectURI:   redirectURI,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}

	if _, err := h.db.NewInsert().Model(authCode).Exec(context.Background()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirect += "&state=" + state
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func generateAuthCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

var loginPageTmpl = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html><head><title>Colign - Sign In</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
body { font-family: -apple-system, sans-serif; background: #0a0a0a; color: #e5e5e5; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; }
.card { background: #171717; border: 1px solid #262626; border-radius: 12px; padding: 2rem; width: 100%; max-width: 400px; }
h1 { font-size: 1.5rem; margin: 0 0 0.5rem; }
p { color: #a3a3a3; font-size: 0.875rem; margin: 0 0 1.5rem; }
label { display: block; font-size: 0.875rem; margin-bottom: 0.25rem; }
input { width: 100%; padding: 0.5rem; background: #0a0a0a; border: 1px solid #262626; border-radius: 6px; color: #e5e5e5; font-size: 0.875rem; margin-bottom: 1rem; box-sizing: border-box; }
button { width: 100%; padding: 0.625rem; background: #e5e5e5; color: #0a0a0a; border: none; border-radius: 6px; font-size: 0.875rem; font-weight: 500; cursor: pointer; }
button:hover { background: #d4d4d4; }
</style></head><body>
<div class="card">
<h1>Sign in to Colign</h1>
<p>Authorize access for {{.ClientID}}</p>
<form method="POST">
<input type="hidden" name="action" value="login">
<input type="hidden" name="client_id" value="{{.ClientID}}">
<input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
<input type="hidden" name="state" value="{{.State}}">
<input type="hidden" name="code_challenge" value="{{.CodeChallenge}}">
<input type="hidden" name="code_challenge_method" value="S256">
<label>Email</label><input type="email" name="email" required autofocus>
<label>Password</label><input type="password" name="password" required>
<button type="submit">Sign In</button>
</form></div></body></html>`))

var consentPageTmpl = template.Must(template.New("consent").Parse(`<!DOCTYPE html>
<html><head><title>Colign - Authorize</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
body { font-family: -apple-system, sans-serif; background: #0a0a0a; color: #e5e5e5; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; }
.card { background: #171717; border: 1px solid #262626; border-radius: 12px; padding: 2rem; width: 100%; max-width: 400px; }
h1 { font-size: 1.5rem; margin: 0 0 0.5rem; }
p { color: #a3a3a3; font-size: 0.875rem; margin: 0 0 1rem; }
.user { color: #10b981; font-weight: 500; }
.perms { background: #0a0a0a; border: 1px solid #262626; border-radius: 6px; padding: 1rem; margin-bottom: 1.5rem; font-size: 0.875rem; }
.perms li { margin-bottom: 0.5rem; }
button { width: 100%; padding: 0.625rem; background: #10b981; color: #fff; border: none; border-radius: 6px; font-size: 0.875rem; font-weight: 500; cursor: pointer; }
button:hover { background: #059669; }
</style></head><body>
<div class="card">
<h1>Authorize Access</h1>
<p>Signed in as <span class="user">{{.Email}}</span></p>
<p><strong>{{.ClientID}}</strong> wants to access your Colign account:</p>
<ul class="perms">
<li>Read your projects and specs</li>
<li>Write and update spec documents</li>
<li>Manage implementation tasks</li>
</ul>
<form method="POST">
<input type="hidden" name="action" value="consent">
<input type="hidden" name="client_id" value="{{.ClientID}}">
<input type="hidden" name="redirect_uri" value="{{.RedirectURI}}">
<input type="hidden" name="state" value="{{.State}}">
<input type="hidden" name="code_challenge" value="{{.CodeChallenge}}">
<input type="hidden" name="code_challenge_method" value="S256">
<button type="submit">Authorize</button>
</form></div></body></html>`))

func (h *AuthorizeHandler) showLoginPage(w http.ResponseWriter, clientID, redirectURI, state, codeChallenge string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = loginPageTmpl.Execute(w, map[string]string{
		"ClientID":      clientID,
		"RedirectURI":   redirectURI,
		"State":         state,
		"CodeChallenge": codeChallenge,
	})
}

func (h *AuthorizeHandler) showConsentPage(w http.ResponseWriter, claims *auth.Claims, clientID, redirectURI, state, codeChallenge string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = consentPageTmpl.Execute(w, map[string]string{
		"Email":         claims.Email,
		"ClientID":      clientID,
		"RedirectURI":   redirectURI,
		"State":         state,
		"CodeChallenge": codeChallenge,
	})
}
