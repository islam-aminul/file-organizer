package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Directories struct {
		Images    string `json:"images"`
		Videos    string `json:"videos"`
		Audios    string `json:"audios"`
		Documents string `json:"documents"`
		Unknown   string `json:"unknown"`
		Hidden    string `json:"hidden"`
	} `json:"directories"`
	
	ImageDirs struct {
		Originals        string `json:"originals"`
		Exports          string `json:"exports"`
		NoExifYearFolder string `json:"no_exif_year_folder"`
	} `json:"image_dirs"`
	
	AudioCategories map[string]struct {
		FolderName string   `json:"folder_name"`
		Extensions []string `json:"extensions"`
		Patterns   []string `json:"patterns"`
	} `json:"audio_categories"`
	
	SkipFiles struct {
		Extensions []string `json:"extensions"`
		Patterns   []string `json:"patterns"`
		Directories []string `json:"directories"`
	} `json:"skip_files"`
	
	Processing struct {
		MaxImageWidth  int `json:"max_image_width"`
		MaxImageHeight int `json:"max_image_height"`
		BufferSize     int `json:"buffer_size"`
		HashChunkSize  int `json:"hash_chunk_size"`
	} `json:"processing"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	config := &Config{}
	
	// Default directory names
	config.Directories.Images = "Images"
	config.Directories.Videos = "Videos"
	config.Directories.Audios = "Audios"
	config.Directories.Documents = "Documents"
	config.Directories.Unknown = "Unknown"
	config.Directories.Hidden = "Hidden"
	
	// Default image subdirectories
	config.ImageDirs.Originals = "Originals"
	config.ImageDirs.Exports = "Exports"
	config.ImageDirs.NoExifYearFolder = "0000"
	
	// Default audio categories
	config.AudioCategories = make(map[string]struct {
		FolderName string   `json:"folder_name"`
		Extensions []string `json:"extensions"`
		Patterns   []string `json:"patterns"`
	})
	
	config.AudioCategories["songs"] = struct {
		FolderName string   `json:"folder_name"`
		Extensions []string `json:"extensions"`
		Patterns   []string `json:"patterns"`
	}{
		FolderName: "Songs",
		Extensions: []string{".mp3", ".flac", ".wav", ".aac", ".ogg", ".m4a", ".wma"},
		Patterns:   []string{},
	}
	
	config.AudioCategories["voice_recordings"] = struct {
		FolderName string   `json:"folder_name"`
		Extensions []string `json:"extensions"`
		Patterns   []string `json:"patterns"`
	}{
		FolderName: "Voice Recordings",
		Extensions: []string{".m4a", ".wav", ".aac", ".3gp"},
		Patterns:   []string{"voice", "memo", "note", "recording", "_rec"},
	}
	
	config.AudioCategories["call_recordings"] = struct {
		FolderName string   `json:"folder_name"`
		Extensions []string `json:"extensions"`
		Patterns   []string `json:"patterns"`
	}{
		FolderName: "Call Recordings",
		Extensions: []string{".m4a", ".wav", ".aac", ".3gp", ".amr"},
		Patterns:   []string{"call", "_call", "phone", "tel", "+", "recording"},
	}
	
	config.AudioCategories["other_audio"] = struct {
		FolderName string   `json:"folder_name"`
		Extensions []string `json:"extensions"`
		Patterns   []string `json:"patterns"`
	}{
		FolderName: "Other Audio",
		Extensions: []string{".mp3", ".wav", ".aac", ".ogg", ".wma", ".au", ".aiff"},
		Patterns:   []string{"podcast", "audiobook", "lecture", "interview", "meeting"},
	}
	
	// Default skip patterns
	config.SkipFiles.Extensions = []string{".tmp", ".temp", ".log", ".cache", ".thumb"}
	config.SkipFiles.Patterns = []string{"~*", ".DS_Store", "Thumbs.db", "*.thumb", "*.thumb[0-9]*"}
	config.SkipFiles.Directories = []string{".git", ".svn", "node_modules"}
	
	// Default processing settings
	config.Processing.MaxImageWidth = 3840
	config.Processing.MaxImageHeight = 2160
	config.Processing.BufferSize = 1024 * 1024 // 1MB
	config.Processing.HashChunkSize = 64 * 1024 // 64KB
	
	return config
}

// LoadConfig loads configuration from file or creates default
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = "zensort-config.json"
	}
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		config := DefaultConfig()
		if err := SaveConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}
	
	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}
