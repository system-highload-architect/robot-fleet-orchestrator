package repository

import (
	"context"
	"sync"

	"robot-fleet-orchestrator/internal/shared/domain"
)

// Repository определяет интерфейс для хранения данных о локациях.
type Repository interface {
	// Save сохраняет или обновляет локацию.
	Save(ctx context.Context, location *domain.Location) error
	// Get возвращает локацию по ID.
	Get(ctx context.Context, id string) (*domain.Location, error)
	// List возвращает все локации.
	List(ctx context.Context) ([]*domain.Location, error)
	// Delete удаляет локацию по ID.
	Delete(ctx context.Context, id string) error
}

// InMemoryRepository — потокобезопасное in-memory хранилище локаций.
type InMemoryRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.Location
}

// NewInMemoryRepository создаёт новый экземпляр in-memory репозитория.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		store: make(map[string]*domain.Location),
	}
}

// Save сохраняет локацию в памяти.
func (r *InMemoryRepository) Save(ctx context.Context, location *domain.Location) error {
	if location == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[location.ID] = location
	return nil
}

// Get возвращает локацию по ID.
func (r *InMemoryRepository) Get(ctx context.Context, id string) (*domain.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	loc, ok := r.store[id]
	if !ok {
		return nil, nil // или можно вернуть ошибку, но для простоты nil
	}
	return loc, nil
}

// List возвращает все локации.
func (r *InMemoryRepository) List(ctx context.Context) ([]*domain.Location, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Location, 0, len(r.store))
	for _, loc := range r.store {
		result = append(result, loc)
	}
	return result, nil
}

// Delete удаляет локацию по ID.
func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.store, id)
	return nil
}
