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
	Config         *config.Config
	Router         *gin.Engine
	DB             *gorm.DB
	Logger         *zap.Logger
	HTTPMiddleware middleware.HTTPMiddleware
	AuthMiddleware middleware.AuthMiddleware
	URLHandler     *url.Handler
}

func NewApp(
	config *config.Config,
	router *gin.Engine,
	db *gorm.DB,
	logger *zap.Logger,
	httpMiddleware middleware.HTTPMiddleware,
	authMiddleware middleware.AuthMiddleware,
	urlHandler *url.Handler,
) *App {
	return &App{
		Config:         config,
		Router:         router,
		DB:             db,
		Logger:         logger,
		HTTPMiddleware: httpMiddleware,
		AuthMiddleware: authMiddleware,
		URLHandler:     urlHandler,
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
