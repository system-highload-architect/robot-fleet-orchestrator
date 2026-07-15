package robot

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"robot-fleet-orchestrator/emulator/internal/adapter"
	"robot-fleet-orchestrator/pkg/logger"
)

// Robot представляет эмулируемого робота.
type Robot struct {
	ID      string
	adapter adapter.Adapter
	log     *logger.Logger
	stopCh  chan struct{}
}

// NewRobot создаёт новый экземпляр робота.
func NewRobot(id string, adapter adapter.Adapter, log *logger.Logger) *Robot {
	return &Robot{
		ID:      id,
		adapter: adapter,
		log:     log,
		stopCh:  make(chan struct{}),
	}
}

// Run запускает цикл эмуляции робота.
// Периодически обновляет статус через адаптер.
func (r *Robot) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	statuses := []string{"idle", "busy", "charging", "offline"}

	for {
		select {
		case <-ctx.Done():
			r.log.Debug("Robot stopped (context)", "robot_id", r.ID)
			return
		case <-r.stopCh:
			r.log.Debug("Robot stopped (stop signal)", "robot_id", r.ID)
			return
		case <-ticker.C:
			// Случайно выбираем новый статус
			newStatus := statuses[rand.Intn(len(statuses))]
			if err := r.adapter.UpdateRobotStatus(ctx, r.ID, newStatus); err != nil {
				r.log.Error("Failed to update status", "robot_id", r.ID, "status", newStatus, "error", err)
			} else {
				r.log.Debug("Updated status", "robot_id", r.ID, "status", newStatus)
			}
		}
	}
}

// Stop останавливает робота.
func (r *Robot) Stop() {
	close(r.stopCh)
}
