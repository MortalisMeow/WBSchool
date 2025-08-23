package main

import (
	"WBSchool/Internal"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("Ошибка загрузки .env файла:", err)
	}
	DbUrl := os.Getenv("DATABASE_URL")
	log.Println("Подключаемся к БД:")
	db, err := sqlx.Connect("postgres", DbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	err = db.Ping() //пингуем бд

	storage := Internal.NewStorage(db)
	if err := storage.RunMigrations(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:9092"
	}
	log.Println("Подключаемся к Kafka:", kafkaBrokers)

	cache := Internal.NewOrderCache(200 * time.Millisecond)
	handler := Internal.NewHandler(db, cache)

	go func() {
		producer, err := Internal.NewProducer("kafka:9092", "orders-topic")
		if err != nil {
			log.Println("Ошибка создания producer: %v", err)
			return
		}
		defer producer.Close()

		log.Println("Producer запущен")

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			order := Internal.GenerateRandomOrder()
			if err := producer.SendOrder(order); err != nil {
				log.Println("Ошибка отправки заказа: %v", err)
			} else {
				log.Println("Заказ отправлен: %s", order.Orders.OrderUid)
			}
		}
	}()

	go func() {
		c, err := Internal.NewConsumer(handler, "kafka:9092", "orders-topic")
		if err != nil {
			log.Println("Ошибка создания consumer:", err)
		}

		log.Println("Consumer запущен")
		c.Start()

	}()
	r := gin.Default() //создание роутера
	log.Println("Создан роутер")
	r.LoadHTMLFiles(
		"./ui/html/index.html",
		"./ui/html/info.html",
	)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.GET("/order", func(c *gin.Context) {
		orderUid := c.Query("order_uid")
		if orderUid == "" {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.Redirect(http.StatusFound, "/order/"+orderUid)
	})

	r.GET("/order/:order_uid", handler.GetOrder)

	r.Static("/static", "./ui/static")

	log.Println("Запуск веб-сервера на http://localhost:8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
