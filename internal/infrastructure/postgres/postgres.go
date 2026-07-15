package postgres

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB — обёртка над sql.DB с дополнительными методами.
type DB struct {
	*sql.DB
}

// Config содержит настройки подключения к PostgreSQL.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string // "disable", "require", "verify-full"
}

// NewDB создаёт новое подключение к PostgreSQL и проверяет его.
func NewDB(cfg Config) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)
	if cfg.SSLMode == "" {
		dsn += " sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверка соединения
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Ping проверяет доступность БД.
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// Close закрывает соединение с БД.
func (db *DB) Close() error {
	return db.DB.Close()
}

// WithTransaction выполняет функцию в рамках транзакции.
// Если функция возвращает ошибку, транзакция откатывается, иначе — фиксируется.
func (db *DB) WithTransaction(fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
