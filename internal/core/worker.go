package core

import (
	"context"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/mem"
)

// WorkerPool manages concurrent file processing
type WorkerPool struct {
	workers    int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// Job represents a file processing task
type Job struct {
	ID       string
	FilePath string
	Type     JobType
}

// JobType defines the type of processing job
type JobType int

const (
	JobTypeProcess JobType = iota
	JobTypeHash
	JobTypeMove
)

// Result represents the outcome of a job
type Result struct {
	JobID    string
	Success  bool
	Error    error
	FilePath string
	Hash     string
	Size     int64
}

// NewWorkerPool creates a new worker pool with optimal worker count
func NewWorkerPool(ctx context.Context) *WorkerPool {
	workers := calculateOptimalWorkers()
	
	poolCtx, cancel := context.WithCancel(ctx)
	
	return &WorkerPool{
		workers: workers,
		jobs:    make(chan Job, workers*2), // Buffer for jobs
		results: make(chan Result, workers*2),
		ctx:     poolCtx,
		cancel:  cancel,
	}
}

// calculateOptimalWorkers determines the optimal number of workers
func calculateOptimalWorkers() int {
	cpuCount := runtime.NumCPU()
	
	// Get available memory
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		// Fallback to CPU count if memory info unavailable
		return cpuCount
	}
	
	// Available memory in GB
	availableGB := float64(memInfo.Available) / (1024 * 1024 * 1024)
	
	// Conservative approach: 1 worker per GB of available memory, max 2x CPU count
	memoryBasedWorkers := int(availableGB)
	maxWorkers := cpuCount * 2
	
	// Use the minimum of memory-based and CPU-based limits
	workers := min(memoryBasedWorkers, maxWorkers)
	
	// Ensure at least 1 worker, max reasonable limit
	if workers < 1 {
		workers = 1
	} else if workers > 16 {
		workers = 16
	}
	
	return workers
}

// Start begins processing jobs
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.jobs)
	wp.wg.Wait()
	close(wp.results)
}

// Submit adds a job to the queue
func (wp *WorkerPool) Submit(job Job) {
	select {
	case wp.jobs <- job:
	case <-wp.ctx.Done():
		// Pool is shutting down
	}
}

// Results returns the results channel
func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

// worker processes jobs from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				return // Jobs channel closed
			}
			
			result := wp.processJob(job)
			
			select {
			case wp.results <- result:
			case <-wp.ctx.Done():
				return
			}
			
		case <-wp.ctx.Done():
			return
		}
	}
}

// processJob handles individual job processing
func (wp *WorkerPool) processJob(job Job) Result {
	result := Result{
		JobID:    job.ID,
		FilePath: job.FilePath,
	}
	
	switch job.Type {
	case JobTypeHash:
		hash, err := calculateFileHash(job.FilePath)
		if err != nil {
			result.Error = err
			return result
		}
		result.Hash = hash
		result.Success = true
		
	case JobTypeProcess:
		// File processing logic is handled by the processor
		// This is just a placeholder - actual processing happens in processor.go
		result.Success = true
		
	case JobTypeMove:
		// File moving logic is handled by the organizer
		result.Success = true
	}
	
	return result
}

// WorkerCount returns the number of workers in the pool
func (wp *WorkerPool) WorkerCount() int {
	return wp.workers
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
