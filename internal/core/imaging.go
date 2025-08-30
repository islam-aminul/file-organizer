package core

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io"
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

	// Save with configurable quality
	quality := ip.config.Processing.JPEGQuality
	if quality <= 0 || quality > 100 {
		quality = 85 // Default fallback
	}
	
	// Preserve EXIF data by copying from original file
	if err := ip.preserveEXIFData(srcPath, outFile, resized, quality); err != nil {
		// Fallback to standard encoding if EXIF preservation fails
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

// preserveEXIFData extracts EXIF from source and embeds it in the processed image
func (ip *ImageProcessor) preserveEXIFData(srcPath string, outFile *os.File, img image.Image, quality int) error {
	// Read original file to extract EXIF data
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Verify EXIF data exists (we don't need to decode it, just check)
	_, err = exif.Decode(srcFile)
	if err != nil {
		return err // No EXIF data or decode error
	}

	// Encode processed image to buffer first
	var imgBuf bytes.Buffer
	if err := jpeg.Encode(&imgBuf, img, &jpeg.Options{Quality: quality}); err != nil {
		return err
	}

	// Read original file completely to get raw EXIF segments
	srcFile.Seek(0, 0)
	originalData, err := io.ReadAll(srcFile)
	if err != nil {
		return err
	}

	// Extract EXIF segment from original JPEG
	exifSegment, err := ip.extractEXIFSegment(originalData)
	if err != nil {
		return err
	}

	// Combine processed image with original EXIF
	return ip.writeJPEGWithEXIF(outFile, imgBuf.Bytes(), exifSegment)
}

// extractEXIFSegment extracts the EXIF APP1 segment from JPEG data
func (ip *ImageProcessor) extractEXIFSegment(jpegData []byte) ([]byte, error) {
	if len(jpegData) < 4 || jpegData[0] != 0xFF || jpegData[1] != 0xD8 {
		return nil, fmt.Errorf("not a valid JPEG file")
	}

	pos := 2
	for pos < len(jpegData)-1 {
		if jpegData[pos] != 0xFF {
			return nil, fmt.Errorf("invalid JPEG marker")
		}
		
		marker := jpegData[pos+1]
		pos += 2
		
		if marker == 0xE1 { // APP1 segment (EXIF)
			if pos+2 > len(jpegData) {
				break
			}
			length := int(jpegData[pos])<<8 | int(jpegData[pos+1])
			if pos+length > len(jpegData) {
				break
			}
			
			// Check if this is EXIF data
			if length > 6 && string(jpegData[pos+2:pos+6]) == "Exif" {
				return jpegData[pos-2:pos+length], nil
			}
		}
		
		if marker == 0xDA { // Start of scan - no more metadata
			break
		}
		
		if pos+2 > len(jpegData) {
			break
		}
		length := int(jpegData[pos])<<8 | int(jpegData[pos+1])
		pos += length
	}
	
	return nil, fmt.Errorf("no EXIF segment found")
}

// writeJPEGWithEXIF writes JPEG with EXIF segment inserted
func (ip *ImageProcessor) writeJPEGWithEXIF(outFile *os.File, jpegData []byte, exifSegment []byte) error {
	if len(jpegData) < 4 || jpegData[0] != 0xFF || jpegData[1] != 0xD8 {
		return fmt.Errorf("invalid JPEG data")
	}

	// Write JPEG SOI marker
	if _, err := outFile.Write(jpegData[0:2]); err != nil {
		return err
	}

	// Write EXIF segment
	if _, err := outFile.Write(exifSegment); err != nil {
		return err
	}

	// Find where to continue writing from processed image (skip SOI and any existing APP segments)
	pos := 2
	for pos < len(jpegData)-1 {
		if jpegData[pos] != 0xFF {
			break
		}
		
		marker := jpegData[pos+1]
		if marker >= 0xE0 && marker <= 0xEF { // APP segments
			if pos+2 >= len(jpegData) {
				break
			}
			length := int(jpegData[pos])<<8 | int(jpegData[pos+1])
			pos += 2 + length
		} else {
			break
		}
	}

	// Write remaining JPEG data
	if pos < len(jpegData) {
		if _, err := outFile.Write(jpegData[pos:]); err != nil {
			return err
		}
	}

	return nil
}
