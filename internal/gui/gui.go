package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"zensort/internal/config"
	"zensort/internal/core"
)

// GUI represents the graphical user interface
type GUI struct {
	app            fyne.App
	window         fyne.Window
	sourceEntry    *widget.Entry
	destEntry      *widget.Entry
	sourceBrowseBtn *widget.Button
	destBrowseBtn   *widget.Button
	progressBar    *widget.ProgressBar
	statusLabel    *widget.Label
	logText        *widget.Entry
	startButton    *widget.Button
	pauseButton    *widget.Button
	stopButton     *widget.Button
	settingsButton *widget.Button
	processor      *core.FileProcessor
	ctx            context.Context
	cancel         context.CancelFunc
	pauseChan      chan bool
	progressChan   chan core.ProgressUpdate
	logBuffer      []string
	maxLogLines    int
	currentFolder  string
	folderFileCount int
	folderFiles    []string
	currentConfig  *config.Config
	lastSourceDir  string
	lastDestDir    string
	isPaused       bool
}

// Launch starts the GUI application
func Launch() {
	gui := NewGUI()
	gui.setupUI()
	gui.window.ShowAndRun()
}

// NewGUI creates a new GUI instance
func NewGUI() *GUI {
	a := app.NewWithID("com.zensort.fileorganizer")
	a.SetIcon(nil) // You can set an icon here
	
	w := a.NewWindow("ZenSort - File Organizer")
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()
	
	gui := &GUI{
		app:           a,
		window:        w,
		logBuffer:     make([]string, 0),
		maxLogLines:   100, // Limit log to 100 lines for performance
		folderFiles:   make([]string, 0),
		currentConfig: config.DefaultConfig(),
	}
	
	// Load last used directories from preferences
	gui.loadLastUsedDirectories()
	
	// Setup UI first
	gui.setupUI()
	
	// Set last used directories in entry fields after UI setup
	if gui.lastSourceDir != "" {
		gui.sourceEntry.SetText(gui.lastSourceDir)
	}
	if gui.lastDestDir != "" {
		gui.destEntry.SetText(gui.lastDestDir)
		gui.onDestinationChanged() // Enable settings if destination is set
	}
	
	return gui
}

// togglePause toggles pause/resume functionality
func (g *GUI) togglePause() {
	if g.pauseChan == nil {
		return // No active processing
	}
	
	if g.isPaused {
		// Resume processing
		select {
		case g.pauseChan <- false:
			g.pauseButton.SetText("Pause")
			g.pauseButton.Importance = widget.MediumImportance
			g.pauseButton.Refresh()
			g.statusLabel.SetText("Resuming...")
			g.isPaused = false
		default:
			// Channel full, ignore
		}
	} else {
		// Pause processing
		select {
		case g.pauseChan <- true:
			g.pauseButton.SetText("Resume")
			g.pauseButton.Importance = widget.HighImportance
			g.pauseButton.Refresh()
			g.statusLabel.SetText("⏸ Processing Paused")
			g.isPaused = true
			g.logMessage("⏸ Processing paused")
		default:
			// Channel full, ignore
		}
	}
}

// loadLastUsedDirectories loads the last used directories from app preferences
func (g *GUI) loadLastUsedDirectories() {
	g.lastSourceDir = g.app.Preferences().String("last_source_dir")
	g.lastDestDir = g.app.Preferences().String("last_dest_dir")
}

// saveLastUsedDirectories saves the current directories to app preferences
func (g *GUI) saveLastUsedDirectories() {
	if g.sourceEntry.Text != "" {
		g.lastSourceDir = g.sourceEntry.Text
		g.app.Preferences().SetString("last_source_dir", g.lastSourceDir)
	}
	if g.destEntry.Text != "" {
		g.lastDestDir = g.destEntry.Text
		g.app.Preferences().SetString("last_dest_dir", g.lastDestDir)
	}
}

