package api

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func (a *API) CompressImage(c echo.Context) error {
	// Get the file from the request
	file, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No image provided").WithInternal(err)
	}

	// Get quality parameter (default to 75 if not specified)
	quality := 75
	if qualityParam := c.FormValue("quality"); qualityParam != "" {
		fmt.Sscanf(qualityParam, "%d", &quality)
		if quality < 1 {
			quality = 1
		} else if quality > 100 {
			quality = 100
		}
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to open image").WithInternal(err)
	}
	defer src.Close()

	// Decode the image
	img, format, err := decodeImage(src, file)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to decode image").WithInternal(err)
	}

	// Prepare the response
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=compressed_%s", file.Filename))

	// Compress and send the image
	if err := encodeImage(c.Response().Writer, img, format, quality); err != nil {
		return err
	}

	return nil
}

func decodeImage(src multipart.File, fileHeader *multipart.FileHeader) (image.Image, string, error) {
	// Reset the file to the beginning
	if _, err := src.Seek(0, 0); err != nil {
		return nil, "", err
	}

	// Read the file content
	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return nil, "", err
	}

	// Create a bytes reader
	reader := bytes.NewReader(fileBytes)

	// Get the file extension
	ext := filepath.Ext(fileHeader.Filename)
	format := ext[1:] // Remove the dot

	// Decode based on format
	var img image.Image
	switch format {
	case "jpg", "jpeg":
		img, err = jpeg.Decode(reader)
		format = "jpeg"
	case "png":
		img, err = png.Decode(reader)
	default:
		return nil, "", fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

func encodeImage(w io.Writer, img image.Image, format string, quality int) error {
	switch format {
	case "jpeg", "jpg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	case "png":
		encoder := png.Encoder{
			CompressionLevel: png.BestCompression,
		}
		return encoder.Encode(w, img)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
