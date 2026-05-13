package postgres

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Gilf4/fog_chess/internal/domain/models"
	"github.com/Gilf4/fog_chess/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LobbyRepository struct {
	db *gorm.DB
}

func NewLobbyRepository(db *gorm.DB) *LobbyRepository {
	return &LobbyRepository{db: db}
}

func (r *LobbyRepository) ListOpen(ctx context.Context, limit, offset int) ([]models.Lobby, int64, error) {
	var rows []lobbyView
	baseQuery := r.baseLobbyQuery(r.db.WithContext(ctx)).
		Where("l.status IN ?", []string{"waiting", "full", "ready"})

	var total int64
	if err := baseQuery.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count open lobbies: %w", err)
	}

	err := baseQuery.Session(&gorm.Session{}).
		Order("l.updated_at DESC").
		Order("l.created_at DESC").
		Offset(offset).
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list open lobbies: %w", err)
	}

	lobbies := make([]models.Lobby, 0, limit)
	for _, row := range rows {
		lobbies = append(lobbies, row.toDomain())
	}

	return lobbies, total, nil
}

func (r *LobbyRepository) GetByCode(ctx context.Context, code string) (*models.Lobby, error) {
	var row lobbyView
	err := r.baseLobbyQuery(r.db.WithContext(ctx)).
		Where("l.code = ?", normalizeCode(code)).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("get lobby by code: %w", err)
	}

	lobby := row.toDomain()
	return &lobby, nil
}

func (r *LobbyRepository) Create(ctx context.Context, hostID int64, timeControl int) (*models.Lobby, error) {
	if timeControl <= 0 {
		return nil, repository.ErrInvalidTimeControl
	}

	if lobby, err := r.GetActiveByUser(ctx, hostID); err == nil && lobby != nil {
		return nil, repository.ErrUserAlreadyInLobby
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	for range 8 {
		code, err := generateLobbyCode(6)
		if err != nil {
			return nil, fmt.Errorf("generate lobby code: %w", err)
		}

		model := lobbyModel{
			Code:        code,
			HostID:      hostID,
			Status:      "waiting",
			TimeControl: timeControl,
			HostReady:   false,
			GuestReady:  false,
		}

		err = r.db.WithContext(ctx).Create(&model).Error
		if err != nil {
			if isDuplicateKeyError(err) {
				continue
			}
			return nil, fmt.Errorf("create lobby: %w", err)
		}

		return r.GetByCode(ctx, code)
	}

	return nil, fmt.Errorf("create lobby: unable to generate unique code")
}

func (r *LobbyRepository) Join(ctx context.Context, code string, userID int64) (*models.Lobby, error) {
	normCode := normalizeCode(code)

	if lobby, err := r.GetActiveByUser(ctx, userID); err == nil && !strings.EqualFold(lobby.Code, normCode) {
		return nil, repository.ErrUserAlreadyInLobby
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lobby lobbyModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ?", normCode).
			Take(&lobby).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrNotFound
			}
			return fmt.Errorf("select join lobby: %w", err)
		}

		switch {
		case lobby.Status == "closed":
			return repository.ErrLobbyClosed
		case lobby.HostID == userID:
			return nil
		case lobby.GuestID != nil && *lobby.GuestID == userID:
			return nil
		case lobby.GuestID != nil:
			return repository.ErrLobbyFull
		}

		lobby.GuestID = &userID
		lobby.GuestReady = false
		lobby.Status = "full"

		if err := tx.Model(&lobby).Updates(map[string]any{
			"guest_id":    lobby.GuestID,
			"guest_ready": lobby.GuestReady,
			"status":      lobby.Status,
			"updated_at":  gorm.Expr("now()"),
		}).Error; err != nil {
			return fmt.Errorf("update join lobby: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByCode(ctx, normCode)
}

func (r *LobbyRepository) Leave(ctx context.Context, code string, userID int64) (*models.Lobby, error) {
	normCode := normalizeCode(code)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lobby lobbyModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ?", normCode).
			Take(&lobby).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrNotFound
			}
			return fmt.Errorf("select leave lobby: %w", err)
		}

		if lobby.Status == "closed" {
			return repository.ErrLobbyClosed
		}

		switch {
		case lobby.HostID == userID:
			return tx.Model(&lobby).Updates(map[string]any{
				"status":      "closed",
				"host_ready":  false,
				"guest_ready": false,
				"updated_at":  gorm.Expr("now()"),
			}).Error
		case lobby.GuestID != nil && *lobby.GuestID == userID:
			return tx.Model(&lobby).Updates(map[string]any{
				"guest_id":    nil,
				"guest_ready": false,
				"status":      "waiting",
				"updated_at":  gorm.Expr("now()"),
			}).Error
		default:
			return repository.ErrLobbyNotParticipant
		}
	})
	if err != nil {
		return nil, err
	}

	return r.GetByCode(ctx, normCode)
}

