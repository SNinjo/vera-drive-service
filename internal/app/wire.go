//go:build wireinject
// +build wireinject

package app

import (
	"vera-identity-service/internal/config"
	"vera-identity-service/internal/db"
	"vera-identity-service/internal/logger"
	"vera-identity-service/internal/middleware"
	"vera-identity-service/internal/router"
	"vera-identity-service/internal/url"

	"github.com/google/wire"
)

func InitApp() (*App, error) {
	wire.Build(
		config.NewConfig,
		router.NewRouter,
		db.NewDatabase,
		logger.NewLogger,
		middleware.NewHTTPHandler,
		middleware.NewCORSHandler,
		middleware.NewAuthHandler,
		url.NewRepository,
		url.NewService,
		url.NewHandler,
		NewApp,
	)
	return &App{}, nil
}
