package api

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	pdftotext "github.com/heussd/pdftotext-go"
	"github.com/labstack/echo/v4"
)

func (a *API) ExtractPdfText(c echo.Context) error {
	// Get the file from the request
	file, err := c.FormFile("pdf")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No PDF file provided").WithInternal(err)
	}

	// Check file extension
	if filepath.Ext(file.Filename) != ".pdf" {
		return echo.NewHTTPError(http.StatusBadRequest, "Only PDF files are supported")
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open PDF file").WithInternal(err)
	}
	defer src.Close()

	// Read file into byte slice
	fileContent, err := io.ReadAll(src)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read PDF file").WithInternal(err)
	}

	// Extract pages texts from the PDF
	pdfPages, err := pdftotext.Extract(fileContent)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read PDF file").WithInternal(err)
	}
	// Concatenate all pages into a single string
	textContent := ""
	for _, page := range pdfPages {
		textContent += fmt.Sprintf("Page %d:\n%s\n", page.Number, page.Content)
	}

	// Set headers for text file download
	c.Response().Header().Set(
		"Content-Type",
		"text/plain; charset=utf-8",
	)
	c.Response().Header().Set(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=%s.txt", fileNameWithoutExt(file.Filename)),
	)

	// Return the extracted text
	return c.String(http.StatusOK, textContent)
}

// fileNameWithoutExt returns the filename without extension
func fileNameWithoutExt(fileName string) string {
	return filepath.Base(fileName[:len(fileName)-len(filepath.Ext(fileName))])
}
