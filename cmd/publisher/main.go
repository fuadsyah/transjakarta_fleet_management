package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type VehicleLocation struct {
	VehicleID string  `json:"vehicle_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

func main() {
	broker := getEnv("MQTT_BROKER", "tcp://localhost:1883")

	// Generate random vehicle plates
	// vehicles := GenerateRandomPlates(1)
	vehicles := []string{"B1234XYZ"}

	// Configure MQTT client
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("mqtt-publisher")
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Println("Connected to MQTT broker")
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect: %v", token.Error())
	}

	log.Println("MQTT Publisher started")
	log.Printf("Publishing to broker: %s", broker)

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Base location (Stasiun Bundaran HI)
	baseLat := -6.1938148
	baseLon := 106.8230342

	// Track current position for each vehicle
	vehiclePositions := make(map[string][2]float64)
	for _, v := range vehicles {
		// Initialize each vehicle at a random position near base
		vehiclePositions[v] = [2]float64{
			baseLat + (rand.Float64()-0.5)*0.01,
			baseLon + (rand.Float64()-0.5)*0.01,
		}
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-quit:
			log.Println("Shutting down publisher...")
			client.Disconnect(250)
			return
		case <-ticker.C:
			for _, vehicleID := range vehicles {
				// Get current position
				pos := vehiclePositions[vehicleID]

				// Randomly move the vehicle (simulate movement)
				pos[0] += (rand.Float64() - 0.5) * 0.0001
				pos[1] += (rand.Float64() - 0.5) * 0.0001

				// Occasionally move vehicle close to geofence center
				// (to trigger geofence events)
				if rand.Float64() < 0.1 {
					pos[0] = baseLat + (rand.Float64()-0.5)*0.0003
					pos[1] = baseLon + (rand.Float64()-0.5)*0.0003
				}

				vehiclePositions[vehicleID] = pos

				location := VehicleLocation{
					VehicleID: vehicleID,
					Latitude:  pos[0],
					Longitude: pos[1],
					Timestamp: time.Now().Unix(),
				}

				payload, err := json.Marshal(location)
				if err != nil {
					log.Printf("Failed to marshal location: %v", err)
					continue
				}

				topic := fmt.Sprintf("/fleet/vehicle/%s/location", vehicleID)
				token := client.Publish(topic, 1, false, payload)
				if token.Wait() && token.Error() != nil {
					log.Printf("Failed to publish: %v", token.Error())
				} else {
					log.Printf("Published to %s: lat=%f, lon=%f", topic, location.Latitude, location.Longitude)
				}
			}
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GenerateRandomPlates(n int) []string {
	letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	plates := make([]string, n)

	for i := 0; i < n; i++ {
		// Random number between 1000-2000
		number := 1000 + rand.Intn(1001)

		// 3 random letters
		suffix := make([]byte, 3)
		for j := 0; j < 3; j++ {
			suffix[j] = letters[rand.Intn(len(letters))]
		}

		plates[i] = fmt.Sprintf("B%d%s", number, string(suffix))
	}

	return plates
}
