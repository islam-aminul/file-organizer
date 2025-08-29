package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"zensort/internal/config"
)

// EXIFData represents extracted EXIF metadata
type EXIFData struct {
	Make         string
	Model        string
	DateTime     time.Time
	HasDateTime  bool
	Orientation  int
}

// ExtractEXIF extracts EXIF data from an image file
func ExtractEXIF(filePath string) (*EXIFData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		// Return empty EXIF data if no EXIF found
		return &EXIFData{}, nil
	}

	data := &EXIFData{}

	// Extract camera make
	if make, err := x.Get(exif.Make); err == nil {
		if makeStr, err := make.StringVal(); err == nil {
			data.Make = strings.TrimSpace(makeStr)
		}
	}

	// Extract camera model
	if model, err := x.Get(exif.Model); err == nil {
		if modelStr, err := model.StringVal(); err == nil {
			data.Model = strings.TrimSpace(modelStr)
		}
	}

	// Extract date/time
	if dateTime, err := x.Get(exif.DateTime); err == nil {
		if dateTimeStr, err := dateTime.StringVal(); err == nil {
			if parsedTime, err := time.Parse("2006:01:02 15:04:05", dateTimeStr); err == nil {
				data.DateTime = parsedTime
				data.HasDateTime = true
			}
		}
	}

	// Extract orientation
	if orientation, err := x.Get(exif.Orientation); err == nil {
		if orientationInt, err := orientation.Int(0); err == nil {
			data.Orientation = int(orientationInt)
		}
	}

	return data, nil
}

// GetImageDestinationPath generates the destination path based on EXIF data
func GetImageDestinationPath(basePath, fileName string, exifData *EXIFData, config *config.Config, isExport bool) string {
	var subDir string
	
	if isExport {
		subDir = config.ImageDirs.Exports
	} else {
		subDir = config.ImageDirs.Originals
	}

	// If no EXIF data or missing camera info, use Collections
	if exifData.Make == "" || exifData.Model == "" {
		return filepath.Join(basePath, config.Directories.Images, subDir, "Collections", fileName)
	}

	// For exports: Images/Exports/Year/Date - HH-MM-SS -- Make - Model -- filename.jpg
	if isExport && exifData.HasDateTime {
		year := fmt.Sprintf("%04d", exifData.DateTime.Year())
		date := exifData.DateTime.Format("2006-01-02")
		time := exifData.DateTime.Format("15-04-05")
		exportName := fmt.Sprintf("%s - %s -- %s - %s -- %s.jpg", 
			date, time, exifData.Make, exifData.Model, 
			strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		return filepath.Join(basePath, config.Directories.Images, subDir, year, exportName)
	}

	// For originals: Images/Originals/Make - Model/Year/filename
	if exifData.HasDateTime {
		cameraDir := fmt.Sprintf("%s - %s", exifData.Make, exifData.Model)
		year := fmt.Sprintf("%04d", exifData.DateTime.Year())
		return filepath.Join(basePath, config.Directories.Images, subDir, cameraDir, year, fileName)
	}

	// Fallback: use camera info but no date
	cameraDir := fmt.Sprintf("%s - %s", exifData.Make, exifData.Model)
	return filepath.Join(basePath, config.Directories.Images, subDir, cameraDir, "Unknown Date", fileName)
}

// IsImageFile checks if a file is an image based on extension
func IsImageFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif", ".webp", ".heic", ".heif", ".raw", ".cr2", ".nef", ".arw", ".dng"}
	
	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}
