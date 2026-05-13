package httpserver

import (
	"log/slog"

	"github.com/Gilf4/fog_chess/internal/config"
	"github.com/Gilf4/fog_chess/internal/service"
	httphandler "github.com/Gilf4/fog_chess/internal/transport/http"
	"github.com/Gilf4/fog_chess/internal/transport/http/middleware"
	ws "github.com/Gilf4/fog_chess/internal/transport/websocket"
	"github.com/go-chi/chi/v5"
)

func registerRoutes(
	r *chi.Mux,
	log *slog.Logger,
	authSvc *service.AuthService,
	userSvc *service.UserService,
	lobbySvc *service.LobbyService,
	matchSvc *service.MatchService,
	jwt *config.JWT,
) {
	authHandler := httphandler.NewAuthHandler(log, authSvc, userSvc, jwt.AccessTTL)
	userHandler := httphandler.NewUserHandler(log, userSvc)
	lobbyHandler := httphandler.NewLobbyHandler(log, lobbySvc)
	matchHandler := httphandler.NewMatchHandler(log, matchSvc)

	wsHandler := ws.New(log, lobbySvc, matchSvc)

	r.Route("/api/v1", func(r chi.Router) {

		// auth
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(jwt.Secret))
				r.Get("/me", authHandler.Me)
			})
		})

		// user
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwt.Secret))
			r.Get("/users/me", userHandler.Me)
			r.Get("/users/stats", userHandler.Stats)
			r.Get("/users/history", userHandler.History)
			r.Get("/matches/active", matchHandler.Active)
			r.Get("/users/{id}", userHandler.GetByID)
		})

		// lobbies
		r.Route("/lobbies", func(r chi.Router) {
			r.Get("/", lobbyHandler.List)

			r.Group(func(r chi.Router) {
				r.Use(middleware.Auth(jwt.Secret))
				r.Get("/active", lobbyHandler.Active)
				r.Post("/", lobbyHandler.Create)
			})

			r.Route("/{code}", func(r chi.Router) {
				r.Get("/", lobbyHandler.GetByCode)
				r.Group(func(r chi.Router) {
					r.Use(middleware.Auth(jwt.Secret))
					r.Post("/join", lobbyHandler.Join)
					r.Post("/leave", lobbyHandler.Leave)
				})
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwt.Secret))
			r.Get("/ws", wsHandler.HandleLobby)
			r.Get("/ws/match", wsHandler.HandleMatch)
		})
	})
}
