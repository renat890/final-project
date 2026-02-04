package config

import (
	"log"
	"os"
)

type Config struct {
	TodoPort     string
	TodoDBFile   string
	TodoPass     string
}

func New() *Config {
	conf := Config{TodoPort: "7540", TodoDBFile: "./scheduler.db", TodoPass: ""}
	if todoPort := os.Getenv("TODO_PORT"); todoPort != "" {
		log.Printf("Определена переменная окружения TODO_PORT пользователем со значением %s", todoPort)
		conf.TodoPort = todoPort
	}
	if todoDBFile := os.Getenv("TODO_DBFILE"); todoDBFile != "" {
		log.Printf("Определена переменная окружения TODO_DBFILE пользователем со значением %s", todoDBFile)
		conf.TodoDBFile = todoDBFile
	}
	if todoPass := os.Getenv("TODO_PASSWORD"); todoPass != "" {
		log.Printf("Определена переменная окружения TODO_PASSWORD пользователем со значением %s", todoPass)
		conf.TodoPass = todoPass
	} else {
		log.Println("В системе не задан пароль, поэтому эндпоинты не требуют авторизации")
	}	

	return &conf
}