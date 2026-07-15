package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"robot-fleet-orchestrator/pkg/logger"
)

// Adapter определяет интерфейс для взаимодействия с бэкендом (оркестратором).
type Adapter interface {
	// UpdateRobotStatus обновляет статус робота.
	UpdateRobotStatus(ctx context.Context, robotID, status string) error
	// CreateTask создаёт новую задачу.
	CreateTask(ctx context.Context, taskType string, priority int, payload interface{}) (string, error)
	// GetRobots возвращает список ID всех роботов.
	GetRobots(ctx context.Context) ([]string, error)
}

// HTTPAdapter реализует Adapter через HTTP-запросы к оркестратору.
type HTTPAdapter struct {
	baseURL string
	client  *http.Client
	log     *logger.Logger
}

// NewHTTPAdapter создаёт новый HTTP-адаптер.
func NewHTTPAdapter(baseURL string, log *logger.Logger) *HTTPAdapter {
	return &HTTPAdapter{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// UpdateRobotStatus отправляет PUT-запрос на обновление статуса робота.
func (a *HTTPAdapter) UpdateRobotStatus(ctx context.Context, robotID, status string) error {
	url := fmt.Sprintf("%s/api/v1/robots/%s/status", a.baseURL, robotID)
	body, _ := json.Marshal(map[string]string{"status": status})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	return nil
}

// CreateTask отправляет POST-запрос на создание задачи.
func (a *HTTPAdapter) CreateTask(ctx context.Context, taskType string, priority int, payload interface{}) (string, error) {
	url := fmt.Sprintf("%s/api/v1/tasks", a.baseURL)
	reqBody := map[string]interface{}{
		"type":     taskType,
		"priority": priority,
		"payload":  payload,
	}
	body, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return result.ID, nil
}

// GetRobots возвращает список ID роботов.
func (a *HTTPAdapter) GetRobots(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/api/v1/robots", a.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var robots []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&robots); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	ids := make([]string, len(robots))
	for i, r := range robots {
		ids[i] = r.ID
	}
	return ids, nil
}
