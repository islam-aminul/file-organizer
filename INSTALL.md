# ZenSort Installation Guide

## Step 1: Install Go

### Method 1: Official Installer (Recommended)
1. Go to https://golang.org/dl/
2. Download "go1.21.x.windows-amd64.msi" (or latest version)
3. Run the installer
4. **Important**: Restart your terminal/PowerShell after installation
5. Verify installation: `go version`

### Method 2: Using Package Managers

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

## Step 2: Install C Compiler (For GUI Version)

The GUI version requires CGO (C bindings). Choose one:

### Option A: TDM-GCC (Lightweight)
1. Download from https://jmeubank.github.io/tdm-gcc/
2. Install TDM-GCC 64-bit
3. Add to PATH: `C:\TDM-GCC-64\bin`

### Option B: Visual Studio Build Tools
1. Download "Build Tools for Visual Studio" from Microsoft
2. Install with "C++ build tools" workload
3. Restart terminal

### Option C: MinGW-w64
```powershell
# Using Chocolatey
choco install mingw

# Using MSYS2
# Download from https://www.msys2.org/
```

## Step 3: Build ZenSort

After installing Go and C compiler:

```powershell
# Navigate to project directory
cd C:\Users\aminu\Workspace\Projects\file-organizer

# Install dependencies
go mod tidy

# Build CLI version (no CGO required)
go build -o zensort-cli.exe cmd/cli/main.go

# Build GUI version (requires CGO)
set CGO_ENABLED=1
go build -o zensort-gui.exe main.go

# Or use the build script
.\build.bat
```

## Step 4: Run ZenSort

### CLI Version
```powershell
.\zensort-cli.exe -source "C:\Source\Folder" -dest "C:\Destination\Folder"
```

### GUI Version
```powershell
.\zensort-gui.exe
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
