package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/moonman369/Shortn/routes"
	"github.com/moonman369/Shortn/errorhandler"
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

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
