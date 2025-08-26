//go:build wireinject
// +build wireinject

package app

import (
	"github.com/vera/vera-drive-service/internal/config"
	"github.com/vera/vera-drive-service/internal/db"
	"github.com/vera/vera-drive-service/internal/logger"
	"github.com/vera/vera-drive-service/internal/middleware"
	"github.com/vera/vera-drive-service/internal/router"
	"github.com/vera/vera-drive-service/internal/url"

	"github.com/google/wire"
)

func InitApp() (*App, error) {
	wire.Build(
		config.NewConfig,
		router.NewRouter,
		db.NewDatabase,
		logger.NewLogger,
		middleware.NewHTTPMiddleware,
		middleware.NewCORSMiddleware,
		middleware.NewAuthMiddleware,
		url.NewRepository,
		url.NewService,
		url.NewHandler,
		NewApp,
	)
	return &App{}, nil
}
