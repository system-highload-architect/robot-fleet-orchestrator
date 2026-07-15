package repository

import (
	"context"
	"sync"

	"robot-fleet-orchestrator/internal/shared/domain"
)

// Repository определяет интерфейс для хранения данных о роботах.
type Repository interface {
	// Save сохраняет или обновляет робота.
	Save(ctx context.Context, robot *domain.Robot) error
	// Get возвращает робота по ID.
	Get(ctx context.Context, id string) (*domain.Robot, error)
	// List возвращает всех роботов.
	List(ctx context.Context) ([]*domain.Robot, error)
	// Delete удаляет робота по ID.
	Delete(ctx context.Context, id string) error
	// UpdateStatus обновляет статус робота.
	UpdateStatus(ctx context.Context, id string, status domain.RobotStatus) error
}

// InMemoryRepository — потокобезопасное in-memory хранилище роботов.
type InMemoryRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.Robot
}

// NewInMemoryRepository создаёт новый экземпляр in-memory репозитория.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		store: make(map[string]*domain.Robot),
	}
}

// Save сохраняет робота в памяти.
func (r *InMemoryRepository) Save(ctx context.Context, robot *domain.Robot) error {
	if robot == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[robot.ID] = robot
	return nil
}

// Get возвращает робота по ID.
func (r *InMemoryRepository) Get(ctx context.Context, id string) (*domain.Robot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	robot, ok := r.store[id]
	if !ok {
		return nil, nil // или можно вернуть ошибку
	}
	return robot, nil
}

// List возвращает всех роботов.
func (r *InMemoryRepository) List(ctx context.Context) ([]*domain.Robot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Robot, 0, len(r.store))
	for _, robot := range r.store {
		result = append(result, robot)
	}
	return result, nil
}

// Delete удаляет робота по ID.
func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.store, id)
	return nil
}

// UpdateStatus обновляет статус робота.
func (r *InMemoryRepository) UpdateStatus(ctx context.Context, id string, status domain.RobotStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	robot, ok := r.store[id]
	if !ok {
		return nil // или ошибка "not found"
	}
	robot.Status = status
	return nil
}
