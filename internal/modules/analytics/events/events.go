package events

// StatsUpdated — событие, публикуемое модулем Analytics при обновлении
// агрегированной статистики по роботам и задачам.
type StatsUpdated struct {
	Timestamp      int64 `json:"timestamp"`
	TotalRobots    int   `json:"total_robots"`
	BusyRobots     int   `json:"busy_robots"`
	OfflineRobots  int   `json:"offline_robots"`
	TotalTasks     int   `json:"total_tasks"`
	CompletedTasks int   `json:"completed_tasks"`
	FailedTasks    int   `json:"failed_tasks"`
}
