package main

import (
	"fmt"
	"log"
	"net/http"
	"tracker/internal/api"
	_ "tracker/internal/api"
	"tracker/internal/config"
	"tracker/internal/db"
)

func main() {
	webDir := "./web"
	conf := config.New()

	if err := db.Init(conf.TodoDBFile); err != nil {
		log.Fatalf("ошибка инициализировать БД: %v", err)
	}

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	api.Init()

	if err := http.ListenAndServe(fmt.Sprintf(":%s", conf.TodoPort), http.DefaultServeMux); err != nil {
		log.Fatal(err)
	}
}