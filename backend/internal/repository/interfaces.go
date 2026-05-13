package repository

import (
	"context"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	domain "github.com/Gilf4/fog_chess/internal/domain/models"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	UpdatePassword(ctx context.Context, id int64, hashedPassword string) error
	Delete(ctx context.Context, id int64) error
	GetStats(ctx context.Context, userID int64) (*domain.UserStats, error)
	GetMatchHistory(ctx context.Context, userID int64, limit, offset int) ([]domain.MatchHistoryItem, int64, error)
}

type LobbyRepository interface {
	ListOpen(ctx context.Context, limit, offset int) ([]domain.Lobby, int64, error)
	GetByCode(ctx context.Context, code string) (*domain.Lobby, error)
	GetActiveByUser(ctx context.Context, userID int64) (*domain.Lobby, error)
	Create(ctx context.Context, hostID int64, timeControl int) (*domain.Lobby, error)
	Join(ctx context.Context, code string, userID int64) (*domain.Lobby, error)
	Leave(ctx context.Context, code string, userID int64) (*domain.Lobby, error)
	SetReady(ctx context.Context, code string, userID int64, ready bool) (*domain.Lobby, error)
	Close(ctx context.Context, code string) error
}

type MatchRepository interface {
	Create(ctx context.Context, whitePlayer, blackPlayer int64) (*models.Match, error)
	GetByID(ctx context.Context, id string) (*models.Match, error)
	GetActiveByUser(ctx context.Context, userID int64) (*models.Match, error)
	Update(ctx context.Context, match *models.Match) error
	Complete(ctx context.Context, match *models.Match, whiteNewRating, blackNewRating int32) error
}
