package flags

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config holds all command-line configuration
// UPDATED: Removed DataDir field as we now use database storage
type Config struct {
	Port string
	Help bool
}

// DefaultConfig returns default configuration values
// UPDATED: Removed DataDir default as we now use database storage
func DefaultConfig() Config {
	return Config{
		Port: "8080",
		Help: false,
	}
}

// Parse parses command-line flags and returns configuration
func Parse() Config {
	config := DefaultConfig()

	// Define flags
	// UPDATED: Removed dataDir flag as we now use database storage
	var (
		port = flag.String("port", config.Port, "Port number")
		help = flag.Bool("help", false, "Show this screen")
	)

	// Custom usage function
	// UPDATED: Removed data directory option from usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Coffee Shop Management System\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  hot-coffee [--port <N>]\n")
		fmt.Fprintf(os.Stderr, "  hot-coffee --help\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  --help       Show this screen.\n")
		fmt.Fprintf(os.Stderr, "  --port N     Port number (1-65535).\n")
	}

	// Parse flags
	flag.Parse()

	// Handle help flag
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Validate port
	if err := validatePort(*port); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	return Config{
		Port: *port,
		Help: *help,
	}
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
// UPDATED: Removed data directory validation as we now use database storage
func (c Config) Validate() error {
	// Validate port
	if err := validatePort(c.Port); err != nil {
		return err
	}

	return nil
}
