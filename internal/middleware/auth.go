package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gobenpark/colign/internal/auth"
)

const (
	ContextKeyUserID = "user_id"
	ContextKeyEmail  = "email"
)

func JWTAuth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Next()
	}
}

func GetUserID(c *gin.Context) int64 {
	v, _ := c.Get(ContextKeyUserID)
	id, _ := v.(int64)
	return id
}
