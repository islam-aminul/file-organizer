# ZenSort Installation Guide

## Step 1: Install Go

### Windows

#### Method 1: Official Installer (Recommended)
1. Go to https://golang.org/dl/
2. Download "go1.21.x.windows-amd64.msi" (or latest version)
3. Run the installer
4. **Important**: Restart your terminal/PowerShell after installation
5. Verify installation: `go version`

#### Method 2: Using Package Managers

**Chocolatey:**
```powershell
choco install golang
```

**Winget:**
```powershell
winget install GoLang.Go
```

**Scoop:**
```powershell
scoop install go
```

### macOS

#### Method 1: Official Installer
1. Go to https://golang.org/dl/
2. Download "go1.21.x.darwin-amd64.pkg" (Intel) or "go1.21.x.darwin-arm64.pkg" (Apple Silicon)
3. Run the installer
4. Verify installation: `go version`

#### Method 2: Using Homebrew
```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Go
brew install go

# Verify installation
go version
```

### Linux

#### Ubuntu/Debian
```bash
# Method 1: Official package
sudo apt update
sudo apt install golang-go

# Method 2: Latest version from official source
wget https://go.dev/dl/go1.21.x.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.x.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

#### CentOS/RHEL/Fedora
```bash
# Method 1: Package manager
sudo yum install golang  # CentOS/RHEL
sudo dnf install golang  # Fedora

# Method 2: Latest version (same as Ubuntu method above)
# Verify installation
go version
```

## Step 2: Install C Compiler (For GUI Version)

The GUI version requires CGO (C bindings).

### Windows

#### Option A: TDM-GCC (Lightweight)
1. Download from https://jmeubank.github.io/tdm-gcc/
2. Install TDM-GCC 64-bit
3. Add to PATH: `C:\TDM-GCC-64\bin`

#### Option B: Visual Studio Build Tools
1. Download "Build Tools for Visual Studio" from Microsoft
2. Install with "C++ build tools" workload
3. Restart terminal

#### Option C: MinGW-w64
```powershell
# Using Chocolatey
choco install mingw

# Using MSYS2
# Download from https://www.msys2.org/
```

### macOS

#### Xcode Command Line Tools (Recommended)
```bash
# Install Xcode Command Line Tools
xcode-select --install

# Verify installation
gcc --version
```

#### Alternative: Full Xcode
1. Install Xcode from App Store
2. Open Xcode and accept license
3. Install additional components when prompted

### Linux

#### Ubuntu/Debian
```bash
# Install build essentials
sudo apt update
sudo apt install build-essential pkg-config

# Install GUI dependencies
sudo apt install libgl1-mesa-dev xorg-dev

# Verify installation
gcc --version
```

#### CentOS/RHEL/Fedora
```bash
# Install development tools
sudo yum groupinstall "Development Tools"  # CentOS/RHEL
sudo dnf groupinstall "Development Tools"  # Fedora

# Install GUI dependencies
sudo yum install libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libGL-devel

# Verify installation
gcc --version
```

## Step 3: Build ZenSort

After installing Go and C compiler:

### Windows
```powershell
# Navigate to project directory
cd C:\Users\aminu\Workspace\Projects\file-organizer

# Install dependencies
go mod tidy

# Build unified version with GUI support (requires CGO)
set CGO_ENABLED=1
go build -o zensort.exe main.go

# Or use the build script (handles CGO automatically)
.\build.bat

# Alternative: CLI-only build (no CGO required)
set CGO_ENABLED=0
go build -tags nocgo -o zensort-cli-only.exe main.go
```

### macOS/Linux
```bash
# Navigate to project directory
cd /path/to/file-organizer

# Install dependencies
go mod tidy

# Build unified version with GUI support (requires CGO)
export CGO_ENABLED=1
go build -o zensort main.go

# Or use the build script (handles CGO automatically)
chmod +x build.sh
./build.sh

# Alternative: CLI-only build (no CGO required)
export CGO_ENABLED=0
go build -tags nocgo -o zensort-cli-only main.go
```

## Step 4: Run ZenSort

### Windows
```powershell
# GUI Mode (default)
.\zensort.exe

# CLI Mode (when source/dest provided)
.\zensort.exe -source "C:\Source\Folder" -dest "C:\Destination\Folder"

# Force CLI mode
.\zensort.exe -cli -source "C:\Source\Folder" -dest "C:\Destination\Folder"
```

### macOS/Linux
```bash
# GUI Mode (default)
./zensort

# CLI Mode (when source/dest provided)
./zensort -source ~/Documents/ToOrganize -dest ~/Documents/Organized

# Force CLI mode
./zensort -cli -source ~/Documents/ToOrganize -dest ~/Documents/Organized
```

## Troubleshooting

### "go: command not found"
- Go is not installed or not in PATH
- Restart terminal after Go installation
- Check PATH: `echo $env:PATH` (PowerShell) or `echo %PATH%` (CMD)

### CGO Build Errors
- Install a C compiler (see Step 2)
- Verify CGO: `go env CGO_ENABLED` (should show "1")
- Use CLI version if GUI build fails

### OpenGL Errors
- Update graphics drivers
- Install Visual C++ Redistributables
- Use CLI version as alternative

### Permission Errors
- Run as Administrator if needed
- Check source/destination folder permissions

## Quick Start (CLI Only)

If you just want to try ZenSort without GUI:

1. Install Go (Step 1)
2. Build CLI: `go build -o zensort-cli.exe cmd/cli/main.go`
3. Run: `.\zensort-cli.exe -source "source_path" -dest "dest_path"`

The CLI version has all the core functionality without GUI dependencies.
