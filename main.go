package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ride4Low/contracts/env"
)

var (
	NotifierAddr = env.GetString("NOTIFIER_ADDR", ":8082")
	DriverAddr   = env.GetString("DRIVER_ADDR", ":9092")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Notifier starting on", NotifierAddr)

	ds, err := NewDriverServiceClient(DriverAddr)
	if err != nil {
		log.Fatalf("Failed to create driver service client: %v", err)
	}

	handler := newHandler(NewConnectionManager(), ds)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws/riders", handler.handleRiders)
	mux.HandleFunc("/ws/drivers", handler.handleDrivers)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	server := &http.Server{
		Addr:    NotifierAddr,
		Handler: mux,
	}

	serverErrors := make(chan error)

	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		log.Printf("Error starting the server: %v", err)

	case sig := <-shutdown:
		log.Printf("Server is shutting down due to %v signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Could not stop the server gracefully: %v", err)
			server.Close()
		}
	}

}
