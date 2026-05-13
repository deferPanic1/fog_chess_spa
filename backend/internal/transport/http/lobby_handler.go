package httphandler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
	"github.com/Gilf4/fog_chess/internal/transport/http/middleware"
	"github.com/go-chi/chi/v5"
)

type lobbyService interface {
	ListOpen(ctx context.Context, limit, offset int) ([]models.Lobby, int64, error)
	GetByCode(ctx context.Context, code string) (*models.Lobby, error)
	GetActiveByUser(ctx context.Context, userID int64) (*models.Lobby, error)
	Create(ctx context.Context, hostID int64, timeControl int) (*models.Lobby, error)
	Join(ctx context.Context, code string, userID int64) (*models.Lobby, error)
	Leave(ctx context.Context, code string, userID int64) (*models.Lobby, error)
}

type LobbyHandler struct {
	log          *slog.Logger
	lobbyService lobbyService
}

func NewLobbyHandler(log *slog.Logger, lobbyService lobbyService) *LobbyHandler {
	return &LobbyHandler{
		log:          log,
		lobbyService: lobbyService,
	}
}

func (h *LobbyHandler) List(w http.ResponseWriter, r *http.Request) {
	pagination := readPagination(r)

	lobbies, total, err := h.lobbyService.ListOpen(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		h.log.Error("failed to list lobbies", slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"lobbies":     lobbies,
		"pagination": newPaginationMeta(pagination.Page, pagination.Limit, total),
	})
}

func (h *LobbyHandler) GetByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	lobby, err := h.lobbyService.GetByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "lobby not found")
			return
		}

		h.log.Error("failed to get lobby", slog.String("code", code), slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": lobby})
}

func (h *LobbyHandler) Active(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	lobby, err := h.lobbyService.GetActiveByUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"lobby": nil})
			return
		}

		h.log.Error("failed to get active lobby", slog.Int64("user_id", userID), slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": lobby})
}

func (h *LobbyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		TimeControl int `json:"timeControl"`
	}

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	lobby, err := h.lobbyService.Create(r.Context(), userID, req.TimeControl)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidTimeControl):
			writeError(w, http.StatusBadRequest, "invalid time control")
		case errors.Is(err, repository.ErrUserAlreadyInLobby):
			writeError(w, http.StatusConflict, "user already has an active lobby")
		default:
			h.log.Error("failed to create lobby", slog.Int64("user_id", userID), slog.Any("err", err))
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"lobby": lobby})
}

func (h *LobbyHandler) Join(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	code := chi.URLParam(r, "code")
	lobby, err := h.lobbyService.Join(r.Context(), code, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			writeError(w, http.StatusNotFound, "lobby not found")
		case errors.Is(err, repository.ErrLobbyClosed):
			writeError(w, http.StatusConflict, "lobby is closed")
		case errors.Is(err, repository.ErrLobbyFull):
			writeError(w, http.StatusConflict, "lobby is full")
		case errors.Is(err, repository.ErrUserAlreadyInLobby):
			writeError(w, http.StatusConflict, "user already has an active lobby")
		default:
			h.log.Error("failed to join lobby", slog.Int64("user_id", userID), slog.String("code", code), slog.Any("err", err))
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": lobby})
}

func (h *LobbyHandler) Leave(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	code := chi.URLParam(r, "code")
	lobby, err := h.lobbyService.Leave(r.Context(), code, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			writeError(w, http.StatusNotFound, "lobby not found")
		case errors.Is(err, repository.ErrLobbyClosed):
			writeError(w, http.StatusConflict, "lobby is closed")
		case errors.Is(err, repository.ErrLobbyNotParticipant):
			writeError(w, http.StatusForbidden, "user is not a lobby participant")
		default:
			h.log.Error("failed to leave lobby", slog.Int64("user_id", userID), slog.String("code", code), slog.Any("err", err))
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"lobby": lobby})
}
