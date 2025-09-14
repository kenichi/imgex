// Package main provides the command-line interface for imgex.
//
// imgex is a tool for extracting Docker image configurations and filesystems
// directly from container registries without requiring a running Docker daemon.
// It supports both public and private registries with authentication.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kenichi/imgex/lib"
	"github.com/spf13/cobra"
)

// Version information
const (
	version = "0.1.0"
	description = "Docker image export tool without Docker daemon"
)

// Global flags for authentication, shared across all commands
var (
	username string // Registry username for private registries
	password string // Registry password for private registries
	registry string // Registry URL (optional, defaults to Docker Hub)
)

// main is the entry point for the imgex CLI application.
// It executes the root command and handles any top-level errors.
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// rootCmd defines the base command when called without any subcommands.
// It provides global flags and serves as the parent for all subcommands.
var rootCmd = &cobra.Command{
	Use:     "imgex",
	Short:   description,
	Version: version,
	Long: `imgex is a tool for extracting Docker image configurations and
filesystems directly from registries without requiring a running Docker daemon.

Examples:
  imgex config nginx:latest
  imgex filesystem alpine:latest > alpine.tar
  imgex filesystem --output nginx.tar nginx:alpine
  imgex --username user --password pass config private.registry.com/image:tag`,
}

// configCmd handles the 'config' subcommand for extracting image configurations.
// It fetches image metadata from registries and outputs the configuration as JSON.
var configCmd = &cobra.Command{
	Use:   "config <image-reference>",
	Short: "Extract image configuration (ENTRYPOINT, CMD, USER, etc.)",
	Long: `Extract the configuration of a Docker image from a registry.

This command fetches the image manifest and configuration blob without
downloading layer data, making it fast and efficient for configuration inspection.

The output includes:
- User: The user account for running the container
- Entrypoint: The default executable and arguments
- Cmd: Default command arguments
- WorkingDir: The working directory for commands
- Env: Environment variables
- Labels: Metadata labels

Examples:
  imgex config nginx:latest
  imgex config --username user --password pass private.registry.com/image:tag`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigCommand,
}

// filesystemCmd handles the 'filesystem' subcommand for exporting image filesystems.
// It downloads all image layers and reconstructs the complete filesystem as a tar archive.
var filesystemCmd = &cobra.Command{
	Use:   "filesystem <image-reference>",
	Short: "Export complete filesystem as tar archive",
	Long: `Export the complete filesystem of a Docker image as a tar archive.

This command downloads all layers of the image and reconstructs the flattened
filesystem, equivalent to what 'docker export' produces. The output can be
written to a file or streamed to stdout for piping to other tools.

Examples:
  imgex filesystem alpine:latest > alpine.tar
  imgex filesystem --output nginx.tar nginx:alpine
  imgex filesystem ubuntu:latest | tar -tv  # List contents`,
	Args: cobra.ExactArgs(1),
	RunE: runFilesystemCommand,
}

// runConfigCommand implements the logic for the 'config' subcommand.
// It creates an authenticated exporter, fetches the image configuration,
// and outputs it as formatted JSON.
func runConfigCommand(cmd *cobra.Command, args []string) error {
	imageRef := args[0]

	// Build authentication configuration if credentials are provided
	auth := buildAuthConfig()

	// Create exporter and fetch image configuration
	exporter := lib.NewImageExporter()
	config, err := exporter.GetImageConfig(imageRef, auth)
	if err != nil {
		return fmt.Errorf("failed to get image config: %w", err)
	}

	// Format and output the configuration as JSON
	output, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// runFilesystemCommand implements the logic for the 'filesystem' subcommand.
// It creates an authenticated exporter and exports the image filesystem,
// either to a specified file or to stdout for streaming.
func runFilesystemCommand(cmd *cobra.Command, args []string) error {
	imageRef := args[0]
	outputPath, _ := cmd.Flags().GetString("output")

	// Build authentication configuration if credentials are provided
	auth := buildAuthConfig()

	// Create exporter
	exporter := lib.NewImageExporter()

	// Export to file or stdout based on flags
	if outputPath != "" {
		// Export to specified file
		err := exporter.ExportImageFilesystem(imageRef, outputPath, auth)
		if err != nil {
			return fmt.Errorf("failed to export filesystem: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Filesystem exported to %s\n", outputPath)
	} else {
		// Stream to stdout for piping
		err := exporter.ExportImageFilesystemToWriter(imageRef, os.Stdout, auth)
		if err != nil {
			return fmt.Errorf("failed to export filesystem: %w", err)
		}
	}

	return nil
}

// buildAuthConfig creates an AuthConfig from global flags if credentials are provided.
// Returns nil if no authentication is configured, which will use system defaults.
func buildAuthConfig() *lib.AuthConfig {
	if username != "" || password != "" {
		return &lib.AuthConfig{
			Username: username,
			Password: password,
			Registry: registry,
		}
	}
	return nil
}

// init sets up the CLI command structure and flags.
// It registers subcommands and configures global and command-specific flags.
func init() {
	// Register subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(filesystemCmd)

	// Global flags for authentication (available to all commands)
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "",
		"Registry username for private registries")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "",
		"Registry password for private registries")
	rootCmd.PersistentFlags().StringVarP(&registry, "registry", "r", "",
		"Registry URL (defaults to Docker Hub)")

	// Command-specific flags
	filesystemCmd.Flags().StringP("output", "o", "",
		"Output file path (default: stdout)")
}
