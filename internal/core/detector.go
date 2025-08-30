package core

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"zensort/internal/config"
)

// FileType represents the category of a file
type FileType int

const (
	FileTypeImage FileType = iota
	FileTypeVideo
	FileTypeAudio
	FileTypeDocument
	FileTypeUnknown
)

// FileTypeDetector handles file type detection using hybrid approach
type FileTypeDetector struct {
	imageExts    map[string]bool
	videoExts    map[string]bool
	audioExts    map[string]bool
	documentExts map[string]bool
	config       *config.Config
}

// NewFileTypeDetector creates a new file type detector
func NewFileTypeDetector() *FileTypeDetector {
	return NewFileTypeDetectorWithConfig(nil)
}

// NewFileTypeDetectorWithConfig creates a new file type detector with configuration
func NewFileTypeDetectorWithConfig(cfg *config.Config) *FileTypeDetector {
	return &FileTypeDetector{
		imageExts: map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
			".tiff": true, ".tif": true, ".webp": true, ".svg": true, ".ico": true,
			".raw": true, ".cr2": true, ".nef": true, ".arw": true, ".dng": true,
			".heic": true, ".heif": true, ".avif": true,
		},
		videoExts: map[string]bool{
			".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".wmv": true,
			".flv": true, ".webm": true, ".m4v": true, ".3gp": true, ".mpg": true,
			".mpeg": true, ".ts": true, ".mts": true, ".m2ts": true,
		},
		audioExts: map[string]bool{
			".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true,
			".wma": true, ".m4a": true, ".opus": true, ".aiff": true, ".au": true,
		},
		documentExts: map[string]bool{
			".pdf": true, ".doc": true, ".docx": true, ".txt": true, ".rtf": true,
			".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true, ".odt": true,
			".ods": true, ".odp": true, ".pages": true, ".numbers": true, ".key": true,
			".html": true, ".htm": true, ".md": true, ".tex": true,
		},
		config: cfg,
	}
}

// DetectFileType determines the file type using hybrid approach
func (d *FileTypeDetector) DetectFileType(filePath string) FileType {
	// Check for Live Photos and Motion Photos first (special case)
	if d.isLiveOrMotionPhoto(filePath) {
		return FileTypeVideo
	}
	
	// First, try extension-based detection (fast)
	ext := strings.ToLower(filepath.Ext(filePath))
	
	if d.imageExts[ext] {
		return FileTypeImage
	}
	if d.videoExts[ext] {
		return FileTypeVideo
	}
	if d.audioExts[ext] {
		return FileTypeAudio
	}
	if d.documentExts[ext] {
		return FileTypeDocument
	}
	
	// If extension is unknown, use MIME type detection (slower but accurate)
	mtype, err := mimetype.DetectFile(filePath)
	if err != nil {
		return FileTypeUnknown
	}
	
	mimeType := mtype.String()
	
	if strings.HasPrefix(mimeType, "image/") {
		return FileTypeImage
	}
	if strings.HasPrefix(mimeType, "video/") {
		return FileTypeVideo
	}
	if strings.HasPrefix(mimeType, "audio/") {
		return FileTypeAudio
	}
	if strings.HasPrefix(mimeType, "text/") || 
	   strings.Contains(mimeType, "document") ||
	   strings.Contains(mimeType, "pdf") ||
	   strings.Contains(mimeType, "spreadsheet") ||
	   strings.Contains(mimeType, "presentation") {
		return FileTypeDocument
	}
	
	return FileTypeUnknown
}

// IsHiddenFile checks if a file is hidden (starts with dot on Unix or has hidden attribute on Windows)
func (d *FileTypeDetector) IsHiddenFile(filePath string) bool {
	filename := filepath.Base(filePath)
	return strings.HasPrefix(filename, ".")
}

// GetFileTypeString returns string representation of file type
func (d *FileTypeDetector) GetFileTypeString(fileType FileType) string {
	switch fileType {
	case FileTypeImage:
		return "Images"
	case FileTypeVideo:
		return "Videos"
	case FileTypeAudio:
		return "Audios"
	case FileTypeDocument:
		return "Documents"
	default:
		return "Unknown"
	}
}

// ShouldSkipFile checks if a file should be skipped based on patterns
func (d *FileTypeDetector) ShouldSkipFile(filePath string, skipExtensions []string, skipPatterns []string, skipDirs []string) bool {
	filename := filepath.Base(filePath)
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Fast path: Check common skip extensions first
	if ext == ".tmp" || ext == ".temp" || ext == ".log" || ext == ".cache" {
		return true
	}
	
	// Special thumb file checks (common case)
	if strings.HasSuffix(ext, ".thumb") {
		return true
	}
	if strings.Contains(ext, ".thumb") && len(ext) > 6 {
		suffix := ext[6:] // Get part after ".thumb"
		if _, err := strconv.Atoi(suffix); err == nil {
			return true
		}
	}
	
	// Check configured extensions
	for _, skipExt := range skipExtensions {
		if ext == strings.ToLower(skipExt) {
			return true
		}
	}
	
	// Check directory patterns (early exit if found)
	if len(skipDirs) > 0 {
		dir := filepath.Dir(filePath)
		for _, skipDir := range skipDirs {
			if strings.Contains(dir, skipDir) {
				return true
			}
		}
	}
	
	// Check filename patterns (most expensive, do last)
	for _, pattern := range skipPatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}
	
	return false
}

