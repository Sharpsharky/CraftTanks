package routes

import (
	"CraftTanks/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/register", handlers.Register)
	api.Post("/login", handlers.Login)
	api.Get("/users", handlers.GetUsers)
	api.Post("/refresh", handlers.Refresh)
}
