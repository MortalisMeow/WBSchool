package Internal

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
)

type KafkaConsumer struct {
	consumer *kafka.Consumer
	handler  *Handler
	stop     bool
}

func NewConsumer(handler *Handler, address string, topic string) (*KafkaConsumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        address,
		"group.id":                 "wb-school-group",
		"session.timeout.ms":       6000,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"heartbeat.interval.ms":    2000,
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
			log.Println("Error receiving message", err)
		}
		if kafkaMsg == nil {
			continue
		}

		if err = c.handler.HandleMessageFrom(kafkaMsg.Value); err != nil {
			log.Println("Error handling message", err)
			continue
		}
		if _, err = c.consumer.StoreMessage(kafkaMsg); err != nil {
			log.Println("Error storing message", err)
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
