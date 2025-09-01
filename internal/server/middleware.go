package server

import (
"net/http"
"os"

"github.com/gin-gonic/gin"
)

func AdminAuth(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok := c.Request.Header.Get("X-Admin-Token")
		if tok == "" {
			// fallback to env if middleware called without token provided
			tok = os.Getenv("GOCHAN_ADMIN_TOKEN")
		}
		if tok == "" || tok != adminToken {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort(); return
		}
		c.Next()
	}
}
