package routes

import (
	"CraftTanks/handlers"
	"os"

	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// Publlc routes
	api.Post("/register", handlers.Register)
	api.Post("/login", handlers.Login)
	api.Post("/refresh", handlers.Refresh)

	protected := api.Group("", jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}))

	// Protected routes
	protected.Post("/logout", handlers.Logout)
	protected.Get("/users", handlers.GetUsers)
	protected.Get("/active-users", handlers.GetActiveUsers)
}
