package api_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"
)

const (
	// test directories and endpoints.
	inputSuccessDir = "in/success"
	inputFailDir    = "in/fail"
	outputSVGDir    = "out/svg"
	outputPNGDir    = "out/png"
	outputPDFDir    = "out/pdf"
	urlSVG          = "http://localhost:3001/svg"
	urlPNG          = "http://localhost:3001/png"
	urlPDF          = "http://localhost:3001/pdf"
)

func TestSuccess(t *testing.T) {
	client := resty.New()
	files, err := os.ReadDir(inputSuccessDir)
	require.NoError(t, err)
	for _, f := range files {
		subTestName := f.Name()
		t.Run(subTestName, func(t *testing.T) {
			tex, err := os.ReadFile(filepath.Join(inputSuccessDir, f.Name()))
			require.NoError(t, err)
			r, err := client.R().SetBody(tex).Post(urlSVG)
			require.NoError(t, err)
			require.Equal(t, 200, r.StatusCode())
			svgPath := filepath.Join(outputSVGDir, strings.TrimSuffix(f.Name(), ".tex")+".svg")
			require.NoError(t, os.WriteFile(svgPath, r.Body(), 0600))
			println("Success")
		})
	}
}
