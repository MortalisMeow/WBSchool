package cache

import (
	"WBSchool/Internal/domain"
	"container/list"
	"fmt"
	"log/slog"
	"sync"
)

type OrderCache struct {
	cache map[string]*list.Element
	mu    sync.RWMutex
	list  *list.List
	cap   int
}

type CacheItem struct {
	order *domain.Order
	key   string
}

func NewOrderCache(capacity int) *OrderCache {
	return &OrderCache{
		cache: make(map[string]*list.Element),
		list:  list.New(),
		cap:   capacity,
	}
}

func (c *OrderCache) Set(order *domain.Order) error {
	if order == nil {
		slog.Error("cache: attempt to add nil order")
		return fmt.Errorf("cannot add nil order to cache")
	}

	uid := order.Orders.OrderUid
	if uid == "" {
		slog.Warn("cache: order_uid is empty, skipping")
		return fmt.Errorf("order_uid is empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exist := c.cache[uid]; exist {
		c.list.MoveToFront(elem)
		item := elem.Value.(*CacheItem)
		item.order = order
		slog.Debug("cache: order updated", "uid", uid)
		return nil
	}

	if c.list.Len() >= c.cap {
		oldest := c.list.Back()
		if oldest != nil {
			item := oldest.Value.(*CacheItem)
			delete(c.cache, item.key)
			c.list.Remove(oldest)
			slog.Info("cache: evicted oldest item", "uid", item.key)
		}
	}

	newItem := &CacheItem{
		order: order,
		key:   uid,
	}
	element := c.list.PushFront(newItem)
	c.cache[uid] = element

	slog.Info("cache: order saved", "uid", uid)
	return nil
}

func (c *OrderCache) Get(uid string) (*domain.Order, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.cache[uid]
	if !ok {
		slog.Debug("cache: order not found", "uid", uid)
		return nil, fmt.Errorf("order %s not found", uid)
	}

	c.list.MoveToFront(elem)
	item := elem.Value.(*CacheItem)

	slog.Debug("cache: order hit", "uid", uid)
	return item.order, nil
}

func (c *OrderCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element)
	c.list.Init()
	slog.Info("cache: all data cleared")
}
