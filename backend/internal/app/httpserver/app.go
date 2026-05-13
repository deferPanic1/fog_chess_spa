package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Gilf4/fog_chess/internal/config"
	"github.com/Gilf4/fog_chess/internal/service"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

type App struct {
	log    *slog.Logger
	server *http.Server
	port   int
}

func New(
	log *slog.Logger,
	cfg *config.Config,
	authSvc *service.AuthService,
	userSvc *service.UserService,
	lobbySvc *service.LobbyService,
	matchSvc *service.MatchService,

) *App {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	registerRoutes(r, log, authSvc, userSvc, lobbySvc, matchSvc, &cfg.JWT)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &App{
		log:    log,
		server: srv,
		port:   cfg.Server.Port,
	}
}

func (app *App) MustRun() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func (app *App) Run() error {
	const op = "httpserver.Run"

	log := app.log.With(
		slog.String("op", op),
		slog.Int("port", app.port),
	)

	log.Info("http server is running")

	if err := app.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Error("http server stopped with error", slog.Any("err", err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("http server stopped")

	return nil
}

func (app *App) Stop() {
	const op = "httpserver.Stop"

	app.log.With(slog.String("op", op)).
		Info("stopping http server", slog.Int("port", app.port))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.server.Shutdown(ctx); err != nil {
		app.log.Error("forced shutdown", slog.Any("err", err))
	}
}
