package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"WBSchool/Internal/adapter/postgres"
	"WBSchool/Internal/controller/consumer"
	httpctrl "WBSchool/Internal/controller/http"
	"WBSchool/cmd/producer"
	"WBSchool/pkg/cache"
)

const (
	serverPort          = ":8081"
	kafkaTopic          = "orders-topic"
	cacheCapacity       = 30
	producerInterval    = 10 * time.Second
	shutdownTimeout     = 15 * time.Second
	consumerWaitTimeout = 5 * time.Second
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go runProducer(ctx, kafkaBrokers, kafkaTopic)
	kafkaConsumer, err := consumer.NewConsumer(handler, kafkaBrokers, kafkaTopic)
	if err != nil {
		slog.Error("create consumer", "error", err)
		os.Exit(1)
	}
	go kafkaConsumer.Start()

	srv := initServer(handler)
	go func() {
		slog.Info("server starting", "addr", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received, stopping gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown", "error", err)
	}
	slog.Info("http server stopped")

	_ = kafkaConsumer.Stop()
	select {
	case <-kafkaConsumer.Done():
		slog.Info("consumer stopped")
	case <-time.After(consumerWaitTimeout):
		slog.Warn("consumer stop timeout")
	}
	slog.Info("graceful shutdown completed")
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

func runProducer(ctx context.Context, brokers, topic string) {
	prod, err := producer.NewProducer(brokers, topic)
	if err != nil {
		slog.Error("create producer", "error", err)
		return
	}
	defer prod.Close()
	slog.Info("producer started")

	ticker := time.NewTicker(producerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("producer stopping")
			return
		case <-ticker.C:
			order := producer.GenerateRandomOrder()
			if err := prod.SendOrder(order); err != nil {
				slog.Error("send order", "error", err)
			} else {
				slog.Info("order sent", "order_uid", order.Orders.OrderUid)
			}
		}
	}
}
