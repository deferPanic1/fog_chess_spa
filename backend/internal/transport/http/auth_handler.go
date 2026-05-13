package httphandler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
	"github.com/Gilf4/fog_chess/internal/transport/http/dto"
	"github.com/Gilf4/fog_chess/internal/transport/http/middleware"
)

type authService interface {
	Register(ctx context.Context, email, username, password string) (*models.User, error)
	Login(ctx context.Context, username, password string) (string, *models.User, error)
}

type AuthHandler struct {
	log          *slog.Logger
	authService  authService
	userService  userService
	lobbyService lobbyService
	accessTTL    time.Duration
}

func NewAuthHandler(log *slog.Logger, authService authService, userService userService, accessTTL time.Duration) *AuthHandler {
	return &AuthHandler{
		log:         log,
		authService: authService,
		userService: userService,
		accessTTL:   accessTTL,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := h.authService.Register(r.Context(), req.Email, req.Username, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user": dto.NewUserResponse(user),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	token, user, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	setAccessTokenCookie(w, token, h.accessTTL)
	writeJSON(w, http.StatusOK, dto.LoginResponse{
		AccessToken: token,
		User:        dto.NewUserResponse(user),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAccessTokenCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		h.log.Error("failed to get user", slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": dto.NewUserResponse(user),
	})
}
