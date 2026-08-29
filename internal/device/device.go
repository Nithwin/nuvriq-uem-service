package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Handler struct {
	repo *Repo
}

func NewDeviceHandler(db *gorm.DB) *Handler {
	repo := NewRepo(db)
	return &Handler{repo: repo}
}

func (h *Handler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "invalid JSON payload",
		})
		return
	}

	// Field Validation
	serial := strings.TrimSpace(req.SerialNumber)
	hostname := strings.TrimSpace(req.Hostname)
	platformRaw := strings.TrimSpace(req.Platform)
	osVersion := strings.TrimSpace(req.OSVersion)
	ownerEmail := strings.TrimSpace(req.OwnerEmail)

	if serial == "" || hostname == "" || platformRaw == "" || osVersion == "" || ownerEmail == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "all fields (serial_number, hostname, platform, os_version, owner_email) are required",
		})
		return
	}

	platform := strings.ToLower(platformRaw)
	if platform != "windows" && platform != "macos" && platform != "linux" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  fmt.Sprintf("invalid platform, must be one of: windows, macos, linux (got '%s')", req.Platform),
		})
		return
	}

	// Duplicate Check in DB
	existing, err := h.repo.GetDeviceBySerialNumber(r.Context(), serial)
	if err == nil && existing != nil {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "device with this serial number already exists",
		})
		return
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "database error checking serial number: " + err.Error(),
		})
		return
	}

	dev := Device{
		SerialNumber: serial,
		Hostname:     hostname,
		Platform:     platform,
		OSVersion:    osVersion,
		OwnerEmail:   ownerEmail,
		Status:       "never_synced",
		LastSyncedAt: nil,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := h.repo.CreateDevice(r.Context(), &dev); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "failed to register device: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Status: "success",
		Data:   dev,
	})
}

func (h *Handler) SyncDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	deviceID := strings.TrimSpace(r.PathValue("id"))
	if deviceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "device ID is required in URL path",
		})
		return
	}

	updatedDev, err := h.repo.UpdateDeviceSync(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(APIResponse{
				Status: "error",
				Error:  fmt.Sprintf("device with ID '%s' not found", deviceID),
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "failed to update device sync: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Status:  "success",
		Message: "Device synchronized successfully",
		Data:    updatedDev,
	})
}
