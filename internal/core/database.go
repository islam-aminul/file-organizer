package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
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

// Database provides efficient file tracking using BadgerDB
type Database struct {
	db     *badger.DB
	dbPath string
	nextID int
}

// NewDatabase creates a new BadgerDB database
func NewDatabase(destDir string) (*Database, error) {
	dbPath := filepath.Join(destDir, "zensort-db")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	
	// Open BadgerDB
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil // Disable BadgerDB logging to avoid noise
	
	badgerDB, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open BadgerDB: %w", err)
	}
	
	db := &Database{
		db:     badgerDB,
		dbPath: dbPath,
		nextID: 1,
	}
	
	// Migrate from JSON if exists
	if err := db.migrateFromJSON(); err != nil {
		badgerDB.Close()
		return nil, fmt.Errorf("failed to migrate from JSON: %w", err)
	}
	
	// Initialize next ID from existing records
	if err := db.initializeNextID(); err != nil {
		badgerDB.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	return db, nil
}

// CheckDuplicate checks if a file hash already exists
func (db *Database) CheckDuplicate(hash string) (bool, string, error) {
	var destinationPath string
	var found bool
	
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("hash:" + hash))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Not found, not an error
			}
			return err
		}
		
		return item.Value(func(val []byte) error {
			var record FileRecord
			if err := json.Unmarshal(val, &record); err != nil {
				return err
			}
			destinationPath = record.DestinationPath
			found = true
			return nil
		})
	})
	
	return found, destinationPath, err
}

// AddFile adds a new file record to the database
func (db *Database) AddFile(hash, originalPath, destinationPath string, size int64) error {
	record := FileRecord{
		ID:              db.nextID,
		Hash:            hash,
		OriginalPath:    originalPath,
		DestinationPath: destinationPath,
		Size:            size,
		ProcessedAt:     time.Now(),
	}
	
	recordData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal record: %w", err)
	}
	
	err = db.db.Update(func(txn *badger.Txn) error {
		// Store record by hash
		if err := txn.Set([]byte("hash:"+hash), recordData); err != nil {
			return err
		}
		
		// Store ID mapping for statistics
		idKey := fmt.Sprintf("id:%d", db.nextID)
		if err := txn.Set([]byte(idKey), []byte(hash)); err != nil {
			return err
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to store record: %w", err)
	}
	
	db.nextID++
	return nil
}

// GetStats returns database statistics
func (db *Database) GetStats() (int, int64, error) {
	var count int
	var totalSize int64
	
	err := db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		
		prefix := []byte("hash:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var record FileRecord
				if err := json.Unmarshal(val, &record); err != nil {
					return err
				}
				totalSize += record.Size
				count++
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	
	return count, totalSize, err
}

// Close closes the BadgerDB database
func (db *Database) Close() error {
	return db.db.Close()
}

// initializeNextID finds the highest ID in the database
func (db *Database) initializeNextID() error {
	maxID := 0
	
	err := db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		
		prefix := []byte("hash:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var record FileRecord
				if err := json.Unmarshal(val, &record); err != nil {
					return err
				}
				if record.ID > maxID {
					maxID = record.ID
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	
	if err != nil {
		return err
	}
	
	db.nextID = maxID + 1
	return nil
}

// migrateFromJSON migrates existing JSON database to BadgerDB (if exists)
func (db *Database) migrateFromJSON() error {
	jsonPath := filepath.Join(filepath.Dir(db.dbPath), "zensort-db.json")
	
	// Check if old JSON database exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return nil // No migration needed
	}
	
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON database: %w", err)
	}
	
	var records []FileRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("failed to parse JSON database: %w", err)
	}
	
	// Migrate records to BadgerDB
	err = db.db.Update(func(txn *badger.Txn) error {
		for _, record := range records {
			recordData, err := json.Marshal(record)
			if err != nil {
				return err
			}
			
			if err := txn.Set([]byte("hash:"+record.Hash), recordData); err != nil {
				return err
			}
			
			idKey := fmt.Sprintf("id:%d", record.ID)
			if err := txn.Set([]byte(idKey), []byte(record.Hash)); err != nil {
				return err
			}
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to migrate records: %w", err)
	}
	
	// Backup and remove old JSON file
	backupPath := jsonPath + ".backup"
	if err := os.Rename(jsonPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup JSON database: %w", err)
	}
	
	return nil
}
