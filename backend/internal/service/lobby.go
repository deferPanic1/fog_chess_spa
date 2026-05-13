package service

import (
	"context"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
)

type LobbyService struct {
	lobbies repository.LobbyRepository
}

func NewLobbyService(lobbies repository.LobbyRepository) *LobbyService {
	return &LobbyService{lobbies: lobbies}
}

func (s *LobbyService) ListOpen(ctx context.Context, limit, offset int) ([]models.Lobby, int64, error) {
	return s.lobbies.ListOpen(ctx, limit, offset)
}

func (s *LobbyService) GetByCode(ctx context.Context, code string) (*models.Lobby, error) {
	return s.lobbies.GetByCode(ctx, code)
}

func (s *LobbyService) GetActiveByUser(ctx context.Context, userID int64) (*models.Lobby, error) {
	return s.lobbies.GetActiveByUser(ctx, userID)
}

func (s *LobbyService) Create(ctx context.Context, hostID int64, timeControl int) (*models.Lobby, error) {
	return s.lobbies.Create(ctx, hostID, timeControl)
}

func (s *LobbyService) Join(ctx context.Context, code string, userID int64) (*models.Lobby, error) {
	return s.lobbies.Join(ctx, code, userID)
}

func (s *LobbyService) Leave(ctx context.Context, code string, userID int64) (*models.Lobby, error) {
	return s.lobbies.Leave(ctx, code, userID)
}

func (s *LobbyService) SetReady(ctx context.Context, code string, userID int64, ready bool) (*models.Lobby, error) {
	return s.lobbies.SetReady(ctx, code, userID, ready)
}

func (s *LobbyService) Close(ctx context.Context, code string) error {
	return s.lobbies.Close(ctx, code)
}
