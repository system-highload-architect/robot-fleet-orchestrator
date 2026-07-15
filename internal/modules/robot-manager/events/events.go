package events

import "time"

// RobotRegistered публикуется при добавлении нового робота в систему.
type RobotRegistered struct {
	RobotID   string    `json:"robot_id"`
	RobotType string    `json:"robot_type"`
	Timestamp time.Time `json:"timestamp"`
}

// RobotUpdated публикуется при обновлении информации о роботе (кроме статуса).
type RobotUpdated struct {
	RobotID   string      `json:"robot_id"`
	Field     string      `json:"field"` // например, "location", "battery"
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value"`
	Timestamp time.Time   `json:"timestamp"`
}

// RobotDeleted публикуется при удалении робота из системы.
type RobotDeleted struct {
	RobotID   string    `json:"robot_id"`
	Timestamp time.Time `json:"timestamp"`
}