func (r *LobbyRepository) SetReady(ctx context.Context, code string, userID int64, ready bool) (*models.Lobby, error) {
	normCode := normalizeCode(code)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lobby lobbyModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ?", normCode).
			Take(&lobby).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return repository.ErrNotFound
			}
			return fmt.Errorf("select set ready: %w", err)
		}

		if lobby.Status == "closed" {
			return repository.ErrLobbyClosed
		}

		switch {
		case lobby.HostID == userID:
			lobby.HostReady = ready
		case lobby.GuestID != nil && *lobby.GuestID == userID:
			lobby.GuestReady = ready
		default:
			return repository.ErrLobbyNotParticipant
		}

		lobby.Status = deriveLobbyStatus(lobby.GuestID, lobby.HostReady, lobby.GuestReady)

		if err := tx.Model(&lobby).Updates(map[string]any{
			"host_ready":  lobby.HostReady,
			"guest_ready": lobby.GuestReady,
			"status":      lobby.Status,
			"updated_at":  gorm.Expr("now()"),
		}).Error; err != nil {
			return fmt.Errorf("update ready state: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.GetByCode(ctx, normCode)
}

func (r *LobbyRepository) GetActiveByUser(ctx context.Context, userID int64) (*models.Lobby, error) {
	var row lobbyView
	err := r.baseLobbyQuery(r.db.WithContext(ctx)).
		Where("l.host_id = @userID OR l.guest_id = @userID", map[string]any{"userID": userID}).
		Where("l.status IN ?", []string{"waiting", "full", "ready"}).
		Order("l.updated_at DESC").
		Limit(1).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("get active lobby by user: %w", err)
	}

	lobby := row.toDomain()
	return &lobby, nil
}

func (r *LobbyRepository) Close(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).
		Model(&lobbyModel{}).
		Where("code = ?", normalizeCode(code)).
		Updates(map[string]any{
			"status":      "closed",
			"host_ready":  false,
			"guest_ready": false,
			"updated_at":  gorm.Expr("now()"),
		})
	if result.Error != nil {
		return fmt.Errorf("close lobby: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return repository.ErrNotFound
	}
	return nil
}

type lobbyView struct {
	ID            string    `gorm:"column:id"`
	Code          string    `gorm:"column:code"`
	HostID        int64     `gorm:"column:host_id"`
	HostUsername  string    `gorm:"column:host_username"`
	GuestID       *int64    `gorm:"column:guest_id"`
	GuestUsername *string   `gorm:"column:guest_username"`
	Status        string    `gorm:"column:status"`
	TimeControl   int       `gorm:"column:time_control"`
	HostReady     bool      `gorm:"column:host_ready"`
	GuestReady    bool      `gorm:"column:guest_ready"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (v lobbyView) toDomain() models.Lobby {
	return models.Lobby{
		ID:            v.ID,
		Code:          v.Code,
		HostID:        v.HostID,
		HostUsername:  v.HostUsername,
		GuestID:       v.GuestID,
		GuestUsername: v.GuestUsername,
		Status:        v.Status,
		TimeControl:   v.TimeControl,
		HostReady:     v.HostReady,
		GuestReady:    v.GuestReady,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
	}
}

func (r *LobbyRepository) baseLobbyQuery(db *gorm.DB) *gorm.DB {
	return db.Table("lobbies AS l").
		Select(`
			l.id,
			l.code,
			l.host_id,
			host_user.username AS host_username,
			l.guest_id,
			guest_user.username AS guest_username,
			l.status,
			l.time_control,
			l.host_ready,
			l.guest_ready,
			l.created_at,
			l.updated_at
		`).
		Joins("JOIN users AS host_user ON host_user.id = l.host_id").
		Joins("LEFT JOIN users AS guest_user ON guest_user.id = l.guest_id")
}

func isDuplicateKeyError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}

func deriveLobbyStatus(guestID *int64, hostReady bool, guestReady bool) string {
	if guestID == nil {
		return "waiting"
	}
	if hostReady && guestReady {
		return "ready"
	}
	return "full"
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func generateLobbyCode(length int) (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var builder strings.Builder
	builder.Grow(length)

	for range length {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		builder.WriteByte(alphabet[n.Int64()])
	}

	return builder.String(), nil
}
