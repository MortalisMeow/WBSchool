package consumer

import (
	"WBSchool/Internal/controller/consumer/mocks"
	"testing"
)

func TestNewConsumer(t *testing.T) {
	handler := &mocks.MessageHandlerMock{}
	_, err := NewConsumer(handler, "localhost:9092", "test-topic")
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
}

func TestNewConsumer_multiple_brokers(t *testing.T) {
	handler := &mocks.MessageHandlerMock{}
	_, err := NewConsumer(handler, "host1:9092, host2:9092", "topic")
	if err != nil {
		t.Fatalf("NewConsumer: %v", err)
	}
}

func TestConsumer_Stop_no_panic(t *testing.T) {
	handler := &mocks.MessageHandlerMock{}
	c, err := NewConsumer(handler, "localhost:9092", "test-topic")
	if err != nil {
		t.Fatal(err)
	}
	// Stop без вызова Start — не должен паниковать
	if err := c.Stop(); err != nil {
		t.Logf("Stop returned (expected if kafka not available): %v", err)
	}
}

func TestMessageHandlerMock(t *testing.T) {
	m := &mocks.MessageHandlerMock{
		HandleMessageFromFunc: func(msg []byte) error {
			if len(msg) == 0 {
				return nil
			}
			return nil
		},
	}
	if err := m.HandleMessageFrom([]byte("test")); err != nil {
		t.Fatal(err)
	}
	if len(m.HandleMessageFromCalls) != 1 {
		t.Errorf("HandleMessageFromCalls = %d", len(m.HandleMessageFromCalls))
	}
}
