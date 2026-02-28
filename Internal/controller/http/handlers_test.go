package http

import (
	"WBSchool/Internal/domain"
	"WBSchool/Internal/controller/http/mocks"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHandler_HandleMessageFrom_validation_error(t *testing.T) {
	storage := &mocks.OrderStorageMock{}
	cache := &mocks.OrderCacheMock{}
	h := NewHandler(storage, cache)

	// Пустой order_uid в JSON
	body := []byte(`{"orders":{"order_uid":""}}`)
	err := h.HandleMessageFrom(body)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if len(storage.CreateCalls) != 0 {
		t.Error("Create should not be called on validation error")
	}
}

func TestHandler_HandleMessageFrom_deserialize_error(t *testing.T) {
	storage := &mocks.OrderStorageMock{}
	cache := &mocks.OrderCacheMock{}
	h := NewHandler(storage, cache)

	err := h.HandleMessageFrom([]byte("invalid json"))
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestHandler_HandleMessageFrom_success(t *testing.T) {
	storage := &mocks.OrderStorageMock{}
	cache := &mocks.OrderCacheMock{}
	h := NewHandler(storage, cache)

	order := domain.Order{
		Orders:  domain.Orders{OrderUid: "valid-uuid-123"},
		Payment: domain.Payment{Transaction: "tx"},
	}
	body, _ := json.Marshal(order)

	err := h.HandleMessageFrom(body)
	if err != nil {
		t.Fatalf("HandleMessageFrom: %v", err)
	}
	if len(storage.CreateCalls) != 1 {
		t.Errorf("Create calls = %d, want 1", len(storage.CreateCalls))
	}
	if storage.CreateCalls[0].Orders.OrderUid != "valid-uuid-123" {
		t.Errorf("Create order_uid = %s", storage.CreateCalls[0].Orders.OrderUid)
	}
	if len(cache.SetCalls) != 1 {
		t.Errorf("Set calls = %d, want 1", len(cache.SetCalls))
	}
}

func TestHandler_HandleMessageFrom_create_error(t *testing.T) {
	storage := &mocks.OrderStorageMock{
		CreateFunc: func(domain.Order) error { return errors.New("db error") },
	}
	cache := &mocks.OrderCacheMock{}
	h := NewHandler(storage, cache)

	order := domain.Order{Orders: domain.Orders{OrderUid: "valid-uuid"}}
	body, _ := json.Marshal(order)

	err := h.HandleMessageFrom(body)
	if err == nil {
		t.Fatal("expected error from Create")
	}
	if len(cache.SetCalls) != 0 {
		t.Error("Set should not be called when Create fails")
	}
}

func TestHandler_GetOrder_validation_empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storage := &mocks.OrderStorageMock{}
	cache := &mocks.OrderCacheMock{}
	h := NewHandler(storage, cache)

	r := gin.New()
	r.GET("/order/:order_uid", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/order/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_GetOrder_validation_invalid_format(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storage := &mocks.OrderStorageMock{}
	cache := &mocks.OrderCacheMock{}
	h := NewHandler(storage, cache)

	r := gin.New()
	r.GET("/order/:order_uid", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/order/bad%20uid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_GetOrder_from_cache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uid := "550e8400-e29b-41d4-a716-446655440000"
	order := &domain.Order{Orders: domain.Orders{OrderUid: uid}}
	storage := &mocks.OrderStorageMock{}
	cache := &mocks.OrderCacheMock{
		GetFunc: func(uid string) (*domain.Order, error) {
			if uid == "550e8400-e29b-41d4-a716-446655440000" {
				return order, nil
			}
			return nil, errors.New("not found")
		},
	}
	h := NewHandler(storage, cache)

	r := gin.New()
	r.GET("/order/:order_uid", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/order/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if len(storage.GetFromDbCalls) != 0 {
		t.Error("GetFromDb should not be called when cache hits")
	}
}

func TestHandler_GetOrder_from_db(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uid := "550e8400-e29b-41d4-a716-446655440000"
	order := domain.Order{Orders: domain.Orders{OrderUid: uid}}
	storage := &mocks.OrderStorageMock{
		GetFromDbFunc: func(orderUid string) (domain.Order, error) {
			if orderUid == uid {
				return order, nil
			}
			return domain.Order{}, errors.New("not found")
		},
	}
	cache := &mocks.OrderCacheMock{
		GetFunc: func(string) (*domain.Order, error) { return nil, errors.New("miss") },
	}
	h := NewHandler(storage, cache)

	r := gin.New()
	r.GET("/order/:order_uid", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/order/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if len(storage.GetFromDbCalls) != 1 {
		t.Errorf("GetFromDb calls = %d", len(storage.GetFromDbCalls))
	}
}

func TestHandler_GetOrder_not_found(t *testing.T) {
	gin.SetMode(gin.TestMode)
	storage := &mocks.OrderStorageMock{
		GetFromDbFunc: func(string) (domain.Order, error) {
			return domain.Order{}, errors.New("not found")
		},
	}
	cache := &mocks.OrderCacheMock{
		GetFunc: func(string) (*domain.Order, error) { return nil, errors.New("miss") },
	}
	h := NewHandler(storage, cache)

	r := gin.New()
	r.GET("/order/:order_uid", h.GetOrder)

	req := httptest.NewRequest(http.MethodGet, "/order/550e8400-e29b-41d4-a716-446655440000", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestHandler_RestoreCacheFromDB(t *testing.T) {
	storage := &mocks.OrderStorageMock{
		GetAllOrdersFunc: func() ([]domain.Order, error) {
			return []domain.Order{
				{Orders: domain.Orders{OrderUid: "uid-1", DateCreated: time.Now()}},
				{Orders: domain.Orders{OrderUid: "uid-2", DateCreated: time.Now()}},
			}, nil
		},
	}
	cache := &mocks.OrderCacheMock{}
	h := NewHandler(storage, cache)
	h.RestoreCacheFromDB()
	if storage.GetAllOrdersCalls != 1 {
		t.Errorf("GetAllOrders calls = %d", storage.GetAllOrdersCalls)
	}
	if len(cache.SetCalls) != 2 {
		t.Errorf("Set calls = %d", len(cache.SetCalls))
	}
}
