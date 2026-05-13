package httphandler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Gilf4/fog_chess/internal/repository"
	"github.com/Gilf4/fog_chess/internal/service"
	"github.com/Gilf4/fog_chess/internal/transport/http/middleware"
)

type matchService interface {
	GetActiveByUser(ctx context.Context, userID int64) (*service.ActiveMatch, error)
}

type MatchHandler struct {
	log          *slog.Logger
	matchService matchService
}

func NewMatchHandler(log *slog.Logger, matchService matchService) *MatchHandler {
	return &MatchHandler{log: log, matchService: matchService}
}

func (h *MatchHandler) Active(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	match, err := h.matchService.GetActiveByUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"match": nil})
			return
		}
		h.log.Error("failed to load active match", slog.Int64("user_id", userID), slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"match": match})
}
