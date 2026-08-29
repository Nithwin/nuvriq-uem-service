package router

import (
	"net/http"

	"gorm.io/gorm"

	"nuvriq-uem-service/internal/health"
)

// RegisterRoutes registers all feature routes on the ServeMux
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB) {
	health.RegisterRoutes(mux, db)
}
