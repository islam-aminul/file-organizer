package core

import (
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
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
}

// NewFileTypeDetector creates a new file type detector
func NewFileTypeDetector() *FileTypeDetector {
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
			".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
			".ppt": true, ".pptx": true, ".txt": true, ".rtf": true, ".odt": true,
			".ods": true, ".odp": true, ".csv": true, ".xml": true, ".json": true,
			".html": true, ".htm": true, ".md": true, ".tex": true,
		},
	}
}

// DetectFileType determines the file type using hybrid approach
func (d *FileTypeDetector) DetectFileType(filePath string) FileType {
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
	dir := filepath.Dir(filePath)
	
	// Check extensions
	for _, skipExt := range skipExtensions {
		if ext == strings.ToLower(skipExt) {
			return true
		}
	}
	
	// Check filename patterns
	for _, pattern := range skipPatterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}
	
	// Check directory patterns
	for _, skipDir := range skipDirs {
		if strings.Contains(dir, skipDir) {
			return true
		}
	}
	
	return false
}