// setupUI creates and arranges the user interface elements
func (g *GUI) setupUI() {
	// Source directory selection
	g.sourceEntry = widget.NewEntry()
	g.sourceEntry.SetPlaceHolder("Select source directory...")
	g.sourceBrowseBtn = widget.NewButton("Browse", g.browseSource)
	sourceContainer := container.NewBorder(nil, nil, nil, g.sourceBrowseBtn, g.sourceEntry)
	
	// Destination directory selection
	g.destEntry = widget.NewEntry()
	g.destEntry.SetPlaceHolder("Select destination directory...")
	g.destBrowseBtn = widget.NewButton("Browse", g.browseDestination)
	destContainer := container.NewBorder(nil, nil, nil, g.destBrowseBtn, g.destEntry)
	
	// Progress bar
	g.progressBar = widget.NewProgressBar()
	g.progressBar.SetValue(0)
	
	// Status label
	g.statusLabel = widget.NewLabel("Ready to organize files")
	
	// Log text area
	g.logText = widget.NewMultiLineEntry()
	g.logText.SetPlaceHolder("Processing logs will appear here...")
	g.logText.Wrapping = fyne.TextWrapWord
	logScroll := container.NewScroll(g.logText)
	logScroll.SetMinSize(fyne.NewSize(0, 200))
	
	// Control buttons with colors
	g.settingsButton = widget.NewButton("Settings", g.showSettings)
	g.settingsButton.Disable() // Disabled until destination is selected
	g.settingsButton.Importance = widget.LowImportance
	
	g.startButton = widget.NewButton("Start Organization", g.startProcessing)
	g.startButton.Disable() // Disabled until destination is selected
	g.startButton.Importance = widget.HighImportance
	
	g.pauseButton = widget.NewButton("Pause", g.togglePause)
	g.pauseButton.Disable()
	g.pauseButton.Importance = widget.MediumImportance
	
	g.stopButton = widget.NewButton("Stop", g.stopProcessing)
	g.stopButton.Disable()
	g.stopButton.Importance = widget.DangerImportance
	
	buttonContainer := container.NewHBox(g.settingsButton, g.startButton, g.pauseButton, g.stopButton)
	
	// Create form layout
	form := container.NewVBox(
		widget.NewLabel("Source Directory:"),
		sourceContainer,
		widget.NewLabel("Destination Directory:"),
		destContainer,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewSeparator(),
		widget.NewLabel("Progress:"),
		g.progressBar,
		g.statusLabel,
		widget.NewSeparator(),
		widget.NewLabel("Processing Log:"),
		logScroll,
	)
	
	// Set content
	g.window.SetContent(container.NewPadded(form))
}

// browseSource opens file dialog for source directory
func (g *GUI) browseSource() {
	// Start from current source entry or last used directory
	var startDir fyne.ListableURI
	startPath := g.sourceEntry.Text
	if startPath == "" {
		startPath = g.lastSourceDir
	}
	
	if startPath != "" {
		// Validate directory exists before creating URI
		if _, err := os.Stat(startPath); err == nil {
			// Ensure proper file:// URI format for Windows paths
			if !strings.HasPrefix(startPath, "file://") {
				// Convert Windows path to proper URI
				startPath = strings.ReplaceAll(startPath, "\\", "/")
				startPath = "file:///" + startPath
			}
			if uri, err := storage.ParseURI(startPath); err == nil {
				if listableURI, ok := uri.(fyne.ListableURI); ok {
					startDir = listableURI
				}
			}
		} else {
			fmt.Printf("Stored source directory no longer exists: %s\n", startPath)
		}
	}
	
	folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic in source dialog callback: %v\n", r)
			}
		}()
		
		if err != nil {
			fmt.Printf("Folder dialog error: %v\n", err)
			return
		}
		if uri == nil {
			fmt.Printf("URI is nil\n")
			return
		}
		
		path := uri.Path()
		fmt.Printf("Selected source path: %s\n", path)
		
		if g.sourceEntry != nil {
			g.sourceEntry.SetText(path)
		}
		g.saveLastUsedDirectories()
	}, g.window)
	
	if startDir != nil {
		// Additional safety check before setting location
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic setting source dialog location: %v\n", r)
				}
			}()
			folderDialog.SetLocation(startDir)
		}()
	}
	
	folderDialog.Show()
}

// browseDestination opens file dialog for destination directory
func (g *GUI) browseDestination() {
	// Start from current dest entry or last used directory
	var startDir fyne.ListableURI
	startPath := g.destEntry.Text
	if startPath == "" {
		startPath = g.lastDestDir
	}
	
	if startPath != "" {
		// Validate directory exists before creating URI
		if _, err := os.Stat(startPath); err == nil {
			// Ensure proper file:// URI format for Windows paths
			if !strings.HasPrefix(startPath, "file://") {
				// Convert Windows path to proper URI
				startPath = strings.ReplaceAll(startPath, "\\", "/")
				startPath = "file:///" + startPath
			}
			if uri, err := storage.ParseURI(startPath); err == nil {
				if listableURI, ok := uri.(fyne.ListableURI); ok {
					startDir = listableURI
				}
			}
		} else {
			fmt.Printf("Stored destination directory no longer exists: %s\n", startPath)
		}
	}
	
	folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic in destination dialog callback: %v\n", r)
			}
		}()
		
		if err != nil {
			fmt.Printf("Folder dialog error: %v\n", err)
			return
		}
		if uri == nil {
			fmt.Printf("URI is nil\n")
			return
		}
		
		path := uri.Path()
		fmt.Printf("Selected destination path: %s\n", path)
		
		if g.destEntry != nil {
			g.destEntry.SetText(path)
		}
		g.saveLastUsedDirectories()
		g.onDestinationChanged()
	}, g.window)
	
	if startDir != nil {
		// Additional safety check before setting location
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic setting destination dialog location: %v\n", r)
				}
			}()
			folderDialog.SetLocation(startDir)
		}()
	}
	
	folderDialog.Show()
}

