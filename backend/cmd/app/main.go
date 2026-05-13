package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	backendapp "github.com/Gilf4/fog_chess/internal/app"
	"github.com/Gilf4/fog_chess/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("starting application",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.Server.Port),
	)

	ctx := context.Background()

	application := backendapp.New(ctx, log, cfg)

	go func() {
		application.HTTPServer.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sig := <-stop

	application.HTTPServer.Stop()
	log.Info("gracefully stopped", "signal", sig)
}

func setupLogger(env string) *slog.Logger {
	switch env {
	case envLocal, envDev:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}
