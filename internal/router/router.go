package router

import (
	"net/http"
	"vera-identity-service/internal/middleware"
	"vera-identity-service/internal/url"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	httpHandler middleware.HTTPHandler,
	corsHandler middleware.CORSHandler,
	authHandler middleware.AuthHandler,
	urlHandler *url.Handler,
) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gin.HandlerFunc(httpHandler),
		gin.HandlerFunc(corsHandler),
	)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})
	r.StaticFile("/docs/swagger.yaml", "./api/swagger.yaml")
	r.StaticFile("/docs", "./api/swagger.html")

	url.RegisterRoutes(r, urlHandler, authHandler)

	return r
}
