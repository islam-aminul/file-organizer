package core

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	"zensort/internal/config"
)

// ImageProcessor handles image resizing and export operations
type ImageProcessor struct {
	config *config.Config
}

// NewImageProcessor creates a new image processor
func NewImageProcessor(config *config.Config) *ImageProcessor {
	return &ImageProcessor{config: config}
}

// ProcessImage handles both original copying and export generation
func (ip *ImageProcessor) ProcessImage(srcPath, destPath string, exifData *EXIFData) error {
	// Always copy original first
	if err := ip.copyOriginal(srcPath, destPath); err != nil {
		return fmt.Errorf("failed to copy original: %w", err)
	}

	// Generate export if enabled and it's a supported format
	if ip.config.Processing.EnableImageExports && ip.shouldCreateExport(srcPath) {
		exportPath := ip.getExportPath(destPath, exifData)
		if err := ip.createExport(srcPath, exportPath, exifData); err != nil {
			// Log error but don't fail the whole operation
			fmt.Printf("Warning: Failed to create export for %s: %v\n", srcPath, err)
		}
	}

	return nil
}

// copyOriginal copies the image file as-is to the originals directory
func (ip *ImageProcessor) copyOriginal(srcPath, destPath string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Copy file
	return copyFile(srcPath, destPath)
}

// createExport creates a resized JPEG export of the image
func (ip *ImageProcessor) createExport(srcPath, exportPath string, exifData *EXIFData) error {
	// Ensure export directory exists
	if err := os.MkdirAll(filepath.Dir(exportPath), 0755); err != nil {
		return err
	}

	// Open and decode image
	src, err := imaging.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Get current dimensions
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check if resizing is needed (only downscale, never upscale)
	maxWidth := ip.config.Processing.MaxImageWidth
	maxHeight := ip.config.Processing.MaxImageHeight
	
	var resized image.Image = src
	
	if width > maxWidth || height > maxHeight {
		// Calculate aspect ratio preserving dimensions
		aspectRatio := float64(width) / float64(height)
		
		var newWidth, newHeight int
		if aspectRatio > float64(maxWidth)/float64(maxHeight) {
			// Width is the limiting factor
			newWidth = maxWidth
			newHeight = int(float64(maxWidth) / aspectRatio)
		} else {
			// Height is the limiting factor
			newHeight = maxHeight
			newWidth = int(float64(maxHeight) * aspectRatio)
		}
		
		resized = imaging.Resize(src, newWidth, newHeight, imaging.Linear)
	}

	// Apply EXIF orientation correction if needed
	if exifData != nil && exifData.Orientation > 1 {
		resized = ip.applyOrientation(resized, exifData.Orientation)
	}

	// Save as JPEG with high quality
	outFile, err := os.Create(exportPath)
	if err != nil {
		return fmt.Errorf("failed to create export file: %w", err)
	}
	defer outFile.Close()

	// Save with configurable quality and preserve EXIF data
	quality := ip.config.Processing.JPEGQuality
	if quality <= 0 || quality > 100 {
		quality = 85 // Default fallback
	}
	
	// Preserve EXIF data by copying from original file
	if err := ip.preserveEXIF(srcPath, outFile, resized, quality); err != nil {
		// Fallback to basic JPEG encoding if EXIF preservation fails
		return jpeg.Encode(outFile, resized, &jpeg.Options{Quality: quality})
	}
	
	return nil
}

// applyOrientation applies EXIF orientation correction
func (ip *ImageProcessor) applyOrientation(img image.Image, orientation int) image.Image {
	switch orientation {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.FlipH(imaging.Rotate270(img))
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.FlipH(imaging.Rotate90(img))
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

// shouldCreateExport determines if an export should be created for this image
func (ip *ImageProcessor) shouldCreateExport(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	// Create exports for common image formats
	exportFormats := []string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp", ".webp"}
	
	for _, format := range exportFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// getExportPath generates the export file path based on EXIF data
func (ip *ImageProcessor) getExportPath(originalDestPath string, exifData *EXIFData) string {
	// Get the base destination directory (not the Images subdirectory)
	// originalDestPath is like: /dest/Images/Originals/Collections/file.jpg
	// We need to get: /dest (base destination)
	
	fileName := filepath.Base(originalDestPath)
	
	// Find the base destination by going up until we're above "Images"
	currentPath := originalDestPath
	for {
		parent := filepath.Dir(currentPath)
		if filepath.Base(parent) != "Images" && filepath.Base(currentPath) == "Images" {
			// parent is the base destination directory
			return GetImageDestinationPath(parent, fileName, exifData, ip.config, true)
		}
		if parent == currentPath {
			// Reached root, fallback
			break
		}
		currentPath = parent
	}
	
	// Fallback: assume standard structure
	baseDir := filepath.Dir(filepath.Dir(filepath.Dir(originalDestPath))) // Go up 3 levels from Collections
	return GetImageDestinationPath(baseDir, fileName, exifData, ip.config, true)
}

// preserveEXIF attempts to preserve EXIF data when creating exports
func (ip *ImageProcessor) preserveEXIF(srcPath string, outFile *os.File, img image.Image, quality int) error {
	// Open original file to read EXIF data
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Try to decode EXIF data
	_, err = exif.Decode(srcFile)
	if err != nil {
		// No EXIF data to preserve, use basic encoding
		return jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
	}

	// For now, we'll use basic JPEG encoding as preserving EXIF in processed images
	// requires more complex handling. The EXIF data is already extracted and stored
	// in the database for organization purposes.
	return jpeg.Encode(outFile, img, &jpeg.Options{Quality: quality})
}
