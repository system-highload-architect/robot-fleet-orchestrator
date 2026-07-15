package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

// KafkaBus реализует EventBus на основе Apache Kafka.
type KafkaBus struct {
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
	groupID  string
	topics   map[string]struct{} // отслеживаем топики, на которые подписались
	mu       sync.Mutex
	closeCh  chan struct{}
	wg       sync.WaitGroup
}

// NewKafkaBus создаёт новый экземпляр KafkaBus.
// brokers - список адресов брокеров Kafka (например, []string{"localhost:9092"}).
// groupID - идентификатор группы потребителей (для масштабирования).
func NewKafkaBus(brokers []string, groupID string) (*KafkaBus, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		_ = producer.Close()
		return nil, fmt.Errorf("failed to create Kafka consumer group: %w", err)
	}

	return &KafkaBus{
		producer: producer,
		consumer: consumer,
		groupID:  groupID,
		topics:   make(map[string]struct{}),
		closeCh:  make(chan struct{}),
	}, nil
}

// Publish публикует событие в Kafka-топик.
// Событие сериализуется в JSON.
func (k *KafkaBus) Publish(ctx context.Context, topic string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = k.producer.SendMessage(msg)
	return err
}

// Subscribe подписывает обработчик на топик.
// Для каждого топика запускается отдельный потребитель в группе.
func (k *KafkaBus) Subscribe(topic string, handler HandlerFunc) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	// Если топик уже подписан, просто возвращаем nil (можно переопределить, но для простоты не переподписываем)
	if _, ok := k.topics[topic]; ok {
		return nil
	}
	k.topics[topic] = struct{}{}

	// Запускаем горутину, которая будет читать сообщения из топика
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		// Создаём хендлер для группы потребителей
		handlerFunc := func(message *sarama.ConsumerMessage) error {
			var event map[string]interface{}
			if err := json.Unmarshal(message.Value, &event); err != nil {
				log.Printf("[KafkaBus] failed to unmarshal message from topic %s: %v", topic, err)
				return nil // не прерываем цикл
			}
			handler(event)
			return nil
		}

		// Используем ConsumerGroupHandler
		consumerHandler := &consumerGroupHandler{
			topic:   topic,
			handler: handlerFunc,
		}

		// Бесконечный цикл потребления
		for {
			select {
			case <-k.closeCh:
				return
			default:
				if err := k.consumer.Consume(context.Background(), []string{topic}, consumerHandler); err != nil {
					log.Printf("[KafkaBus] consume error on topic %s: %v", topic, err)
				}
			}
		}
	}()

	return nil
}

// Close закрывает продюсер и потребителя, освобождая ресурсы.
func (k *KafkaBus) Close() error {
	close(k.closeCh)
	k.wg.Wait()

	var errs []error
	if err := k.producer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("producer close: %w", err))
	}
	if err := k.consumer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("consumer close: %w", err))
	}
	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}
	return nil
}

// consumerGroupHandler реализует sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	topic   string
	handler func(message *sarama.ConsumerMessage) error
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.handler(msg); err != nil {
			log.Printf("[KafkaBus] handler error on topic %s: %v", h.topic, err)
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
