package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/config"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/models"
)

const (
	ExchangeName = "fleet.events"
	QueueName    = "geofence_alerts"
	RoutingKey   = "geofence.entry"
)

// Publisher handles RabbitMQ publishing for geofence events
type Publisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     *config.Config
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher(cfg *config.Config) (*Publisher, error) {
	return &Publisher{cfg: cfg}, nil
}

// Connect establishes connection to RabbitMQ
func (p *Publisher) Connect() error {
	var err error

	cfg := amqp.Config{
		Properties: amqp.Table{
			"connection_name": "fleet_mgmt_publisher",
		},
	}

	p.conn, err = amqp.DialConfig(p.cfg.RabbitMQURL, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	p.channel, err = p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = p.channel.ExchangeDeclare(
		ExchangeName, // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = p.channel.QueueDeclare(
		QueueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = p.channel.QueueBind(
		QueueName,    // queue name
		RoutingKey,   // routing key
		ExchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Println("Successfully connected to RabbitMQ")
	return nil
}

// PublishGeofenceEvent sends a geofence event to RabbitMQ
func (p *Publisher) PublishGeofenceEvent(event *models.GeofenceEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = p.channel.PublishWithContext(
		ctx,
		ExchangeName, // exchange
		RoutingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published geofence event for vehicle: %s", event.VehicleID)
	return nil
}

// Close closes the RabbitMQ connection
func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	log.Println("Disconnected from RabbitMQ")
}
