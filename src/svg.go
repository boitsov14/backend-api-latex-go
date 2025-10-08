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
	outLatex, err := cmd.CombinedOutput()
	if err != nil {
		log.Warn(err)
		slog.Error("LaTeX error", "output", string(outLatex))
		c.Set("App-Error-Code", "LATEX_ERROR")
		// scan latex output and pick first line starting with '!'
		errMsg := ""
		for line := range strings.SplitSeq(string(outLatex), "\n") {
			if after, ok := strings.CutPrefix(line, "! "); ok {
				errMsg = strings.TrimSpace(after)
				break
			}
		}
		if errMsg == "" {
			// no '!' lines
			errMsg = "Unexpected LaTeX Error"
		}
		log.Warn(errMsg)
		return c.Status(fiber.StatusBadRequest).SendString(errMsg)
	}
	dvi := filepath.Join(tmp, "out.dvi")
	// compile dvi to svg
	log.Info("Compiling DVI to SVG")
	cmd = exec.Command("dvisvgm", "--bbox=preview", "--bitmap-format=none", "--font-format=woff2", "--optimize", "--relative", "-o", filepath.Join(tmp, "out.svg"), dvi) // #nosec G204
	outDvisvgm, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err)
		slog.Error("dvisvgm error", "output", string(outDvisvgm))
		c.Set("App-Error-Code", "LATEX_ERROR")
		return c.Status(fiber.StatusBadRequest).SendString("Unexpected LaTeX Error")
	}
	svg, err := os.ReadFile(filepath.Join(tmp, "out.svg")) // #nosec G304
	if err != nil {
		log.Error(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	// success
	log.Info("SVG generated successfully")
	// set Content-Type to image/svg+xml
	c.Type("svg")
	return c.Send(svg)
}
