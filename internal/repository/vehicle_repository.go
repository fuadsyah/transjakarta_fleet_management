package repository

import (
	"database/sql"
	"fmt"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/models"
)

// VehicleRepository handles database operations for vehicle locations
type VehicleRepository struct {
	db *sql.DB
}

// NewVehicleRepository creates a new VehicleRepository instance
func NewVehicleRepository(db *sql.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

// SaveLocation inserts a new vehicle location into the database
func (r *VehicleRepository) SaveLocation(loc *models.VehicleLocation) error {
	query := `
		INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(query, loc.VehicleID, loc.Latitude, loc.Longitude, loc.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to save location: %w", err)
	}

	return nil
}

// GetLatestLocation retrieves the most recent location for a vehicle
func (r *VehicleRepository) GetLatestLocation(vehicleID string) (*models.VehicleLocation, error) {
	query := `
		SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var loc models.VehicleLocation
	err := r.db.QueryRow(query, vehicleID).Scan(
		&loc.VehicleID,
		&loc.Latitude,
		&loc.Longitude,
		&loc.Timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get latest location: %w", err)
	}

	return &loc, nil
}

// GetLocationHistory retrieves location history for a vehicle within a time range
func (r *VehicleRepository) GetLocationHistory(vehicleID string, startTime, endTime int64) ([]models.VehicleLocation, error) {
	query := `
		SELECT vehicle_id, latitude, longitude, timestamp
		FROM vehicle_locations
		WHERE vehicle_id = $1 AND timestamp >= $2 AND timestamp <= $3
		ORDER BY timestamp ASC
	`

	rows, err := r.db.Query(query, vehicleID, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get location history: %w", err)
	}
	defer rows.Close()

	var locations []models.VehicleLocation
	for rows.Next() {
		var loc models.VehicleLocation
		if err := rows.Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		locations = append(locations, loc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return locations, nil
}
