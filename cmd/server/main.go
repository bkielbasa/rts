package main

import (
	"flag"
	"log"

	"github.com/bklimczak/tanks/server"
)

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	flag.Parse()

	log.Println("=================================")
	log.Println("  Tanks RTS Multiplayer Server")
	log.Println("=================================")

	srv := server.New()
	if err := srv.Start(*addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
