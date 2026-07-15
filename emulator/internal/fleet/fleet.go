package fleet

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"robot-fleet-orchestrator/emulator/internal/adapter"
	"robot-fleet-orchestrator/emulator/internal/robot"
	"robot-fleet-orchestrator/pkg/logger"
)

// Fleet управляет парком эмулируемых роботов.
type Fleet struct {
	robots   []*robot.Robot
	adapter  adapter.Adapter
	log      *logger.Logger
	wg       sync.WaitGroup
	stopChan chan struct{}
}

// NewFleet создаёт новый парк роботов.
func NewFleet(adapter adapter.Adapter, log *logger.Logger) *Fleet {
	return &Fleet{
		adapter:  adapter,
		log:      log,
		stopChan: make(chan struct{}),
	}
}

// AddRobot добавляет робота в парк.
func (f *Fleet) AddRobot(r *robot.Robot) {
	f.robots = append(f.robots, r)
}

// Start запускает всех роботов в парке.
func (f *Fleet) Start(ctx context.Context) {
	f.log.Info("Starting fleet", "robot_count", len(f.robots))
	for _, r := range f.robots {
		f.wg.Add(1)
		go r.Run(ctx, &f.wg)
	}
}

// Stop останавливает всех роботов (посылает сигнал завершения).
func (f *Fleet) Stop() {
	f.log.Info("Stopping fleet...")
	close(f.stopChan)
	f.wg.Wait()
	f.log.Info("Fleet stopped")
}

// SpawnRobots создаёт роботов с указанными ID.
func (f *Fleet) SpawnRobots(ids []string) {
	for _, id := range ids {
		r := robot.NewRobot(id, f.adapter, f.log)
		f.AddRobot(r)
	}
}

// GenerateTasks периодически создаёт задачи через адаптер.
func (f *Fleet) GenerateTasks(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	taskTypes := []string{"delivery", "inspection", "move", "charge"}
	payloads := []interface{}{
		"Package A to warehouse",
		"Check shelf B",
		"Move to docking station",
		"Go to charging point",
	}

	for {
		select {
		case <-ctx.Done():
			f.log.Info("Task generator stopped")
			return
		case <-f.stopChan:
			f.log.Info("Task generator stopped by fleet stop")
			return
		case <-ticker.C:
			taskType := taskTypes[rand.Intn(len(taskTypes))]
			payload := payloads[rand.Intn(len(payloads))]
			priority := rand.Intn(5) + 1

			taskID, err := f.adapter.CreateTask(ctx, taskType, priority, payload)
			if err != nil {
				f.log.Error("Failed to create task", "error", err)
			} else {
				f.log.Info("Created task", "id", taskID, "type", taskType, "priority", priority)
			}
		}
	}
}
