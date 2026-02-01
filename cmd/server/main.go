package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bklimczak/tanks/server"
)

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	flag.Parse()

	log.Println("=================================")
	log.Println("  Tanks RTS Multiplayer Server")
	log.Println("=================================")

	srv := server.New()

	// Channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Channel to listen for server errors
	errChan := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := srv.Start(*addr); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case err := <-errChan:
		log.Fatalf("Server error: %v", err)
	case sig := <-sigChan:
		log.Printf("Received signal: %v", sig)
	}

	// Graceful shutdown with 10 second timeout
	log.Println("Initiating graceful shutdown...")
	if err := srv.GracefulShutdown(10 * time.Second); err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Println("Server stopped gracefully")
}
