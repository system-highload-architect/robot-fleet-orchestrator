package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"robot-fleet-orchestrator/internal/modules/location-service/service"
	"robot-fleet-orchestrator/pkg/logger"
)

// Handler предоставляет HTTP-эндпоинты для модуля Location Service.
type Handler struct {
	svc *service.Service
	log *logger.Logger
}

// NewHandler создаёт новый HTTP-обработчик для Location Service.
func NewHandler(svc *service.Service, log *logger.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

// RegisterRoutes регистрирует маршруты на переданном мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/locations", h.handleListLocations)
	mux.HandleFunc("/api/v1/locations/", h.handleGetLocation)
}

// handleListLocations возвращает список всех локаций.
func (h *Handler) handleListLocations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	locations, err := h.svc.ListLocations(ctx)
	if err != nil {
		h.log.Error("Failed to list locations", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(locations); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}

// handleGetLocation возвращает информацию о конкретной локации по ID.
func (h *Handler) handleGetLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Извлекаем ID из пути: /api/v1/locations/{id}
	id := r.URL.Path[len("/api/v1/locations/"):]
	if id == "" {
		http.Error(w, "Location ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	location, err := h.svc.GetLocation(ctx, id)
	if err != nil {
		h.log.Error("Failed to get location", "id", id, "error", err)
		http.Error(w, "Not found or internal error", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(location); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}
