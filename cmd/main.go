package main

import (
	"WBSchool/Internal"
	"fmt"
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
	DbUrl := os.Getenv("DATABASE_URL") //берем данные для подключения к БД
	db, err := sqlx.Connect("postgres", DbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	err = db.Ping() //пингуем бд
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	cache := Internal.NewOrderCache(200 * time.Millisecond)
	handler := Internal.NewHandler(db, cache)

	go func() {
		c, err := Internal.NewConsumer(handler, "localhost:9092", "test")
		if err != nil {
			fmt.Sprintf("Ошибка создания consumer:", err)
		}

		fmt.Sprintf("Consumer запущен")
		c.Start()

	}()
	r := gin.Default() //создание роутера
	r.LoadHTMLFiles(
		"./ui/html/index.html",
		"./ui/html/info.html",
	)
	r.GET("/order", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/order/:order_uid", handler.GetOrder)

	r.Static("/static", "./ui/static")

	log.Println("Запуск веб-сервера на http://localhost:8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}
}
