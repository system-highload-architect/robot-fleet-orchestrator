package eventbus

import (
	"context"
	"sync"
)

// InMemoryBus реализует EventBus в памяти (для монолита).
type InMemoryBus struct {
	subscribers map[string][]HandlerFunc
	mu          sync.RWMutex
}

// NewInMemoryBus создаёт новый экземпляр in-memory шины.
func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		subscribers: make(map[string][]HandlerFunc),
	}
}

// Publish публикует событие в топик.
// Обработчики вызываются асинхронно (в отдельных горутинах).
func (b *InMemoryBus) Publish(ctx context.Context, topic string, event interface{}) error {
	b.mu.RLock()
	handlers := b.subscribers[topic]
	b.mu.RUnlock()

	for _, h := range handlers {
		go h(event)
	}
	return nil
}

// Subscribe подписывает обработчик на топик.
func (b *InMemoryBus) Subscribe(topic string, handler HandlerFunc) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[topic] = append(b.subscribers[topic], handler)
	return nil
}

// Close ничего не делает для in-memory шины.
func (b *InMemoryBus) Close() error {
	return nil
}
