package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"robot-fleet-orchestrator/internal/modules/task-dispatcher/service"
	"robot-fleet-orchestrator/pkg/logger"
)

// Handler предоставляет HTTP-эндпоинты для модуля Task Dispatcher.
type Handler struct {
	svc *service.Service
	log *logger.Logger
}

// NewHandler создаёт новый HTTP-обработчик для Task Dispatcher.
func NewHandler(svc *service.Service, log *logger.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

// RegisterRoutes регистрирует маршруты на переданном мультиплексоре.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/tasks", h.handleTasks)
	mux.HandleFunc("/api/v1/tasks/", h.handleTaskByID)
}

// handleTasks обрабатывает запросы к списку задач.
func (h *Handler) handleTasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleListTasks(w, r)
	case http.MethodPost:
		h.handleCreateTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTaskByID обрабатывает запросы к конкретной задаче.
func (h *Handler) handleTaskByID(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID из пути: /api/v1/tasks/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
	if path == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}
	// Если в пути есть слеш (например, /tasks/{id}/status), берём только ID
	id := strings.Split(path, "/")[0]

	switch r.Method {
	case http.MethodGet:
		h.handleGetTask(w, r, id)
	case http.MethodPut:
		// Если путь заканчивается на /status, обрабатываем обновление статуса
		if strings.HasSuffix(r.URL.Path, "/status") {
			h.handleUpdateTaskStatus(w, r, id)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListTasks возвращает список всех задач.
func (h *Handler) handleListTasks(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tasks, err := h.svc.ListTasks(ctx)
	if err != nil {
		h.log.Error("Failed to list tasks", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}

// handleGetTask возвращает информацию о конкретной задаче.
func (h *Handler) handleGetTask(w http.ResponseWriter, r *http.Request, id string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	task, err := h.svc.GetTask(ctx, id)
	if err != nil {
		h.log.Error("Failed to get task", "id", id, "error", err)
		if err == service.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(task); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}

// handleCreateTask создаёт новую задачу.
func (h *Handler) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type     string      `json:"type"`
		Priority int         `json:"priority"`
		Payload  interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "Task type is required", http.StatusBadRequest)
		return
	}
	if req.Priority < 1 || req.Priority > 5 {
		req.Priority = 3 // значение по умолчанию
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	task, err := h.svc.CreateTask(ctx, req.Type, req.Priority, req.Payload)
	if err != nil {
		h.log.Error("Failed to create task", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(task); err != nil {
		h.log.Error("Failed to encode response", "error", err)
	}
}

// handleUpdateTaskStatus обновляет статус задачи.
func (h *Handler) handleUpdateTaskStatus(w http.ResponseWriter, r *http.Request, id string) {
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

	if err := h.svc.UpdateTaskStatus(ctx, id, req.Status); err != nil {
		h.log.Error("Failed to update task status", "id", id, "status", req.Status, "error", err)
		if err == service.ErrTaskNotFound {
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"updated"}`))
}
