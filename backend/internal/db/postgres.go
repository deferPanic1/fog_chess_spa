package db

import (
	"context"
	"fmt"

	"github.com/Gilf4/fog_chess/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(ctx context.Context, cfg config.DB) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("db.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("sqlDB.PingContext: %w", err)
	}

	return db.WithContext(ctx), nil
}
