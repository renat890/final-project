package config

import (
	"log"
	"os"
)

type Config struct {
	TodoPort   string
	TodoDBFile string
}

func New() *Config {
	conf := Config{TodoPort: "7540", TodoDBFile: "./scheduler.db"}
	if todoPort := os.Getenv("TODO_PORT"); todoPort != "" {
		log.Printf("Определена переменная окружения TODO_PORT пользователем со значением %s", todoPort)
		conf.TodoPort = todoPort
	}
	if todoDBFile := os.Getenv("TODO_DBFILE"); todoDBFile != "" {
		log.Printf("Определена переменная окружения TODO_DBFILE пользователем со значением %s", todoDBFile)
		conf.TodoDBFile = todoDBFile
	} 

	return &conf
}