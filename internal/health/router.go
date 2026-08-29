package health

import (
	"net/http"

	"gorm.io/gorm"
)

func RegisterRoutes(mux *http.ServeMux, db *gorm.DB) {
	mux.HandleFunc("GET /health", HandleHealth(db))
}
