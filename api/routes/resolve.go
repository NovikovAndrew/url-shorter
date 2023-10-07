package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"net/http"
	"url-shorter/database"
)

func (s *Server) ResolveURL(ctx *fiber.Ctx) error {
	url := ctx.Params("url")
	value, err := s.redis.Get(database.Ctx, url).Result()

	if err == redis.Nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "short not found",
		})
	}

	if err != nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	_ = s.redis.Incr(database.Ctx, "counter")

	return ctx.Redirect(value, http.StatusMovedPermanently)
}
