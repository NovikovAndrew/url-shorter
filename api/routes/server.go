package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	app   *fiber.App
	redis *redis.Client
	addr  string
}

func NewServer(redis *redis.Client, addr string) *Server {
	app := fiber.New()
	app.Use(logger.New())
	server := Server{
		app:   app,
		redis: redis,
		addr:  addr,
	}
	server.setupRoutes()

	return &server
}

func (s *Server) Start() error {
	return s.app.Listen(s.addr)
}

func (s *Server) setupRoutes() {
	s.app.Get("/:url", s.ResolveURL)
	s.app.Post("/api/v1", s.ShortenURL)
}
