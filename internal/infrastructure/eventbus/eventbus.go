package eventbus

import "context"

// HandlerFunc определяет функцию-обработчик события.
type HandlerFunc func(event interface{})

// EventBus — интерфейс шины событий.
type EventBus interface {
	// Publish публикует событие в топик.
	// context используется для отмены и таймаутов.
	Publish(ctx context.Context, topic string, event interface{}) error

	// Subscribe подписывает обработчик на топик.
	// Возвращает ошибку, если подписка невозможна.
	Subscribe(topic string, handler HandlerFunc) error

	// Close закрывает шину, освобождая ресурсы.
	Close() error
}
