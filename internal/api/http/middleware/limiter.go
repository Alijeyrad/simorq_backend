package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	fiberredis "github.com/gofiber/storage/redis/v3"
	"github.com/redis/go-redis/v9"
)

func NewLimiterWithRedis(rdb *redis.Client) fiber.Handler {
	storage := fiberredis.NewFromConnection(rdb)
	return limiter.New(limiter.Config{
		Storage: storage,

		// sliding window
		Max:               20,
		Expiration:        30 * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
	})
}
