package http

import (
	"WBSchool/Internal/domain"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"time"
)

type OrderStorage interface {
	Create(order domain.Order) error
	GetFromDb(orderUid string) (domain.Order, error)
	GetAllOrders() ([]domain.Order, error)
}

type OrderCache interface {
	Set(order *domain.Order) error
	Get(uid string) (*domain.Order, error)
}

type Handler struct {
	db    OrderStorage
	cache OrderCache
}

func NewHandler(db OrderStorage, cache OrderCache) *Handler {
	return &Handler{
		db:    db,
		cache: cache,
	}
}

func (h *Handler) HandleMessageFrom(message []byte) error {
	var order domain.Order
	if err := json.Unmarshal(message, &order); err != nil {
		slog.Error("deserialization error", "error", err)
		return err
	}

	if order.Orders.OrderUid == "" {
		slog.Warn("empty order_uid")

	}
	if err := h.db.Create(order); err != nil {
		slog.Error("failed to create order", "error", err)
		return err
	}

	slog.Info("order created successfully", "order_uid", order.Orders.OrderUid)

	if err := h.cache.Set(&order); err != nil {
		slog.Error("failed to add order to cache", "order_uid", order.Orders.OrderUid, "error", err)
		return nil
	}

	slog.Info("order added to cache", "order_uid", order.Orders.OrderUid)
	return nil
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderUid := c.Param("order_uid")
	start := time.Now()
	if orderUid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "empty order id",
		})
		slog.Warn("empty order id")
		return
	}

	var source string
	if cachedOrder, err := h.cache.Get(orderUid); err == nil {
		source = "cache"
		since := time.Since(start)
		slog.Info("order retrieved from cache", "order_uid", orderUid, "source", source, "elapsed", since)
		c.HTML(http.StatusOK, "info.html", cachedOrder)
		return
	} else {
		slog.Debug("order not found in cache", "order_uid", orderUid, "error", err)
	}

	order, err := h.db.GetFromDb(orderUid)
	if err != nil {
		slog.Error("failed to get order from DB", "order_uid", orderUid, "error", err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "order not found",
		})
		return
	}
	source = "database"
	since := time.Since(start)
	slog.Info("order retrieved from database", "order_uid", orderUid, "source", source, "elapsed", since)
	if err := h.cache.Set(&order); err != nil {
		slog.Error("failed to add order to cache", "order_uid", order.Orders.OrderUid, "error", err)
	}

	c.HTML(http.StatusOK, "info.html", order)
}

func (h *Handler) RestoreCacheFromDB() {
	orders, err := h.db.GetAllOrders()
	if err != nil {
		slog.Error("failed to load orders from DB", "error", err)
		return
	}

	for i := range orders {
		order := orders[i]
		if err := h.cache.Set(&order); err != nil {
			slog.Error("failed to add order to cache", "order_uid", order.Orders.OrderUid, "error", err)
			continue
		}

	}

	slog.Info("cache restored")
}
