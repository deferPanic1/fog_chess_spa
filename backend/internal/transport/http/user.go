package httphandler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
	"github.com/Gilf4/fog_chess/internal/transport/http/dto"
	"github.com/Gilf4/fog_chess/internal/transport/http/middleware"
	"github.com/go-chi/chi/v5"
)

type userService interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetStats(ctx context.Context, userID int64) (*models.UserStats, error)
	GetMatchHistory(ctx context.Context, userID int64, limit, offset int) ([]models.MatchHistoryItem, int64, error)
}

type UserHandler struct {
	log         *slog.Logger
	userService userService
}

func NewUserHandler(log *slog.Logger, userService userService) *UserHandler {
	return &UserHandler{log: log, userService: userService}
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("failed to load user", slog.Int64("user_id", userID), slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": dto.NewUserResponse(user),
	})
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		h.log.Error("failed to load current user", slog.Int64("user_id", userID), slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": dto.NewUserResponse(user),
	})
}

func (h *UserHandler) Stats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	stats, err := h.userService.GetStats(r.Context(), userID)
	if err != nil {
		h.log.Error("failed to load user stats", slog.Int64("user_id", userID), slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"stats": stats})
}

func (h *UserHandler) History(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	pagination := readPagination(r)

	history, total, err := h.userService.GetMatchHistory(r.Context(), userID, pagination.Limit, pagination.Offset)
	if err != nil {
		h.log.Error("failed to load user history", slog.Int64("user_id", userID), slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"history":     history,
		"pagination": newPaginationMeta(pagination.Page, pagination.Limit, total),
	})
}
