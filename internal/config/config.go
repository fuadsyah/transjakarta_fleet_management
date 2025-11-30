package config

import (
	"os"
)

type Config struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string

	MQTTBroker   string
	MQTTClientID string

	RabbitMQURL string

	HTTPPort string

	// Geofence configuration
	GeofenceLatitude  float64
	GeofenceLongitude float64
	GeofenceRadius    float64 // in meters
}

func Load() *Config {
	return &Config{
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnv("POSTGRES_PORT", "5432"),
		PostgresDB:       getEnv("POSTGRES_DB", "fleet_management"),
		PostgresUser:     getEnv("POSTGRES_USER", "postgres"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "postgres"),

		// Local username/password for development
		// PostgresUser:     getEnv("POSTGRES_USER", "messagehub"),
		// PostgresPassword: getEnv("POSTGRES_PASSWORD", "Tunggulhitam1234"),

		MQTTBroker:   getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "fleet-backend"),

		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		// RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://messagehub:Tunggulhitam1234@localhost:5672/"),

		HTTPPort: getEnv("HTTP_PORT", "3000"),

		// Default geofence: Stasiun Bundaran HI
		GeofenceLatitude:  -6.1938148,
		GeofenceLongitude: 106.8230342,
		GeofenceRadius:    50.0, // 50 meters
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
