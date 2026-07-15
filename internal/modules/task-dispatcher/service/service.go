package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"robot-fleet-orchestrator/internal/infrastructure/eventbus"
	"robot-fleet-orchestrator/internal/modules/task-dispatcher/repository"
	"robot-fleet-orchestrator/internal/shared/domain"
	"robot-fleet-orchestrator/internal/shared/events"
	"robot-fleet-orchestrator/pkg/logger"
	"robot-fleet-orchestrator/pkg/utils"
)

// ErrTaskNotFound возвращается, когда задача не найдена.
var ErrTaskNotFound = errors.New("task not found")

// Service предоставляет бизнес-логику для управления задачами.
type Service struct {
	repo repository.Repository
	bus  eventbus.EventBus
	log  *logger.Logger
}

// NewService создаёт новый экземпляр сервиса задач.
func NewService(repo repository.Repository, bus eventbus.EventBus, log *logger.Logger) *Service {
	s := &Service{
		repo: repo,
		bus:  bus,
		log:  log,
	}

	// Подписываемся на события изменения статуса роботов (для перераспределения)
	if err := bus.Subscribe("robot.status.changed", s.handleRobotStatusChanged); err != nil {
		log.Error("Failed to subscribe to robot.status.changed", "error", err)
	}

	return s
}

// ListTasks возвращает все задачи.
func (s *Service) ListTasks(ctx context.Context) ([]*domain.Task, error) {
	s.log.Debug("Listing all tasks")
	return s.repo.List(ctx)
}

// GetTask возвращает задачу по ID.
func (s *Service) GetTask(ctx context.Context, id string) (*domain.Task, error) {
	s.log.Debug("Getting task", "id", id)
	if id == "" {
		return nil, fmt.Errorf("task ID cannot be empty")
	}
	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

// CreateTask создаёт новую задачу и пытается назначить её роботу.
func (s *Service) CreateTask(ctx context.Context, taskType string, priority int, payload interface{}) (*domain.Task, error) {
	s.log.Debug("Creating task", "type", taskType, "priority", priority)
	if taskType == "" {
		return nil, fmt.Errorf("task type cannot be empty")
	}
	if priority < 1 || priority > 5 {
		priority = 3
	}

	task := &domain.Task{
		ID:        utils.GenerateID("task"),
		Type:      taskType,
		Priority:  priority,
		Status:    domain.TaskPending,
		Payload:   payload,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Save(ctx, task); err != nil {
		s.log.Error("Failed to save task", "error", err)
		return nil, err
	}

	// Публикуем событие о создании задачи (опционально)
	// Здесь можно опубликовать локальное событие TaskCreated, если нужно.
	// Но основное — назначить задачу.
	go s.assignTask(task.ID) // асинхронно, чтобы не блокировать

	s.log.Info("Task created", "id", task.ID)
	return task, nil
}

// UpdateTaskStatus обновляет статус задачи.
func (s *Service) UpdateTaskStatus(ctx context.Context, id, statusStr string) error {
	s.log.Debug("Updating task status", "id", id, "status", statusStr)
	if id == "" {
		return fmt.Errorf("task ID cannot be empty")
	}
	status := domain.TaskStatus(statusStr)
	if status != domain.TaskPending && status != domain.TaskAssigned &&
		status != domain.TaskCompleted && status != domain.TaskFailed {
		return fmt.Errorf("invalid task status: %s", statusStr)
	}

	task, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}
	oldStatus := task.Status

	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		s.log.Error("Failed to update task status in repo", "id", id, "error", err)
		return err
	}

	// Публикуем событие в зависимости от нового статуса
	switch status {
	case domain.TaskCompleted:
		if task.AssignedTo != nil {
			s.bus.Publish(ctx, "task.completed", events.TaskCompleted{
				TaskID:    id,
				RobotID:   *task.AssignedTo,
				Timestamp: time.Now(),
			})
		}
	case domain.TaskFailed:
		if task.AssignedTo != nil {
			s.bus.Publish(ctx, "task.failed", events.TaskFailed{
				TaskID:    id,
				RobotID:   *task.AssignedTo,
				Reason:    "manual failure",
				Timestamp: time.Now(),
			})
		}
	}

	s.log.Info("Task status updated", "id", id, "old", oldStatus, "new", status)
	return nil
}

// assignTask пытается назначить задачу свободному роботу.
func (s *Service) assignTask(taskID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Получаем задачу
	task, err := s.repo.Get(ctx, taskID)
	if err != nil || task == nil {
		s.log.Error("Failed to get task for assignment", "task_id", taskID, "error", err)
		return
	}
	if task.Status != domain.TaskPending {
		return // уже назначена или завершена
	}

	// В демо-версии всегда назначаем robot-1
	// В реальности здесь будет логика выбора подходящего робота
	robotID := "robot-1"

	if err := s.repo.AssignTask(ctx, taskID, robotID); err != nil {
		s.log.Error("Failed to assign task to robot", "task_id", taskID, "robot_id", robotID, "error", err)
		return
	}

	// Публикуем событие назначения
	s.bus.Publish(ctx, "task.assigned", events.TaskAssigned{
		TaskID:    taskID,
		RobotID:   robotID,
		Timestamp: time.Now(),
	})

	s.log.Info("Task assigned", "task_id", taskID, "robot_id", robotID)
}

// handleRobotStatusChanged обрабатывает изменения статуса роботов.
// Если робот стал idle, можно попытаться назначить ему новую задачу.
func (s *Service) handleRobotStatusChanged(event interface{}) {
	ev, ok := event.(events.RobotStatusChanged)
	if !ok {
		s.log.Warn("Invalid event type for robot.status.changed")
		return
	}
	if ev.NewStatus == string(domain.StatusIdle) {
		// Робот освободился, можно назначить новую задачу
		s.log.Debug("Robot is idle, checking for pending tasks", "robot_id", ev.RobotID)
		// В простой версии просто ищем первую pending задачу и назначаем этому роботу
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		tasks, err := s.repo.List(ctx)
		if err != nil {
			s.log.Error("Failed to list tasks for reassignment", "error", err)
			return
		}
		for _, task := range tasks {
			if task.Status == domain.TaskPending {
				// Назначаем этому роботу
				if err := s.repo.AssignTask(ctx, task.ID, ev.RobotID); err != nil {
					s.log.Error("Failed to assign pending task to robot", "task_id", task.ID, "robot_id", ev.RobotID, "error", err)
					continue
				}
				s.bus.Publish(ctx, "task.assigned", events.TaskAssigned{
					TaskID:    task.ID,
					RobotID:   ev.RobotID,
					Timestamp: time.Now(),
				})
				s.log.Info("Assigned pending task to idle robot", "task_id", task.ID, "robot_id", ev.RobotID)
				break // назначаем только одну задачу для простоты
			}
		}
	}
}