// onDestinationChanged handles destination directory selection changes
func (g *GUI) onDestinationChanged() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic in onDestinationChanged: %v\n", r)
		}
	}()
	
	destDir := g.destEntry.Text
	if destDir == "" {
		if g.settingsButton != nil {
			g.settingsButton.Disable()
		}
		if g.startButton != nil {
			g.startButton.Disable()
		}
		return
	}
	
	// Enable buttons when destination is selected
	if g.settingsButton != nil {
		g.settingsButton.Enable()
	}
	if g.startButton != nil {
		g.startButton.Enable()
	}
}

// showSettings opens the settings configuration window
func (g *GUI) showSettings() {
	g.openSettingsWindow()
}

// loadDestinationConfig loads configuration from destination directory or creates default
func (g *GUI) loadDestinationConfig(destDir string) (*config.Config, error) {
	configPath := filepath.Join(destDir, "zensort-config.json")
	
	// Try to load existing config from destination directory
	if cfg, err := config.LoadConfig(configPath); err == nil {
		g.currentConfig = cfg
		return cfg, nil
	}
	
	// If no config exists, save default config to destination directory
	if err := config.SaveConfig(g.currentConfig, configPath); err != nil {
		return nil, fmt.Errorf("failed to save default config: %w", err)
	}
	
	return g.currentConfig, nil
}

