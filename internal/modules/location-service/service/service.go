package service

import (
	"context"
	"fmt"
	"time"

	"robot-fleet-orchestrator/internal/infrastructure/eventbus"
	locationEvents "robot-fleet-orchestrator/internal/modules/location-service/events"
	"robot-fleet-orchestrator/internal/modules/location-service/repository"
	"robot-fleet-orchestrator/internal/shared/domain"
	"robot-fleet-orchestrator/pkg/logger"
	"robot-fleet-orchestrator/pkg/utils"
)

// Service предоставляет бизнес-логику для управления локациями.
type Service struct {
	repo repository.Repository
	bus  eventbus.EventBus
	log  *logger.Logger
}

// NewService создаёт новый экземпляр сервиса локаций.
func NewService(repo repository.Repository, bus eventbus.EventBus, log *logger.Logger) *Service {
	s := &Service{
		repo: repo,
		bus:  bus,
		log:  log,
	}
	return s
}

// ListLocations возвращает список всех локаций.
func (s *Service) ListLocations(ctx context.Context) ([]*domain.Location, error) {
	s.log.Debug("Listing all locations")
	return s.repo.List(ctx)
}

// GetLocation возвращает локацию по ID.
func (s *Service) GetLocation(ctx context.Context, id string) (*domain.Location, error) {
	s.log.Debug("Getting location", "id", id)
	if id == "" {
		return nil, fmt.Errorf("location ID cannot be empty")
	}
	loc, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if loc == nil {
		return nil, fmt.Errorf("location %s not found", id)
	}
	return loc, nil
}

// CreateLocation создаёт новую локацию.
func (s *Service) CreateLocation(ctx context.Context, name, locType string) (*domain.Location, error) {
	if name == "" {
		return nil, fmt.Errorf("location name cannot be empty")
	}
	if locType == "" {
		locType = "warehouse" // значение по умолчанию
	}

	// Преобразуем строку в доменный тип
	locationType := domain.LocationType(locType)

	location := &domain.Location{
		ID:        utils.GenerateID("loc"),
		Name:      name,
		Type:      locationType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Save(ctx, location); err != nil {
		s.log.Error("Failed to save location", "error", err)
		return nil, err
	}

	// Публикуем событие о создании (преобразуем тип в строку)
	event := locationEvents.LocationCreated{
		ID:        location.ID,
		Name:      location.Name,
		Type:      string(location.Type),
		Timestamp: time.Now(),
	}
	if err := s.bus.Publish(ctx, "location.created", event); err != nil {
		s.log.Error("Failed to publish location.created", "error", err)
		// Не возвращаем ошибку, т.к. локация уже сохранена
	}

	s.log.Info("Location created", "id", location.ID, "name", location.Name)
	return location, nil
}

// UpdateLocation обновляет существующую локацию.
func (s *Service) UpdateLocation(ctx context.Context, id, name, locType string) (*domain.Location, error) {
	if id == "" {
		return nil, fmt.Errorf("location ID cannot be empty")
	}

	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("location %s not found", id)
	}

	// Обновляем поля
	if name != "" {
		existing.Name = name
	}
	if locType != "" {
		existing.Type = domain.LocationType(locType)
	}
	existing.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, existing); err != nil {
		s.log.Error("Failed to update location", "id", id, "error", err)
		return nil, err
	}

	// Публикуем событие об обновлении
	event := locationEvents.LocationUpdated{
		ID:        existing.ID,
		Name:      existing.Name,
		Type:      string(existing.Type),
		Timestamp: time.Now(),
	}
	if err := s.bus.Publish(ctx, "location.updated", event); err != nil {
		s.log.Error("Failed to publish location.updated", "error", err)
	}

	s.log.Info("Location updated", "id", existing.ID)
	return existing, nil
}

// DeleteLocation удаляет локацию по ID.
func (s *Service) DeleteLocation(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("location ID cannot be empty")
	}

	existing, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("location %s not found", id)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete location", "id", id, "error", err)
		return err
	}

	// Публикуем событие об удалении
	event := locationEvents.LocationDeleted{
		ID:        id,
		Timestamp: time.Now(),
	}
	if err := s.bus.Publish(ctx, "location.deleted", event); err != nil {
		s.log.Error("Failed to publish location.deleted", "error", err)
	}

	s.log.Info("Location deleted", "id", id)
	return nil
}
