package main

import (
	"WBSchool/Internal"
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load() //загрузка переменных окружения
	if err != nil {
		log.Println("Ошибка загрузки .env файла:", err)
	}
	DbUrl := os.Getenv("DATABASE_URL") //берем данные для подключения к БД
	db, err := sql.Open("postgres", DbUrl)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	err = db.Ping() //пингуем бд
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	//mux := http.NewServeMux()
	r := mux.NewRouter() //создание роутера
	r.HandleFunc("/", Internal.HomePage).Methods("GET")
	r.HandleFunc("/order/{order_uid}", Internal.GetOrder).Methods("GET")

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fileServer))

	log.Println("Запуск веб-сервера на http://localhost:8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}

}
