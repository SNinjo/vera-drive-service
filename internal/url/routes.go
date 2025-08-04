package url

import (
	"vera-identity-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware middleware.AuthMiddleware) {
	g := r.Group("/urls")
	g.Use(gin.HandlerFunc(authMiddleware))
	{
		g.POST("", handler.CreateURL)
		g.GET("/root-id", handler.GetRootID)
		g.GET("/:id", handler.GetURL)
		g.PUT("/:id", handler.ReplaceURL)
		g.DELETE("/:id", handler.DeleteURL)
	}
}
