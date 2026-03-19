package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	service     *OAuthService
	frontendURL string
}

func NewOAuthHandler(service *OAuthService, frontendURL string) *OAuthHandler {
	return &OAuthHandler{service: service, frontendURL: frontendURL}
}

func (h *OAuthHandler) Redirect(c *gin.Context) {
	provider := c.Param("provider")

	stateBytes := make([]byte, 16)
	rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	c.SetCookie("oauth_state", state, 600, "/", "", false, true)

	url, err := h.service.GetAuthURL(provider, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	cookieState, err := c.Cookie("oauth_state")
	if err != nil || cookieState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	tokenPair, err := h.service.HandleCallback(c.Request.Context(), provider, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "oauth failed"})
		return
	}

	// Redirect to frontend with tokens
	c.Redirect(http.StatusTemporaryRedirect,
		h.frontendURL+"/auth/callback?access_token="+tokenPair.AccessToken+"&refresh_token="+tokenPair.RefreshToken)
}

func (h *OAuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	oauth := r.Group("/auth")
	oauth.GET("/:provider", h.Redirect)
	oauth.GET("/:provider/callback", h.Callback)
}
