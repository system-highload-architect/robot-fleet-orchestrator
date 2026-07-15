package service

import (
	"context"
	"sync"
	"time"

	"robot-fleet-orchestrator/internal/infrastructure/eventbus"
	analyticsEvents "robot-fleet-orchestrator/internal/modules/analytics/events"
	"robot-fleet-orchestrator/internal/modules/analytics/repository"
	"robot-fleet-orchestrator/internal/shared/domain"
	"robot-fleet-orchestrator/internal/shared/events"
	"robot-fleet-orchestrator/pkg/logger"
)

// Service — бизнес-логика модуля Analytics.
type Service struct {
	repo       repository.Repository
	bus        eventbus.EventBus
	log        *logger.Logger
	mu         sync.RWMutex
	stats      *repository.Stats
	shutdownCh chan struct{}
	wg         sync.WaitGroup
}

// NewService создаёт новый экземпляр сервиса аналитики.
func NewService(repo repository.Repository, bus eventbus.EventBus, log *logger.Logger) *Service {
	s := &Service{
		repo:       repo,
		bus:        bus,
		log:        log,
		stats:      &repository.Stats{},
		shutdownCh: make(chan struct{}),
	}

	// Подписываемся на события
	s.subscribe()

	// Запускаем фоновую задачу периодического сохранения
	s.wg.Add(1)
	go s.persistLoop()

	return s
}

// subscribe подписывается на события, необходимые для аналитики.
func (s *Service) subscribe() {
	// Подписываемся на изменения статуса роботов
	if err := s.bus.Subscribe("robot.status.changed", s.handleRobotStatusChanged); err != nil {
		s.log.Error("Failed to subscribe to robot.status.changed", "error", err)
	}

	// Подписываемся на завершение задач
	if err := s.bus.Subscribe("task.completed", s.handleTaskCompleted); err != nil {
		s.log.Error("Failed to subscribe to task.completed", "error", err)
	}
}

// handleRobotStatusChanged обновляет внутреннюю статистику при изменении статуса робота.
func (s *Service) handleRobotStatusChanged(event interface{}) {
	ev, ok := event.(events.RobotStatusChanged)
	if !ok {
		s.log.Warn("Invalid event type for robot.status.changed")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Обновляем счётчики в зависимости от нового статуса (в реальности нужно знать старый статус,
	// но для демо мы просто инкрементируем общие счётчики)
	// В продакшене нужно хранить карту роботов с их статусами для точного подсчёта.
	s.stats.TotalRobots++
	if ev.NewStatus == string(domain.StatusBusy) {
		s.stats.BusyRobots++
	}
	// Можно добавить логику для offline и idle, но для простоты оставляем так.
	s.stats.Timestamp = time.Now()

	s.log.Debug("Updated robot stats",
		"total", s.stats.TotalRobots,
		"busy", s.stats.BusyRobots,
	)
}

// handleTaskCompleted обновляет внутреннюю статистику при завершении задачи.
func (s *Service) handleTaskCompleted(event interface{}) {
	ev, ok := event.(events.TaskCompleted)
	if !ok {
		s.log.Warn("Invalid event type for task.completed")
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.stats.TotalTasks++
	s.stats.CompletedTasks++
	s.stats.Timestamp = time.Now()

	s.log.Debug("Task completed",
		"task_id", ev.TaskID,
		"robot_id", ev.RobotID,
	)
}

// persistLoop периодически сохраняет агрегированную статистику в репозиторий.
func (s *Service) persistLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdownCh:
			s.log.Info("Persist loop stopped")
			return
		case <-ticker.C:
			s.flushStats()
		}
	}
}

// flushStats сохраняет текущую статистику в репозиторий.
func (s *Service) flushStats() {
	s.mu.RLock()
	statsCopy := *s.stats
	s.mu.RUnlock()

	if statsCopy.TotalRobots == 0 && statsCopy.TotalTasks == 0 {
		return // нет данных для сохранения
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.repo.SaveStats(ctx, statsCopy); err != nil {
		s.log.Error("Failed to save analytics stats", "error", err)
		return
	}

	// Публикуем событие об обновлении статистики
	analyticsEvent := analyticsEvents.StatsUpdated{
		Timestamp:      statsCopy.Timestamp.Unix(),
		TotalRobots:    statsCopy.TotalRobots,
		BusyRobots:     statsCopy.BusyRobots,
		OfflineRobots:  statsCopy.OfflineRobots,
		TotalTasks:     statsCopy.TotalTasks,
		CompletedTasks: statsCopy.CompletedTasks,
		FailedTasks:    statsCopy.FailedTasks,
	}

	if err := s.bus.Publish(ctx, "analytics.stats.updated", analyticsEvent); err != nil {
		s.log.Error("Failed to publish analytics.stats.updated", "error", err)
	} else {
		s.log.Info("Analytics stats published",
			"total_robots", analyticsEvent.TotalRobots,
			"total_tasks", analyticsEvent.TotalTasks,
		)
	}
}

// GetRobotStats возвращает текущую статистику по роботам.
func (s *Service) GetRobotStats(ctx context.Context) (*repository.Stats, error) {
	// Пытаемся получить из репозитория, если там нет — возвращаем из памяти
	stats, err := s.repo.GetLatestStats(ctx)
	if err != nil {
		return nil, err
	}
	if stats != nil {
		return stats, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	copyStats := *s.stats
	return &copyStats, nil
}

// GetTaskStats возвращает текущую статистику по задачам.
func (s *Service) GetTaskStats(ctx context.Context) (*repository.Stats, error) {
	// В данном случае статистика та же, что и по роботам (общая).
	// Можно расширить для разных типов.
	return s.GetRobotStats(ctx)
}

// Shutdown gracefully останавливает сервис.
func (s *Service) Shutdown() {
	s.log.Info("Shutting down analytics service...")
	close(s.shutdownCh)
	s.wg.Wait()
	s.log.Info("Analytics service stopped")
}
