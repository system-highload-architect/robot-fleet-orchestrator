package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"robot-fleet-orchestrator/internal/modules/robot-manager/service"
	"robot-fleet-orchestrator/pkg/logger"
)

// Handler предоставляет HTTP-эндпоинты для модуля Robot Manager.
type Handler struct {
	svc *service.Service
	log *logger.Logger
}

// NewHandler создаёт новый HTTP-обработчик для Robot Manager.
func NewHandler(svc *service.Service, log *logger.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

// RegisterRoutes регистрирует маршруты на переданном мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/robots", h.handleRobots)
	mux.HandleFunc("/api/v1/robots/", h.handleRobotByID)
}

// handleRobots обрабатывает запросы к списку роботов.
func (h *Handler) handleRobots(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleListRobots(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRobotByID обрабатывает запросы к конкретному роботу.
func (h *Handler) handleRobotByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/robots/")
	if path == "" {
		http.Error(w, "Robot ID is required", http.StatusBadRequest)
		return
	}
	id := strings.Split(path, "/")[0]

	switch r.Method {
	case http.MethodGet:
		h.handleGetRobot(w, r, id)
	case http.MethodPut:
		if strings.HasSuffix(r.URL.Path, "/status") {
			h.handleUpdateRobotStatus(w, r, id)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListRobots возвращает список всех роботов.
func (h *Handler) handleListRobots(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	robots, err := h.svc.ListRobots(ctx)
	if err != nil {
		h.log.Error("Failed to list robots", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(robots); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}

// handleGetRobot возвращает информацию о конкретном роботе.
func (h *Handler) handleGetRobot(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	robot, err := h.svc.GetRobot(ctx, id)
	if err != nil {
		h.log.Error("Failed to get robot", "id", id, "error", err)
		if err == service.ErrRobotNotFound {
			http.Error(w, "Robot not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(robot); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}

// handleUpdateRobotStatus обновляет статус робота.
func (h *Handler) handleUpdateRobotStatus(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.svc.UpdateRobotStatus(ctx, id, req.Status); err != nil {
		h.log.Error("Failed to update robot status", "id", id, "status", req.Status, "error", err)
		if err == service.ErrRobotNotFound {
			http.Error(w, "Robot not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"updated"}`))
}
