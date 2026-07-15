package events

import "time"

// TaskCreated публикуется при создании новой задачи.
type TaskCreated struct {
	TaskID    string    `json:"task_id"`
	Type      string    `json:"type"`
	Priority  int       `json:"priority"`
	Timestamp time.Time `json:"timestamp"`
}

// TaskUpdated публикуется при изменении задачи (например, статуса, назначения).
type TaskUpdated struct {
	TaskID    string      `json:"task_id"`
	Field     string      `json:"field"`
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value"`
	Timestamp time.Time   `json:"timestamp"`
}
