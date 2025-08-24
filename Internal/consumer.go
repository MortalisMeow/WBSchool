package Internal

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"time"
)

type KafkaConsumer struct {
	consumer *kafka.Consumer
	handler  *Handler
	stop     bool
}

func NewConsumer(handler *Handler, address string, topic string) (*KafkaConsumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":               address,
		"group.id":                        "wb-school-group",
		"session.timeout.ms":              6000,
		"enable.auto.commit":              true,
		"enable.auto.offset.store":        false,
		"isolation.level":                 "read_committed",
		"go.application.rebalance.enable": true,
		"auto.offset.reset":               "earliest",
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, err
	}
	if err = c.Subscribe(topic, nil); err != nil {
		return nil, err
	}
	return &KafkaConsumer{
		consumer: c,
		handler:  handler,
	}, nil

}

func (c *KafkaConsumer) Start() {
	log.Println("Consumer начал работу")

	for {
		if c.stop {
			break
		}
		ev := c.consumer.Poll(100)
		if ev == nil {
			continue
		}

		switch e := ev.(type) {
		case *kafka.Message:
			log.Printf("Cообщение из %s [%d]",
				*e.TopicPartition.Topic, e.TopicPartition.Partition)

			var lastErr error
			for attempt := 1; attempt <= 3; attempt++ {
				err := c.handler.HandleMessageFrom(e.Value)
				if err == nil {
					break
				}
				lastErr = err
				log.Printf("Попытка %d провалена: %v", attempt, err)
				time.Sleep(time.Duration(attempt) * time.Second)
			}

			if lastErr != nil {
				log.Printf("Все попытки провалились: %v", lastErr)
				continue
			}

			if _, err := c.consumer.StoreMessage(e); err != nil {
				log.Println("Не удалось сохранить offset:", err)
			}

		case kafka.AssignedPartitions:
			log.Printf("Назначены новые партиции: %v", e.Partitions)
			c.consumer.Assign(e.Partitions)

		case kafka.RevokedPartitions:
			log.Printf("Отозваны партиции: %v", e.Partitions)
			c.consumer.Unassign()

		case kafka.Error:
			log.Printf("Ошибка Kafka: %v", e)

		default:
			log.Printf("Игнорируем: %v", e)
		}
	}
}

func (c *KafkaConsumer) Stop() error {
	c.stop = true
	if _, err := c.consumer.Commit(); err != nil {
		return err
	}
	return c.consumer.Close()
}
