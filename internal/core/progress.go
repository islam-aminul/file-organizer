package core

import (
	"sync"
	"time"
)

// ProgressTracker manages progress reporting for file operations
type ProgressTracker struct {
	mu                sync.RWMutex
	totalFiles        int64
	processedFiles    int64
	totalSize         int64
	processedSize     int64
	currentFile       string
	startTime         time.Time
	errors            []string
	subscribers       []chan ProgressUpdate
	done              bool
}

// ProgressUpdate contains current progress information
type ProgressUpdate struct {
	TotalFiles      int64
	ProcessedFiles  int64
	TotalSize       int64
	ProcessedSize   int64
	CurrentFile     string
	Percentage      float64
	ElapsedTime     time.Duration
	EstimatedTime   time.Duration
	FilesPerSecond  float64
	BytesPerSecond  float64
	ErrorCount      int
	Done            bool
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker() *ProgressTracker {
	return &ProgressTracker{
		startTime:   time.Now(),
		subscribers: make([]chan ProgressUpdate, 0),
	}
}

// SetTotal sets the total number of files and size
func (pt *ProgressTracker) SetTotal(files int64, size int64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.totalFiles = files
	pt.totalSize = size
	pt.notifySubscribers()
}

// UpdateProgress updates the current progress
func (pt *ProgressTracker) UpdateProgress(filesProcessed int64, sizeProcessed int64, currentFile string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.processedFiles = filesProcessed
	pt.processedSize = sizeProcessed
	pt.currentFile = currentFile
	pt.notifySubscribers()
}

// IncrementProgress increments progress by one file
func (pt *ProgressTracker) IncrementProgress(fileSize int64, fileName string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.processedFiles++
	pt.processedSize += fileSize
	pt.currentFile = fileName
	pt.notifySubscribers()
}

// AddError adds an error to the tracker
func (pt *ProgressTracker) AddError(err string) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.errors = append(pt.errors, err)
	pt.notifySubscribers()
}

// SetDone marks the operation as complete
func (pt *ProgressTracker) SetDone() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	pt.done = true
	pt.notifySubscribers()
}

// Subscribe adds a channel to receive progress updates
func (pt *ProgressTracker) Subscribe() chan ProgressUpdate {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	ch := make(chan ProgressUpdate, 10)
	pt.subscribers = append(pt.subscribers, ch)
	return ch
}

// GetProgress returns the current progress state
func (pt *ProgressTracker) GetProgress() ProgressUpdate {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	elapsed := time.Since(pt.startTime)
	
	var percentage float64
	if pt.totalFiles > 0 {
		percentage = float64(pt.processedFiles) / float64(pt.totalFiles) * 100
	}
	
	var filesPerSecond, bytesPerSecond float64
	if elapsed.Seconds() > 0 {
		filesPerSecond = float64(pt.processedFiles) / elapsed.Seconds()
		bytesPerSecond = float64(pt.processedSize) / elapsed.Seconds()
	}
	
	var estimatedTime time.Duration
	if pt.processedFiles > 0 && pt.totalFiles > pt.processedFiles {
		remaining := pt.totalFiles - pt.processedFiles
		avgTimePerFile := elapsed / time.Duration(pt.processedFiles)
		estimatedTime = avgTimePerFile * time.Duration(remaining)
	}
	
	return ProgressUpdate{
		TotalFiles:      pt.totalFiles,
		ProcessedFiles:  pt.processedFiles,
		TotalSize:       pt.totalSize,
		ProcessedSize:   pt.processedSize,
		CurrentFile:     pt.currentFile,
		Percentage:      percentage,
		ElapsedTime:     elapsed,
		EstimatedTime:   estimatedTime,
		FilesPerSecond:  filesPerSecond,
		BytesPerSecond:  bytesPerSecond,
		ErrorCount:      len(pt.errors),
		Done:            pt.done,
	}
}

// GetErrors returns all recorded errors
func (pt *ProgressTracker) GetErrors() []string {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	errors := make([]string, len(pt.errors))
	copy(errors, pt.errors)
	return errors
}

// notifySubscribers sends updates to all subscribers
func (pt *ProgressTracker) notifySubscribers() {
	update := pt.getProgressUnsafe()
	
	for i := len(pt.subscribers) - 1; i >= 0; i-- {
		select {
		case pt.subscribers[i] <- update:
		default:
			// Channel is full or closed, remove it
			close(pt.subscribers[i])
			pt.subscribers = append(pt.subscribers[:i], pt.subscribers[i+1:]...)
		}
	}
}

// getProgressUnsafe returns progress without locking (internal use)
func (pt *ProgressTracker) getProgressUnsafe() ProgressUpdate {
	elapsed := time.Since(pt.startTime)
	
	var percentage float64
	if pt.totalFiles > 0 {
		percentage = float64(pt.processedFiles) / float64(pt.totalFiles) * 100
	}
	
	var filesPerSecond, bytesPerSecond float64
	if elapsed.Seconds() > 0 {
		filesPerSecond = float64(pt.processedFiles) / elapsed.Seconds()
		bytesPerSecond = float64(pt.processedSize) / elapsed.Seconds()
	}
	
	var estimatedTime time.Duration
	if pt.processedFiles > 0 && pt.totalFiles > pt.processedFiles {
		remaining := pt.totalFiles - pt.processedFiles
		avgTimePerFile := elapsed / time.Duration(pt.processedFiles)
		estimatedTime = avgTimePerFile * time.Duration(remaining)
	}
	
	return ProgressUpdate{
		TotalFiles:      pt.totalFiles,
		ProcessedFiles:  pt.processedFiles,
		TotalSize:       pt.totalSize,
		ProcessedSize:   pt.processedSize,
		CurrentFile:     pt.currentFile,
		Percentage:      percentage,
		ElapsedTime:     elapsed,
		EstimatedTime:   estimatedTime,
		FilesPerSecond:  filesPerSecond,
		BytesPerSecond:  bytesPerSecond,
		ErrorCount:      len(pt.errors),
		Done:            pt.done,
	}
}
