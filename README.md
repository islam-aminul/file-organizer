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
- **Deduplication**: JSON/SQLite database prevents duplicate files using SHA256 hashing
- **Conflict Resolution**: Automatic file renaming with " -- n" suffix for naming conflicts
- **Category-Based Organization**: Images, Videos, Audios, Documents, Unknown files
- **Hidden File Handling**: Dedicated subdirectories for hidden files

### Advanced Features
- **Configurable Everything**: JSON configuration for all directory names and settings
- **Skip Patterns**: Ignore files by extensions, patterns, or directory paths
- **Detailed Logging**: Comprehensive error and operation logs in `zensort-logs/` directory
- **Status Reports**: JSON and TXT reports with statistics stored in `zensort-logs/` folder
- **EXIF Processing**: Image organization based on camera make, model, and date with time stamps

## Installation

### Prerequisites
- Go 1.21 or later
- CGO enabled (for SQLite support)

### Build from Source

#### Windows
```powershell
# Clone the repository
git clone <repository-url>
cd file-organizer

# Install dependencies
go mod tidy

# Build with GUI support (requires C compiler)
.\build.bat

# Manual build
set CGO_ENABLED=1
go build -o zensort.exe main.go
```

#### macOS
```bash
# Clone the repository
git clone <repository-url>
cd file-organizer

# Install dependencies
go mod tidy

# Install Xcode Command Line Tools (for CGO)
xcode-select --install

# Build with GUI support
chmod +x build.sh
./build.sh

# Manual build
CGO_ENABLED=1 go build -o zensort main.go
```

#### Linux
```bash
# Clone the repository
git clone <repository-url>
cd file-organizer

# Install dependencies
go mod tidy

# Install build essentials (Ubuntu/Debian)
sudo apt update
sudo apt install build-essential pkg-config libgl1-mesa-dev xorg-dev

# Install build essentials (CentOS/RHEL/Fedora)
sudo yum groupinstall "Development Tools"
sudo yum install libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libGL-devel

# Build with GUI support
chmod +x build.sh
./build.sh

# Manual build
CGO_ENABLED=1 go build -o zensort main.go
```

#### Cross-compilation
```bash
# Windows from Unix
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -o zensort.exe main.go

# macOS from Linux (requires osxcross)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o zensort-mac main.go

# Linux from macOS
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o zensort-linux main.go
```

## Usage

### GUI Mode
```bash
# Launch GUI (default behavior)
./zensort
```

### CLI Mode
```bash
# Basic usage
./zensort -source /path/to/source -dest /path/to/destination

# Force CLI mode
./zensort -cli -source /path/to/source -dest /path/to/destination

# With custom configuration
./zensort -source /path/to/source -dest /path/to/destination -config /path/to/config.json
```

### Platform-specific Examples

#### Windows
```powershell
# GUI Mode
.\zensort.exe

# CLI Mode
.\zensort.exe -source "C:\Source\Folder" -dest "C:\Organized\Files"
```

#### macOS/Linux
```bash
# GUI Mode
./zensort

# CLI Mode
./zensort -source ~/Documents/ToOrganize -dest ~/Documents/Organized
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
│   ├── operations_2024-08-29_22-30-15.log
│   ├── zensort-report_2024-08-29_22-35-42.json
│   └── zensort-report_2024-08-29_22-35-42.txt
└── zensort-db.json
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

## Logging and Database System

### Centralized Logging (`zensort-logs/` folder)
- **Error Logs**: `errors_YYYY-MM-DD_HH-MM-SS.log` - Comprehensive error tracking with file paths and timestamps
- **Operation Logs**: `operations_YYYY-MM-DD_HH-MM-SS.log` - Success/failure status for each file operation
- **JSON Reports**: `zensort-report_YYYY-MM-DD_HH-MM-SS.json` - Machine-readable statistics and metadata
- **Text Reports**: `zensort-report_YYYY-MM-DD_HH-MM-SS.txt` - Human-readable summary with file counts and performance metrics

### Duplicate Detection Database
- **JSON Database** (Default): `zensort-db.json` - CGO-free, cross-platform hash storage
- **SQLite Database** (Enhanced): `zensort-db.sqlite` - Professional database with indexing (requires CGO)
- **Hash-Based Deduplication**: SHA256 content fingerprinting with memory-efficient streaming
- **Persistent Storage**: Remembers processed files across application restarts
- **Thread-Safe Operations**: Concurrent access protection with mutex locks

### Report Contents
- **File Statistics**: Total, processed, skipped, duplicate, and error counts
- **Size Information**: Total bytes processed with human-readable formatting
- **Performance Metrics**: Processing duration, files per second, throughput rates
- **Category Breakdown**: Statistics per file type (Images, Videos, Audios, Documents)
- **Error Analysis**: Grouped error types with sample file paths and frequencies
- **EXIF Processing**: Camera-based organization statistics and export counts

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
