package api_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

const (
	// test directories and endpoints.
	inputSuccessDir = "in/success"
	inputFailDir    = "in/fail"
	outputDir       = "out"
	urlSVG          = "http://localhost:3001/svg"
	urlPNG          = "http://localhost:3001/png"
	urlPDF          = "http://localhost:3001/pdf"
)

func TestSuccess(t *testing.T) {
	// run tests in parallel
	// t.Parallel()
	client := resty.New()
	client.SetTimeout(10 * time.Minute)
	// walk recursively under inputSuccessDir and pick .tex files
	err := filepath.WalkDir(inputSuccessDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".tex" {
			return nil
		}

		// get relative path
		inRel, err := filepath.Rel(inputSuccessDir, path)
		if err != nil {
			return err
		}

		t.Run(filepath.ToSlash(inRel)+": svg", func(t *testing.T) {
			// run sub tests in parallel
			// t.Parallel()

			// read tex file
			tex, err := os.ReadFile(path) // #nosec G304
			require.NoError(t, err)

			// post to /svg endpoint
			r, err := client.R().SetBody(tex).Post(urlSVG)
			require.NoError(t, err)
			if r.StatusCode() != 200 {
				// print response body for debugging
				t.Log("Response body:", string(r.Body()))
			}
			// expect 200 OK
			require.Equal(t, 200, r.StatusCode())

			// mirror input folders
			outRel := strings.TrimSuffix(inRel, ".tex") + ".svg"
			svgPath := filepath.Join(outputDir, outRel)
			require.NoError(t, os.MkdirAll(filepath.Dir(svgPath), 0600))
			require.NoError(t, os.WriteFile(svgPath, r.Body(), 0600))
		})

		t.Run(filepath.ToSlash(inRel)+": png", func(t *testing.T) {
			// run sub tests in parallel
			// t.Parallel()

			// read tex file
			tex, err := os.ReadFile(path) // #nosec G304
			require.NoError(t, err)

			// post to /png endpoint
			r, err := client.R().SetBody(tex).Post(urlPNG)
			require.NoError(t, err)
			if r.StatusCode() != 200 {
				// print response body for debugging
				t.Log("Response body:", string(r.Body()))
			}
			// expect 200 OK
			require.Equal(t, 200, r.StatusCode())

			// mirror input folders
			outRel := strings.TrimSuffix(inRel, ".tex") + ".png"
			pngPath := filepath.Join(outputDir, outRel)
			require.NoError(t, os.MkdirAll(filepath.Dir(pngPath), 0600))
			require.NoError(t, os.WriteFile(pngPath, r.Body(), 0600))
		})

		t.Run(filepath.ToSlash(inRel)+": pdf", func(t *testing.T) {
			// run sub tests in parallel
			// t.Parallel()

			// read tex file
			tex, err := os.ReadFile(path) // #nosec G304
			require.NoError(t, err)

			// post to /pdf endpoint
			r, err := client.R().SetBody(tex).Post(urlPDF)
			require.NoError(t, err)
			if r.StatusCode() != 200 {
				// print response body for debugging
				t.Log("Response body:", string(r.Body()))
			}
			// expect 200 OK
			require.Equal(t, 200, r.StatusCode())

			// mirror input folders
			outRel := strings.TrimSuffix(inRel, ".tex") + ".pdf"
			pdfPath := filepath.Join(outputDir, outRel)
			require.NoError(t, os.MkdirAll(filepath.Dir(pdfPath), 0600))
			require.NoError(t, os.WriteFile(pdfPath, r.Body(), 0600))
		})

		return nil
	})
	require.NoError(t, err)
}

func TestFail(t *testing.T) {
	// run tests in parallel
	// t.Parallel()
	client := resty.New()
	// walk recursively under inputFailDir and pick .tex files
	err := filepath.WalkDir(inputFailDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".tex" {
			return nil
		}

		// get relative path
		inRel, err := filepath.Rel(inputFailDir, path)
		if err != nil {
			return err
		}

		t.Run(filepath.ToSlash(inRel), func(t *testing.T) {
			// run sub tests in parallel
			// t.Parallel()

			// read tex file
			tex, err := os.ReadFile(path) // #nosec G304
			require.NoError(t, err)

			// post to /svg endpoint
			r, err := client.R().SetBody(tex).Post(urlSVG)
			require.NoError(t, err)
			// log response body for debugging
			t.Log("Response body:", string(r.Body()))
			// expect 400 Bad Request with LATEX_ERROR
			require.Equal(t, 400, r.StatusCode())
			require.Equal(t, "LATEX_ERROR", r.Header().Get("App-Error-Code"))
			// assert response body does not contain "Unexpected"
			require.NotContains(t, string(r.Body()), "Unexpected")
		})

		return nil
	})
	require.NoError(t, err)
}
