package flags

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Config holds all command-line configuration
type Config struct {
	Port    string
	DataDir string
	Help    bool
}

// DefaultConfig returns default configuration values
func DefaultConfig() Config {
	return Config{
		Port:    "8080",
		DataDir: "./data",
		Help:    false,
	}
}

// Parse parses command-line flags and returns configuration
func Parse() Config {
	config := DefaultConfig()

	// Define flags
	var (
		port    = flag.String("port", config.Port, "Port number")
		dataDir = flag.String("dir", config.DataDir, "Path to the data directory")
		help    = flag.Bool("help", false, "Show this screen")
	)

	// Custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Coffee Shop Management System\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  hot-coffee [--port <N>] [--dir <S>]\n")
		fmt.Fprintf(os.Stderr, "  hot-coffee --help\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  --help       Show this screen.\n")
		fmt.Fprintf(os.Stderr, "  --port N     Port number (1-65535).\n")
		fmt.Fprintf(os.Stderr, "  --dir S      Path to the data directory.\n")
	}

	// Parse flags
	flag.Parse()

	// Handle help flag
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate and clean data directory path
	validatedDir := validateDataDir(*dataDir)

	// Validate port
	if err := validatePort(*port); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	return Config{
		Port:    *port,
		DataDir: validatedDir,
		Help:    *help,
	}
}

// validateDataDir validates and normalizes the data directory path
func validateDataDir(dir string) string {
	// Convert to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid data directory path: %v\n", err)
		os.Exit(1)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Cannot create data directory: %v\n", err)
		os.Exit(1)
	}

	return absDir
}

// validatePort validates the port number
func validatePort(port string) error {
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}

	// Convert to integer
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number '%s': must be a number", port)
	}

	// Check port range (1-65535)
	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port number %d is out of range: must be between 1 and 65535", portNum)
	}

	// Warn about privileged ports (1-1023)
	if portNum < 1024 {
		fmt.Fprintf(os.Stderr, "Warning: Port %d is a privileged port (1-1023). You may need administrator privileges.\n", portNum)
	}

	return nil
}

// Validate validates the parsed configuration
func (c Config) Validate() error {
	// Validate port
	if err := validatePort(c.Port); err != nil {
		return err
	}

	// Validate data directory
	if c.DataDir == "" {
		return fmt.Errorf("data directory cannot be empty")
	}

	return nil
}
