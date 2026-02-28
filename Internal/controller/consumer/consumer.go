package consumer

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type MessageHandler interface {
	HandleMessageFrom(message []byte) error
}

type Consumer struct {
	reader *kafka.Reader
	handler MessageHandler
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func NewConsumer(handler MessageHandler, address string, topic string) (*Consumer, error) {
	brokers := strings.Split(address, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  "wb-school-group",
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	ctx, cancel := context.WithCancel(context.Background())
	slog.Info("kafka consumer created and subscribed", "topic", topic)
	return &Consumer{
		reader:  reader,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
		done:    make(chan struct{}),
	}, nil
}

func (c *Consumer) Done() <-chan struct{} { return c.done }

func (c *Consumer) Start() {
	defer close(c.done)
	slog.Info("kafka consumer started")

	for {
		msg, err := c.reader.ReadMessage(c.ctx)
		if err != nil {
			if c.ctx.Err() != nil {
				slog.Info("stopping kafka consumer loop")
				break
			}
			slog.Error("error reading message from kafka", "error", err)
			continue
		}

		slog.Debug("message received", "partition", msg.Partition, "offset", msg.Offset)

		var lastErr error
		for attempt := 1; attempt <= 3; attempt++ {
			err = c.handler.HandleMessageFrom(msg.Value)
			if err == nil {
				slog.Info("message processed successfully", "attempt", attempt)
				break
			}
			lastErr = err
			slog.Warn("failed to handle message", "attempt", attempt, "error", err)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		if err != nil {
			slog.Error("all attempts to process message failed", "error", lastErr)
			continue
		}

		if err = c.reader.CommitMessages(c.ctx, msg); err != nil {
			slog.Error("failed to commit message offset", "error", err)
		}
	}
}

func (c *Consumer) Stop() error {
	c.cancel()
	slog.Info("closing consumer")
	err := c.reader.Close()
	if err != nil {
		slog.Error("error closing kafka consumer", "error", err)
	} else {
		slog.Info("kafka consumer closed gracefully")
	}
	return err
}
