package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
)

func main() {
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", order)
	mux.HandleFunc("/order", orderInfo)

	fileServer := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	log.Println("Запуск веб-сервера на http://localhost:8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Ошибка сервера: %v", err)
	}

}
