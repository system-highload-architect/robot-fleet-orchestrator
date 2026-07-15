package repository

import (
	"context"
	"sync"
	"time"
)

// Stats представляет агрегированную статистику по роботам и задачам.
type Stats struct {
	Timestamp      time.Time
	TotalRobots    int
	BusyRobots     int
	OfflineRobots  int
	TotalTasks     int
	CompletedTasks int
	FailedTasks    int
}

// Repository определяет интерфейс для хранения и получения данных аналитики.
type Repository interface {
	// SaveStats сохраняет агрегированную статистику.
	SaveStats(ctx context.Context, stats Stats) error
	// GetLatestStats возвращает последнюю сохранённую статистику.
	GetLatestStats(ctx context.Context) (*Stats, error)
	// GetStatsHistory возвращает статистику за последние N записей (опционально).
	GetStatsHistory(ctx context.Context, limit int) ([]Stats, error)
}

// InMemoryRepository — in-memory реализация Repository для демо-целей.
type InMemoryRepository struct {
	mu      sync.RWMutex
	latest  *Stats
	history []Stats
}

// NewInMemoryRepository создаёт новый экземпляр in-memory репозитория.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		history: make([]Stats, 0),
	}
}

// SaveStats сохраняет статистику в памяти.
func (r *InMemoryRepository) SaveStats(ctx context.Context, stats Stats) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Сохраняем как последнюю запись
	r.latest = &stats
	// Добавляем в историю (можно ограничить размер при необходимости)
	r.history = append(r.history, stats)
	if len(r.history) > 1000 { // ограничим историю 1000 записями
		r.history = r.history[1:]
	}
	return nil
}

// GetLatestStats возвращает последнюю сохранённую статистику.
func (r *InMemoryRepository) GetLatestStats(ctx context.Context) (*Stats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.latest == nil {
		return nil, nil // нет данных
	}
	// Возвращаем копию
	stats := *r.latest
	return &stats, nil
}

// GetStatsHistory возвращает историю статистики (последние N записей).
func (r *InMemoryRepository) GetStatsHistory(ctx context.Context, limit int) ([]Stats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.history) == 0 {
		return nil, nil
	}
	start := len(r.history) - limit
	if start < 0 {
		start = 0
	}
	// Возвращаем копию среза
	result := make([]Stats, len(r.history[start:]))
	copy(result, r.history[start:])
	return result, nil
}
