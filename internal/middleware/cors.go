package middleware

import (
	"net/http"
	"slices"
	"vera-identity-service/internal/config"

	"github.com/gin-gonic/gin"
)

type CORSHandler gin.HandlerFunc

func NewCORSHandler(config *config.Config) CORSHandler {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if slices.Contains(config.ALLOWED_ORIGINS, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
		}

		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Max-Age", "86400") // 24 hours
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