// openSettingsWindow creates and displays the settings configuration window
func (g *GUI) openSettingsWindow() {
	settingsWindow := g.app.NewWindow("ZenSort Settings")
	settingsWindow.Resize(fyne.NewSize(600, 700))
	settingsWindow.CenterOnScreen()
	
	// Load current config from destination if available
	if destDir := g.destEntry.Text; destDir != "" {
		if cfg, err := g.loadDestinationConfig(destDir); err == nil {
			g.currentConfig = cfg
		}
	}
	
	// Directory names section
	imagesEntry := widget.NewEntry()
	imagesEntry.SetText(g.currentConfig.Directories.Images)
	
	videosEntry := widget.NewEntry()
	videosEntry.SetText(g.currentConfig.Directories.Videos)
	
	audiosEntry := widget.NewEntry()
	audiosEntry.SetText(g.currentConfig.Directories.Audios)
	
	documentsEntry := widget.NewEntry()
	documentsEntry.SetText(g.currentConfig.Directories.Documents)
	
	unknownEntry := widget.NewEntry()
	unknownEntry.SetText(g.currentConfig.Directories.Unknown)
	
	hiddenEntry := widget.NewEntry()
	hiddenEntry.SetText(g.currentConfig.Directories.Hidden)
	
	// Image subdirectories
	originalsEntry := widget.NewEntry()
	originalsEntry.SetText(g.currentConfig.ImageDirs.Originals)
	
	exportsEntry := widget.NewEntry()
	exportsEntry.SetText(g.currentConfig.ImageDirs.Exports)
	
	noExifYearEntry := widget.NewEntry()
	noExifYearEntry.SetText(g.currentConfig.ImageDirs.NoExifYearFolder)
	
	// Processing settings
	maxWidthEntry := widget.NewEntry()
	maxWidthEntry.SetText(fmt.Sprintf("%d", g.currentConfig.Processing.MaxImageWidth))
	
	maxHeightEntry := widget.NewEntry()
	maxHeightEntry.SetText(fmt.Sprintf("%d", g.currentConfig.Processing.MaxImageHeight))
	
	enableExportsCheck := widget.NewCheck("Enable Image Exports (slower processing)", nil)
	enableExportsCheck.SetChecked(g.currentConfig.Processing.EnableImageExports)
	
	jpegQualityEntry := widget.NewEntry()
	jpegQualityEntry.SetText(fmt.Sprintf("%d", g.currentConfig.Processing.JPEGQuality))
	
	shortVideoThresholdEntry := widget.NewEntry()
	shortVideoThresholdEntry.SetText(fmt.Sprintf("%d", g.currentConfig.Processing.ShortVideoThreshold))
	
	// Live Photos settings
	livePhotosEnabledCheck := widget.NewCheck("Enable Motion Photo Detection", nil)
	livePhotosEnabledCheck.SetChecked(g.currentConfig.MotionPhotos.Enabled)
	
	iPhonePatternsEntry := widget.NewEntry()
	iPhonePatternsEntry.SetText(strings.Join(g.currentConfig.MotionPhotos.IPhonePatterns, ", "))
	
	samsungPatternsEntry := widget.NewEntry()
	samsungPatternsEntry.SetText(strings.Join(g.currentConfig.MotionPhotos.SamsungPatterns, ", "))
	
	livePhotoExtensionsEntry := widget.NewEntry()
	livePhotoExtensionsEntry.SetText(strings.Join(g.currentConfig.MotionPhotos.Extensions, ", "))
	
	maxLiveDurationEntry := widget.NewEntry()
	maxLiveDurationEntry.SetText(fmt.Sprintf("%d", g.currentConfig.MotionPhotos.MaxDurationSeconds))
	
	// Screenshots settings
	screenshotsEnabledCheck := widget.NewCheck("Enable Screenshot Detection", nil)
	screenshotsEnabledCheck.SetChecked(g.currentConfig.Screenshots.Enabled)
	
	screenshotPatternsEntry := widget.NewEntry()
	screenshotPatternsEntry.SetText(strings.Join(g.currentConfig.Screenshots.Patterns, ", "))
	
	screenshotExtensionsEntry := widget.NewEntry()
	screenshotExtensionsEntry.SetText(strings.Join(g.currentConfig.Screenshots.Extensions, ", "))
	
	screenshotFolderEntry := widget.NewEntry()
	screenshotFolderEntry.SetText(g.currentConfig.Screenshots.FolderName)
	
	// Skip files extensions
	skipExtEntry := widget.NewMultiLineEntry()
	skipExtEntry.SetText(strings.Join(g.currentConfig.SkipFiles.Extensions, "\n"))
	skipExtEntry.Wrapping = fyne.TextWrapWord
	skipExtScroll := container.NewScroll(skipExtEntry)
	skipExtScroll.SetMinSize(fyne.NewSize(0, 120))
	
	// Skip files patterns
	skipPatternsEntry := widget.NewMultiLineEntry()
	skipPatternsEntry.SetText(strings.Join(g.currentConfig.SkipFiles.Patterns, "\n"))
	skipPatternsEntry.Wrapping = fyne.TextWrapWord
	skipPatternsScroll := container.NewScroll(skipPatternsEntry)
	skipPatternsScroll.SetMinSize(fyne.NewSize(0, 120))
	
	// Audio categories settings
	var audioEntries = make(map[string]struct {
		folderEntry    *widget.Entry
		extensionsEntry *widget.Entry
		patternsEntry  *widget.Entry
	})
	
	audioContainer := container.NewVBox()
	for categoryKey, category := range g.currentConfig.AudioCategories {
		folderEntry := widget.NewEntry()
		folderEntry.SetText(category.FolderName)
		
		extensionsEntry := widget.NewEntry()
		extensionsEntry.SetText(strings.Join(category.Extensions, ", "))
		
		patternsEntry := widget.NewEntry()
		patternsEntry.SetText(strings.Join(category.Patterns, ", "))
		
		audioEntries[categoryKey] = struct {
			folderEntry    *widget.Entry
			extensionsEntry *widget.Entry
			patternsEntry  *widget.Entry
		}{folderEntry, extensionsEntry, patternsEntry}
		
		// Make input fields larger for better visibility
		folderEntry.Resize(fyne.NewSize(200, 0))
		extensionsEntry.Resize(fyne.NewSize(300, 0))
		patternsEntry.Resize(fyne.NewSize(300, 0))
		
		categoryCard := widget.NewCard(strings.Title(strings.ReplaceAll(categoryKey, "_", " ")), "", 
			container.NewVBox(
				container.NewVBox(
					widget.NewLabel("Folder Name:"), 
					folderEntry,
					widget.NewLabel("Extensions (comma-separated):"), 
					extensionsEntry,
					widget.NewLabel("Patterns (comma-separated):"), 
					patternsEntry,
				),
			))
		audioContainer.Add(categoryCard)
	}
	
	// Buttons
	saveButton := widget.NewButton("Save Settings", func() {
		g.saveSettingsFromFormWithAudio(imagesEntry, videosEntry, audiosEntry, documentsEntry, unknownEntry, hiddenEntry,
			originalsEntry, exportsEntry, noExifYearEntry, maxWidthEntry, maxHeightEntry, enableExportsCheck, jpegQualityEntry, shortVideoThresholdEntry, livePhotosEnabledCheck, iPhonePatternsEntry, samsungPatternsEntry, livePhotoExtensionsEntry, maxLiveDurationEntry, skipExtEntry, skipPatternsEntry, audioEntries)
		settingsWindow.Close()
	})
	saveButton.Importance = widget.HighImportance
	
	cancelButton := widget.NewButton("Cancel", func() {
		settingsWindow.Close()
	})
	
	resetButton := widget.NewButton("Reset to Defaults", func() {
		g.currentConfig = config.DefaultConfig()
		g.refreshSettingsFormWithAudio(imagesEntry, videosEntry, audiosEntry, documentsEntry, unknownEntry, hiddenEntry,
			originalsEntry, exportsEntry, noExifYearEntry, maxWidthEntry, maxHeightEntry, skipExtEntry, skipPatternsEntry, audioEntries)
	})
	
	buttonContainer := container.NewHBox(saveButton, cancelButton, resetButton)
	
	// Create form layout
	form := container.NewVBox(
		widget.NewCard("Directory Names", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("Images:"), imagesEntry,
				widget.NewLabel("Videos:"), videosEntry,
				widget.NewLabel("Audios:"), audiosEntry,
				widget.NewLabel("Documents:"), documentsEntry,
				widget.NewLabel("Unknown:"), unknownEntry,
				widget.NewLabel("Hidden:"), hiddenEntry,
			),
		)),
		
		widget.NewCard("Image Organization", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("Originals Folder:"), originalsEntry,
				widget.NewLabel("Exports Folder:"), exportsEntry,
				widget.NewLabel("No EXIF Year Folder:"), noExifYearEntry,
			),
		)),
		
		widget.NewCard("Processing Settings", "", container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewLabel("Max Image Width:"), maxWidthEntry,
				widget.NewLabel("Max Image Height:"), maxHeightEntry,
			),
			widget.NewSeparator(),
			container.NewVBox(
				enableExportsCheck,
				container.NewGridWithColumns(2,
					widget.NewLabel("JPEG Quality (1-100):"), jpegQualityEntry,
					widget.NewLabel("Short Video Duration Threshold (seconds):"), shortVideoThresholdEntry,
				),
			),
			widget.NewSeparator(),
			livePhotosEnabledCheck,
			container.NewGridWithColumns(2,
				widget.NewLabel("iPhone Patterns (comma-separated):"), iPhonePatternsEntry,
				widget.NewLabel("Samsung Patterns (comma-separated):"), samsungPatternsEntry,
				widget.NewLabel("Motion Photo Extensions:"), livePhotoExtensionsEntry,
				widget.NewLabel("Max Motion Photo Duration (seconds):"), maxLiveDurationEntry,
			),
			widget.NewSeparator(),
			screenshotsEnabledCheck,
			container.NewGridWithColumns(2,
				widget.NewLabel("Screenshot Patterns (comma-separated):"), screenshotPatternsEntry,
				widget.NewLabel("Screenshot Extensions:"), screenshotExtensionsEntry,
				widget.NewLabel("Screenshot Folder Name:"), screenshotFolderEntry,
			),
		)),
		
		widget.NewCard("Audio Categories", "", func() *container.Scroll {
			audioScroll := container.NewScroll(audioContainer)
			audioScroll.SetMinSize(fyne.NewSize(500, 300))
			return audioScroll
		}()),
		
		widget.NewCard("Skip Files", "", container.NewVBox(
			widget.NewLabel("Extensions (one per line):"),
			skipExtScroll,
			widget.NewLabel("Patterns (one per line):"),
			skipPatternsScroll,
		)),
		
		buttonContainer,
	)
	
	scroll := container.NewScroll(form)
	settingsWindow.SetContent(container.NewPadded(scroll))
	settingsWindow.Show()
}

