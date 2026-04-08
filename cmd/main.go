package main

import (
	"log"

	"github.com/jelsin/http-server/internal/server"
)

func main() {
	if err := server.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
