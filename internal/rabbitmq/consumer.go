package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/config"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/models"
)

// Consumer handles RabbitMQ consumption for geofence alerts
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     *config.Config
}

// NewConsumer creates a new RabbitMQ consumer
func NewConsumer(cfg *config.Config) (*Consumer, error) {
	return &Consumer{cfg: cfg}, nil
}

// Connect establishes connection to RabbitMQ
func (c *Consumer) Connect() error {
	var err error

	cfg := amqp.Config{
		Properties: amqp.Table{
			"connection_name": "fleet_mgmt_consumer",
		},
	}

	c.conn, err = amqp.DialConfig(c.cfg.RabbitMQURL, cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange (ensure it exists)
	err = c.channel.ExchangeDeclare(
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

	// Declare queue (ensure it exists)
	_, err = c.channel.QueueDeclare(
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
	err = c.channel.QueueBind(
		QueueName,    // queue name
		RoutingKey,   // routing key
		ExchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Println("Consumer connected to RabbitMQ")
	return nil
}

// Consume starts consuming messages from the geofence_alerts queue
func (c *Consumer) Consume(handler func(*models.GeofenceEvent)) error {
	msgs, err := c.channel.Consume(
		QueueName, // queue
		"",        // consumer tag
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Println("Worker started, waiting for geofence alerts...")

	for msg := range msgs {
		var event models.GeofenceEvent
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			msg.Nack(false, false) // Reject message
			continue
		}

		handler(&event)
		msg.Ack(false)
	}

	return nil
}

// Close closes the RabbitMQ connection
func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	log.Println("Consumer disconnected from RabbitMQ")
}
