package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zensort/internal/config"
)

// FileProcessor handles the main file organization logic
type FileProcessor struct {
	config          *config.Config
	destDir         string
	db              *Database
	detector        *FileTypeDetector
	workerPool      *WorkerPool
	progressTracker *ProgressTracker
	logger          *Logger
	reportGen       *ReportGenerator
	categoryStats   map[string]CategoryStats
}

// NewFileProcessor creates a new file processor
func NewFileProcessor(cfg *config.Config, destDir string) (*FileProcessor, error) {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}
	
	// Initialize database
	db, err := NewDatabase(destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	// Initialize logger
	logger, err := NewLogger(destDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	
	fp := &FileProcessor{
		config:          cfg,
		destDir:         destDir,
		db:              db,
		detector:        NewFileTypeDetector(),
		progressTracker: NewProgressTracker(),
		logger:          logger,
		reportGen:       NewReportGenerator(destDir),
		categoryStats:   make(map[string]CategoryStats),
	}
	
	return fp, nil
}

// ProcessDirectory processes all files in the source directory
func (fp *FileProcessor) ProcessDirectory(ctx context.Context, sourceDir string) error {
	startTime := time.Now()
	
	// Create worker pool
	fp.workerPool = NewWorkerPool(ctx)
	defer fp.workerPool.Stop()
	
	fp.logger.LogOperation("INFO", fmt.Sprintf("Starting processing with %d workers", fp.workerPool.WorkerCount()), sourceDir)
	
	// Scan directory to get file list and total size
	files, totalSize, err := fp.scanDirectory(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}
	
	fp.progressTracker.SetTotal(int64(len(files)), totalSize)
	fp.logger.LogOperation("INFO", fmt.Sprintf("Found %d files (%s total)", len(files), formatBytes(totalSize)), "")
	
	// Start worker pool
	fp.workerPool.Start()
	
	// Process files
	stats := ProcessingStats{
		StartTime:  startTime,
		TotalFiles: int64(len(files)),
		TotalSize:  totalSize,
	}
	
	err = fp.processFiles(ctx, files, &stats)
	if err != nil {
		return err
	}
	
	// Finalize processing
	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)
	
	fp.progressTracker.SetDone()
	fp.logger.LogStatistics(stats)
	
	// Generate final report
	errors := fp.progressTracker.GetErrors()
	err = fp.reportGen.GenerateReport(stats, fp.categoryStats, errors, fp.workerPool.WorkerCount(), sourceDir, "")
	if err != nil {
		fp.logger.LogError(LogLevelWarning, "Failed to generate report", "", err)
	}
	
	return nil
}

// scanDirectory recursively scans directory and returns file list
func (fp *FileProcessor) scanDirectory(sourceDir string) ([]string, int64, error) {
	var files []string
	var totalSize int64
	
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fp.logger.LogError(LogLevelWarning, "Error accessing file", path, err)
			return nil // Continue processing other files
		}
		
		// Skip directories
		if info.IsDir() {
			return nil
		}
		
		// Check if file should be skipped
		if fp.detector.ShouldSkipFile(path, fp.config.SkipFiles.Extensions, fp.config.SkipFiles.Patterns, fp.config.SkipFiles.Directories) {
			fp.logger.LogFileSkipped(path, "matches skip pattern")
			return nil
		}
		
		files = append(files, path)
		totalSize += info.Size()
		return nil
	})
	
	return files, totalSize, err
}

// processFiles processes the list of files using worker pool
func (fp *FileProcessor) processFiles(ctx context.Context, files []string, stats *ProcessingStats) error {
	// Create file organizer
	organizer := NewFileOrganizer(fp.config, fp.destDir, fp.db, fp.logger)
	
	// Process files directly (simplified approach)
	for _, filePath := range files {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Get file info for progress tracking
			fileInfo, err := os.Stat(filePath)
			var fileSize int64
			if err == nil {
				fileSize = fileInfo.Size()
			}
			
			// Process the file
			err = organizer.OrganizeFile(filePath)
			if err != nil {
				stats.ErrorFiles++
				fp.logger.LogError(LogLevelError, "Failed to process file", filePath, err)
				fp.progressTracker.AddError(fmt.Sprintf("Error processing %s: %v", filePath, err))
			} else {
				stats.ProcessedFiles++
				stats.ProcessedSize += fileSize
			}
			
			// Update progress
			fp.progressTracker.IncrementProgress(fileSize, filePath)
		}
	}
	
	return nil
}

// GetProgressTracker returns the progress tracker
func (fp *FileProcessor) GetProgressTracker() *ProgressTracker {
	return fp.progressTracker
}

// GetWorkerCount returns the number of workers
func (fp *FileProcessor) GetWorkerCount() int {
	if fp.workerPool != nil {
		return fp.workerPool.WorkerCount()
	}
	return calculateOptimalWorkers()
}

// Close closes the processor and cleans up resources
func (fp *FileProcessor) Close() error {
	var err error
	
	if fp.logger != nil {
		if closeErr := fp.logger.Close(); closeErr != nil {
			err = closeErr
		}
	}
	
	if fp.db != nil {
		if closeErr := fp.db.Close(); closeErr != nil {
			err = closeErr
		}
	}
	
	return err
}

