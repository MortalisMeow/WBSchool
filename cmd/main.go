package main

import (
	"log"
	"net/http"
)

func title(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("BugBerries"))

}

func order_info(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Введите ID вашего заказа"))

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", title)
	mux.HandleFunc("/order", order_info)

	log.Println("Запуск веб-сервера на http://localhost:8081")
	err := http.ListenAndServe("localhost:8081", mux)
	log.Fatal(err)

}
