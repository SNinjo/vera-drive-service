package app

import (
	"vera-identity-service/internal/config"
	"vera-identity-service/internal/middleware"
	"vera-identity-service/internal/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Config      *config.Config
	Router      *gin.Engine
	DB          *gorm.DB
	Logger      *zap.Logger
	HTTPHandler middleware.HTTPHandler
	CORSHandler middleware.CORSHandler
	AuthHandler middleware.AuthHandler
	URLHandler  *url.Handler
}

func NewApp(
	config *config.Config,
	router *gin.Engine,
	db *gorm.DB,
	logger *zap.Logger,
	httpHandler middleware.HTTPHandler,
	corsHandler middleware.CORSHandler,
	authHandler middleware.AuthHandler,
	urlHandler *url.Handler,
) *App {
	return &App{
		Config:      config,
		Router:      router,
		DB:          db,
		Logger:      logger,
		HTTPHandler: httpHandler,
		CORSHandler: corsHandler,
		AuthHandler: authHandler,
		URLHandler:  urlHandler,
	}
}

func (a *App) Close() {
	a.Logger.Sync()
}

func (a *App) Run() {
	if err := a.Router.Run(a.Config.Domain + ":" + a.Config.Port); err != nil {
		a.Logger.Fatal("failed to run server", zap.Error(err))
	}
}
