package Internal

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
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
		log.Println("Failed to unmarshal message: %v", err)
		return err
	}

	if order.Orders.OrderUid == "" {
		log.Println("Пустое значение order_uid: %v")

	}
	if err := h.db.Create(order); err != nil {
		log.Println("Failed to create order: %v", err)
		return err
	}

	log.Printf("Order created: %s", order.Orders.OrderUid)

	if err := h.cache.AddToCache(&order); err != nil {
		log.Println("Failed to add order to cache: %s", &order.Orders.OrderUid)
		return nil
	}

	log.Printf("Order added to cache: %s", order.Orders.OrderUid)
	return nil
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderUid := c.Param("order_uid")
	log.Printf("Order struct: %+v", orderUid)
	log.Printf("Does order have OrderUid field? %v", orderUid)
	if orderUid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Пустое значение id заказа",
		})
		log.Println("Empty order_uid")
		return
	}

	// Пробуем получить из кэша
	if h.cache.ExistInCache(orderUid) {
		order, err := h.cache.GetFromCache(orderUid)
		if err != nil {
			log.Printf("Failed to get order from cache: %v", err)

		} else {
			log.Printf("Get order from cache: %s", orderUid)
			c.HTML(http.StatusOK, "info.html", order)
			return
		}
	}

	order, err := h.db.GetFromDb(orderUid)
	if err != nil {
		log.Printf("Error getting order from DB: %v", err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Заказ не найден",
		})
		return
	}

	log.Printf("Get order from db: %s", orderUid)

	if err := h.cache.AddToCache(&order); err != nil {
		log.Printf("Failed to add cache order: %v", err)
	}

	log.Printf("Successfully processed order: %s", order.Orders.OrderUid)
	c.HTML(http.StatusOK, "info.html", order)
}
