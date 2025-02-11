package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"battle-arena/pkg/server"
)

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Parse command line flags
	port := flag.String("port", "8080", "port to listen on")
	flag.Parse()

	// Create and start server
	gameServer := server.NewGameServer()
	if err := gameServer.Start(*port); err != nil {
		log.Fatal(err)
	}
}
