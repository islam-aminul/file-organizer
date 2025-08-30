# Motion Photos Detection and Organization

## Overview

ZenSort automatically detects and organizes Motion Photos (formerly Live Photos) - short video files created by iPhone and Samsung devices that capture a few seconds of video alongside still photos.

## How Motion Photos Work

Motion Photos are **video files** (`.mov`, `.mp4`) with specific filename patterns that indicate they were created as part of a Live Photo capture. ZenSort detects these based on:

1. **File Type**: Must be a video file (`.mov` or `.mp4`)
2. **Filename Patterns**: Contains specific patterns indicating Motion Photo origin
3. **Duration**: Typically short videos (configurable maximum duration)

## Detection Patterns

### iPhone Motion Photos
- Files containing: `live`, `livephoto`, `_live`
- Files starting with: `img_`
- Extension: `.mov`

### Samsung Motion Photos  
- Files containing: `motion`, `_motion`, `motionphoto`
- Files starting with: `mvimg_`
- Extension: `.mp4`

## Organization Structure

Motion Photos are organized separately from regular videos:

```
Videos/
├── Motion Photos/
│   ├── 2023/
│   │   ├── IMG_1234_live.mov
│   │   └── MVIMG_20231201_motion.mp4
│   └── 2024/
├── Short Videos/
│   ├── 2023/
│   │   └── quick_video.mp4 (duration-based detection)
│   └── 2024/
├── 2023/ (regular videos)
└── 2024/ (regular videos)
```

## Configuration

Motion Photos detection can be customized in the GUI settings or `zensort-config.json`:

```json
{
  "motion_photos": {
    "enabled": true,
    "iphone_patterns": ["live", "livephoto", "_live", "img_"],
    "samsung_patterns": ["motion", "_motion", "motionphoto", "mvimg_"],
    "extensions": [".mov", ".mp4"],
    "max_duration_seconds": 10
  }
}
```

### Configuration Options

- **enabled**: Enable/disable Motion Photos detection
- **iphone_patterns**: Filename patterns for iPhone Motion Photos
- **samsung_patterns**: Filename patterns for Samsung Motion Photos  
- **extensions**: Video file extensions to check (video files only)
- **max_duration_seconds**: Maximum duration for Motion Photo classification

## Processing Logic

1. **File Type Check**: Only video files are considered for Motion Photos
2. **Pattern Matching**: Filename is checked against configured patterns
3. **Duration Validation**: Video duration must be under the maximum threshold
4. **Organization**: Matching files go to `Videos/Motion Photos/Year/`

## Key Differences from Regular Videos

- **Motion Photos**: Pattern-based detection → `Videos/Motion Photos/Year/`
- **Short Videos**: Duration-based detection → `Videos/Short Videos/Year/`
- **Regular Videos**: All other videos → `Videos/Year/`

## Important Notes

- **Images are NOT Motion Photos**: Only video files can be Motion Photos
- **Separate from Short Videos**: Motion Photos use pattern detection, Short Videos use duration detection
- **Year-based Organization**: All video types are organized by year within their respective folders
- **No Image Processing**: Images are handled completely separately based on EXIF data
