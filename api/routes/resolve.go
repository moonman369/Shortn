package routes

import (
	"github.com/redis/go-redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/moonman369/Shortn/database"
	"github.com/moonman369/Shortn/errorhandler"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")

	r := database.CreateClient(0)
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		errorhandler.ErrorHandler(err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Short was not found in the database"})
	} else if err != nil {
		errorhandler.ErrorHandler(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not connect to DB"})
	}

	rInr := database.CreateClient(1)
	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")

	return c.Redirect(value, 301)
}
