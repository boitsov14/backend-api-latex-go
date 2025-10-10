package main

import (
	"image"
	_ "image/png"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

const pngMaxDimension = 32768

func png(c *fiber.Ctx) error {
	log.Info("Request received: /png")
	// get latex from body
	s := string(c.Body())
	// tmp directory
	tmp, err := os.MkdirTemp(".", "png-")
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
	// compile tex to pdf
	log.Info("Generating PDF")
	cmd := exec.Command("pdflatex", "-halt-on-error", "-interaction=nonstopmode", "-output-directory", tmp, filepath.ToSlash(tex)) // #nosec G204
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
	pdf := filepath.Join(tmp, "out.pdf")
	// convert pdf to png using ghostscript
	log.Info("Generating PNG")
	gs := "gs"
	if runtime.GOOS == "windows" {
		gs = "gswin64c"
	}
	png := filepath.Join(tmp, "out.png")
	dpis := []int{600, 300, 200, 150, 100, 72, 16}
	for _, dpi := range dpis {
		// run ghostscript
		cmd = exec.Command(gs, "-dBATCH", "-dNOPAUSE", "-r"+strconv.Itoa(dpi), "-sDEVICE=pngmono", "-o", filepath.ToSlash(png), filepath.ToSlash(pdf)) // #nosec G204
		outGS, err := cmd.CombinedOutput()
		if err != nil {
			log.Error(err)
			slog.Error("ghostscript error", "dpi", dpi, "output", string(outGS))
			c.Set("App-Error-Code", "LATEX_ERROR")
			return c.Status(fiber.StatusBadRequest).SendString("Unexpected LaTeX Error")
		}
		// open png
		f, err := os.Open(png)
		if err != nil {
			log.Error(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		// get config
		cfg, _, err := image.DecodeConfig(f)
		// close png
		if err := f.Close(); err != nil {
			log.Error(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		if err != nil {
			log.Error(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		log.Info("Generated PNG: dpi=", dpi, ", width=", cfg.Width, ", height=", cfg.Height)
		if cfg.Width <= pngMaxDimension && cfg.Height <= pngMaxDimension {
			// read png
			b, err := os.ReadFile(png) // #nosec G304
			if err != nil {
				log.Error(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			log.Info("PNG generated successfully")
			c.Type("png")
			return c.Send(b)
		}
		log.Warn("PNG too large, trying lower dpi")
	}
	log.Warn("PNG too large")
	// read png
	b, err := os.ReadFile(png) // #nosec G304
	if err != nil {
		log.Error(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	c.Type("png")
	return c.Send(b)
}
