package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
	"gorm.io/gorm"
)

type MatchRepository struct {
	db *gorm.DB
}

func NewMatchRepository(db *gorm.DB) *MatchRepository {
	return &MatchRepository{db: db}
}

func (r *MatchRepository) Create(ctx context.Context, whitePlayer, blackPlayer int64) (*models.Match, error) {
	m := matchModel{
		WhitePlayer: whitePlayer,
		BlackPlayer: blackPlayer,
		Status:      "active",
		BoardFEN:    "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		CurrentTurn: "white",
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, fmt.Errorf("create match: %w", err)
	}
	return &models.Match{
		ID:               m.ID,
		WhitePlayer:      m.WhitePlayer,
		BlackPlayer:      m.BlackPlayer,
		Status:           m.Status,
		Result:           m.Result,
		BoardFEN:         m.BoardFEN,
		CurrentTurn:      m.CurrentTurn,
		WhiteRatingDelta: m.WhiteRatingDelta,
		BlackRatingDelta: m.BlackRatingDelta,
		CreatedAt:        m.CreatedAt,
		FinishedAt:       m.FinishedAt,
	}, nil
}

func (r *MatchRepository) GetByID(ctx context.Context, id string) (*models.Match, error) {
	var m matchModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("get match: %w", err)
	}

	return &models.Match{
		ID:               m.ID,
		WhitePlayer:      m.WhitePlayer,
		BlackPlayer:      m.BlackPlayer,
		Status:           m.Status,
		Result:           m.Result,
		BoardFEN:         m.BoardFEN,
		CurrentTurn:      m.CurrentTurn,
		WhiteRatingDelta: m.WhiteRatingDelta,
		BlackRatingDelta: m.BlackRatingDelta,
		CreatedAt:        m.CreatedAt,
		FinishedAt:       m.FinishedAt,
	}, nil
}

func (r *MatchRepository) GetActiveByUser(ctx context.Context, userID int64) (*models.Match, error) {
	var m matchModel
	if err := r.db.WithContext(ctx).
		Where("(white_player = ? OR black_player = ?) AND status = ?", userID, userID, "active").
		Order("created_at DESC").
		Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("get active match: %w", err)
	}

	return &models.Match{
		ID:               m.ID,
		WhitePlayer:      m.WhitePlayer,
		BlackPlayer:      m.BlackPlayer,
		Status:           m.Status,
		Result:           m.Result,
		BoardFEN:         m.BoardFEN,
		CurrentTurn:      m.CurrentTurn,
		WhiteRatingDelta: m.WhiteRatingDelta,
		BlackRatingDelta: m.BlackRatingDelta,
		CreatedAt:        m.CreatedAt,
		FinishedAt:       m.FinishedAt,
	}, nil
}

func (r *MatchRepository) Update(ctx context.Context, match *models.Match) error {
	updates := map[string]any{
		"status":             match.Status,
		"result":             match.Result,
		"board_fen":          match.BoardFEN,
		"current_turn":       match.CurrentTurn,
		"white_rating_delta": match.WhiteRatingDelta,
		"black_rating_delta": match.BlackRatingDelta,
		"finished_at":        match.FinishedAt,
	}

	if err := r.db.WithContext(ctx).
		Model(&matchModel{}).
		Where("id = ?", match.ID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("update match: %w", err)
	}

	return nil
}

func (r *MatchRepository) Complete(ctx context.Context, match *models.Match, whiteNewRating, blackNewRating int32) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"status":             match.Status,
			"result":             match.Result,
			"board_fen":          match.BoardFEN,
			"current_turn":       match.CurrentTurn,
			"white_rating_delta": match.WhiteRatingDelta,
			"black_rating_delta": match.BlackRatingDelta,
			"finished_at":        match.FinishedAt,
		}

		if err := tx.Model(&matchModel{}).
			Where("id = ?", match.ID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("complete match update: %w", err)
		}

		if err := tx.Model(&userModel{}).
			Where("id = ?", match.WhitePlayer).
			Update("rating", whiteNewRating).Error; err != nil {
			return fmt.Errorf("update white rating: %w", err)
		}

		if err := tx.Model(&userModel{}).
			Where("id = ?", match.BlackPlayer).
			Update("rating", blackNewRating).Error; err != nil {
			return fmt.Errorf("update black rating: %w", err)
		}

		return nil
	})
}
