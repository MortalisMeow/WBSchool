package Internal

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"net/http"
)

type Handler struct {
	db    *Storage
	cache *OrderCache
}

func NewHandler(db *sqlx.DB, cache *OrderCache) *Handler {
	return &Handler{
		db:    NewStorage(db),
		cache: cache,
	}
}

func (h *Handler) HandleMessageFrom(message []byte) error {
	var order Order
	if err := json.Unmarshal(message, &order); err != nil {
		fmt.Sprintf("Failed to unmarshal message: %v", err)
		return err
	}

	if order.OrderUid == "" {
		fmt.Sprintf("Пустое значение order_uid: %v")

	}
	if err := h.db.Create(order); err != nil {
		fmt.Sprintf("Failed to create order: %v", err)
		return err
	}

	if err := h.cache.AddToCache(&order); err != nil {
		fmt.Sprintf("Successfully processed order: %s", &order.OrderUid)
		return nil
	}
	return nil
}

func (h *Handler) GetOrder(c *gin.Context) {
	var order Order
	orderUID := c.Param("order_uid")
	if orderUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Пустое значение id заказа",
		})
		return
	}
	if h.cache.ExistInCache(orderUID) {
		order, err := h.cache.GetFromCache(orderUID)
		if err != nil {
			fmt.Sprintf("Failed to get order from cache: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get order from cache",
			})
		} else {
			c.JSON(http.StatusOK, order)
			return
		}
	}

	order, err := h.db.GetFromDb(orderUID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Заказ не найден",
		})
		return
	}

	if err := h.cache.AddToCache(&order); err != nil {
		fmt.Sprintf("Failed to cache order: %v", err)
	}

	c.JSON(http.StatusOK, order)
}
