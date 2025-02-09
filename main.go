package main

import (
	"CraftTanks/database"
	"CraftTanks/routes"
	"CraftTanks/utils"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	utils.LoadEnv()

	database.InitDB()
	database.InitRedis()

	app := fiber.New()
	app.Use(logger.New())
	app.Use(cors.New())
	//app.Use(handlers.SessionMiddleware())

	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
