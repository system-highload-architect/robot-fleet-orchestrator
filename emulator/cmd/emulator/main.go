package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"robot-fleet-orchestrator/pkg/logger"
)

// Config для эмулятора (читаем из переменных окружения)
type Config struct {
	OrchestratorURL string
	RobotCount      int
	TaskInterval    time.Duration
}

func loadConfig() *Config {
	url := os.Getenv("ORCHESTRATOR_URL")
	if url == "" {
		url = "http://localhost:8080"
	}
	count := 5
	if v := os.Getenv("ROBOT_COUNT"); v != "" {
		fmt.Sscanf(v, "%d", &count)
	}
	interval := 10 * time.Second
	if v := os.Getenv("TASK_INTERVAL_SEC"); v != "" {
		var sec int
		fmt.Sscanf(v, "%d", &sec)
		interval = time.Duration(sec) * time.Second
	}
	return &Config{
		OrchestratorURL: url,
		RobotCount:      count,
		TaskInterval:    interval,
	}
}

func main() {
	logger := logger.Default()
	cfg := loadConfig()
	logger.Info("Starting emulator",
		"orchestrator_url", cfg.OrchestratorURL,
		"robot_count", cfg.RobotCount,
		"task_interval", cfg.TaskInterval,
	)

	// Регистрируем роботов (создаём их через API?)
	// Для простоты будем считать, что роботы уже зарегистрированы в системе.
	// Мы будем использовать заранее известные ID (например, robot-1, robot-2, ...)
	// или можем создать их через POST к /api/v1/robots (если бы такой эндпоинт был).
	// В текущей реализации robot-manager не имеет создания, только ручное добавление в repo.
	// Поэтому мы просто используем существующих роботов из оркестратора (они создаются при старте).

	// Получаем список роботов из оркестратора
	robots, err := fetchRobots(cfg.OrchestratorURL)
	if err != nil {
		logger.Error("Failed to fetch robots", "error", err)
		// Продолжаем с пустым списком? Или завершаем?
		// Для демо можно сгенерировать фейковые ID
		robots = []string{"robot-1", "robot-2", "robot-3"}
		logger.Warn("Using fallback robot IDs", "ids", robots)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Запускаем эмуляторы для каждого робота
	for _, id := range robots {
		wg.Add(1)
		go runRobotEmulator(ctx, &wg, cfg.OrchestratorURL, id, logger)
	}

	// Запускаем генератор задач (периодически создаём новые задачи)
	wg.Add(1)
	go taskGenerator(ctx, &wg, cfg.OrchestratorURL, cfg.TaskInterval, logger)

	// Ожидаем сигнал завершения
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	logger.Info("Shutting down emulator...")
	cancel()
	wg.Wait()
	logger.Info("Emulator stopped")
}

func fetchRobots(url string) ([]string, error) {
	resp, err := http.Get(url + "/api/v1/robots")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}
	var robots []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&robots); err != nil {
		return nil, err
	}
	ids := make([]string, len(robots))
	for i, r := range robots {
		ids[i] = r.ID
	}
	return ids, nil
}

func runRobotEmulator(ctx context.Context, wg *sync.WaitGroup, baseURL, robotID string, log *logger.Logger) {
	defer wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	statuses := []string{"idle", "busy", "charging", "offline"}

	for {
		select {
		case <-ctx.Done():
			log.Debug("Robot emulator stopped", "robot_id", robotID)
			return
		case <-ticker.C:
			// Случайно меняем статус
			newStatus := statuses[rand.Intn(len(statuses))]
			// Отправляем PUT на /api/v1/robots/{id}/status
			reqBody, _ := json.Marshal(map[string]string{"status": newStatus})
			url := fmt.Sprintf("%s/api/v1/robots/%s/status", baseURL, robotID)
			req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(reqBody))
			if err != nil {
				log.Error("Failed to create request", "error", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Error("Failed to update status", "robot_id", robotID, "error", err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Warn("Unexpected status code", "status", resp.StatusCode)
			} else {
				log.Debug("Updated robot status", "robot_id", robotID, "status", newStatus)
			}
		}
	}
}

func taskGenerator(ctx context.Context, wg *sync.WaitGroup, baseURL string, interval time.Duration, log *logger.Logger) {
	defer wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	taskTypes := []string{"delivery", "inspection", "move", "charge"}
	payloads := []string{"Package A to warehouse", "Check shelf B", "Move to docking station", "Go to charging point"}

	for {
		select {
		case <-ctx.Done():
			log.Info("Task generator stopped")
			return
		case <-ticker.C:
			// Создаём задачу
			taskType := taskTypes[rand.Intn(len(taskTypes))]
			payload := payloads[rand.Intn(len(payloads))]
			priority := rand.Intn(5) + 1

			reqBody, _ := json.Marshal(map[string]interface{}{
				"type":     taskType,
				"priority": priority,
				"payload":  payload,
			})
			url := baseURL + "/api/v1/tasks"
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
			if err != nil {
				log.Error("Failed to create task request", "error", err)
				continue
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Error("Failed to create task", "error", err)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusCreated {
				log.Info("Created task", "type", taskType, "priority", priority)
			} else {
				log.Warn("Failed to create task", "status", resp.StatusCode)
			}
		}
	}
}
