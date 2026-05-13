package httphandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Gilf4/fog_chess/internal/repository"
	"github.com/Gilf4/fog_chess/internal/service"
)

func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrEmailExists):
		writeError(w, http.StatusConflict, "email already exists")
	case errors.Is(err, repository.ErrUsernameExists):
		writeError(w, http.StatusConflict, "username already exists")
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid username or password")
	default:
		h.log.Error("internal error", slog.Any("err", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
