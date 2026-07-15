package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"robot-fleet-orchestrator/internal/infrastructure/eventbus"
	"robot-fleet-orchestrator/internal/modules/robot-manager/repository"
	"robot-fleet-orchestrator/internal/shared/domain"
	"robot-fleet-orchestrator/internal/shared/events"
	"robot-fleet-orchestrator/pkg/logger"
)

// ErrRobotNotFound возвращается, когда робот не найден.
var ErrRobotNotFound = errors.New("robot not found")

// Service предоставляет бизнес-логику для управления роботами.
type Service struct {
	repo repository.Repository
	bus  eventbus.EventBus
	log  *logger.Logger
}

// NewService создаёт новый экземпляр сервиса роботов.
func NewService(repo repository.Repository, bus eventbus.EventBus, log *logger.Logger) *Service {
	return &Service{
		repo: repo,
		bus:  bus,
		log:  log,
	}
}

// ListRobots возвращает список всех роботов.
func (s *Service) ListRobots(ctx context.Context) ([]*domain.Robot, error) {
	s.log.Debug("Listing all robots")
	return s.repo.List(ctx)
}

// GetRobot возвращает робота по ID.
func (s *Service) GetRobot(ctx context.Context, id string) (*domain.Robot, error) {
	s.log.Debug("Getting robot", "id", id)
	if id == "" {
		return nil, fmt.Errorf("robot ID cannot be empty")
	}
	robot, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if robot == nil {
		return nil, ErrRobotNotFound
	}
	return robot, nil
}

// UpdateRobotStatus обновляет статус робота и публикует событие.
func (s *Service) UpdateRobotStatus(ctx context.Context, id, statusStr string) error {
	s.log.Debug("Updating robot status", "id", id, "status", statusStr)
	if id == "" {
		return fmt.Errorf("robot ID cannot be empty")
	}
	status := domain.RobotStatus(statusStr)
	if status != domain.StatusIdle && status != domain.StatusBusy &&
		status != domain.StatusOffline && status != domain.StatusCharging {
		return fmt.Errorf("invalid status: %s", statusStr)
	}

	robot, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if robot == nil {
		return ErrRobotNotFound
	}
	oldStatus := string(robot.Status)

	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		s.log.Error("Failed to update status in repo", "id", id, "error", err)
		return err
	}

	event := events.RobotStatusChanged{
		RobotID:   id,
		OldStatus: oldStatus,
		NewStatus: string(status),
		Timestamp: time.Now(),
	}
	if err := s.bus.Publish(ctx, "robot.status.changed", event); err != nil {
		s.log.Error("Failed to publish robot.status.changed", "error", err)
		// Не возвращаем ошибку, т.к. статус уже обновлён
	}

	s.log.Info("Robot status updated", "id", id, "old", oldStatus, "new", status)
	return nil
}
