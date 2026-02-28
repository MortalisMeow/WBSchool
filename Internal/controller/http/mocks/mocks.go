package mocks

import (
	"WBSchool/Internal/domain"
	"sync"
)

// OrderStorageMock — мок для OrderStorage.
type OrderStorageMock struct {
	mu sync.Mutex

	CreateFunc        func(order domain.Order) error
	CreateCalls       []domain.Order
	GetFromDbFunc     func(orderUid string) (domain.Order, error)
	GetFromDbCalls    []string
	GetAllOrdersFunc  func() ([]domain.Order, error)
	GetAllOrdersCalls int
}

func (m *OrderStorageMock) Create(order domain.Order) error {
	m.mu.Lock()
	m.CreateCalls = append(m.CreateCalls, order)
	f := m.CreateFunc
	m.mu.Unlock()
	if f != nil {
		return f(order)
	}
	return nil
}

func (m *OrderStorageMock) GetFromDb(orderUid string) (domain.Order, error) {
	m.mu.Lock()
	m.GetFromDbCalls = append(m.GetFromDbCalls, orderUid)
	f := m.GetFromDbFunc
	m.mu.Unlock()
	if f != nil {
		return f(orderUid)
	}
	return domain.Order{}, nil
}

func (m *OrderStorageMock) GetAllOrders() ([]domain.Order, error) {
	m.mu.Lock()
	m.GetAllOrdersCalls++
	f := m.GetAllOrdersFunc
	m.mu.Unlock()
	if f != nil {
		return f()
	}
	return nil, nil
}

// OrderCacheMock — мок для OrderCache.
type OrderCacheMock struct {
	mu sync.Mutex

	SetFunc  func(order *domain.Order) error
	SetCalls []*domain.Order
	GetFunc  func(uid string) (*domain.Order, error)
	GetCalls []string
}

func (m *OrderCacheMock) Set(order *domain.Order) error {
	m.mu.Lock()
	m.SetCalls = append(m.SetCalls, order)
	f := m.SetFunc
	m.mu.Unlock()
	if f != nil {
		return f(order)
	}
	return nil
}

func (m *OrderCacheMock) Get(uid string) (*domain.Order, error) {
	m.mu.Lock()
	m.GetCalls = append(m.GetCalls, uid)
	f := m.GetFunc
	m.mu.Unlock()
	if f != nil {
		return f(uid)
	}
	return nil, nil
}
