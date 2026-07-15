package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client — обёртка над redis.Client с дополнительными методами.
type Client struct {
	*redis.Client
}

// Config содержит настройки подключения к Redis.
type Config struct {
	Addr     string // host:port, например "localhost:6379"
	Password string
	DB       int // номер базы данных (0..15)
}

// NewClient создаёт новое подключение к Redis и проверяет его.
func NewClient(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Проверка соединения с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{Client: rdb}, nil
}

// Ping проверяет доступность Redis.
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// Close закрывает соединение с Redis.
func (c *Client) Close() error {
	return c.Client.Close()
}

// GetString получает строковое значение по ключу.
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	return c.Client.Get(ctx, key).Result()
}

// SetString устанавливает строковое значение с TTL.
func (c *Client) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return c.Client.Set(ctx, key, value, ttl).Err()
}

// Del удаляет один или несколько ключей.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.Client.Del(ctx, keys...).Err()
}

// Exists проверяет существование ключа.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.Client.Exists(ctx, key).Result()
	return n > 0, err
}

// GetJSON получает значение и десериализует его в target.
// Пример использования: var robot domain.Robot; err := c.GetJSON(ctx, "robot:1", &robot)
func (c *Client) GetJSON(ctx context.Context, key string, target interface{}) error {
	return c.Client.Get(ctx, key).Scan(target)
}

// SetJSON сериализует value в JSON и сохраняет с TTL.
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.Client.Set(ctx, key, value, ttl).Err()
}
