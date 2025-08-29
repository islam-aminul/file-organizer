package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger handles detailed error and operation logging
type Logger struct {
	errorLog    *log.Logger
	operationLog *log.Logger
	errorFile   *os.File
	opFile      *os.File
	destDir     string
}

// LogLevel defines the severity of log messages
type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelWarning
	LogLevelError
	LogLevelCritical
)

// NewLogger creates a new logger that writes to the destination directory
func NewLogger(destDir string) (*Logger, error) {
	// Create logs directory in destination
	logDir := filepath.Join(destDir, "zensort-logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	
	// Create error log file
	errorLogPath := filepath.Join(logDir, fmt.Sprintf("errors_%s.log", timestamp))
	errorFile, err := os.OpenFile(errorLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create error log file: %w", err)
	}
	
	// Create operation log file
	opLogPath := filepath.Join(logDir, fmt.Sprintf("operations_%s.log", timestamp))
	opFile, err := os.OpenFile(opLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		errorFile.Close()
		return nil, fmt.Errorf("failed to create operation log file: %w", err)
	}
	
	logger := &Logger{
		errorLog:    log.New(errorFile, "", log.LstdFlags|log.Lmicroseconds),
		operationLog: log.New(opFile, "", log.LstdFlags|log.Lmicroseconds),
		errorFile:   errorFile,
		opFile:      opFile,
		destDir:     destDir,
	}
	
	// Log session start
	logger.LogOperation("INFO", "ZenSort session started", "")
	
	return logger, nil
}

// LogError logs an error message
func (l *Logger) LogError(level LogLevel, message string, filePath string, err error) {
	levelStr := l.levelToString(level)
	
	var logMsg string
	if err != nil {
		logMsg = fmt.Sprintf("[%s] %s | File: %s | Error: %v", levelStr, message, filePath, err)
	} else {
		logMsg = fmt.Sprintf("[%s] %s | File: %s", levelStr, message, filePath)
	}
	
	l.errorLog.Println(logMsg)
}

// LogOperation logs a general operation
func (l *Logger) LogOperation(level string, message string, filePath string) {
	logMsg := fmt.Sprintf("[%s] %s", level, message)
	if filePath != "" {
		logMsg += fmt.Sprintf(" | File: %s", filePath)
	}
	
	l.operationLog.Println(logMsg)
}

// LogFileProcessed logs a successfully processed file
func (l *Logger) LogFileProcessed(sourcePath, destPath, hash string, size int64) {
	l.LogOperation("SUCCESS", 
		fmt.Sprintf("File processed - Size: %d bytes, Hash: %s, Destination: %s", size, hash, destPath), 
		sourcePath)
}

// LogFileDuplicate logs a duplicate file detection
func (l *Logger) LogFileDuplicate(sourcePath, existingPath, hash string) {
	l.LogOperation("DUPLICATE", 
		fmt.Sprintf("Duplicate detected - Hash: %s, Existing: %s", hash, existingPath), 
		sourcePath)
}

// LogFileSkipped logs a skipped file
func (l *Logger) LogFileSkipped(filePath, reason string) {
	l.LogOperation("SKIPPED", fmt.Sprintf("Reason: %s", reason), filePath)
}

// LogStatistics logs processing statistics
func (l *Logger) LogStatistics(stats ProcessingStats) {
	l.LogOperation("STATS", fmt.Sprintf(
		"Processing complete - Total: %d, Processed: %d, Skipped: %d, Duplicates: %d, Errors: %d, Duration: %v",
		stats.TotalFiles, stats.ProcessedFiles, stats.SkippedFiles, stats.DuplicateFiles, stats.ErrorFiles, stats.Duration), "")
}

// Close closes the log files
func (l *Logger) Close() error {
	l.LogOperation("INFO", "ZenSort session ended", "")
	
	var err1, err2 error
	if l.errorFile != nil {
		err1 = l.errorFile.Close()
	}
	if l.opFile != nil {
		err2 = l.opFile.Close()
	}
	
	if err1 != nil {
		return err1
	}
	return err2
}

// levelToString converts LogLevel to string
func (l *Logger) levelToString(level LogLevel) string {
	switch level {
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarning:
		return "WARNING"
	case LogLevelError:
		return "ERROR"
	case LogLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ProcessingStats holds statistics about the processing operation
type ProcessingStats struct {
	TotalFiles     int64
	ProcessedFiles int64
	SkippedFiles   int64
	DuplicateFiles int64
	ErrorFiles     int64
	TotalSize      int64
	ProcessedSize  int64
	Duration       time.Duration
	StartTime      time.Time
	EndTime        time.Time
}
