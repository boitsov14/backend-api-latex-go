package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

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
	cmd := exec.Command("latex", "-halt-on-error", "-interaction=nonstopmode", "-output-directory", tmp, filepath.ToSlash(tex)) // #nosec G204
	out, err := cmd.CombinedOutput()
	// dimension too large
	if strings.Contains(string(out), "Dimension too large") {
		log.Info("Dimension too large")
		return c.JSON(fiber.Map{"dimensionTooLarge": true})
	}
	if err != nil {
		// unexpected latex error
		log.Error(err)
		slog.Error("Unexpected latex error", "output", string(out))
		return c.JSON(fiber.Map{"err": true})
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
	// set Content-Type to image/svg+xml
	c.Type("svg")
	return c.Send(svg)
}
