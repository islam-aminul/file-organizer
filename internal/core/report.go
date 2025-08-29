package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// StatusReport contains the final processing report
type StatusReport struct {
	SessionInfo struct {
		StartTime    time.Time `json:"start_time"`
		EndTime      time.Time `json:"end_time"`
		Duration     string    `json:"duration"`
		SourceDir    string    `json:"source_directory"`
		DestDir      string    `json:"destination_directory"`
		ConfigFile   string    `json:"config_file"`
		WorkerCount  int       `json:"worker_count"`
	} `json:"session_info"`
	
	FileCounts struct {
		Total       int64 `json:"total_files"`
		Processed   int64 `json:"processed_files"`
		Skipped     int64 `json:"skipped_files"`
		Duplicates  int64 `json:"duplicate_files"`
		Errors      int64 `json:"error_files"`
	} `json:"file_counts"`
	
	SizeInfo struct {
		TotalBytes     int64  `json:"total_bytes"`
		ProcessedBytes int64  `json:"processed_bytes"`
		TotalHuman     string `json:"total_human"`
		ProcessedHuman string `json:"processed_human"`
	} `json:"size_info"`
	
	CategoryBreakdown map[string]CategoryStats `json:"category_breakdown"`
	
	Performance struct {
		FilesPerSecond  float64 `json:"files_per_second"`
		BytesPerSecond  int64   `json:"bytes_per_second"`
		ThroughputHuman string  `json:"throughput_human"`
	} `json:"performance"`
	
	ErrorSummary []ErrorInfo `json:"error_summary"`
}

