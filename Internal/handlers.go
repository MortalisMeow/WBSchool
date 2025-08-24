package Internal

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
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
		log.Println("Ошибка десериализации: %v", err)
		return err
	}

	if order.OrderUid == "" {
		log.Println("Пустое значение order_uid: %v")

	}
	log.Println("Попытка добавить в заказ в ______________БД:", order.OrderUid)
	if err := h.db.Create(order); err != nil {
		log.Println("Не удалось создать заказ: %v", err)
		return err
	}

	log.Printf("Успешно создан заказ_______________________: %s", order.OrderUid)

	if err := h.cache.AddToCache(&order); err != nil {
		log.Println("Не удалось добавить заказ в кэш: %s", &order.OrderUid)
		return nil
	}

	log.Printf("Добавлен в ____________________________ КЭШ: %s", order.OrderUid)
	return nil
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderUid := c.Param("order_uid")
	start := time.Now()
	if orderUid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Пустое значение id заказа",
		})
		log.Println("Пустое значение id заказа")
		return
	}

	var source string
	if h.cache.ExistInCache(orderUid) {
		order, err := h.cache.GetFromCache(orderUid)
		if err != nil {
			log.Printf("Не удалось получить из кэша: %v", err)

		} else {
			source = "КЭША"
			since := time.Since(start)
			log.Printf("ВРЕМЯ ПОЛУЧЕНИЯ ЗАКАЗА %s ИЗ +++++++ %s ++++++++++%v", source, orderUid, since)
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
	source = "БАЗЫ ДАННЫХ"
	since := time.Since(start)
	log.Printf("ВРЕМЯ ПОЛУЧЕНИЯ ЗАКАЗА %s ИЗ +++++++ %s ++++++++++%v", source, orderUid, since)
	if err := h.cache.AddToCache(&order); err != nil {
		log.Printf("Failed to add cache order: %v", err)
	}

	c.HTML(http.StatusOK, "info.html", order)
}

func (h *Handler) RestoreCacheFromDB() {
	orders, err := h.db.GetAllOrders()
	if err != nil {
		log.Printf("Не удалось достать из БД")
		return
	}

	for _, order := range orders {
		if err := h.cache.AddToCache(&order); err != nil {
			log.Printf("Не получилось добавить заказ %s в кэш: %v", order.OrderUid, err)
			continue
		}

	}

	log.Printf("КЭШ восстановлен")
}
