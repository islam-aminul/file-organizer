package main

import (
	"flag"
	"fmt"
	"os"

	"zensort/internal/cli"
	"zensort/internal/gui"
)

func main() {
	var useCLI = flag.Bool("cli", false, "Force command-line interface")
	var source = flag.String("source", "", "Source directory path")
	var dest = flag.String("dest", "", "Destination directory path")
	var config = flag.String("config", "", "Configuration file path")
	
	flag.Parse()

	// Determine interface mode
	if *useCLI || (*source != "" && *dest != "") {
		// Use CLI mode if explicitly requested or if source/dest provided
		if *source == "" || *dest == "" {
			fmt.Println("ZenSort - Cross-Platform File Organizer")
			fmt.Println("=====================================")
			fmt.Println()
			fmt.Println("Usage:")
			fmt.Println("  GUI Mode (default):    zensort")
			fmt.Println("  GUI Mode (explicit):   zensort -gui")
			fmt.Println("  CLI Mode:              zensort -source <path> -dest <path> [-config <path>]")
			fmt.Println("  CLI Mode (explicit):   zensort -cli -source <path> -dest <path>")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  zensort")
			fmt.Println("  zensort -gui")
			fmt.Println("  zensort -source \"C:\\Source\" -dest \"C:\\Organized\"")
			fmt.Println("  zensort -cli -source \"./files\" -dest \"./sorted\"")
			os.Exit(1)
		}
		
		cli.Run(*source, *dest, *config)
	} else {
		// Default to GUI mode
		gui.Launch()
	}
}
