package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/raitses/ask/internal/config"
	"github.com/raitses/ask/internal/context"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	// Define flags
	analyze := flag.Bool("analyze", false, "Analyze directory structure before responding")
	analyzeShort := flag.Bool("a", false, "Analyze directory structure before responding (short)")
	reset := flag.Bool("reset", false, "Clear conversation context for current directory")
	resetShort := flag.Bool("r", false, "Clear conversation context for current directory (short)")
	info := flag.Bool("info", false, "Show context information")
	infoShort := flag.Bool("i", false, "Show context information (short)")
	showVersion := flag.Bool("version", false, "Show version information")
	versionShort := flag.Bool("v", false, "Show version information (short)")
	showHelp := flag.Bool("help", false, "Show help message")
	helpShort := flag.Bool("h", false, "Show help message (short)")

	flag.Parse()

	// Combine short and long flags
	*analyze = *analyze || *analyzeShort
	*reset = *reset || *resetShort
	*info = *info || *infoShort
	*showVersion = *showVersion || *versionShort
	*showHelp = *showHelp || *helpShort

	// Handle special flags
	if *showVersion {
		fmt.Printf("ask version %s\n", version)
		if commit != "unknown" {
			fmt.Printf("commit: %s\n", commit)
		}
		if date != "unknown" {
			fmt.Printf("built: %s\n", date)
		}
		os.Exit(0)
	}

	if *showHelp {
		printHelp()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		os.Exit(2)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintf(os.Stderr, "Set it with: export ASK_API_KEY='your-api-key'\n")
		os.Exit(2)
	}

	// Create context manager
	manager, err := context.NewManager(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize context: %v\n", err)
		os.Exit(3)
	}

	// Handle reset command
	if *reset {
		if err := manager.Reset(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to reset context: %v\n", err)
			os.Exit(3)
		}
		fmt.Println("Context reset successfully")
		os.Exit(0)
	}

	// Handle info command
	if *info {
		fmt.Print(manager.GetInfo())
		os.Exit(0)
	}

	// Get query from remaining arguments
	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	query := strings.Join(args, " ")

	// Perform analysis if requested
	if *analyze {
		fmt.Fprintln(os.Stderr, "Analyzing directory structure...")
		err := manager.Analyze()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Analysis failed: %v\n", err)
			// Continue with query even if analysis fails
		}
		if err == nil {
			fmt.Fprintln(os.Stderr, "Analysis complete.")
		}
	}

	// Execute query
	response, err := manager.Query(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(response)
}

func printUsage() {
	fmt.Println("Usage: ask [OPTIONS] <query>")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -a, --analyze      Analyze directory structure before responding")
	fmt.Println("  -r, --reset        Clear conversation context for current directory")
	fmt.Println("  -i, --info         Show context information")
	fmt.Println("  -h, --help         Show this help message")
	fmt.Println("  -v, --version      Show version information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ask how do I run tests")
	fmt.Println("  ask \"how does this work?\"")
	fmt.Println("  ask --analyze what is the project structure")
	fmt.Println("  ask --reset")
	fmt.Println("  ask --info")
}

func printHelp() {
	printUsage()
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  ASK_API_KEY        API key for LLM provider (required for OpenAI)")
	fmt.Println("  ASK_MODEL          Model to use (default: gpt-4o)")
	fmt.Println("  ASK_OS             Operating system (default: macOS)")
	fmt.Println("  ASK_API_URL        API endpoint (default: OpenAI)")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Config files are loaded in this order:")
	fmt.Println("  1. ~/.config/ask/.env (global)")
	fmt.Println("  2. ./.env (local, overrides global)")
	fmt.Println("  3. Environment variables (highest priority)")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/raitses/ask")
}
