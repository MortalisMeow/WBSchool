package cache

import (
	"WBSchool/Internal/domain"
	"testing"
	"time"
)

func TestNewOrderCache(t *testing.T) {
	c := NewOrderCache(10)
	if c == nil || c.cap != 10 {
		t.Fatal("NewOrderCache(10) failed")
	}
}

func TestOrderCache_Set_Get(t *testing.T) {
	c := NewOrderCache(5)
	uid := "test-uid-1"
	order := &domain.Order{Orders: domain.Orders{OrderUid: uid}}

	if err := c.Set(order); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := c.Get(uid)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Orders.OrderUid != uid {
		t.Errorf("Get: order_uid = %s, want %s", got.Orders.OrderUid, uid)
	}
}

func TestOrderCache_Set_nil(t *testing.T) {
	c := NewOrderCache(5)
	if err := c.Set(nil); err == nil {
		t.Error("Set(nil) expected error")
	}
}

func TestOrderCache_Set_empty_uid(t *testing.T) {
	c := NewOrderCache(5)
	order := &domain.Order{Orders: domain.Orders{OrderUid: ""}}
	if err := c.Set(order); err == nil {
		t.Error("Set(empty uid) expected error")
	}
}

func TestOrderCache_Get_not_found(t *testing.T) {
	c := NewOrderCache(5)
	_, err := c.Get("nonexistent")
	if err == nil {
		t.Error("Get(nonexistent) expected error")
	}
}

func TestOrderCache_eviction(t *testing.T) {
	cap := 3
	c := NewOrderCache(cap)
	for i := 0; i < cap+1; i++ {
		uid := string(rune('a' + i))
		order := &domain.Order{Orders: domain.Orders{OrderUid: uid}}
		_ = c.Set(order)
	}
	// oldest 'a' should be evicted
	_, err := c.Get("a")
	if err == nil {
		t.Error("expected 'a' to be evicted")
	}
	for _, uid := range []string{"b", "c", "d"} {
		if _, err := c.Get(uid); err != nil {
			t.Errorf("Get(%q): %v", uid, err)
		}
	}
}

func TestOrderCache_Update_move_to_front(t *testing.T) {
	c := NewOrderCache(3)
	for _, uid := range []string{"first", "second", "third"} {
		_ = c.Set(&domain.Order{Orders: domain.Orders{OrderUid: uid}})
	}
	// update "first" so it moves to front; then add "fourth" — "second" should be evicted
	_ = c.Set(&domain.Order{Orders: domain.Orders{OrderUid: "first"}})
	_ = c.Set(&domain.Order{Orders: domain.Orders{OrderUid: "fourth"}})
	_, errSecond := c.Get("second")
	_, errOthers := c.Get("first")
	if errSecond == nil {
		t.Error("expected 'second' to be evicted")
	}
	if errOthers != nil {
		t.Errorf("Get(first): %v", errOthers)
	}
}

func TestOrderCache_Clear(t *testing.T) {
	c := NewOrderCache(5)
	_ = c.Set(&domain.Order{Orders: domain.Orders{OrderUid: "x"}})
	c.Clear()
	_, err := c.Get("x")
	if err == nil {
		t.Error("after Clear, Get should fail")
	}
}

// Проверка, что порядок полей не ломает LRU (используем реальную структуру с датой)
func TestOrderCache_Order_with_date(t *testing.T) {
	c := NewOrderCache(2)
	order := &domain.Order{
		Orders:  domain.Orders{OrderUid: "uid-1", DateCreated: time.Now()},
		Payment: domain.Payment{Transaction: "tx"},
	}
	if err := c.Set(order); err != nil {
		t.Fatal(err)
	}
	got, _ := c.Get("uid-1")
	if got.Orders.OrderUid != "uid-1" || got.Transaction != "tx" {
		t.Errorf("order fields: uid=%s tx=%s", got.Orders.OrderUid, got.Transaction)
	}
}
