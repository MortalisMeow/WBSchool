package Internal

import (
	"fmt"
	"sync"
	"time"
)

type OrderCache struct {
	Cache map[string]*Order
	mu    sync.RWMutex
	ttl   time.Duration
}

func NewOrderCache(ttl time.Duration) *OrderCache {
	return &OrderCache{
		Cache: make(map[string]*Order),
		mu:    sync.RWMutex{},
		ttl:   ttl,
	}
}

func (c *OrderCache) ExistInCache(orderUid string) bool {

	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.Cache[orderUid]
	return ok
}

func (c *OrderCache) AddToCache(order *Order) error {
	if order == nil {
		return fmt.Errorf("cannot add nil order to cache")
	}

	if order.Orders.OrderUid == "" {
		return fmt.Errorf("order_uid is empty, cannot add to cache")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.Cache[order.Orders.OrderUid] = order
	time.AfterFunc(c.ttl, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.Cache, order.Orders.OrderUid)
	})
	return nil
}

func (c *OrderCache) GetFromCache(orderUid string) (*Order, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	order, exists := c.Cache[orderUid]
	if !exists {
		return nil, fmt.Errorf("order %s not found in cache", orderUid)
	}

	return order, nil
}
