package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/models"
	"github.com/fuadsyah/transjakarta_fleet_management/internal/repository"
)

// VehicleHandler handles HTTP requests for vehicle endpoints
type VehicleHandler struct {
	repo *repository.VehicleRepository
}

// NewVehicleHandler creates a new VehicleHandler
func NewVehicleHandler(repo *repository.VehicleRepository) *VehicleHandler {
	return &VehicleHandler{repo: repo}
}

// GetLatestLocation handles GET /vehicles/:vehicle_id/location
func (h *VehicleHandler) GetLatestLocation(c *fiber.Ctx) error {
	vehicleID := c.Params("vehicle_id")
	if vehicleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "vehicle_id is required",
		})
	}

	location, err := h.repo.GetLatestLocation(vehicleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to get location",
		})
	}

	if location == nil {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "vehicle not found",
		})
	}

	return c.JSON(location)
}

// GetLocationHistory handles GET /vehicles/:vehicle_id/history
func (h *VehicleHandler) GetLocationHistory(c *fiber.Ctx) error {
	vehicleID := c.Params("vehicle_id")
	if vehicleID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "vehicle_id is required",
		})
	}

	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr == "" || endStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "start and end query parameters are required",
		})
	}

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid start timestamp",
		})
	}

	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "invalid end timestamp",
		})
	}

	if start > end {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "start timestamp must be less than or equal to end timestamp",
		})
	}

	locations, err := h.repo.GetLocationHistory(vehicleID, start, end)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "failed to get location history",
		})
	}

	if locations == nil {
		locations = []models.VehicleLocation{}
	}

	return c.JSON(locations)
}
