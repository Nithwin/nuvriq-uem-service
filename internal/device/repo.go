package device

import (
	"context"

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
