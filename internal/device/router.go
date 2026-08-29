package device

import (
	"net/http"

	"gorm.io/gorm"
)

func RegisterRoutes(mux *http.ServeMux, db *gorm.DB) {
	handler := NewDeviceHandler(db)
	mux.HandleFunc("POST /api/v1/devices", handler.RegisterDevice)
}