// CategoryStats holds statistics for each file category
type CategoryStats struct {
	Count int64 `json:"count"`
	Size  int64 `json:"size"`
	SizeHuman string `json:"size_human"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Type        string `json:"type"`
	Count       int    `json:"count"`
	SampleFiles []string `json:"sample_files"`
}

// ReportGenerator creates final status reports
type ReportGenerator struct {
	destDir string
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(destDir string) *ReportGenerator {
	return &ReportGenerator{
		destDir: destDir,
	}
}

// GenerateReport creates and saves the final status report
func (rg *ReportGenerator) GenerateReport(stats ProcessingStats, categoryStats map[string]CategoryStats, errors []string, workerCount int, sourceDir, configFile string) error {
	report := StatusReport{}
	
	// Session info
	report.SessionInfo.StartTime = stats.StartTime
	report.SessionInfo.EndTime = stats.EndTime
	report.SessionInfo.Duration = stats.Duration.String()
	report.SessionInfo.SourceDir = sourceDir
	report.SessionInfo.DestDir = rg.destDir
	report.SessionInfo.ConfigFile = configFile
	report.SessionInfo.WorkerCount = workerCount
	
	// File counts
	report.FileCounts.Total = stats.TotalFiles
	report.FileCounts.Processed = stats.ProcessedFiles
	report.FileCounts.Skipped = stats.SkippedFiles
	report.FileCounts.Duplicates = stats.DuplicateFiles
	report.FileCounts.Errors = stats.ErrorFiles
	
	// Size info
	report.SizeInfo.TotalBytes = stats.TotalSize
	report.SizeInfo.ProcessedBytes = stats.ProcessedSize
	report.SizeInfo.TotalHuman = formatBytes(stats.TotalSize)
	report.SizeInfo.ProcessedHuman = formatBytes(stats.ProcessedSize)
	
	// Category breakdown
	report.CategoryBreakdown = make(map[string]CategoryStats)
	for category, catStats := range categoryStats {
		report.CategoryBreakdown[category] = CategoryStats{
			Count:     catStats.Count,
			Size:      catStats.Size,
			SizeHuman: formatBytes(catStats.Size),
		}
	}
	
	// Performance metrics
	if stats.Duration.Seconds() > 0 {
		report.Performance.FilesPerSecond = float64(stats.ProcessedFiles) / stats.Duration.Seconds()
		report.Performance.BytesPerSecond = int64(float64(stats.ProcessedSize) / stats.Duration.Seconds())
		report.Performance.ThroughputHuman = formatBytes(report.Performance.BytesPerSecond) + "/s"
	}
	
	// Error summary
	report.ErrorSummary = rg.summarizeErrors(errors)
	
	// Save JSON report
	if err := rg.saveJSONReport(report); err != nil {
		return fmt.Errorf("failed to save JSON report: %w", err)
	}
	
	// Save human-readable report
	if err := rg.saveTextReport(report); err != nil {
		return fmt.Errorf("failed to save text report: %w", err)
	}
	
	return nil
}

// saveJSONReport saves the report in JSON format
func (rg *ReportGenerator) saveJSONReport(report StatusReport) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("zensort-report_%s.json", timestamp)
	filepath := filepath.Join(rg.destDir, "zensort-logs", filename)
	
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath, data, 0644)
}

// saveTextReport saves a human-readable report
func (rg *ReportGenerator) saveTextReport(report StatusReport) error {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("zensort-report_%s.txt", timestamp)
	filepath := filepath.Join(rg.destDir, "zensort-logs", filename)
	
	content := rg.formatTextReport(report)
	return os.WriteFile(filepath, []byte(content), 0644)
}

// formatTextReport creates a human-readable report
func (rg *ReportGenerator) formatTextReport(report StatusReport) string {
	content := fmt.Sprintf(`ZenSort Processing Report
========================

Session Information:
  Start Time: %s
  End Time: %s
  Duration: %s
  Source Directory: %s
  Destination Directory: %s
  Configuration File: %s
  Worker Threads: %d

File Processing Summary:
  Total Files Found: %d
  Successfully Processed: %d
  Skipped Files: %d
  Duplicate Files: %d
  Files with Errors: %d

Data Processing Summary:
  Total Data Size: %s (%d bytes)
  Processed Data Size: %s (%d bytes)

Performance Metrics:
  Processing Speed: %.2f files/second
  Data Throughput: %s

Category Breakdown:
`,
		report.SessionInfo.StartTime.Format("2006-01-02 15:04:05"),
		report.SessionInfo.EndTime.Format("2006-01-02 15:04:05"),
		report.SessionInfo.Duration,
		report.SessionInfo.SourceDir,
		report.SessionInfo.DestDir,
		report.SessionInfo.ConfigFile,
		report.SessionInfo.WorkerCount,
		report.FileCounts.Total,
		report.FileCounts.Processed,
		report.FileCounts.Skipped,
		report.FileCounts.Duplicates,
		report.FileCounts.Errors,
		report.SizeInfo.TotalHuman,
		report.SizeInfo.TotalBytes,
		report.SizeInfo.ProcessedHuman,
		report.SizeInfo.ProcessedBytes,
		report.Performance.FilesPerSecond,
		report.Performance.ThroughputHuman,
	)
	
	// Add category breakdown
	for category, stats := range report.CategoryBreakdown {
		content += fmt.Sprintf("  %s: %d files (%s)\n", category, stats.Count, stats.SizeHuman)
	}
	
	// Add error summary if there are errors
	if len(report.ErrorSummary) > 0 {
		content += "\nError Summary:\n"
		for _, errorInfo := range report.ErrorSummary {
			content += fmt.Sprintf("  %s: %d occurrences\n", errorInfo.Type, errorInfo.Count)
			if len(errorInfo.SampleFiles) > 0 {
				content += "    Sample files:\n"
				for _, file := range errorInfo.SampleFiles {
					content += fmt.Sprintf("      - %s\n", file)
				}
			}
		}
	}
	
	content += "\nFor detailed logs, check the zensort-logs directory.\n"
	
	return content
}

// summarizeErrors creates a summary of errors
func (rg *ReportGenerator) summarizeErrors(errors []string) []ErrorInfo {
	errorMap := make(map[string][]string)
	
	// Group errors by type (simplified approach)
	for _, err := range errors {
		errorType := "General Error"
		if len(err) > 50 {
			errorType = err[:50] + "..."
		} else {
			errorType = err
		}
		
		errorMap[errorType] = append(errorMap[errorType], err)
	}
	
	var errorSummary []ErrorInfo
	for errorType, errorList := range errorMap {
		info := ErrorInfo{
			Type:  errorType,
			Count: len(errorList),
		}
		
		// Add up to 3 sample files
		sampleCount := min(len(errorList), 3)
		info.SampleFiles = errorList[:sampleCount]
		
		errorSummary = append(errorSummary, info)
	}
	
	return errorSummary
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}
