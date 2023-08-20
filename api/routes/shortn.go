package routes

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/redis/go-redis/v9"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/moonman369/Shortn/database"
	"github.com/moonman369/Shortn/errorhandler"
	"github.com/moonman369/Shortn/helpers"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortnURL(c *fiber.Ctx) error {
	err := godotenv.Load("../.env")
	if err != nil {
		errorhandler.ErrorHandler(err)
	}
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		errorhandler.ErrorHandler(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot Parse JSON"})
	}

	// implementing rate limiting
	r0 := database.CreateClient(0)
	defer r0.Close()
	val, err := r0.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil || val == "" {
		errorhandler.ErrorHandler(err)
		err0 := r0.Set(database.Ctx, c.IP(), 10, 30*60*time.Second).Err()
		if err0 != nil {
			errorhandler.ErrorHandler(err0)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "1: Unable to connect to server"})
		}
		fmt.Println(os.Getenv("API_QUOTA"))
	} else {
		val, _ := r0.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r0.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Rate Limit Exceeded", "rate_limit_reset": limit / time.Nanosecond / time.Minute})
		}
	}

	// check validity of request url
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	// check for domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Domain Exploit Detected"})
	}

	// enforce https, SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(1)
	defer r.Close()

	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Custom short is already in use"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		fmt.Println(err)
		errorhandler.ErrorHandler(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "0: Unable to connect to server"})
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	r0.Decr(database.Ctx, c.IP())
	val, _ = r0.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r0.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
