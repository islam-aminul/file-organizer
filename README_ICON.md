# Adding an Icon to ZenSort

## Method 1: Using Python Script (Recommended)

1. Install Python and Pillow:
```bash
pip install Pillow
```

2. Run the icon creation script:
```bash
python create_icon.py
```

This creates `icon.ico` with a blue folder design and "Z" letter.

## Method 2: Manual Icon Creation

1. Create or download a 256x256 PNG icon
2. Convert to ICO format using online tools or software like GIMP
3. Save as `icon.ico` in the project root

## Method 3: Using Go Resource Embedding (Windows)

The project includes:
- `icon.rc` - Windows resource file
- `embed_icon.go` - Go build constraint for Windows

To embed the icon:

1. Install a resource compiler (windres comes with MinGW/TDM-GCC)
2. Compile resource file:
```bash
windres -i icon.rc -o icon.syso
```

3. Build with embedded icon:
```bash
go build -ldflags="-H windowsgui" -o zensort.exe main.go
```

## Build Scripts Updated

The `build.bat` script now includes the `-H windowsgui` flag to create a proper Windows GUI application without console window.

## Icon Design

The default icon features:
- Blue folder representing file organization
- White sorting lines showing file categorization
- "Z" letter for ZenSort branding
- Multiple sizes (16x16 to 256x256) for Windows compatibility