// IsLiveOrMotionPhoto detects iPhone Live Photos and Samsung Motion Photos (exported method)
func (d *FileTypeDetector) IsLiveOrMotionPhoto(filePath string) bool {
	return d.isLiveOrMotionPhoto(filePath)
}

// IsMotionPhoto checks if a video file is a Motion Photo
func (d *FileTypeDetector) IsMotionPhoto(filePath string) bool {
	// Only check video files for Motion Photos
	fileType := d.DetectFileType(filePath)
	if fileType != FileTypeVideo {
		return false
	}
	
	// Use config if available, otherwise fall back to defaults
	if d.config != nil && d.config.MotionPhotos.Enabled {
		return d.isMotionPhotoWithConfig(filePath, &d.config.MotionPhotos)
	}
	
	// Fallback to hardcoded patterns if no config
	return d.isMotionPhotoFallback(filePath)
}

// isLiveOrMotionPhoto detects iPhone Live Photos and Samsung Motion Photos using config
func (d *FileTypeDetector) isLiveOrMotionPhoto(filePath string) bool {
	// Use config if available, otherwise fall back to defaults
	if d.config != nil && d.config.MotionPhotos.Enabled {
		return d.isMotionPhotoWithConfig(filePath, &d.config.MotionPhotos)
	}
	
	// Fallback to hardcoded patterns if no config
	return d.isMotionPhotoFallback(filePath)
}

// isMotionPhotoWithConfig uses configuration-based detection for Motion Photos
func (d *FileTypeDetector) isMotionPhotoWithConfig(filePath string, motionConfig *struct {
	Enabled           bool     `json:"enabled"`
	IPhonePatterns    []string `json:"iphone_patterns"`
	SamsungPatterns   []string `json:"samsung_patterns"`
	Extensions        []string `json:"extensions"`
	MaxDurationSeconds int     `json:"max_duration_seconds"`
}) bool {
	if !motionConfig.Enabled {
		return false
	}
	
	filename := strings.ToLower(filepath.Base(filePath))
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Check if extension is supported
	extensionSupported := false
	for _, supportedExt := range motionConfig.Extensions {
		if ext == strings.ToLower(supportedExt) {
			extensionSupported = true
			break
		}
	}
	if !extensionSupported {
		return false
	}
	
	// Check iPhone patterns
	for _, pattern := range motionConfig.IPhonePatterns {
		if strings.Contains(filename, strings.ToLower(pattern)) {
			return true
		}
	}
	
	// Check Samsung patterns
	for _, pattern := range motionConfig.SamsungPatterns {
		if strings.Contains(filename, strings.ToLower(pattern)) {
			return true
		}
	}
	
	return false
}

// isMotionPhotoFallback provides fallback detection when no config is available
func (d *FileTypeDetector) isMotionPhotoFallback(filePath string) bool {
	filename := strings.ToLower(filepath.Base(filePath))
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// iPhone Live Photos detection
	if ext == ".mov" {
		if strings.Contains(filename, "live") ||
		   strings.Contains(filename, "livephoto") ||
		   strings.Contains(filename, "_live") ||
		   strings.HasPrefix(filename, "img_") {
			return true
		}
	}
	
	// Samsung Motion Photos detection
	if ext == ".mp4" {
		if strings.Contains(filename, "motion") ||
		   strings.Contains(filename, "_motion") ||
		   strings.Contains(filename, "motionphoto") ||
		   strings.HasPrefix(filename, "mvimg_") {
			return true
		}
	}
	
	// Remove image file detection - Motion Photos are video-only
	
	return false
}

// IsScreenshot checks if an image file is a screenshot based on filename patterns
func (d *FileTypeDetector) IsScreenshot(filePath string) bool {
	// Only check image files for screenshots
	fileType := d.DetectFileType(filePath)
	if fileType != FileTypeImage {
		return false
	}
	
	// Use config if available, otherwise fall back to defaults
	if d.config != nil && d.config.Screenshots.Enabled {
		return d.isScreenshotWithConfig(filePath, &d.config.Screenshots)
	}
	
	// Fallback to hardcoded patterns if no config
	return d.isScreenshotFallback(filePath)
}

// isScreenshotWithConfig checks screenshot patterns using provided config
func (d *FileTypeDetector) isScreenshotWithConfig(filePath string, screenshotConfig *struct {
	Enabled    bool     `json:"enabled"`
	Patterns   []string `json:"patterns"`
	Extensions []string `json:"extensions"`
	FolderName string   `json:"folder_name"`
}) bool {
	filename := strings.ToLower(filepath.Base(filePath))
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Check if extension is in allowed list
	extensionAllowed := false
	for _, allowedExt := range screenshotConfig.Extensions {
		if ext == strings.ToLower(allowedExt) {
			extensionAllowed = true
			break
		}
	}
	
	if !extensionAllowed {
		return false
	}
	
	// Check patterns
	for _, pattern := range screenshotConfig.Patterns {
		if strings.Contains(filename, strings.ToLower(pattern)) {
			return true
		}
	}
	
	return false
}

// isScreenshotFallback provides fallback detection when no config is available
func (d *FileTypeDetector) isScreenshotFallback(filePath string) bool {
	filename := strings.ToLower(filepath.Base(filePath))
	ext := strings.ToLower(filepath.Ext(filePath))
	
	// Check common image extensions
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return false
	}
	
	// Common screenshot patterns
	patterns := []string{"screenshot", "screen shot", "screen_shot", "screencapture", "screen capture"}
	for _, pattern := range patterns {
		if strings.Contains(filename, pattern) {
			return true
		}
	}
	
	return false
}
