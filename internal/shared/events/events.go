package events

import "time"

// RobotStatusChanged публикуется при изменении статуса робота.
type RobotStatusChanged struct {
	RobotID   string    `json:"robot_id"`
	OldStatus string    `json:"old_status,omitempty"`
	NewStatus string    `json:"new_status"`
	Timestamp time.Time `json:"timestamp"`
}

// === Task Events ===

// TaskAssigned публикуется при назначении задачи роботу.
type TaskAssigned struct {
	TaskID    string    `json:"task_id"`
	RobotID   string    `json:"robot_id"`
	Timestamp time.Time `json:"timestamp"`
}

// TaskCompleted публикуется при успешном завершении задачи.
type TaskCompleted struct {
	TaskID    string    `json:"task_id"`
	RobotID   string    `json:"robot_id"`
	Timestamp time.Time `json:"timestamp"`
}

// TaskFailed публикуется при провале задачи.
type TaskFailed struct {
	TaskID    string    `json:"task_id"`
	RobotID   string    `json:"robot_id"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

// === Location Events ===

// LocationUpdated публикуется при изменении данных локации.
type LocationUpdated struct {
	LocationID string    `json:"location_id"`
	Timestamp  time.Time `json:"timestamp"`
}

// === Analytics Events ===

// StatsUpdated публикуется модулем Analytics при обновлении статистики.
// (Используется analytics/events, но дублируем здесь для общности, если потребуется)
// Вместо дублирования мы импортируем из analytics/events.
// Этот файл содержит только события, используемые несколькими модулями.
