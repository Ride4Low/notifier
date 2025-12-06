package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ride4Low/contracts/env"
	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/contracts/pkg/rabbitmq"
)

var (
	notifierAddr = env.GetString("NOTIFIER_ADDR", ":8082")
	driverAddr   = env.GetString("DRIVER_SERVICE_ADDR", "driver-service:9092")
	rabbitMQURI  = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("Notifier starting on", notifierAddr)

	cm := NewConnectionManager()

	// RabbitMQ Session
	rmq, err := rabbitmq.NewRabbitMQ(rabbitMQURI)
	if err != nil {
		log.Fatalf("Failed to create connection manager: %v", err)
	}

	publisher := rabbitmq.NewPublisher(rmq)

	eventHandler := NewEventHandler(cm)
	consumer := rabbitmq.NewConsumer(rmq, eventHandler)

	queues := []string{
		// for riders
		events.NotifyDriverNoDriversFoundQueue,
		events.NotifyDriverAssignQueue,

		// for drivers
		events.DriverCmdTripRequestQueue,
		events.NotifyPaymentSessionCreatedQueue,
	}
	for _, queue := range queues {
		if err := consumer.Consume(context.Background(), queue); err != nil {
			log.Fatalf("Failed to consume queue: %v", err)
		}
	}

	ds, err := NewDriverServiceClient(driverAddr)
	if err != nil {
		log.Fatalf("Failed to create driver service client: %v", err)
	}
	defer ds.Close()

	handler := newHandler(cm, ds, publisher)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws/riders", handler.handleRiders)
	mux.HandleFunc("/ws/drivers", handler.handleDrivers)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	server := &http.Server{
		Addr:    notifierAddr,
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
