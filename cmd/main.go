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
	"sync"
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

	err = db.Ping()

	storage := Internal.NewStorage(db)
	if err := storage.RunMigrations(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:9092"
	}
	log.Println("Подключаемся к Kafka:", kafkaBrokers)

	cache := Internal.NewOrderCache(30 * time.Second)
	handler := Internal.NewHandler(db, cache)
	handler.RestoreCacheFromDB()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		c, err := Internal.NewConsumer(handler, "kafka:9092", "orders-topic")
		if err != nil {
			log.Println("Ошибка создания consumer:", err)
		}
		log.Println("Consumer запущен")
		c.Start()
		defer c.Stop()

	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		producer, err := Internal.NewProducer("kafka:9092", "orders-topic")
		if err != nil {
			log.Println("Ошибка создания producer: %v", err)
			return
		}
		defer producer.Close()

		log.Println("Producer запущен")

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		if err := producer.SendSampleOrder("sample_order.json"); err != nil {
			log.Printf("Ошибка отправки тестового заказа: %v", err)
		} else {
			log.Println("Отправка тестового заказа прошла успешно")
		}

		for range ticker.C {
			order := Internal.GenerateRandomOrder()
			if err := producer.SendOrder(order); err != nil {
				log.Println("Ошибка отправки заказа: %v", err)
			} else {
				log.Println("Заказ отправлен из producer______________:", order.OrderUid)
			}
		}
	}()

	r := gin.Default()
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

	go func() {
		log.Println("Запуск веб-сервера на http://localhost:8081")
		if err := r.Run(":8081"); err != nil {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	wg.Wait()
}
