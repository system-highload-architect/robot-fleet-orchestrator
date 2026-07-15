package eventbus

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryBus_PublishSubscribe(t *testing.T) {
	bus := NewInMemoryBus()

	received := make(chan interface{}, 1)
	err := bus.Subscribe("test.topic", func(event interface{}) {
		received <- event
	})
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	testEvent := "hello world"
	ctx := context.Background()
	if err := bus.Publish(ctx, "test.topic", testEvent); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case ev := <-received:
		if ev != testEvent {
			t.Errorf("Expected %v, got %v", testEvent, ev)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestInMemoryBus_MultipleSubscribers(t *testing.T) {
	bus := NewInMemoryBus()

	count1 := 0
	count2 := 0
	bus.Subscribe("test.topic", func(interface{}) { count1++ })
	bus.Subscribe("test.topic", func(interface{}) { count2++ })

	ctx := context.Background()
	if err := bus.Publish(ctx, "test.topic", "event"); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond) // даём время на обработку
	if count1 != 1 || count2 != 1 {
		t.Errorf("Expected both handlers to be called once, got %d, %d", count1, count2)
	}
}
