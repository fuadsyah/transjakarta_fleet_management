package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/fuadsyah/transjakarta_fleet_management/internal/handlers"
)

// SetupRouter configures the Fiber app with routes and middleware
func SetupRouter(vehicleHandler *handlers.VehicleHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Fleet Management API",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	})

	// API routes
	vehicles := app.Group("/vehicles")
	vehicles.Get("/:vehicle_id/location", vehicleHandler.GetLatestLocation)
	vehicles.Get("/:vehicle_id/history", vehicleHandler.GetLocationHistory)

	return app
}
