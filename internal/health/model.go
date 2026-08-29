package health

import "time"

// HealthResponse represents the JSON payload returned by /healthz endpoint
type HealthResponse struct {
	Status    string    `json:"status"`
	Database  string    `json:"database"`
	Timestamp time.Time `json:"timestamp"`
}
