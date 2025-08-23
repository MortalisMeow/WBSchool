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
		"enable.auto.commit":              false,
		"enable.auto.offset.store":        false,
		"isolation.level":                 "read_committed",
		"go.application.rebalance.enable": true,
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
	for {
		if c.stop {
			break
		}
		kafkaMsg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			log.Println("Ошибка чтения сообщения", err)
		}
		if kafkaMsg == nil {
			continue
		}

		var lastErr error
		for attempt := 1; attempt <= 3; attempt++ {
			err = c.handler.HandleMessageFrom(kafkaMsg.Value)
			if err == nil {
				break
			}
			lastErr = err
			log.Printf("Попытка %d провалена: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		if err != nil {
			log.Printf("Все попотки провалились: %v", lastErr)
			continue
		}

		if _, err = c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Println("Не удалось сохранить сообщение", err)
			continue
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
