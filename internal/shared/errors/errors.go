package errors

import "errors"

// Общие ошибки для всех модулей.
var (
	// ErrNotFound — ресурс не найден.
	ErrNotFound = errors.New("resource not found")
	// ErrAlreadyExists — ресурс уже существует.
	ErrAlreadyExists = errors.New("resource already exists")
	// ErrInvalidInput — неверные входные данные.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnauthorized — неавторизованный доступ.
	ErrUnauthorized = errors.New("unauthorized")
	// ErrInternal — внутренняя ошибка.
	ErrInternal = errors.New("internal error")
	// ErrTimeout — превышен таймаут.
	ErrTimeout = errors.New("operation timeout")
	// ErrBusy — ресурс занят.
	ErrBusy = errors.New("resource busy")
	// ErrUnavailable — ресурс недоступен.
	ErrUnavailable = errors.New("resource unavailable")
)

// NewError создаёт новую ошибку с сообщением и обёртывает базовую ошибку.
// Позволяет добавлять контекст: fmt.Errorf("robot %s: %w", id, errors.ErrNotFound)
// Использовать напрямую fmt.Errorf с %w.
