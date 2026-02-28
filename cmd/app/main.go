package main

import (
	"WBSchool/Internal/adapter/postgres"
	"WBSchool/Internal/controller/consumer"
	httpctrl "WBSchool/Internal/controller/http"
	"WBSchool/cmd/producer"
	"WBSchool/pkg/cache"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	serverPort       = ":8081"
	kafkaTopic       = "orders-topic"
	cacheCapacity    = 30
	producerInterval = 10 * time.Second
)

func main() {
	_ = godotenv.Load()

	dbURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5433/wb_school?sslmode=disable")
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:29092")

	db, err := initDB(dbURL)
	if err != nil {
		slog.Error("init db", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	storage := postgres.NewStorage(db)
	if err := storage.RunMigrations(); err != nil {
		slog.Error("migrations", "error", err)
		os.Exit(1)
	}

	orderCache := cache.NewOrderCache(cacheCapacity)
	handler := httpctrl.NewHandler(storage, orderCache)
	handler.RestoreCacheFromDB()

	go runProducer(kafkaBrokers, kafkaTopic)
	go runConsumer(kafkaBrokers, kafkaTopic, handler)

	srv := initServer(handler)
	slog.Info("server starting", "addr", serverPort)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server", "error", err)
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func initDB(url string) (*sqlx.DB, error) {
	slog.Info("connecting to database")
	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func initServer(handler *httpctrl.Handler) *http.Server {
	r := gin.Default()
	r.LoadHTMLFiles("./pkg/ui/html/index.html", "./pkg/ui/html/info.html")
	r.GET("/", func(c *gin.Context) { c.HTML(http.StatusOK, "index.html", nil) })
	r.GET("/order", func(c *gin.Context) {
		orderUid := c.Query("order_uid")
		if orderUid == "" {
			c.Redirect(http.StatusFound, "/")
			return
		}
		c.Redirect(http.StatusFound, "/order/"+orderUid)
	})
	r.GET("/order/:order_uid", handler.GetOrder)
	r.Static("/static", "./pkg/ui/static")

	return &http.Server{
		Addr:    serverPort,
		Handler: r,
	}
}

func runProducer(brokers, topic string) {
	prod, err := producer.NewProducer(brokers, topic)
	if err != nil {
		slog.Error("create producer", "error", err)
		return
	}
	defer prod.Close()
	slog.Info("producer started")

	ticker := time.NewTicker(producerInterval)
	defer ticker.Stop()

	for range ticker.C {
		order := producer.GenerateRandomOrder()
		if err := prod.SendOrder(order); err != nil {
			slog.Error("send order", "error", err)
		} else {
			slog.Info("order sent", "order_uid", order.Orders.OrderUid)
		}
	}
}

func runConsumer(brokers, topic string, handler *httpctrl.Handler) {
	c, err := consumer.NewConsumer(handler, brokers, topic)
	if err != nil {
		slog.Error("create consumer", "error", err)
		return
	}
	slog.Info("consumer started")
	c.Start()
}
