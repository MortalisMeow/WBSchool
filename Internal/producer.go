package Internal

import (
	"encoding/json"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"math/rand"
	"os"
	"time"
)

type KafkaProducer struct {
	producer *kafka.Producer
	topic    string
}

func NewProducer(address string, topic string) (*KafkaProducer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": address,
	}

	p, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{
		producer: p,
		topic:    topic,
	}, nil
}

func (p *KafkaProducer) SendOrder(order Order) error {
	jsonData, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("Ошибка сериализации: %w", err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(order.OrderUid),
		Value: jsonData,
	}
	return p.producer.Produce(msg, nil)
}

func (p *KafkaProducer) Close() {
	p.producer.Close()
}

func generateOrderUid() string {
	part1 := fmt.Sprintf("%x", rand.Intn(99999999))
	part2 := fmt.Sprintf("%x", rand.Intn(99999999))
	part3 := fmt.Sprintf("%x", rand.Intn(9999))
	return fmt.Sprintf("%s%s%s", part1, part2, part3)
}

func generateTrackNumber() string {
	return fmt.Sprintf("WBILMTESTTRACK%d", rand.Intn(10000))
}

func GenerateRandomOrder() Order {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	orderUid := generateOrderUid()
	trackNumber := generateTrackNumber()

	var internalSignature string
	if rand.Intn(2) == 0 {
		internalSignature = fmt.Sprintf("sig-%d", rand.Intn(1000))
	}

	var requestID string
	if rand.Intn(2) == 0 {
		requestID = fmt.Sprintf("req-%d", rand.Intn(1000))
	}

	names := []string{"Test Testov", "John Doe", "Jane Smith", "Иван Иванов", "Петр Петров"}
	cities := []string{"Moscow", "Saint Petersburg", "Kazan", "Novosibirsk", "Yekaterinburg"}
	regions := []string{"Moscow Oblast", "Leningrad Oblast", "Tatarstan", "Sverdlovsk Oblast"}
	products := []string{"Mascaras", "Lipstick", "Foundation", "Eyeshadow", "Blush"}
	brands := []string{"Vivienne Sabo", "Maybelline", "L'Oreal", "NYX", "MAC"}

	return Order{
		OrderUid:          orderUid,
		TrackNumber:       trackNumber,
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: internalSignature,
		CustomerID:        fmt.Sprintf("customer-%d", rand.Intn(1000)),
		DeliveryService:   "meest",
		Shardkey:          fmt.Sprintf("%d", rand.Intn(10)),
		SmID:              rand.Intn(100),
		DateCreated:       time.Now(),
		OofShard:          fmt.Sprintf("%d", rand.Intn(10)),

		Payment: Payment{
			Transaction:  orderUid,
			RequestID:    requestID,
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       rand.Intn(10000) + 500,
			PaymentDt:    time.Now().Unix(),
			Bank:         "alpha",
			DeliveryCost: int64(rand.Intn(2000) + 500),
			GoodsTotal:   rand.Intn(5000) + 100,
			CustomFee:    rand.Intn(100),
			OrderUid:     orderUid,
		},
		Delivery: Delivery{
			Name:     names[rand.Intn(len(names))],
			Phone:    fmt.Sprintf("+7%d", 900000000+rand.Intn(100000000)),
			Zip:      fmt.Sprintf("%d", 100000+rand.Intn(900000)),
			City:     cities[rand.Intn(len(cities))],
			Address:  fmt.Sprintf("Street %d, apt %d", rand.Intn(100)+1, rand.Intn(200)+1),
			Region:   regions[rand.Intn(len(regions))],
			Email:    fmt.Sprintf("test%d@gmail.com", rand.Intn(1000)),
			OrderUid: orderUid,
		},
		Items: []Item{
			{
				ChrtID:      int64(rand.Intn(1000000)),
				TrackNumber: trackNumber,
				Price:       int64(rand.Intn(1000) + 100),
				Rid:         fmt.Sprintf("rid-%d", rand.Intn(1000000)),
				Name:        products[rand.Intn(len(products))],
				Sale:        int64(rand.Intn(50)),
				Size:        fmt.Sprintf("%d", rand.Intn(5)),
				TotalPrice:  int64(rand.Intn(5000) + 100),
				NmID:        int64(rand.Intn(1000000)),
				Brand:       brands[rand.Intn(len(brands))],
				Status:      202,
				OrderUid:    orderUid,
			},
		},
	}
}

func LoadSampleOrder(file string) (Order, error) {
	var order Order

	data, err := os.ReadFile(file)
	if err != nil {
		return order, err
	}

	if err := json.Unmarshal(data, &order); err != nil {
		log.Printf("Ошибка парсинга JSON: %v", err)
		return order, err
	}

	return order, nil
}

func (p *KafkaProducer) SendSampleOrder(file string) error {
	order, err := LoadSampleOrder(file)
	if err != nil {
		return err
	}
	log.Printf("OrderUid: %s", order.OrderUid)
	return p.SendOrder(order)
}
