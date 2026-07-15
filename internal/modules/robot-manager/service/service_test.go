package service

import (
	"context"
	"testing"

	"robot-fleet-orchestrator/internal/infrastructure/eventbus"
	"robot-fleet-orchestrator/internal/shared/domain"
	"robot-fleet-orchestrator/pkg/logger"
)

type mockRepo struct {
	robots map[string]*domain.Robot
}

func (m *mockRepo) Save(ctx context.Context, robot *domain.Robot) error {
	m.robots[robot.ID] = robot
	return nil
}
func (m *mockRepo) Get(ctx context.Context, id string) (*domain.Robot, error) {
	return m.robots[id], nil
}
func (m *mockRepo) List(ctx context.Context) ([]*domain.Robot, error) {
	var list []*domain.Robot
	for _, r := range m.robots {
		list = append(list, r)
	}
	return list, nil
}
func (m *mockRepo) Delete(ctx context.Context, id string) error {
	delete(m.robots, id)
	return nil
}
func (m *mockRepo) UpdateStatus(ctx context.Context, id string, status domain.RobotStatus) error {
	if r, ok := m.robots[id]; ok {
		r.Status = status
	}
	return nil
}

func TestService_UpdateRobotStatus(t *testing.T) {
	repo := &mockRepo{robots: make(map[string]*domain.Robot)}
	bus := eventbus.NewInMemoryBus()
	log := logger.Default()

	svc := NewService(repo, bus, log)

	// Добавляем робота
	robot := &domain.Robot{ID: "robot-1", Status: domain.StatusIdle}
	repo.Save(context.Background(), robot)

	err := svc.UpdateRobotStatus(context.Background(), "robot-1", "busy")
	if err != nil {
		t.Fatalf("UpdateRobotStatus failed: %v", err)
	}

	updated, _ := repo.Get(context.Background(), "robot-1")
	if updated.Status != domain.StatusBusy {
		t.Errorf("Expected status busy, got %s", updated.Status)
	}
}

func TestService_UpdateRobotStatus_NotFound(t *testing.T) {
	repo := &mockRepo{robots: make(map[string]*domain.Robot)}
	bus := eventbus.NewInMemoryBus()
	log := logger.Default()

	svc := NewService(repo, bus, log)

	err := svc.UpdateRobotStatus(context.Background(), "nonexistent", "busy")
	if err != ErrRobotNotFound {
		t.Errorf("Expected ErrRobotNotFound, got %v", err)
	}
}
