package device

import (
	"context"
	"math"
	"strings"
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

func (r *Repo) ListDevices(ctx context.Context, filter ListDevicesFilter) ([]Device, int64, error) {
	var devices []Device
	var total int64

	query := r.db.WithContext(ctx).Model(&Device{})

	if strings.TrimSpace(filter.Platform) != "" {
		query = query.Where("platform = ?", strings.ToLower(strings.TrimSpace(filter.Platform)))
	}
	if strings.TrimSpace(filter.Status) != "" {
		query = query.Where("status = ?", strings.ToLower(strings.TrimSpace(filter.Status)))
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err := query.Order("created_at DESC").Offset(offset).Limit(filter.Limit).Find(&devices).Error
	if err != nil {
		return nil, 0, err
	}

	return devices, total, nil
}

func (r *Repo) GetInactiveDevices(ctx context.Context, thresholdSeconds int) ([]Device, InactiveSummary, error) {
	cutoff := time.Now().UTC().Add(-time.Duration(thresholdSeconds) * time.Second)

	var devices []Device
	err := r.db.WithContext(ctx).
		Where("last_synced_at IS NULL OR last_synced_at < ?", cutoff).
		Order("last_synced_at ASC NULLS FIRST").
		Find(&devices).Error
	if err != nil {
		return nil, InactiveSummary{}, err
	}

	var summary InactiveSummary
	summary.TotalInactiveOrUnSynced = len(devices)

	if len(devices) > 0 {
		var ids []string
		for i := range devices {
			if devices[i].LastSyncedAt == nil {
				summary.NeverSyncedCount++
				devices[i].Status = "never_synced"
			} else {
				summary.InactiveCount++
				if devices[i].Status != "inactive" {
					ids = append(ids, devices[i].ID)
					devices[i].Status = "inactive"
				}
			}
		}
		if len(ids) > 0 {
			_ = r.db.WithContext(ctx).Model(&Device{}).Where("id IN ?", ids).Update("status", "inactive").Error
		}
	}

	return devices, summary, nil
}

func (r *Repo) GetFleetSummary(ctx context.Context) (*FleetSummaryData, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&Device{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var active, inactive, neverSynced int64
	_ = r.db.WithContext(ctx).Model(&Device{}).Where("status = ?", "active").Count(&active).Error
	_ = r.db.WithContext(ctx).Model(&Device{}).Where("status = ?", "inactive").Count(&inactive).Error
	_ = r.db.WithContext(ctx).Model(&Device{}).Where("status = ?", "never_synced").Count(&neverSynced).Error

	var windows, macos, linux int64
	_ = r.db.WithContext(ctx).Model(&Device{}).Where("platform = ?", "windows").Count(&windows).Error
	_ = r.db.WithContext(ctx).Model(&Device{}).Where("platform = ?", "macos").Count(&macos).Error
	_ = r.db.WithContext(ctx).Model(&Device{}).Where("platform = ?", "linux").Count(&linux).Error

	healthPercentage := 0.0
	if total > 0 {
		healthPercentage = math.Round((float64(active)/float64(total))*10000.0) / 100.0
	}

	return &FleetSummaryData{
		TotalDevices: total,
		StatusCounts: StatusCounts{
			Active:      active,
			Inactive:    inactive,
			NeverSynced: neverSynced,
		},
		PlatformBreakdown: PlatformBreakdown{
			Windows: windows,
			MacOS:   macos,
			Linux:   linux,
		},
		HealthPercentage: healthPercentage,
	}, nil
}
