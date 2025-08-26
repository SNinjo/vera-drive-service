package url

import (
	"github.com/vera/vera-drive-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *Handler, authMiddleware middleware.AuthMiddleware) {
	g := r.Group("/urls")
	g.Use(gin.HandlerFunc(authMiddleware))
	{
		g.POST("", h.CreateURL)
		g.GET("/root-id", h.GetRootID)
		g.GET("/:id", h.GetURL)
		g.PUT("/:id", h.ReplaceURL)
		g.DELETE("/:id", h.DeleteURL)
	}
}
