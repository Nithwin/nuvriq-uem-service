package device

import (
	"net/http"

	"gorm.io/gorm"
)

func RegisterRoutes(mux *http.ServeMux, db *gorm.DB) {
	handler := NewDeviceHandler(db)
	mux.HandleFunc("POST /api/v1/devices", handler.RegisterDevice)
	mux.HandleFunc("POST /api/v1/devices/{id}/sync", handler.SyncDevice)
	mux.HandleFunc("GET /api/v1/devices/inactive", handler.GetInactiveDevices)
	mux.HandleFunc("GET /api/v1/devices/{id}", handler.GetDevice)
	mux.HandleFunc("GET /api/v1/devices", handler.ListDevices)
	mux.HandleFunc("GET /api/v1/fleet/summary", handler.GetFleetSummary)
}