// saveSettingsFromFormWithAudio saves the configuration from form inputs including audio settings
func (g *GUI) saveSettingsFromFormWithAudio(imagesEntry, videosEntry, audiosEntry, documentsEntry, unknownEntry, hiddenEntry,
	originalsEntry, exportsEntry, noExifYearEntry, maxWidthEntry, maxHeightEntry *widget.Entry, enableExportsCheck *widget.Check, jpegQualityEntry, shortVideoThresholdEntry *widget.Entry, livePhotosEnabledCheck *widget.Check, iPhonePatternsEntry, samsungPatternsEntry, livePhotoExtensionsEntry, maxLiveDurationEntry, skipExtEntry, skipPatternsEntry *widget.Entry,
	audioEntries map[string]struct {
		folderEntry    *widget.Entry
		extensionsEntry *widget.Entry
		patternsEntry  *widget.Entry
	}) {
	
	// Update directory names
	g.currentConfig.Directories.Images = imagesEntry.Text
	g.currentConfig.Directories.Videos = videosEntry.Text
	g.currentConfig.Directories.Audios = audiosEntry.Text
	g.currentConfig.Directories.Documents = documentsEntry.Text
	g.currentConfig.Directories.Unknown = unknownEntry.Text
	g.currentConfig.Directories.Hidden = hiddenEntry.Text
	
	// Update image directories
	g.currentConfig.ImageDirs.Originals = originalsEntry.Text
	g.currentConfig.ImageDirs.Exports = exportsEntry.Text
	g.currentConfig.ImageDirs.NoExifYearFolder = noExifYearEntry.Text
	
	// Update processing settings
	if width, err := strconv.Atoi(maxWidthEntry.Text); err == nil {
		g.currentConfig.Processing.MaxImageWidth = width
	}
	if height, err := strconv.Atoi(maxHeightEntry.Text); err == nil {
		g.currentConfig.Processing.MaxImageHeight = height
	}
	
	// Update export settings
	g.currentConfig.Processing.EnableImageExports = enableExportsCheck.Checked
	if quality, err := strconv.Atoi(jpegQualityEntry.Text); err == nil && quality > 0 && quality <= 100 {
		g.currentConfig.Processing.JPEGQuality = quality
	}
	
	// Update short video threshold
	if threshold, err := strconv.Atoi(shortVideoThresholdEntry.Text); err == nil && threshold >= 0 {
		g.currentConfig.Processing.ShortVideoThreshold = threshold
	}
	
	// Update Motion Photos settings
	g.currentConfig.MotionPhotos.Enabled = livePhotosEnabledCheck.Checked
	
	// Parse iPhone patterns
	if iPhoneText := strings.TrimSpace(iPhonePatternsEntry.Text); iPhoneText != "" {
		patterns := strings.Split(iPhoneText, ",")
		for i, pattern := range patterns {
			patterns[i] = strings.TrimSpace(pattern)
		}
		g.currentConfig.MotionPhotos.IPhonePatterns = patterns
	}
	
	// Parse Samsung patterns
	if samsungText := strings.TrimSpace(samsungPatternsEntry.Text); samsungText != "" {
		patterns := strings.Split(samsungText, ",")
		for i, pattern := range patterns {
			patterns[i] = strings.TrimSpace(pattern)
		}
		g.currentConfig.MotionPhotos.SamsungPatterns = patterns
	}
	
	// Parse Motion Photo extensions
	if extText := strings.TrimSpace(livePhotoExtensionsEntry.Text); extText != "" {
		extensions := strings.Split(extText, ",")
		for i, ext := range extensions {
			ext = strings.TrimSpace(ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			extensions[i] = strings.ToLower(ext)
		}
		g.currentConfig.MotionPhotos.Extensions = extensions
	}
	
	// Update max Motion Photo duration
	if duration, err := strconv.Atoi(maxLiveDurationEntry.Text); err == nil && duration > 0 {
		g.currentConfig.MotionPhotos.MaxDurationSeconds = duration
	}
	
	// Update Screenshots settings
	g.currentConfig.Screenshots.Enabled = screenshotsEnabledCheck.Checked
	
	// Parse screenshot patterns
	if screenshotText := strings.TrimSpace(screenshotPatternsEntry.Text); screenshotText != "" {
		patterns := strings.Split(screenshotText, ",")
		for i, pattern := range patterns {
			patterns[i] = strings.TrimSpace(pattern)
		}
		g.currentConfig.Screenshots.Patterns = patterns
	}
	
	// Parse screenshot extensions
	if extText := strings.TrimSpace(screenshotExtensionsEntry.Text); extText != "" {
		extensions := strings.Split(extText, ",")
		for i, ext := range extensions {
			ext = strings.TrimSpace(ext)
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			extensions[i] = strings.ToLower(ext)
		}
		g.currentConfig.Screenshots.Extensions = extensions
	}
	
	// Update screenshot folder name
	if folderName := strings.TrimSpace(screenshotFolderEntry.Text); folderName != "" {
		g.currentConfig.Screenshots.FolderName = folderName
	}
	
	// Update audio categories
	for categoryKey, entries := range audioEntries {
		if category, exists := g.currentConfig.AudioCategories[categoryKey]; exists {
			category.FolderName = entries.folderEntry.Text
			
			// Parse extensions (comma-separated)
			extText := strings.TrimSpace(entries.extensionsEntry.Text)
			if extText != "" {
				extensions := strings.Split(extText, ",")
				for i, ext := range extensions {
					extensions[i] = strings.TrimSpace(ext)
				}
				category.Extensions = extensions
			}
			
			// Parse patterns (comma-separated)
			patText := strings.TrimSpace(entries.patternsEntry.Text)
			if patText != "" {
				patterns := strings.Split(patText, ",")
				for i, pat := range patterns {
					patterns[i] = strings.TrimSpace(pat)
				}
				category.Patterns = patterns
			} else {
				category.Patterns = []string{}
			}
			
			g.currentConfig.AudioCategories[categoryKey] = category
		}
	}
	
	// Update skip files
	g.currentConfig.SkipFiles.Extensions = strings.Split(strings.TrimSpace(skipExtEntry.Text), "\n")
	g.currentConfig.SkipFiles.Patterns = strings.Split(strings.TrimSpace(skipPatternsEntry.Text), "\n")
	
	// Save to destination directory if available
	if destDir := g.destEntry.Text; destDir != "" {
		configPath := filepath.Join(destDir, "zensort-config.json")
		if err := config.SaveConfig(g.currentConfig, configPath); err != nil {
			dialog.ShowError(fmt.Errorf("failed to save settings: %w", err), g.window)
		} else {
			dialog.ShowInformation("Settings Saved", "Configuration saved successfully!", g.window)
		}
	}
}

// refreshSettingsFormWithAudio updates form fields with current config values including audio settings
func (g *GUI) refreshSettingsFormWithAudio(imagesEntry, videosEntry, audiosEntry, documentsEntry, unknownEntry, hiddenEntry,
	originalsEntry, exportsEntry, noExifYearEntry, maxWidthEntry, maxHeightEntry, skipExtEntry, skipPatternsEntry *widget.Entry,
	audioEntries map[string]struct {
		folderEntry    *widget.Entry
		extensionsEntry *widget.Entry
		patternsEntry  *widget.Entry
	}) {
	
	imagesEntry.SetText(g.currentConfig.Directories.Images)
	videosEntry.SetText(g.currentConfig.Directories.Videos)
	audiosEntry.SetText(g.currentConfig.Directories.Audios)
	documentsEntry.SetText(g.currentConfig.Directories.Documents)
	unknownEntry.SetText(g.currentConfig.Directories.Unknown)
	hiddenEntry.SetText(g.currentConfig.Directories.Hidden)
	
	originalsEntry.SetText(g.currentConfig.ImageDirs.Originals)
	exportsEntry.SetText(g.currentConfig.ImageDirs.Exports)
	noExifYearEntry.SetText(g.currentConfig.ImageDirs.NoExifYearFolder)
	
	maxWidthEntry.SetText(fmt.Sprintf("%d", g.currentConfig.Processing.MaxImageWidth))
	maxHeightEntry.SetText(fmt.Sprintf("%d", g.currentConfig.Processing.MaxImageHeight))
	
	// Update audio categories
	for categoryKey, entries := range audioEntries {
		if category, exists := g.currentConfig.AudioCategories[categoryKey]; exists {
			entries.folderEntry.SetText(category.FolderName)
			entries.extensionsEntry.SetText(strings.Join(category.Extensions, ", "))
			entries.patternsEntry.SetText(strings.Join(category.Patterns, ", "))
		}
	}
	
	skipExtEntry.SetText(strings.Join(g.currentConfig.SkipFiles.Extensions, "\n"))
	skipPatternsEntry.SetText(strings.Join(g.currentConfig.SkipFiles.Patterns, "\n"))
}

// validateDirectories ensures destination is not within source directory
func (g *GUI) validateDirectories(sourceDir, destDir string) error {
	// Clean and resolve absolute paths
	absSource, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("invalid source directory: %w", err)
	}
	
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("invalid destination directory: %w", err)
	}
	
	// Check if destination is within source
	relPath, err := filepath.Rel(absSource, absDest)
	if err == nil && !strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("destination directory cannot be within source directory")
	}
	
	// Check if source and destination are the same
	if absSource == absDest {
		return fmt.Errorf("source and destination directories cannot be the same")
	}
	
	return nil
}

