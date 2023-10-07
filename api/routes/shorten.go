package routes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"url-shorter/database"
	"url-shorter/helpers"

	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"customShort"`
	ExpiredAt   time.Duration `json:"expired_at"`
}

type urlResponse struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"custom_short"`
	ExpiredAt      time.Duration `json:"expired_at"`
	XRateRemaining int           `json:"x_rate_remaining"`
	XRateLimitRest time.Duration `json:"x_rate_limit_rest"`
}

func (s *Server) ShortenURL(ctx *fiber.Ctx) error {
	req := new(request)
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"failed": fmt.Sprintf("failed to parse json, err: %s", err.Error()),
		})
	}

	val, err := s.redis.Get(database.Ctx, ctx.IP()).Result()

	if err == redis.Nil {
		_ = s.redis.Set(database.Ctx, ctx.IP(), os.Getenv("APP_QUOTA"), 30*60*time.Minute).Err()
	} else {
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := s.redis.TTL(database.Ctx, ctx.IP()).Result()
			return ctx.Status(http.StatusServiceUnavailable).JSON(fiber.Map{
				"error":           "limited request",
				"rate_limit_rest": limit / time.Nanosecond / time.Second,
			})
		}
	}

	if !govalidator.IsURL(req.URL) {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"failed": fmt.Sprintf("the field [%s] is not url", req.URL),
		})
	}

	if !helpers.RemoveDomainError(req.URL) {
		return ctx.Status(http.StatusServiceUnavailable).JSON(fiber.Map{
			"failed": fmt.Sprintf("the domain [%s] is not unavailable", req.URL),
		})
	}

	req.URL = helpers.EnforceHTTP(req.URL)

	var id string

	if req.CustomShort == "" {
		uuid, err := uuid.NewRandom()
		if err != nil {
			return internalErrJSON(ctx, err)
		}
		id = uuid.String()[:6]
	} else {
		id = req.CustomShort
	}

	val, _ = s.redis.Get(database.Ctx, id).Result()

	if val != "" {
		return ctx.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": fmt.Sprintf("url custom short already has: %s", id),
		})
	}

	if req.ExpiredAt == 0 {
		req.ExpiredAt = 24
	}

	err = s.redis.Set(database.Ctx, id, req.URL, 3600*req.ExpiredAt*time.Second).Err()

	if err != nil {
		return internalErrJSON(ctx, err)
	}

	response := urlResponse{
		URL:            req.URL,
		CustomShort:    "",
		ExpiredAt:      req.ExpiredAt,
		XRateRemaining: 10,
		XRateLimitRest: 30,
	}

	err = s.redis.Decr(database.Ctx, ctx.IP()).Err()

	if err != nil {
		return internalErrJSON(ctx, err)
	}

	remaining, err := s.redis.Get(database.Ctx, ctx.IP()).Result()

	if err != nil {
		return internalErrJSON(ctx, err)
	}

	intValue, err := strconv.Atoi(remaining)

	if err != nil {
		return internalErrJSON(ctx, err)
	}

	response.XRateRemaining = intValue

	ttl, err := s.redis.TTL(database.Ctx, ctx.IP()).Result()

	if err != nil {
		return internalErrJSON(ctx, err)
	}

	response.ExpiredAt = ttl / time.Nanosecond / time.Minute

	response.URL = os.Getenv("DOMAIN") + "/" + id

	return ctx.Status(http.StatusOK).JSON(response)
}

func internalErrJSON(ctx *fiber.Ctx, err error) error {
	return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
		"error": err.Error(),
	})
}
