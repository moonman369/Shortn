package database

import (
	"context"
	"fmt"

	"github.com/moonman369/Shortn/errorhandler"
	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func CreateClient(dbNo int) *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // os.Getenv("DB_ADDRESS"),
		Password: "",           // os.Getenv("DB_PASSWORD"),
		DB:       0,
		Protocol: 3,
	})

	ping, err := rdb.Ping(Ctx).Result()
	if err != nil {
		errorhandler.ErrorHandler(err)
	}

	fmt.Println(ping)

	return rdb
}
