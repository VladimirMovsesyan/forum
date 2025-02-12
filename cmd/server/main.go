package main

import (
	"github.com/VladimirMovsesyan/forum/internal/application/process"
	"log"
	"os"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dsn := os.Getenv("DSN")

	p := process.New(port, dsn)

	err := p.Run()
	if err != nil {
		log.Println(err)
		return
	}
}
