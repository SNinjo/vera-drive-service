package router

import (
	"net/http"

	"github.com/vera/vera-drive-service/internal/middleware"
	"github.com/vera/vera-drive-service/internal/url"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	httpMiddleware middleware.HTTPMiddleware,
	corsMiddleware middleware.CORSMiddleware,
	authMiddleware middleware.AuthMiddleware,
	urlHandler *url.Handler,
) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gin.HandlerFunc(httpMiddleware),
		gin.HandlerFunc(corsMiddleware),
	)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})
	r.StaticFile("/docs/swagger.yaml", "./api/swagger.yaml")
	r.StaticFile("/docs", "./api/swagger.html")

	url.RegisterRoutes(r, urlHandler, authMiddleware)

	return r
}