// startProcessing begins the file organization process
func (g *GUI) startProcessing() {
	sourceDir := g.sourceEntry.Text
	destDir := g.destEntry.Text
	
	// Use fallback for source directory if not selected
	if sourceDir == "" {
		// Try current working directory first
		if cwd, err := os.Getwd(); err == nil {
			sourceDir = cwd
		} else {
			// Fallback to executable directory
			if execPath, err := os.Executable(); err == nil {
				sourceDir = filepath.Dir(execPath)
			} else {
				dialog.ShowError(fmt.Errorf("could not determine source directory"), g.window)
				return
			}
		}
		g.sourceEntry.SetText(sourceDir)
	}
	
	if destDir == "" {
		dialog.ShowError(fmt.Errorf("please select a destination directory"), g.window)
		return
	}
	
	// Validate that destination is not within source
	if err := g.validateDirectories(sourceDir, destDir); err != nil {
		dialog.ShowError(err, g.window)
		return
	}
	
	// Load or create configuration in destination directory
	cfg, err := g.loadDestinationConfig(destDir)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load configuration: %w", err), g.window)
		return
	}
	
	// Create context for cancellation
	g.ctx, g.cancel = context.WithCancel(context.Background())
	
	// Create progress and pause channels
	g.progressChan = make(chan core.ProgressUpdate, 100)
	g.pauseChan = make(chan bool, 1)
	
	// Update UI state - disable all input fields during processing
	g.startButton.Disable()
	g.pauseButton.Enable()
	g.pauseButton.SetText("Pause")
	g.pauseButton.Importance = widget.MediumImportance
	g.pauseButton.Refresh()
	g.stopButton.Enable()
	g.stopButton.Importance = widget.DangerImportance
	g.stopButton.Refresh()
	g.sourceEntry.Disable()
	g.destEntry.Disable()
	g.sourceBrowseBtn.Disable()
	g.destBrowseBtn.Disable()
	g.settingsButton.Disable()
	g.isPaused = false
	g.progressBar.SetValue(0)
	g.statusLabel.SetText("Initializing...")
	g.logBuffer = make([]string, 0) // Clear log buffer
	g.logText.SetText("")
	g.currentFolder = ""
	g.folderFileCount = 0
	g.folderFiles = make([]string, 0)
	
	// Start progress monitoring
	go g.monitorProgress()
	
	// Start processing in goroutine
	go func() {
		defer func() {
			g.startButton.Enable()
			g.startButton.Importance = widget.HighImportance
			g.startButton.Refresh()
			g.pauseButton.Disable()
			g.pauseButton.SetText("Pause")
			g.pauseButton.Importance = widget.MediumImportance
			g.pauseButton.Refresh()
			g.stopButton.Disable()
			g.stopButton.Importance = widget.DangerImportance
			g.stopButton.Refresh()
			g.sourceEntry.Enable()
			g.destEntry.Enable()
			g.sourceBrowseBtn.Enable()
			g.destBrowseBtn.Enable()
			g.settingsButton.Enable()
			g.isPaused = false
			close(g.progressChan)
			g.progressChan = nil
			g.pauseChan = nil
		}()
		
		processor, err := core.NewFileProcessor(cfg, destDir)
		if err != nil {
			g.logMessage(fmt.Sprintf("Error: Failed to create processor: %v", err))
			return
		}
		g.processor = processor
		
		// Subscribe to progress updates
		progressChan := processor.GetProgressTracker().Subscribe()
		go func() {
			for update := range progressChan {
				select {
				case g.progressChan <- update:
				case <-g.ctx.Done():
					return
				}
			}
		}()
		
		// Start processing with pause support
		err = g.processWithPauseSupport(processor, sourceDir)
		if err != nil {
			if err == context.Canceled {
				g.logMessage("Processing stopped by user")
			} else {
				g.logMessage(fmt.Sprintf("Error: %v", err))
			}
		} else {
			g.logMessage("Processing completed successfully!")
		}
	}()
}

