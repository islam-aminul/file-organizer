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
	Software     string
}

// ExtractEXIF extracts EXIF data from an image file with timeout protection
func ExtractEXIF(filePath string) (*EXIFData, error) {
	// Set timeout to prevent hanging on corrupted files
	done := make(chan *EXIFData, 1)
	errChan := make(chan error, 1)
	
	go func() {
		file, err := os.Open(filePath)
		if err != nil {
			errChan <- err
			return
		}
		defer file.Close()

		x, err := exif.Decode(file)
		if err != nil {
			// Return empty EXIF data if no EXIF found
			done <- &EXIFData{}
			return
		}
		
		data := extractEXIFFields(x)
		done <- data
	}()
	
	select {
	case data := <-done:
		return data, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(5 * time.Second):
		return &EXIFData{}, fmt.Errorf("EXIF extraction timeout after 5 seconds for file: %s", filePath)
	}
}

// extractEXIFFields extracts EXIF fields from decoded EXIF data
func extractEXIFFields(x *exif.Exif) *EXIFData {

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

	// Extract date/time with multiple format support
	if dateTime, err := x.Get(exif.DateTime); err == nil {
		if dateTimeStr, err := dateTime.StringVal(); err == nil {
			if parsedTime := parseDateTime(dateTimeStr); !parsedTime.IsZero() {
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

	// Extract software
	if software, err := x.Get(exif.Software); err == nil {
		if softwareStr, err := software.StringVal(); err == nil {
			data.Software = strings.TrimSpace(softwareStr)
		}
	}

	return data
}

// IsEditedImage checks if an image has been edited based on EXIF software field
func IsEditedImage(exifData *EXIFData, config *config.Config) bool {
	if exifData.Software == "" {
		return false
	}
	
	// Check against configured editing software patterns
	softwareLower := strings.ToLower(exifData.Software)
	for _, pattern := range config.EditedImages.SoftwarePatterns {
		if strings.Contains(softwareLower, strings.ToLower(pattern)) {
			return true
		}
	}
	
	return false
}

// parseDateTime attempts to parse datetime string with multiple formats
func parseDateTime(dateTimeStr string) time.Time {
	// Try standard EXIF format first: "2006:01:02 15:04:05"
	if parsedTime, err := time.Parse("2006:01:02 15:04:05", dateTimeStr); err == nil {
		return parsedTime
	}
	
	// Try ISO format: "2006-01-02 15:04:05"
	if parsedTime, err := time.Parse("2006-01-02 15:04:05", dateTimeStr); err == nil {
		return parsedTime
	}
	
	// Try other common formats
	formats := []string{
		"2006:01:02T15:04:05",
		"2006-01-02T15:04:05",
		"2006:01:02 15:04:05Z",
		"2006-01-02 15:04:05Z",
		"2006:01:02T15:04:05Z",
		"2006-01-02T15:04:05Z",
	}
	
	for _, format := range formats {
		if parsedTime, err := time.Parse(format, dateTimeStr); err == nil {
			return parsedTime
		}
	}
	
	return time.Time{} // Return zero time if no format matches
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
	
	// For exports without EXIF date: use configured no-EXIF year folder
	if isExport {
		exportName := fmt.Sprintf("%s - %s -- %s.jpg", 
			exifData.Make, exifData.Model, 
			strings.TrimSuffix(fileName, filepath.Ext(fileName)))
		return filepath.Join(basePath, config.Directories.Images, subDir, config.ImageDirs.NoExifYearFolder, exportName)
	}

	// For originals: Images/Originals/Make - Model/Year/filename
	if exifData.HasDateTime {
		cameraDir := fmt.Sprintf("%s - %s", exifData.Make, exifData.Model)
		year := fmt.Sprintf("%04d", exifData.DateTime.Year())
		return filepath.Join(basePath, config.Directories.Images, subDir, cameraDir, year, fileName)
	}

	// Fallback: use camera info but no date - use configured no-EXIF year folder
	cameraDir := fmt.Sprintf("%s - %s", exifData.Make, exifData.Model)
	return filepath.Join(basePath, config.Directories.Images, subDir, cameraDir, config.ImageDirs.NoExifYearFolder, fileName)
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
