package api

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/young2j/oxmltotext/pdftotext"
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

	// Create PDF reader
	pdfReader, err := pdftotext.OpenReader(src)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to read PDF file").WithInternal(err)
	}
	defer pdfReader.Close()

	textContent, err := pdfReader.ExtractTexts()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to extract text from PDF").WithInternal(err)
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
