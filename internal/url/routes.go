package url

import (
	"vera-identity-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, urlHandler *Handler, authHandler middleware.AuthHandler) {
	g := r.Group("/urls")
	g.Use(gin.HandlerFunc(authHandler))
	{
		g.POST("", urlHandler.CreateURL)
		g.GET("/root-id", urlHandler.GetRootID)
		g.GET("/:id", urlHandler.GetURL)
		g.PUT("/:id", urlHandler.ReplaceURL)
		g.DELETE("/:id", urlHandler.DeleteURL)
	}
}
