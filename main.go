package main

import (
	"CraftTanks/database"
	"CraftTanks/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	database.InitDB()
	database.InitRedis()

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/routes", func(c *fiber.Ctx) error {
		return c.JSON(app.Stack())
	})

	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
