package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/moonman369/Shortn/errorhandler"
	"github.com/moonman369/Shortn/routes"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveURL)
	app.Post("/api/v1", routes.ShortnURL)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
		errorhandler.ErrorHandler(err)
	}

	app := fiber.New()

	app.Use(logger.New())

	setupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
