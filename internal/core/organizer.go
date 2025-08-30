package core

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zensort/internal/config"
)


// FileOrganizer handles the actual file organization logic
type FileOrganizer struct {
	config   *config.Config
	destDir  string
	detector *FileTypeDetector
	db       *Database
	logger   *Logger
}

// NewFileOrganizer creates a new file organizer
func NewFileOrganizer(cfg *config.Config, destDir string, db *Database, logger *Logger) *FileOrganizer {
	return &FileOrganizer{
		config:   cfg,
		destDir:  destDir,
		detector: NewFileTypeDetectorWithConfig(cfg),
		db:       db,
		logger:   logger,
	}
}

// OrganizeFile processes and organizes a single file
func (fo *FileOrganizer) OrganizeFile(sourcePath string) error {
	// Get file info
	fileInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Skip directories
	if fileInfo.IsDir() {
		return nil
	}

	// Check if file should be skipped
	if fo.detector.ShouldSkipFile(sourcePath, fo.config.SkipFiles.Extensions, fo.config.SkipFiles.Patterns, fo.config.SkipFiles.Directories) {
		fo.logger.LogFileSkipped(sourcePath, "matches skip pattern")
		return nil
	}

	// Calculate file hash for duplicate detection
	hash, err := calculateFileHash(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to calculate file hash: %w", err)
	}

	// Check for duplicates
	isDuplicate, existingPath, err := fo.db.CheckDuplicate(hash)
	if err != nil {
		return fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if isDuplicate {
		fo.logger.LogFileDuplicate(sourcePath, existingPath, hash)
		return nil // Skip duplicate files
	}

	// Detect file type
	fileType := fo.detector.DetectFileType(sourcePath)
	
	// Determine destination path
	destPath, err := fo.getDestinationPath(sourcePath, fileType, fileInfo)
	if err != nil {
		return fmt.Errorf("failed to determine destination path: %w", err)
	}

	// Create destination directory
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Handle naming conflicts
	finalDestPath := fo.resolveNamingConflict(destPath)

	// Copy file to destination
	if err := fo.copyFile(sourcePath, finalDestPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Add to database
	if err := fo.db.AddFile(hash, sourcePath, finalDestPath, fileInfo.Size()); err != nil {
		fo.logger.LogError(LogLevelWarning, "Failed to add file to database", sourcePath, err)
		// Don't fail the operation if database update fails
	}

	// Log successful processing
	fo.logger.LogFileProcessed(sourcePath, finalDestPath, hash, fileInfo.Size())

	return nil
}

// getDestinationPath determines where a file should be placed
func (fo *FileOrganizer) getDestinationPath(sourcePath string, fileType FileType, fileInfo os.FileInfo) (string, error) {
	filename := filepath.Base(sourcePath)
	
	// Check if file is hidden
	isHidden := fo.detector.IsHiddenFile(sourcePath)
	
	// Handle hidden files first (common case optimization)
	if isHidden {
		var categoryDir string
		switch fileType {
		case FileTypeImage:
			categoryDir = fo.config.Directories.Images
		case FileTypeVideo:
			categoryDir = fo.config.Directories.Videos
		case FileTypeAudio:
			categoryDir = fo.config.Directories.Audios
		case FileTypeDocument:
			categoryDir = fo.config.Directories.Documents
		default:
			categoryDir = fo.config.Directories.Unknown
		}
		return filepath.Join(fo.destDir, categoryDir, fo.config.Directories.Hidden, filename), nil
	}

	// Handle non-hidden files by type
	switch fileType {
	case FileTypeImage:
		categoryDir := fo.config.Directories.Images
		// For images, use EXIF-based organization
		exifData, err := ExtractEXIF(sourcePath)
		if err != nil {
			// No EXIF data, use collections folder
			return filepath.Join(fo.destDir, categoryDir, fo.config.ImageDirs.Originals, "Collections", filename), nil
		}
		return GetImageDestinationPath(fo.destDir, filename, exifData, fo.config, false), nil
		
	case FileTypeVideo:
		categoryDir := fo.config.Directories.Videos
		year := fo.extractYear(fileInfo.ModTime())
		
		// Special handling for Motion Photos (video files with specific patterns)
		if fo.detector.IsMotionPhoto(sourcePath) {
			// Organize Motion Photos: Videos/Motion Photos/Year/
			return filepath.Join(fo.destDir, categoryDir, "Motion Photos", year, filename), nil
		}
		
		// Check if it's a short video (after Live Photo detection)
		if fo.config.Processing.ShortVideoThreshold > 0 {
			videoAnalyzer := NewVideoAnalyzer()
			if videoAnalyzer.IsShortVideo(sourcePath, fo.config.Processing.ShortVideoThreshold) {
				// Organize Short Videos: Videos/Short Videos/Year/
				return filepath.Join(fo.destDir, categoryDir, "Short Videos", year, filename), nil
			}
		}
		
		// Regular video organization by year
		return filepath.Join(fo.destDir, categoryDir, year, filename), nil
		
	case FileTypeAudio:
		categoryDir := fo.config.Directories.Audios
		// Categorize audio files
		audioCategory := fo.categorizeAudio(filename)
		return filepath.Join(fo.destDir, categoryDir, audioCategory, filename), nil
		
	case FileTypeDocument:
		categoryDir := fo.config.Directories.Documents
		// Organize by file extension
		ext := strings.ToUpper(strings.TrimPrefix(filepath.Ext(filename), "."))
		if ext == "" {
			ext = "Other Documents"
		}
		return filepath.Join(fo.destDir, categoryDir, ext, filename), nil
		
	default:
		categoryDir := fo.config.Directories.Unknown
		return filepath.Join(fo.destDir, categoryDir, filename), nil
	}
}

// extractYear extracts year from file modification time
func (fo *FileOrganizer) extractYear(modTime time.Time) string {
	if modTime.IsZero() {
		return "0000"
	}
	return fmt.Sprintf("%04d", modTime.Year())
}

// categorizeAudio categorizes audio files based on configurable patterns and extensions
func (fo *FileOrganizer) categorizeAudio(filename string) string {
	lowerName := strings.ToLower(filename)
	ext := strings.ToLower(filepath.Ext(filename))
	
	// First pass: Check extension + pattern combinations (most specific)
	for _, category := range fo.config.AudioCategories {
		// Quick extension check first
		hasMatchingExt := false
		for _, configExt := range category.Extensions {
			if ext == strings.ToLower(configExt) {
				hasMatchingExt = true
				break
			}
		}
		
		if hasMatchingExt {
			// Extension matches - check patterns if any
			if len(category.Patterns) == 0 {
				// No patterns required, extension match is sufficient
				return category.FolderName
			}
			
			// Check patterns for this category
			for _, pattern := range category.Patterns {
				if strings.Contains(lowerName, strings.ToLower(pattern)) {
					return category.FolderName
				}
			}
		}
	}
	
	// Second pass: Pattern-only matching (fallback)
	for _, category := range fo.config.AudioCategories {
		for _, pattern := range category.Patterns {
			if strings.Contains(lowerName, strings.ToLower(pattern)) {
				return category.FolderName
			}
		}
	}
	
	// Third pass: Extension-only matching for categories without patterns
	for _, category := range fo.config.AudioCategories {
		if len(category.Patterns) == 0 {
			for _, configExt := range category.Extensions {
				if ext == strings.ToLower(configExt) {
					return category.FolderName
				}
			}
		}
	}
	
	// Default fallbacks with early returns
	if songsCategory, exists := fo.config.AudioCategories["songs"]; exists {
		return songsCategory.FolderName
	}
	
	// Return first available category
	for _, category := range fo.config.AudioCategories {
		return category.FolderName
	}
	
	return "Songs" // Ultimate fallback
}

// resolveNamingConflict handles file naming conflicts by appending " -- n"
func (fo *FileOrganizer) resolveNamingConflict(destPath string) string {
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		return destPath // No conflict
	}
	
	dir := filepath.Dir(destPath)
	filename := filepath.Base(destPath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)
	
	counter := 1
	for {
		newFilename := fmt.Sprintf("%s -- %d%s", nameWithoutExt, counter, ext)
		newPath := filepath.Join(dir, newFilename)
		
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

// copyFile copies a file from source to destination with image processing support
func (fo *FileOrganizer) copyFile(src, dst string) error {
	// Check if this is an image file that needs special processing
	if IsImageFile(src) {
		// Check if this is a hidden file - hidden images should not be exported
		isHidden := fo.detector.IsHiddenFile(src)
		if isHidden {
			// Hidden images get regular copy only (no exports)
			return fo.regularCopy(src, dst)
		}
		
		// Extract EXIF data for image processing
		exifData, err := ExtractEXIF(src)
		if err != nil {
			// If EXIF extraction fails, fall back to regular copy
			return fo.regularCopy(src, dst)
		}
		
		// Use image processor for images
		imageProcessor := NewImageProcessor(fo.config)
		return imageProcessor.ProcessImage(src, dst, exifData)
	}
	
	// Regular file copy for non-images
	return fo.regularCopy(src, dst)
}

// regularCopy performs a standard file copy
func (fo *FileOrganizer) regularCopy(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}

// copyFile is a utility function for image processing
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}
