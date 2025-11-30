package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/config"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/models"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/rabbitmq"
)

func main() {
	log.Println("Starting Geofence Worker...")

	// Load configuration
	cfg := config.Load()

	// Create RabbitMQ consumer
	consumer, err := rabbitmq.NewConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	// Connect to RabbitMQ
	if err := consumer.Connect(); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer consumer.Close()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start consuming in a goroutine
	go func() {
		err := consumer.Consume(handleGeofenceEvent)
		if err != nil {
			log.Fatalf("Failed to consume messages: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down worker...")
}

func handleGeofenceEvent(event *models.GeofenceEvent) {
	log.Printf("=== GEOFENCE ALERT ===")
	log.Printf("Vehicle ID: %s", event.VehicleID)
	log.Printf("Event: %s", event.Event)
	log.Printf("Location: lat=%f, lon=%f", event.Location.Latitude, event.Location.Longitude)
	log.Printf("Timestamp: %d", event.Timestamp)
	log.Printf("======================")
}
