package middleware

import (
	"net/http"
	"strings"

	"github.com/edberthansel/BtechDevCases/backend/internal/token"
	"github.com/gin-gonic/gin"
)

const (
	ctxUserID = "userID"
	ctxEmail  = "email"

	RefreshedTokenHeader = "X-Refreshed-Token"
)

func Auth(tokens *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed Authorization header"})
			return
		}

		claims, err := tokens.Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxEmail, claims.Email)

		// Sliding expiration: setiap request yang lolos mendapat token baru.
		// Header harus di-set SEBELUM handler menulis response.
		if fresh, _, err := tokens.Generate(claims.UserID, claims.Email); err == nil {
			c.Header(RefreshedTokenHeader, fresh)
		}

		c.Next()
	}
}

// UserID dan Email dipakai handler untuk membaca identitas dari token.
func UserID(c *gin.Context) string { return c.GetString(ctxUserID) }
func Email(c *gin.Context) string  { return c.GetString(ctxEmail) }