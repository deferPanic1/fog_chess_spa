package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *models.User) (*models.User, error) {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if isUniqueViolation(result.Error, "users_email_key") {
			return nil, repository.ErrEmailExists
		}
		if isUniqueViolation(result.Error, "users_username_key") {
			return nil, repository.ErrUsernameExists
		}
		return nil, result.Error
	}
	return user, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *UserRepo) List(ctx context.Context, username string, limit, offset int) ([]models.User, int64, error) {
	var users []models.User

	query := r.db.WithContext(ctx).Model(&models.User{})
	if username != "" {
		query = query.Where("LOWER(username) LIKE ?", "%"+strings.ToLower(username)+"%")
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	if err := query.Session(&gorm.Session{}).
		Order("username ASC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}

	return users, total, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, hashedPassword string) error {
	result := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.User{}, id)
	if result.Error != nil {
		if isForeignKeyViolation(result.Error) {
			return repository.ErrUserInUse
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *UserRepo) GetStats(ctx context.Context, userID int64) (*models.UserStats, error) {
	var stats models.UserStats

	err := r.db.WithContext(ctx).
		Model(&matchModel{}).
		Select(`
            COUNT(*) AS games,
            SUM(CASE
                WHEN (white_player = ? AND result IN ('white', 'white_win')) OR (black_player = ? AND result IN ('black', 'black_win')) THEN 1
                ELSE 0 END) AS wins,
            SUM(CASE
                WHEN (white_player = ? AND result IN ('black', 'black_win')) OR (black_player = ? AND result IN ('white', 'white_win')) THEN 1
                ELSE 0 END) AS losses,
            SUM(CASE WHEN result = 'draw' THEN 1 ELSE 0 END) AS draws
        `, userID, userID, userID, userID).
		Where("(white_player = ? OR black_player = ?) AND status = 'finished'", userID, userID).
		Scan(&stats).Error
	if err != nil {
		return nil, fmt.Errorf("get user stats: %w", err)
	}

	return &stats, nil
}

func (r *UserRepo) GetMatchHistory(ctx context.Context, userID int64, limit, offset int) ([]models.MatchHistoryItem, int64, error) {
	type historyRow struct {
		FinishedAt   time.Time `gorm:"column:finished_at"`
		Opponent     string    `gorm:"column:opponent"`
		Result       string    `gorm:"column:result"`
		RatingChange int       `gorm:"column:rating_change"`
	}

	var rows []historyRow
	baseQuery := r.db.WithContext(ctx).
		Table("matches m").
		Joins("JOIN users u_white ON u_white.id = m.white_player").
		Joins("JOIN users u_black ON u_black.id = m.black_player").
		Where("(m.white_player = ? OR m.black_player = ?) AND m.status = 'finished'", userID, userID)

	var total int64
	if err := baseQuery.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count match history: %w", err)
	}

	err := baseQuery.Session(&gorm.Session{}).
		Select(`
            m.finished_at,
            CASE
                WHEN m.white_player = ? THEN u_black.username
                ELSE u_white.username
            END AS opponent,
            CASE
                WHEN m.result = 'draw' THEN 'draw'
                WHEN (m.white_player = ? AND m.result IN ('white', 'white_win')) OR (m.black_player = ? AND m.result IN ('black', 'black_win')) THEN 'win'
                ELSE 'loss'
            END AS result,
            CASE
                WHEN m.white_player = ? THEN m.white_rating_delta
                ELSE m.black_rating_delta
            END AS rating_change
        `, userID, userID, userID, userID).
		Order("m.finished_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("get match history: %w", err)
	}

	history := make([]models.MatchHistoryItem, 0, len(rows))
	for _, row := range rows {
		history = append(history, models.MatchHistoryItem{
			Date:         row.FinishedAt,
			Opponent:     row.Opponent,
			Result:       row.Result,
			RatingChange: strconv.Itoa(row.RatingChange),
		})
	}

	return history, total, nil
}
