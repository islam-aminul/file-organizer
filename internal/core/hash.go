package core

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// calculateFileHash computes SHA256 hash of a file using streaming
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	hasher := sha256.New()
	
	// Use a buffer for memory-efficient streaming
	buffer := make([]byte, 64*1024) // 64KB chunks
	
	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
	}
	
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
