package geofence

import (
	"math"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/config"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/models"
)

// Checker handles geofence checking logic
type Checker struct {
	cfg *config.Config
}

// NewChecker creates a new geofence checker
func NewChecker(cfg *config.Config) *Checker {
	return &Checker{cfg: cfg}
}

// IsInsideGeofence checks if a location is within the geofence radius
func (c *Checker) IsInsideGeofence(loc *models.VehicleLocation) bool {
	distance := haversineDistance(
		c.cfg.GeofenceLatitude,
		c.cfg.GeofenceLongitude,
		loc.Latitude,
		loc.Longitude,
	)

	return distance <= c.cfg.GeofenceRadius
}

// haversineDistance calculates the distance between two points in meters
// using the Haversine formula
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // Earth's radius in meters

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
