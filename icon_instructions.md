# Icon Application Instructions

The executable icon requires proper resource compilation. Follow these steps:

## Method 1: Automatic (Recommended)
Run the build script which will automatically handle icon compilation:

### Windows:
```cmd
.\build.bat
```

### macOS/Linux:
```bash
chmod +x build.sh
./build.sh
```

## Method 2: Manual Icon Compilation

### Prerequisites:
- **Windows**: Install TDM-GCC, MinGW, or Visual Studio Build Tools (includes windres)
- **macOS**: Install Xcode Command Line Tools (`xcode-select --install`)
- **Linux**: Install build-essential and mingw-w64 for cross-compilation

### Steps:
1. **Compile resource file**:
   ```cmd
   windres -i icon.rc -o icon.syso
   ```

2. **Build with icon**:
   ```cmd
   set CGO_ENABLED=1
   go build -ldflags="-H windowsgui" -o zensort.exe main.go
   ```

## Method 3: Create Custom Icon
1. **Generate icon** (if you don't have one):
   ```cmd
   pip install Pillow
   python create_icon.py
   ```

2. **Follow Method 2** steps above

## Troubleshooting:

### "windres: command not found"
- **Windows**: Install TDM-GCC or MinGW-w64
- **macOS/Linux**: Install mingw-w64 for Windows cross-compilation

### Icon not appearing
1. Ensure `icon.ico` exists in project root
2. Verify `icon.syso` was created after windres compilation
3. Check that CGO is enabled during build
4. Use `-ldflags="-H windowsgui"` flag for Windows GUI applications

### Build without icon
If icon compilation fails, the build will continue without icon but show a warning.

## Files Required:
- `icon.ico` - The actual icon file
- `icon.rc` - Resource definition file
- `icon.syso` - Compiled resource (auto-generated)
- `embed_icon.go` - Go build constraints for Windows
