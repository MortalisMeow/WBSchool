package producer

import (
	"WBSchool/Internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewProducer(address string, topic string) (*KafkaProducer, error) {
	brokers := strings.Split(address, ",")
	for i := range brokers {
		brokers[i] = strings.TrimSpace(brokers[i])
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		RequiredAcks: kafka.RequireOne,
	}

	slog.Info("kafka producer created", "address", address, "topic", topic)
	return &KafkaProducer{
		writer: writer,
		topic:  topic,
	}, nil
}

func (p *KafkaProducer) SendOrder(order domain.Order) error {
	jsonData, err := json.Marshal(order)
	if err != nil {
		slog.Error("failed to marshal order", "order_uid", order.Orders.OrderUid, "error", err)
		return fmt.Errorf("marshalling error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(order.Orders.OrderUid),
		Value: jsonData,
	})
	if err != nil {
		slog.Error("failed to produce message to kafka", "order_uid", order.Orders.OrderUid, "error", err)
		return err
	}

	slog.Info("order sent to kafka", "order_uid", order.Orders.OrderUid, "topic", p.topic)
	return nil
}

func (p *KafkaProducer) Close() {
	slog.Info("closing kafka producer")
	_ = p.writer.Close()
}

func GenerateRandomOrder() domain.Order {
	var order domain.Order

	err := gofakeit.Struct(&order)
	if err != nil {
		slog.Error("failed to generate random order structure", "error", err)
		return domain.Order{}
	}

	uid := order.Orders.OrderUid
	track := order.Orders.TrackNumber

	order.Payment.OrderUid = uid
	order.Delivery.OrderUid = uid
	order.Payment.PaymentDt = order.Orders.DateCreated.Unix()

	for i := range order.Items {
		order.Items[i].OrderUid = uid
		order.Items[i].TrackNumber = track
	}

	slog.Debug("random order generated", "order_uid", uid)
	return order
}
