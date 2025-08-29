package core

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDatabase handles SQLite-based deduplication
type SQLiteDatabase struct {
	db *sql.DB
}

// NewSQLiteDatabase creates a new SQLite database connection
func NewSQLiteDatabase(destPath string) (*SQLiteDatabase, error) {
	dbPath := filepath.Join(destPath, "zensort-db.sqlite")
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqliteDB := &SQLiteDatabase{db: db}
	
	// Initialize database schema
	if err := sqliteDB.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return sqliteDB, nil
}

// initSchema creates the necessary tables
func (s *SQLiteDatabase) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT UNIQUE NOT NULL,
		original_path TEXT NOT NULL,
		destination_path TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_hash ON files(hash);
	CREATE INDEX IF NOT EXISTS idx_original_path ON files(original_path);
	`

	_, err := s.db.Exec(schema)
	return err
}

// FileExists checks if a file with the given hash already exists
func (s *SQLiteDatabase) FileExists(hash string) (bool, string, error) {
	var destPath string
	err := s.db.QueryRow("SELECT destination_path FROM files WHERE hash = ?", hash).Scan(&destPath)
	
	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}
	
	return true, destPath, nil
}

// AddFile adds a file record to the database
func (s *SQLiteDatabase) AddFile(hash, originalPath, destinationPath string, fileSize int64) error {
	_, err := s.db.Exec(
		"INSERT INTO files (hash, original_path, destination_path, file_size) VALUES (?, ?, ?, ?)",
		hash, originalPath, destinationPath, fileSize,
	)
	return err
}

// GetStats returns database statistics
func (s *SQLiteDatabase) GetStats() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM files").Scan(&count)
	return count, err
}

// Close closes the database connection
func (s *SQLiteDatabase) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetAllFiles returns all files in the database (for migration/debugging)
func (s *SQLiteDatabase) GetAllFiles() ([]SQLiteFileRecord, error) {
	rows, err := s.db.Query("SELECT hash, original_path, destination_path, file_size FROM files")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []SQLiteFileRecord
	for rows.Next() {
		var record SQLiteFileRecord
		err := rows.Scan(&record.Hash, &record.OriginalPath, &record.DestinationPath, &record.FileSize)
		if err != nil {
			return nil, err
		}
		files = append(files, record)
	}

	return files, rows.Err()
}

// SQLiteFileRecord represents a file record in the SQLite database
type SQLiteFileRecord struct {
	Hash            string `json:"hash"`
	OriginalPath    string `json:"original_path"`
	DestinationPath string `json:"destination_path"`
	FileSize        int64  `json:"file_size"`
}
