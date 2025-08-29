package main

import (
	"flag"
	"fmt"
	"os"

	"zensort/internal/cli"
)

func main() {
	var source = flag.String("source", "", "Source directory path")
	var dest = flag.String("dest", "", "Destination directory path")
	var config = flag.String("config", "", "Configuration file path")
	
	flag.Parse()

	if *source == "" || *dest == "" {
		fmt.Println("ZenSort CLI - File Organizer")
		fmt.Println("Usage: zensort-cli -source <path> -dest <path> [-config <path>]")
		os.Exit(1)
	}
	
	cli.Run(*source, *dest, *config)
}
