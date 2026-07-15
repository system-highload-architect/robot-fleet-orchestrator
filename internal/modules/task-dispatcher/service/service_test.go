package service

import (
	"context"
	"testing"

	"robot-fleet-orchestrator/internal/infrastructure/eventbus"
	"robot-fleet-orchestrator/internal/shared/domain"
	"robot-fleet-orchestrator/pkg/logger"
)

// mockTaskRepo — in-memory реализация репозитория для тестов.
type mockTaskRepo struct {
	tasks map[string]*domain.Task
}

func (m *mockTaskRepo) Save(ctx context.Context, task *domain.Task) error {
	m.tasks[task.ID] = task
	return nil
}
func (m *mockTaskRepo) Get(ctx context.Context, id string) (*domain.Task, error) {
	return m.tasks[id], nil
}
func (m *mockTaskRepo) List(ctx context.Context) ([]*domain.Task, error) {
	var list []*domain.Task
	for _, t := range m.tasks {
		list = append(list, t)
	}
	return list, nil
}
func (m *mockTaskRepo) Delete(ctx context.Context, id string) error {
	delete(m.tasks, id)
	return nil
}
func (m *mockTaskRepo) UpdateStatus(ctx context.Context, id string, status domain.TaskStatus) error {
	if t, ok := m.tasks[id]; ok {
		t.Status = status
	}
	return nil
}
func (m *mockTaskRepo) AssignTask(ctx context.Context, taskID, robotID string) error {
	if t, ok := m.tasks[taskID]; ok {
		t.Status = domain.TaskAssigned
		t.AssignedTo = &robotID
	}
	return nil
}

func TestTaskDispatcher_CreateTask(t *testing.T) {
	repo := &mockTaskRepo{tasks: make(map[string]*domain.Task)}
	bus := eventbus.NewInMemoryBus()
	log := logger.Default()

	svc := NewService(repo, bus, log)

	task, err := svc.CreateTask(context.Background(), "delivery", 3, "test payload")
	if err != nil {
		t.Fatalf("CreateTask failed: %v", err)
	}
	if task.Status != domain.TaskPending {
		t.Errorf("Expected status pending, got %s", task.Status)
	}

	// Проверяем, что задача сохранилась в репозитории
	saved, _ := repo.Get(context.Background(), task.ID)
	if saved == nil {
		t.Error("Task not saved in repository")
	}
}
