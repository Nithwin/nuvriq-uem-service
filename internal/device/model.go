package device

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Device represents the GORM model for the devices table in PostgreSQL
type Device struct {
	ID           string     `gorm:"type:uuid;primaryKey" json:"id"`
	SerialNumber string     `gorm:"size:100;uniqueIndex;not null" json:"serial_number"`
	Hostname     string     `gorm:"size:255;not null" json:"hostname"`
	Platform     string     `gorm:"size:50;index;not null" json:"platform"`
	OSVersion    string     `gorm:"size:100;not null" json:"os_version"`
	OwnerEmail   string     `gorm:"size:255;not null" json:"owner_email"`
	Status       string     `gorm:"size:50;index;default:'never_synced'" json:"status"`
	LastSyncedAt *time.Time `gorm:"index" json:"last_synced_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// BeforeCreate hook generates a UUID for the device if not provided
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}

// RegisterDeviceRequest DTO
type RegisterDeviceRequest struct {
	SerialNumber string `json:"serial_number"`
	Hostname     string `json:"hostname"`
	Platform     string `json:"platform"`
	OSVersion    string `json:"os_version"`
	OwnerEmail   string `json:"owner_email"`
}

// Response Wrapper DTO
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