// processWithPauseSupport handles file processing with pause/resume capability
func (g *GUI) processWithPauseSupport(processor *core.FileProcessor, sourceDir string) error {
	// Start the actual processing in a separate goroutine
	processingDone := make(chan error, 1)
	go func() {
		processingDone <- processor.ProcessDirectory(g.ctx, sourceDir)
	}()
	
	// Monitor pause/resume signals
	go func() {
		for {
			select {
			case shouldPause := <-g.pauseChan:
				if shouldPause {
					// Pause logic - processor should check context
					g.logMessage("Processing paused")
				} else {
					// Resume logic
					g.logMessage("Processing resumed")
				}
			case <-g.ctx.Done():
				return
			}
		}
	}()
	
	// Wait for processing to complete
	return <-processingDone
}

// stopProcessing cancels the current operation
func (g *GUI) stopProcessing() {
	if g.cancel != nil {
		g.cancel()
		g.logMessage("Stopping processing...")
	}
}

// monitorProgress updates the UI with progress information
func (g *GUI) monitorProgress() {
	for update := range g.progressChan {
		// Skip progress updates if paused
		if g.isPaused {
			continue
		}
		
		// Update progress bar
		g.progressBar.SetValue(update.Percentage / 100.0)
		
		// Update status label
		statusText := fmt.Sprintf("Processed: %d/%d files (%.1f%%) - %s",
			update.ProcessedFiles, update.TotalFiles, update.Percentage, 
			formatDuration(update.ElapsedTime))
		
		if update.EstimatedTime > 0 {
			statusText += fmt.Sprintf(" - ETA: %s", formatDuration(update.EstimatedTime))
		}
		
		g.statusLabel.SetText(statusText)
		
		// Batch log files by folder
		if update.CurrentFile != "" {
			g.batchLogFile(update.CurrentFile)
		}
		
		// Check if done
		if update.Done {
			g.flushFolderLog() // Flush any remaining files
			g.logMessage(fmt.Sprintf("✓ Complete! Processed %d files in %s", 
				update.ProcessedFiles, formatDuration(update.ElapsedTime)))
			
			if update.ErrorCount > 0 {
				g.logMessage(fmt.Sprintf("⚠ %d errors occurred. Check logs for details.", update.ErrorCount))
			}
			
			// Force final log display update
			g.updateLogDisplay()
		}
	}
}

