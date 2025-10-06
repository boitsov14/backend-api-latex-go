package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func svg(c *fiber.Ctx) error {
	log.Info("Request received: /svg")
	// get latex from body
	s := string(c.Body())
	// check if empty
	if strings.TrimSpace(s) == "" {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	// tmp directory
	tmp, err := os.MkdirTemp(".", "svg-")
	if err != nil {
		log.Error(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	// cleanup
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			log.Error(err)
		}
	}()
	// write tex to file
	tex := filepath.Join(tmp, "out.tex")
	if err := os.WriteFile(tex, []byte(s), 0400); err != nil {
		log.Error(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	// compile tex to dvi
	log.Info("Compiling LaTeX to DVI")
	cmd := exec.Command("latex", "-halt-on-error", "-interaction=nonstopmode", "-output-directory", tmp, tex) // #nosec G204
	out, err := cmd.CombinedOutput()
	if err != nil {
		// unexpected latex error
		log.Error(err)
		slog.Error("Unexpected latex error", "output", string(out))
		return c.JSON(fiber.Map{"err": true})
	}
	// dimension too large
	if strings.Contains(string(out), "Dimension too large") {
		log.Info("Dimension too large")
		return c.JSON(fiber.Map{"dimensionTooLarge": true})
	}
	dvi := filepath.Join(tmp, "out.dvi")
	// check if dvi exists
	if _, err := os.Stat(dvi); err != nil {
		// unexpected latex error
		log.Error(err)
		slog.Error("Unexpected latex error", "output", string(out))
		return c.JSON(fiber.Map{"err": true})
	}
	// compile dvi to svg
	log.Info("Compiling DVI to SVG")
	cmd = exec.Command("dvisvgm", "--bbox=preview", "--bitmap-format=none", "--font-format=woff2", "--optimize", "--relative", "-o", filepath.Join(tmp, "out.svg"), dvi) // #nosec G204
	out, err = cmd.CombinedOutput()
	if err != nil {
		// unexpected latex error
		log.Error(err)
		slog.Error("Unexpected latex error", "output", string(out))
		return c.JSON(fiber.Map{"err": true})
	}
	svg, err := os.ReadFile(filepath.Join(tmp, "out.svg")) // #nosec G304
	if err != nil {
		// unexpected latex error
		log.Error(err)
		slog.Error("Unexpected latex error", "output", string(out))
		return c.JSON(fiber.Map{"err": true})
	}
	// success
	log.Info("SVG generated successfully")
	return c.JSON(fiber.Map{"svg": string(svg)})
}
