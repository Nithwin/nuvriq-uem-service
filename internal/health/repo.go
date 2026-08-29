package health

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repository interface {
	CheckDB(ctx context.Context) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CheckDB(ctx context.Context) error {
	if r.db == nil {
		return gorm.ErrInvalidDB
	}

	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctxTimeout)
}
