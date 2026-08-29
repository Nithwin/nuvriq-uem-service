package device

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CreateDevice(ctx context.Context, dev *Device) error {
	return r.db.WithContext(ctx).Create(dev).Error
}

func (r *Repo) GetDeviceBySerialNumber(ctx context.Context, serial string) (*Device, error) {
	var dev Device
	err := r.db.WithContext(ctx).Where("serial_number = ?", serial).First(&dev).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

func (r *Repo) GetDeviceByID(ctx context.Context, id string) (*Device, error) {
	var dev Device
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&dev).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

func (r *Repo) UpdateDeviceSync(ctx context.Context, id string) (*Device, error) {
	now := time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&Device{}).Where("id = ?", id).
		Updates(map[string]any{
			"last_synced_at": now,
			"status":         "active",
			"updated_at":     now,
		})

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return r.GetDeviceByID(ctx, id)
}
