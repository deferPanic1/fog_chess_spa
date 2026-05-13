package service

import (
	"context"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *UserService) GetStats(ctx context.Context, userID int64) (*models.UserStats, error) {
	return s.users.GetStats(ctx, userID)
}

func (s *UserService) GetMatchHistory(ctx context.Context, userID int64, limit, offset int) ([]models.MatchHistoryItem, int64, error) {
	return s.users.GetMatchHistory(ctx, userID, limit, offset)
}
