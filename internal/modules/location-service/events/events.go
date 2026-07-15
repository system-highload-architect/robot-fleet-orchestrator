package events

import "time"

// LocationCreated публикуется при создании новой локации.
type LocationCreated struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // warehouse, city, drone_zone
	Timestamp time.Time `json:"timestamp"`
}

// LocationUpdated публикуется при обновлении информации о локации.
type LocationUpdated struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
}

// LocationDeleted публикуется при удалении локации.
type LocationDeleted struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}
