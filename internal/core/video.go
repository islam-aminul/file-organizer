package core

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// VideoMetadata represents extracted video metadata
type VideoMetadata struct {
	Make         string
	Model        string
	CreationTime time.Time
	HasDateTime  bool
	Duration     time.Duration
}

// VideoAnalyzer handles video file analysis
type VideoAnalyzer struct{}

// NewVideoAnalyzer creates a new video analyzer
func NewVideoAnalyzer() *VideoAnalyzer {
	return &VideoAnalyzer{}
}

// GetVideoDuration extracts video duration using ffprobe if available
func (va *VideoAnalyzer) GetVideoDuration(filePath string) (time.Duration, error) {
	// Try ffprobe first (most accurate)
	if duration, err := va.getFFProbeDuration(filePath); err == nil {
		return duration, nil
	}
	
	// Fallback: estimate from file size and common bitrates
	return va.estimateDurationFromSize(filePath)
}

// getFFProbeDuration uses ffprobe to get exact video duration
func (va *VideoAnalyzer) getFFProbeDuration(filePath string) (time.Duration, error) {
	// Check if ffprobe is available
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "format=duration", "-of", "csv=p=0", filePath)
	
	// Hide console window on Windows to prevent flashing
	hideConsoleWindow(cmd)
	
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe not available or failed: %w", err)
	}
	
	// Parse duration
	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}
	
	return time.Duration(durationFloat * float64(time.Second)), nil
}

// hideConsoleWindow hides the console window on Windows to prevent flashing
func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

// estimateDurationFromSize provides a rough estimate based on file size
func (va *VideoAnalyzer) estimateDurationFromSize(filePath string) (time.Duration, error) {
	// This is a fallback estimation - not very accurate but better than nothing
	// We'll use conservative estimates to avoid misclassifying long videos as short
	
	// For other videos, we can't reliably estimate without metadata
	// Return a duration that won't trigger short video classification
	// Use a high value to ensure videos with unknown duration go to regular Videos folder
	return 300 * time.Second, nil // 5 minutes - well above typical short video thresholds
}

// IsShortVideo determines if a video is shorter than the threshold
func (va *VideoAnalyzer) IsShortVideo(filePath string, thresholdSeconds int) bool {
	if thresholdSeconds <= 0 {
		return false // Feature disabled
	}
	
	duration, err := va.GetVideoDuration(filePath)
	if err != nil {
		// If we can't determine duration, don't classify as short video
		// This prevents misclassification of regular videos
		return false
	}
	
	threshold := time.Duration(thresholdSeconds) * time.Second
	
	// Debug logging to help identify the issue
	fmt.Printf("DEBUG: Video %s - Duration: %.2f seconds, Threshold: %d seconds, IsShort: %v\n", 
		filePath, duration.Seconds(), thresholdSeconds, duration < threshold)
	
	return duration < threshold
}

// ExtractVideoMetadata extracts metadata from video file using ffprobe
func (va *VideoAnalyzer) ExtractVideoMetadata(filePath string) (*VideoMetadata, error) {
	// Use ffprobe to get comprehensive metadata
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filePath)
	
	// Hide console window on Windows
	hideConsoleWindow(cmd)
	
	output, err := cmd.Output()
	if err != nil {
		return &VideoMetadata{}, fmt.Errorf("ffprobe not available or failed: %w", err)
	}
	
	// Parse JSON output
	var probe struct {
		Format struct {
			Duration string            `json:"duration"`
			Tags     map[string]string `json:"tags"`
		} `json:"format"`
		Streams []struct {
			Tags map[string]string `json:"tags"`
		} `json:"streams"`
	}
	
	if err := json.Unmarshal(output, &probe); err != nil {
		return &VideoMetadata{}, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}
	
	metadata := &VideoMetadata{}
	
	// Extract duration
	if probe.Format.Duration != "" {
		if durationFloat, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
			metadata.Duration = time.Duration(durationFloat * float64(time.Second))
		}
	}
	
	// Extract tags from format and streams
	allTags := make(map[string]string)
	
	// Add format tags
	for key, value := range probe.Format.Tags {
		allTags[strings.ToLower(key)] = value
	}
	
	// Add stream tags
	for _, stream := range probe.Streams {
		for key, value := range stream.Tags {
			allTags[strings.ToLower(key)] = value
		}
	}
	
	// Extract metadata using tag patterns
	extractedData := extractVideoTags(allTags)
	metadata.Make = extractedData["make"]
	metadata.Model = extractedData["model"]
	
	if creationTime, exists := extractedData["creation_time"]; exists {
		if parsedTime := parseVideoDateTime(creationTime); !parsedTime.IsZero() {
			metadata.CreationTime = parsedTime
			metadata.HasDateTime = true
		}
	}
	
	return metadata, nil
}

// extractVideoTags extracts relevant metadata from video tags
func extractVideoTags(tags map[string]string) map[string]string {
	videoDetails := make(map[string]string)
	
	for tag, value := range tags {
		// Check for make/manufacturer
		if strings.Contains(tag, "make") || strings.Contains(tag, "manufacturer") {
			videoDetails["make"] = value
		}
		// Check for model
		if strings.Contains(tag, "model") {
			videoDetails["model"] = value
		}
		// Check for creation date/time
		if strings.Contains(tag, "creationdate") || strings.Contains(tag, "creation_time") || 
		   strings.Contains(tag, "date") || strings.Contains(tag, "datetime") {
			videoDetails["creation_time"] = value
		}
	}
	
	return videoDetails
}

// parseVideoDateTime attempts to parse video datetime with multiple formats
func parseVideoDateTime(dateTimeStr string) time.Time {
	// Common video metadata datetime formats
	formats := []string{
		"2006-01-02T15:04:05.000000Z",  // ISO with microseconds
		"2006-01-02T15:04:05Z",         // ISO basic
		"2006-01-02 15:04:05",          // Standard format
		"2006:01:02 15:04:05",          // EXIF-like format
		"2006-01-02T15:04:05",          // ISO without timezone
		"2006/01/02 15:04:05",          // Slash format
		"Mon Jan 2 15:04:05 2006",      // RFC822 format
	}
	
	for _, format := range formats {
		if parsedTime, err := time.Parse(format, dateTimeStr); err == nil {
			return parsedTime
		}
	}
	
	return time.Time{} // Return zero time if no format matches
}
