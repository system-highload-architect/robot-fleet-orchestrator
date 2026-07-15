package repository

import (
	"context"
	"sync"

	"robot-fleet-orchestrator/internal/shared/domain"
)

// Repository определяет интерфейс для хранения данных о задачах.
type Repository interface {
	// Save сохраняет или обновляет задачу.
	Save(ctx context.Context, task *domain.Task) error
	// Get возвращает задачу по ID.
	Get(ctx context.Context, id string) (*domain.Task, error)
	// List возвращает все задачи.
	List(ctx context.Context) ([]*domain.Task, error)
	// Delete удаляет задачу по ID.
	Delete(ctx context.Context, id string) error
	// UpdateStatus обновляет статус задачи.
	UpdateStatus(ctx context.Context, id string, status domain.TaskStatus) error
	// AssignTask назначает задачу роботу.
	AssignTask(ctx context.Context, taskID, robotID string) error
}

// InMemoryRepository — потокобезопасное in-memory хранилище задач.
type InMemoryRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.Task
}

// NewInMemoryRepository создаёт новый экземпляр in-memory репозитория.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		store: make(map[string]*domain.Task),
	}
}

// Save сохраняет задачу в памяти.
func (r *InMemoryRepository) Save(ctx context.Context, task *domain.Task) error {
	if task == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[task.ID] = task
	return nil
}

// Get возвращает задачу по ID.
func (r *InMemoryRepository) Get(ctx context.Context, id string) (*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, ok := r.store[id]
	if !ok {
		return nil, nil
	}
	return task, nil
}

// List возвращает все задачи.
func (r *InMemoryRepository) List(ctx context.Context) ([]*domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*domain.Task, 0, len(r.store))
	for _, task := range r.store {
		result = append(result, task)
	}
	return result, nil
}

// Delete удаляет задачу по ID.
func (r *InMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.store, id)
	return nil
}

// UpdateStatus обновляет статус задачи.
func (r *InMemoryRepository) UpdateStatus(ctx context.Context, id string, status domain.TaskStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.store[id]
	if !ok {
		return nil
	}
	task.Status = status
	return nil
}

// AssignTask назначает задачу роботу.
func (r *InMemoryRepository) AssignTask(ctx context.Context, taskID, robotID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.store[taskID]
	if !ok {
		return nil
	}
	task.Status = domain.TaskAssigned
	task.AssignedTo = &robotID
	return nil
}
