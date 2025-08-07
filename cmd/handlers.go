package main

import (
	"html/template"
	"log"
	"net/http"
)

func order(w http.ResponseWriter, r *http.Request) {
	//w.Write([]byte("BugBerries"))
	ts, err := template.ParseFiles("./ui/html/order.tmpl")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}

}

func orderInfo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Введите ID вашего заказа"))

}
