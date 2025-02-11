package main

import (
	"flag"
	"log"

	"battle-arena/pkg/client"

	"github.com/faiface/pixel/pixelgl"
)

func run() {
	// Parse command line flags
	serverAddr := flag.String("server", "localhost:8080", "server address")
	flag.Parse()

	// Create and start client
	gameClient, err := client.NewGameClient(*serverAddr)
	if err != nil {
		log.Fatal(err)
	}

	if err := gameClient.Start(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	pixelgl.Run(run)
}
