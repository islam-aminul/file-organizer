package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileRecord represents a file entry in the database
type FileRecord struct {
	ID              int       `json:"id"`
	Hash            string    `json:"hash"`
	OriginalPath    string    `json:"original_path"`
	DestinationPath string    `json:"destination_path"`
	Size            int64     `json:"size"`
	ProcessedAt     time.Time `json:"processed_at"`
}

// Database provides CGO-free file tracking using JSON
type Database struct {
	mu       sync.RWMutex
	dbPath   string
	records  map[string]FileRecord // hash -> record
	nextID   int
}

// NewDatabase creates a new CGO-free database
func NewDatabase(destDir string) (*Database, error) {
	dbPath := filepath.Join(destDir, "zensort-db.json")
	
	db := &Database{
		dbPath:  dbPath,
		records: make(map[string]FileRecord),
		nextID:  1,
	}
	
	// Load existing database if it exists
	if err := db.load(); err != nil {
		// If file doesn't exist, that's OK - we'll create it
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load database: %w", err)
		}
	}
	
	return db, nil
}

// CheckDuplicate checks if a file hash already exists
func (db *Database) CheckDuplicate(hash string) (bool, string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	if record, exists := db.records[hash]; exists {
		return true, record.DestinationPath, nil
	}
	
	return false, "", nil
}

// AddFile adds a new file record to the database
func (db *Database) AddFile(hash, originalPath, destinationPath string, size int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	
	record := FileRecord{
		ID:              db.nextID,
		Hash:            hash,
		OriginalPath:    originalPath,
		DestinationPath: destinationPath,
		Size:            size,
		ProcessedAt:     time.Now(),
	}
	
	db.records[hash] = record
	db.nextID++
	
	return db.save()
}

// GetStats returns database statistics
func (db *Database) GetStats() (int, int64, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	
	var totalSize int64
	for _, record := range db.records {
		totalSize += record.Size
	}
	
	return len(db.records), totalSize, nil
}

// Close saves the database (no-op for JSON implementation)
func (db *Database) Close() error {
	return db.save()
}

// load reads the database from JSON file
func (db *Database) load() error {
	data, err := os.ReadFile(db.dbPath)
	if err != nil {
		return err
	}
	
	var records []FileRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("failed to parse database JSON: %w", err)
	}
	
	// Convert slice to map and find next ID
	db.records = make(map[string]FileRecord)
	maxID := 0
	
	for _, record := range records {
		db.records[record.Hash] = record
		if record.ID > maxID {
			maxID = record.ID
		}
	}
	
	db.nextID = maxID + 1
	return nil
}

// save writes the database to JSON file
func (db *Database) save() error {
	// Convert map to slice for JSON serialization
	records := make([]FileRecord, 0, len(db.records))
	for _, record := range db.records {
		records = append(records, record)
	}
	
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(db.dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}
	
	// Write to temporary file first, then rename for atomic operation
	tempPath := db.dbPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write database file: %w", err)
	}
	
	if err := os.Rename(tempPath, db.dbPath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename database file: %w", err)
	}
	
	return nil
}
