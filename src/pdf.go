package main

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

func pdf(c *fiber.Ctx) error {
	log.Info("Request received: /pdf")
	// get latex from body
	s := string(c.Body())
	// tmp directory
	tmp, err := os.MkdirTemp(".", "pdf-")
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
	// compress pdf using ghostscript
	log.Info("Compressing PDF")
	gs := "gs"
	if runtime.GOOS == "windows" {
		gs = "gswin64c"
	}
	pdfComp := filepath.Join(tmp, "out-comp.pdf")
	cmd = exec.Command(gs, "-dBATCH", "-dCompatibilityLevel=1.5", "-dNOPAUSE", "-sDEVICE=pdfwrite", "-o", filepath.ToSlash(pdfComp), filepath.ToSlash(pdf)) // #nosec G204
	outGS, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(err)
		slog.Error("ghostscript error", "output", string(outGS))
		c.Set("App-Error-Code", "LATEX_ERROR")
		return c.Status(fiber.StatusBadRequest).SendString("Unexpected LaTeX Error")
	}
	// read pdf
	b, err := os.ReadFile(pdfComp) // #nosec G304
	if err != nil {
		log.Error(err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	// success
	log.Info("PDF generated successfully")
	// set Content-Type to application/pdf
	c.Type("pdf")
	return c.Send(b)
}
