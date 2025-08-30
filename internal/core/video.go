package core

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

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
