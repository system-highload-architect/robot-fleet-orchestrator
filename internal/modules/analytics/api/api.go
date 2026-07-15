package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"robot-fleet-orchestrator/internal/modules/analytics/service"
	"robot-fleet-orchestrator/pkg/logger"
)

// Handler предоставляет HTTP-эндпоинты для модуля Analytics.
type Handler struct {
	svc *service.Service
	log *logger.Logger
}

// NewHandler создаёт новый HTTP-обработчик для Analytics.
func NewHandler(svc *service.Service, log *logger.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

// RegisterRoutes регистрирует маршруты на переданном мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/analytics/robot-stats", h.handleRobotStats)
	mux.HandleFunc("/api/v1/analytics/task-stats", h.handleTaskStats)
}

// handleRobotStats возвращает статистику по роботам.
func (h *Handler) handleRobotStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, err := h.svc.GetRobotStats(ctx)
	if err != nil {
		h.log.Error("Failed to get robot stats", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}

// handleTaskStats возвращает статистику по задачам.
func (h *Handler) handleTaskStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, err := h.svc.GetTaskStats(ctx)
	if err != nil {
		h.log.Error("Failed to get task stats", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}