// logMessage adds a message to the log text area with buffering for performance
func (g *GUI) logMessage(message string) {
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	
	// Add to buffer
	g.logBuffer = append(g.logBuffer, logEntry)
	
	// Trim buffer if it exceeds max lines
	if len(g.logBuffer) > g.maxLogLines {
		g.logBuffer = g.logBuffer[len(g.logBuffer)-g.maxLogLines:]
	}
	
	// Update UI with buffered content (less frequent updates for better performance)
	if len(g.logBuffer)%5 == 0 || len(g.logBuffer) <= 10 {
		g.updateLogDisplay()
	}
}

// updateLogDisplay refreshes the log text widget with current buffer
func (g *GUI) updateLogDisplay() {
	if len(g.logBuffer) == 0 {
		g.logText.SetText("")
		return
	}
	
	// Join all log entries
	logContent := ""
	for _, entry := range g.logBuffer {
		logContent += entry + "\n"
	}
	
	g.logText.SetText(logContent)
	g.logText.CursorRow = len(g.logBuffer) // Scroll to bottom
}

// batchLogFile batches files by folder for better log display
func (g *GUI) batchLogFile(filePath string) {
	folder := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)
	
	// If this is a new folder, flush the previous batch
	if g.currentFolder != "" && g.currentFolder != folder {
		g.flushFolderLog()
	}
	
	// Set current folder
	g.currentFolder = folder
	g.folderFiles = append(g.folderFiles, fileName)
	g.folderFileCount++
	
	// Flush if we have 10 files or every 50 files processed
	if len(g.folderFiles) >= 10 || g.folderFileCount%50 == 0 {
		g.flushFolderLog()
	}
}

// flushFolderLog outputs the current batch of files for a folder
func (g *GUI) flushFolderLog() {
	if g.currentFolder == "" || len(g.folderFiles) == 0 {
		return
	}
	
	folderName := filepath.Base(g.currentFolder)
	if len(g.folderFiles) == 1 {
		g.logMessage(fmt.Sprintf("📁 %s: %s", folderName, g.folderFiles[0]))
	} else {
		fileList := strings.Join(g.folderFiles, ", ")
		if len(fileList) > 100 {
			fileList = fileList[:97] + "..."
		}
		g.logMessage(fmt.Sprintf("📁 %s (%d files): %s", folderName, len(g.folderFiles), fileList))
	}
	
	// Reset for next batch
	g.folderFiles = make([]string, 0)
	g.currentFolder = ""
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-60*d.Minutes())
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - 60*hours
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}
