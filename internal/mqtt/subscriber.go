package mqtt

import (
	"encoding/json"
	"fmt"
	"log"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/config"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/models"
)

// Subscriber handles MQTT subscription for vehicle locations
type Subscriber struct {
	client  pahomqtt.Client
	handler func(*models.VehicleLocation)
	cfg     *config.Config
}

// NewSubscriber creates a new MQTT subscriber
func NewSubscriber(cfg *config.Config, handler func(*models.VehicleLocation)) (*Subscriber, error) {
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(cfg.MQTTClientID)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(func(c pahomqtt.Client) {
		log.Println("Connected to MQTT broker")
	})
	opts.SetConnectionLostHandler(func(c pahomqtt.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	})

	client := pahomqtt.NewClient(opts)

	return &Subscriber{
		client:  client,
		handler: handler,
		cfg:     cfg,
	}, nil
}

// Connect establishes connection to MQTT broker
func (s *Subscriber) Connect() error {
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}
	return nil
}

// Subscribe starts listening to vehicle location topic
func (s *Subscriber) Subscribe() error {
	topic := "/fleet/vehicle/+/location"

	messageHandler := func(client pahomqtt.Client, msg pahomqtt.Message) {
		log.Printf("Received message on topic: %s", msg.Topic())

		var location models.VehicleLocation
		if err := json.Unmarshal(msg.Payload(), &location); err != nil {
			log.Printf("Failed to parse location data: %v", err)
			return
		}

		// Validate data
		if err := validateLocation(&location); err != nil {
			log.Printf("Invalid location data: %v", err)
			return
		}

		s.handler(&location)
	}

	token := s.client.Subscribe(topic, 1, messageHandler)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", token.Error())
	}

	log.Printf("Subscribed to topic: %s", topic)
	return nil
}

// Disconnect closes the MQTT connection
func (s *Subscriber) Disconnect() {
	s.client.Disconnect(250)
	log.Println("Disconnected from MQTT broker")
}

// validateLocation validates the location data
func validateLocation(loc *models.VehicleLocation) error {
	if loc.VehicleID == "" {
		return fmt.Errorf("vehicle_id is required")
	}

	if loc.Latitude < -90 || loc.Latitude > 90 {
		return fmt.Errorf("invalid latitude: %f", loc.Latitude)
	}

	if loc.Longitude < -180 || loc.Longitude > 180 {
		return fmt.Errorf("invalid longitude: %f", loc.Longitude)
	}

	if loc.Timestamp <= 0 {
		return fmt.Errorf("invalid timestamp: %d", loc.Timestamp)
	}

	return nil
}
