package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// fiber instance
	app := fiber.New(fiber.Config{
		// disable startup message
		DisableStartupMessage: true,
	})

	// add middlewares
	app.Use(recover.New())     // recover from panics
	app.Use(helmet.New())      // security
	app.Use(logger.New())      // logging
	app.Use(compress.New())    // compression
	app.Use(healthcheck.New()) // healthcheck at /livez

	// setup json logger
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(l)

	// main API
	app.Post("/svg", svg)

	// init port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	// use localhost in dev environment
	host := ""
	if os.Getenv("ENV") == "dev" {
		host = "localhost"
	}

	// start server
	log.Info("Starting server on port: ", port)
	log.Fatal(app.Listen(host + ":" + port))
}
