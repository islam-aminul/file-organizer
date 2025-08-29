# ZenSort - Cross-Platform File Organizer

ZenSort is a powerful, cross-platform file organization application built in Go with both GUI and CLI interfaces. It automatically categorizes and organizes files from a source directory into a structured destination folder using intelligent file type detection and EXIF data analysis.

## Features

### Core Functionality
- **Cross-Platform Single Executable**: Runs on Windows, macOS, and Linux (AMD64, ARM64)
- **Dual Interface**: Modern GUI with Fyne framework and efficient CLI mode
- **Smart Worker Pools**: Auto-detects optimal worker count based on CPU cores and available memory
- **Memory-Efficient Processing**: Streaming file processing with configurable buffer sizes
- **Real-Time Progress**: Live progress bars, status updates, and performance metrics

### File Organization
- **Hybrid File Detection**: Fast extension-based detection with MIME type fallback
- **Deduplication**: SQLite database prevents duplicate files using SHA256 hashing
- **Conflict Resolution**: Automatic file renaming with " -- n" suffix for naming conflicts
- **Category-Based Organization**: Images, Videos, Audios, Documents, Unknown files
- **Hidden File Handling**: Dedicated subdirectories for hidden files

### Advanced Features
- **Configurable Everything**: JSON configuration for all directory names and settings
- **Skip Patterns**: Ignore files by extensions, patterns, or directory paths
- **Detailed Logging**: Comprehensive error and operation logs in destination directory
- **Status Reports**: JSON and human-readable reports with statistics and performance metrics
- **EXIF Processing**: Image organization based on camera make, model, and date

## Installation

### Prerequisites
- Go 1.21 or later
- CGO enabled (for SQLite support)

### Build from Source
```bash
# Clone the repository
git clone <repository-url>
cd file-organizer

# Install dependencies
go mod tidy

# Build for current platform
go build -o zensort main.go

# Cross-compile for different platforms
# Windows
GOOS=windows GOARCH=amd64 go build -o zensort.exe main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -o zensort-mac main.go

# Linux
GOOS=linux GOARCH=amd64 go build -o zensort-linux main.go
```

## Usage

### GUI Mode
```bash
# Launch GUI (default behavior)
./zensort

# Explicitly launch GUI
./zensort -gui
```

### CLI Mode
```bash
# Basic usage
./zensort -source /path/to/source -dest /path/to/destination

# With custom configuration
./zensort -source /path/to/source -dest /path/to/destination -config /path/to/config.json
```

### Configuration

ZenSort automatically creates a default configuration file (`zensort-config.json`) on first run. You can customize:

```json
{
  "directories": {
    "images": "Images",
    "videos": "Videos",
    "audios": "Audios",
    "documents": "Documents",
    "unknown": "Unknown",
    "hidden": "Hidden"
  },
  "image_dirs": {
    "originals": "Originals",
    "exports": "Exports"
  },
  "skip_files": {
    "extensions": [".tmp", ".temp", ".log", ".cache"],
    "patterns": ["~*", ".DS_Store", "Thumbs.db"],
    "directories": [".git", ".svn", "node_modules"]
  },
  "processing": {
    "max_image_width": 3840,
    "max_image_height": 2160,
    "buffer_size": 1048576,
    "hash_chunk_size": 65536
  }
}
```

## Output Structure

ZenSort creates an organized directory structure in your destination folder:

```
destination/
├── Images/
│   ├── Originals/
│   │   ├── Canon - EOS R5/
│   │   │   └── 2024/
│   │   └── Collections/
│   ├── Exports/
│   │   └── 2024/
│   └── Hidden/
├── Videos/
│   ├── 2024/
│   └── Hidden/
├── Audios/
│   ├── Songs/
│   ├── Voice Recordings/
│   └── Hidden/
├── Documents/
│   ├── PDF/
│   ├── Word/
│   └── Hidden/
├── Unknown/
├── zensort-logs/
│   ├── errors_2024-08-29_22-30-15.log
│   └── operations_2024-08-29_22-30-15.log
├── zensort-report_2024-08-29_22-35-42.json
├── zensort-report_2024-08-29_22-35-42.txt
└── zensort.db
```

## Performance Features

### Auto-Scaling Worker Pools
- Automatically detects CPU cores and available memory
- Scales worker count between 1-16 based on system resources
- Conservative memory usage (1 worker per GB available RAM)

### Memory-Efficient Processing
- Streaming file hash calculation (64KB chunks)
- Configurable buffer sizes for different file operations
- Minimal memory footprint even with large files

### Progress Tracking
- Real-time progress bars with percentage completion
- Files per second and throughput metrics
- Estimated time remaining calculations
- Live status updates for current file being processed

## Logging and Reports

### Detailed Logging
- **Error Logs**: Comprehensive error tracking with file paths and timestamps
- **Operation Logs**: Success/failure status for each file operation
- **Performance Logs**: Processing statistics and timing information

### Status Reports
- **JSON Format**: Machine-readable statistics and metadata
- **Human-Readable**: Summary reports with file counts and performance metrics
- **Category Breakdown**: Statistics per file type with size information
- **Error Summary**: Grouped error types with sample file paths

## Dependencies

- **fyne.io/fyne/v2**: Modern cross-platform GUI framework
- **github.com/mattn/go-sqlite3**: SQLite database driver
- **github.com/shirou/gopsutil/v3**: System information for worker pool optimization
- **github.com/gabriel-vasile/mimetype**: MIME type detection
- **github.com/rwcarlsen/goexif**: EXIF data extraction
- **github.com/disintegration/imaging**: Image processing and resizing

## System Requirements

- **Memory**: Minimum 512MB RAM, 1GB+ recommended for large file sets
- **Storage**: Sufficient space in destination directory for organized files
- **CPU**: Any modern CPU (auto-scales worker count based on cores)
- **OS**: Windows 7+, macOS 10.12+, Linux (any modern distribution)

## Troubleshooting

### Common Issues

1. **CGO Build Errors**: Ensure CGO is enabled and C compiler is available
2. **Permission Errors**: Run with appropriate permissions for source/destination directories
3. **Memory Issues**: Reduce worker count in low-memory environments
4. **GUI Not Starting**: Check display environment and Fyne dependencies

### Debug Mode
Enable verbose logging by checking the operation logs in the destination directory's `zensort-logs/` folder.

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]
