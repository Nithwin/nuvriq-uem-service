package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gorm.io/gorm"
)

func HandleHealth(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		dbStatus := "connected"
		status := "ok"
		statusCode := http.StatusOK

		if db != nil {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				dbStatus = "disconnected"
				status = "degraded"
				statusCode = http.StatusServiceUnavailable
			}
		} else {
			dbStatus = "disconnected"
			status = "degraded"
			statusCode = http.StatusServiceUnavailable
		}

		response := HealthResponse{
			Status:    status,
			Database:  dbStatus,
			Timestamp: time.Now().UTC(),
		}

		w.WriteHeader(statusCode)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			fmt.Println("Error encoding health response:", err)
		}
	}
}
