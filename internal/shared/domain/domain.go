package domain

import "time"

// RobotStatus — статус робота.
type RobotStatus string

const (
	StatusIdle     RobotStatus = "idle"
	StatusBusy     RobotStatus = "busy"
	StatusOffline  RobotStatus = "offline"
	StatusCharging RobotStatus = "charging"
)

// RobotType — тип робота.
type RobotType string

const (
	TypeDelivery RobotType = "delivery"
	TypeDrone    RobotType = "drone"
	TypeLoader   RobotType = "loader"
)

// Point — координаты на плоскости.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Robot — модель робота.
type Robot struct {
	ID       string      `json:"id"`
	Type     RobotType   `json:"type"`
	Status   RobotStatus `json:"status"`
	Location Point       `json:"location"`
	Battery  float64     `json:"battery"`  // 0..100
	Capacity float64     `json:"capacity"` // грузоподъёмность
}

// TaskStatus — статус задачи.
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskAssigned  TaskStatus = "assigned"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

// Task — модель задачи.
type Task struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`     // delivery, inspection, move, charge
	Priority   int         `json:"priority"` // 1 — высокий, 5 — низкий
	Status     TaskStatus  `json:"status"`
	AssignedTo *string     `json:"assigned_to,omitempty"` // ID робота
	Payload    interface{} `json:"payload"`               // содержимое задачи
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// LocationType — тип локации.
type LocationType string

const (
	LocationWarehouse LocationType = "warehouse"
	LocationCity      LocationType = "city"
	LocationDroneZone LocationType = "drone_zone"
)

// Location — модель локации (зоны работы).
type Location struct {
	ID     string       `json:"id"`
	Name   string       `json:"name"`
	Type   LocationType `json:"type"`
	Bounds struct {     // прямоугольная область, ограничивающая зону
		MinX float64 `json:"min_x"`
		MaxX float64 `json:"max_x"`
		MinY float64 `json:"min_y"`
		MaxY float64 `json:"max_y"`
	} `json:"bounds"`
	AllowedRobotTypes []RobotType `json:"allowed_robot_types"` // какие типы роботов могут работать в этой зоне
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}
