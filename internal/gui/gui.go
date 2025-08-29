package gui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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
	configEntry    *widget.Entry
	progressBar    *widget.ProgressBar
	statusLabel    *widget.Label
	logText        *widget.Entry
	startButton    *widget.Button
	stopButton     *widget.Button
	processor      *core.FileProcessor
	ctx            context.Context
	cancel         context.CancelFunc
	progressChan   chan core.ProgressUpdate
}

// Launch starts the GUI application
func Launch() {
	gui := NewGUI()
	gui.setupUI()
	gui.window.ShowAndRun()
}

// NewGUI creates a new GUI instance
func NewGUI() *GUI {
	a := app.New()
	a.SetIcon(nil) // You can set an icon here
	
	w := a.NewWindow("ZenSort - File Organizer")
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()
	
	return &GUI{
		app:    a,
		window: w,
	}
}

// setupUI creates and arranges the user interface elements
func (g *GUI) setupUI() {
	// Source directory selection
	g.sourceEntry = widget.NewEntry()
	g.sourceEntry.SetPlaceHolder("Select source directory...")
	sourceBrowse := widget.NewButton("Browse", g.browseSource)
	sourceContainer := container.NewBorder(nil, nil, nil, sourceBrowse, g.sourceEntry)
	
	// Destination directory selection
	g.destEntry = widget.NewEntry()
	g.destEntry.SetPlaceHolder("Select destination directory...")
	destBrowse := widget.NewButton("Browse", g.browseDestination)
	destContainer := container.NewBorder(nil, nil, nil, destBrowse, g.destEntry)
	
	// Configuration file selection
	g.configEntry = widget.NewEntry()
	g.configEntry.SetPlaceHolder("Optional: Select configuration file...")
	configBrowse := widget.NewButton("Browse", g.browseConfig)
	configContainer := container.NewBorder(nil, nil, nil, configBrowse, g.configEntry)
	
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
	
	// Control buttons
	g.startButton = widget.NewButton("Start Organization", g.startProcessing)
	g.startButton.Importance = widget.HighImportance
	
	g.stopButton = widget.NewButton("Stop", g.stopProcessing)
	g.stopButton.Disable()
	
	buttonContainer := container.NewHBox(g.startButton, g.stopButton)
	
	// Create form layout
	form := container.NewVBox(
		widget.NewLabel("Source Directory:"),
		sourceContainer,
		widget.NewLabel("Destination Directory:"),
		destContainer,
		widget.NewLabel("Configuration File (Optional):"),
		configContainer,
		widget.NewSeparator(),
		widget.NewLabel("Progress:"),
		g.progressBar,
		g.statusLabel,
		widget.NewSeparator(),
		buttonContainer,
		widget.NewLabel("Processing Log:"),
		logScroll,
	)
	
	// Set content
	g.window.SetContent(container.NewPadded(form))
}

// browseSource opens file dialog for source directory
func (g *GUI) browseSource() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		g.sourceEntry.SetText(uri.Path())
	}, g.window)
}

// browseDestination opens file dialog for destination directory
func (g *GUI) browseDestination() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil || uri == nil {
			return
		}
		g.destEntry.SetText(uri.Path())
	}, g.window)
}

// browseConfig opens file dialog for configuration file
func (g *GUI) browseConfig() {
	dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
		if err != nil || uri == nil {
			return
		}
		defer uri.Close()
		g.configEntry.SetText(uri.URI().Path())
	}, g.window)
}

// startProcessing begins the file organization process
func (g *GUI) startProcessing() {
	sourceDir := g.sourceEntry.Text
	destDir := g.destEntry.Text
	configFile := g.configEntry.Text
	
	// Validate inputs
	if sourceDir == "" {
		dialog.ShowError(fmt.Errorf("please select a source directory"), g.window)
		return
	}
	
	if destDir == "" {
		dialog.ShowError(fmt.Errorf("please select a destination directory"), g.window)
		return
	}
	
	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to load configuration: %w", err), g.window)
		return
	}
	
	// Create context for cancellation
	g.ctx, g.cancel = context.WithCancel(context.Background())
	
	// Create progress channel
	g.progressChan = make(chan core.ProgressUpdate, 100)
	
	// Update UI state
	g.startButton.Disable()
	g.stopButton.Enable()
	g.progressBar.SetValue(0)
	g.statusLabel.SetText("Initializing...")
	g.logText.SetText("")
	
	// Start progress monitoring
	go g.monitorProgress()
	
	// Start processing in goroutine
	go func() {
		defer func() {
			g.startButton.Enable()
			g.stopButton.Disable()
			close(g.progressChan)
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
		
		// Start processing
		err = processor.ProcessDirectory(g.ctx, sourceDir)
		if err != nil {
			g.logMessage(fmt.Sprintf("Error: Processing failed: %v", err))
		} else {
			g.logMessage("Processing completed successfully!")
		}
	}()
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
		
		// Log current file
		if update.CurrentFile != "" {
			g.logMessage(fmt.Sprintf("Processing: %s", filepath.Base(update.CurrentFile)))
		}
		
		// Check if done
		if update.Done {
			g.logMessage(fmt.Sprintf("✓ Complete! Processed %d files in %s", 
				update.ProcessedFiles, formatDuration(update.ElapsedTime)))
			
			if update.ErrorCount > 0 {
				g.logMessage(fmt.Sprintf("⚠ %d errors occurred. Check logs for details.", update.ErrorCount))
			}
		}
	}
}

// logMessage adds a message to the log text area
func (g *GUI) logMessage(message string) {
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)
	
	// Append to existing text
	currentText := g.logText.Text
	g.logText.SetText(currentText + logEntry)
	
	// Auto-scroll to bottom
	g.logText.CursorRow = len(g.logText.Text)
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
