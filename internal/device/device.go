package device

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
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

func (h *Handler) GetDevice(w http.ResponseWriter, r *http.Request) {
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

	dev, err := h.repo.GetDeviceByID(r.Context(), deviceID)
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
			Error:  "failed to retrieve device: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Status: "success",
		Data:   dev,
	})
}

func (h *Handler) ListDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()
	platform := strings.TrimSpace(query.Get("platform"))
	status := strings.TrimSpace(query.Get("status"))

	page := 1
	if pageStr := strings.TrimSpace(query.Get("page")); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := strings.TrimSpace(query.Get("limit")); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100
			}
		}
	}

	filter := ListDevicesFilter{
		Platform: platform,
		Status:   status,
		Page:     page,
		Limit:    limit,
	}

	devices, total, err := h.repo.ListDevices(r.Context(), filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "failed to list devices: " + err.Error(),
		})
		return
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(APIResponse{
		Status: "success",
		Data:   devices,
		Meta: &PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

func (h *Handler) GetInactiveDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	days := 60
	thresholdSeconds := days * 86400

	query := r.URL.Query()
	if daysStr := strings.TrimSpace(query.Get("days")); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
			thresholdSeconds = days * 86400
		}
	}
	if thresholdStr := strings.TrimSpace(query.Get("threshold_seconds")); thresholdStr != "" {
		if t, err := strconv.Atoi(thresholdStr); err == nil && t > 0 {
			thresholdSeconds = t
			days = t / 86400
		}
	}

	devices, summary, err := h.repo.GetInactiveDevices(r.Context(), thresholdSeconds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(APIResponse{
			Status: "error",
			Error:  "failed to fetch inactive devices: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(InactiveDevicesResponse{
		Status:        "success",
		ThresholdDays: days,
		Summary:       summary,
		Data:          devices,
	})
}
