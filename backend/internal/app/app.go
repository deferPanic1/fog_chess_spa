package app

import (
	"context"
	"log/slog"

	"github.com/Gilf4/fog_chess/internal/app/httpserver"
	"github.com/Gilf4/fog_chess/internal/config"
	"github.com/Gilf4/fog_chess/internal/db"
	"github.com/Gilf4/fog_chess/internal/repository/postgres"
	"github.com/Gilf4/fog_chess/internal/service"
	"gorm.io/gorm"
)

type App struct {
	DB         *gorm.DB
	HTTPServer *httpserver.App
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *App {
	database, err := db.New(ctx, cfg.DB)
	if err != nil {
		panic(err)
	}

	userRepo := postgres.NewUserRepo(database)
	lobbyRepo := postgres.NewLobbyRepository(database)
	matchRepo := postgres.NewMatchRepository(database)

	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.AccessTTL)
	userService := service.NewUserService(userRepo)
	lobbyService := service.NewLobbyService(lobbyRepo)
	matchService := service.NewMatchService(matchRepo, userRepo)

	return &App{
		DB:         database,
		HTTPServer: httpserver.New(log, cfg, authService, userService, lobbyService, matchService),
	}
}
