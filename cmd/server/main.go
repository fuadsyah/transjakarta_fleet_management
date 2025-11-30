package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/api"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/config"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/database"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/geofence"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/handlers"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/models"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/mqtt"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/rabbitmq"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/repository"
)

func main() {
	log.Println("Starting Fleet Management Backend...")

	// Load configuration
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	if err := database.InitSchema(db); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Create repository
	vehicleRepo := repository.NewVehicleRepository(db)

	// Create RabbitMQ publisher
	rabbitPublisher, err := rabbitmq.NewPublisher(cfg)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ publisher: %v", err)
	}
	if err := rabbitPublisher.Connect(); err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitPublisher.Close()

	// Create geofence checker
	geofenceChecker := geofence.NewChecker(cfg)

	// Create MQTT subscriber with location handler
	mqttSubscriber, err := mqtt.NewSubscriber(cfg, func(loc *models.VehicleLocation) {
		// Save location to database
		if err := vehicleRepo.SaveLocation(loc); err != nil {
			log.Printf("Failed to save location: %v", err)
			return
		}
		log.Printf("Saved location for vehicle %s: lat=%f, lon=%f", loc.VehicleID, loc.Latitude, loc.Longitude)

		// Check geofence
		if geofenceChecker.IsInsideGeofence(loc) {
			log.Printf("Vehicle %s entered geofence!", loc.VehicleID)

			event := &models.GeofenceEvent{
				VehicleID: loc.VehicleID,
				Event:     "geofence_entry",
				Location: models.Location{
					Latitude:  loc.Latitude,
					Longitude: loc.Longitude,
				},
				Timestamp: loc.Timestamp,
			}

			if err := rabbitPublisher.PublishGeofenceEvent(event); err != nil {
				log.Printf("Failed to publish geofence event: %v", err)
			}
		}
	})
	if err != nil {
		log.Fatalf("Failed to create MQTT subscriber: %v", err)
	}

	// Connect to MQTT broker
	if err := mqttSubscriber.Connect(); err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}
	defer mqttSubscriber.Disconnect()

	// Subscribe to vehicle location topic
	if err := mqttSubscriber.Subscribe(); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// Setup API handler
	vehicleHandler := handlers.NewVehicleHandler(vehicleRepo)
	app := api.SetupRouter(vehicleHandler)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	go func() {
		if err := app.Listen(":" + cfg.HTTPPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server running on port %s", cfg.HTTPPort)

	<-quit
	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
}
