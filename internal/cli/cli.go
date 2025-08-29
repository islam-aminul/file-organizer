package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"zensort/internal/config"
	"zensort/internal/core"
)

// Run executes the CLI version of the file organizer
func Run(sourceDir, destDir, configFile string) {
	fmt.Println("ZenSort File Organizer - CLI Mode")
	fmt.Println("================================")
	
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	
	// Create processor
	processor, err := core.NewFileProcessor(cfg, destDir)
	if err != nil {
		fmt.Printf("Error creating processor: %v\n", err)
		os.Exit(1)
	}
	
	// Subscribe to progress updates
	progressChan := processor.GetProgressTracker().Subscribe()
	
	// Start progress monitoring in background
	go monitorProgress(progressChan)
	
	// Create context
	ctx := context.Background()
	
	fmt.Printf("Source: %s\n", sourceDir)
	fmt.Printf("Destination: %s\n", destDir)
	if configFile != "" {
		fmt.Printf("Config: %s\n", configFile)
	}
	fmt.Printf("Workers: %d\n", processor.GetWorkerCount())
	fmt.Println()
	
	// Start processing
	startTime := time.Now()
	err = processor.ProcessDirectory(ctx, sourceDir)
	duration := time.Since(startTime)
	
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("\nProcessing completed in %v\n", duration)
	fmt.Println("Check the destination directory for detailed logs and reports.")
}

// monitorProgress displays progress updates in CLI
func monitorProgress(progressChan <-chan core.ProgressUpdate) {
	var lastUpdate time.Time
	
	for update := range progressChan {
		// Throttle updates to avoid spam
		if time.Since(lastUpdate) < 500*time.Millisecond && !update.Done {
			continue
		}
		lastUpdate = time.Now()
		
		// Clear line and show progress
		fmt.Printf("\r\033[K") // Clear line
		
		if update.TotalFiles > 0 {
			fmt.Printf("Progress: %d/%d files (%.1f%%) - %s",
				update.ProcessedFiles, update.TotalFiles, update.Percentage,
				formatDuration(update.ElapsedTime))
			
			if update.EstimatedTime > 0 {
				fmt.Printf(" - ETA: %s", formatDuration(update.EstimatedTime))
			}
			
			if update.FilesPerSecond > 0 {
				fmt.Printf(" - %.1f files/s", update.FilesPerSecond)
			}
		}
		
		if update.Done {
			fmt.Printf("\n✓ Complete! Processed %d files", update.ProcessedFiles)
			if update.ErrorCount > 0 {
				fmt.Printf(" (%d errors)", update.ErrorCount)
			}
			fmt.Println()
			break
		}
	}
}

// formatDuration formats a duration for CLI display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-60*d.Minutes())
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - 60*hours
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}
