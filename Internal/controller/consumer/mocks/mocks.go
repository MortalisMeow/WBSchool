package mocks

import (
	"sync"
)

// MessageHandlerMock — мок для MessageHandler.
type MessageHandlerMock struct {
	mu sync.Mutex

	HandleMessageFromFunc func(message []byte) error
	HandleMessageFromCalls [][]byte
}

func (m *MessageHandlerMock) HandleMessageFrom(message []byte) error {
	m.mu.Lock()
	m.HandleMessageFromCalls = append(m.HandleMessageFromCalls, message)
	f := m.HandleMessageFromFunc
	m.mu.Unlock()
	if f != nil {
		return f(message)
	}
	return nil
}
