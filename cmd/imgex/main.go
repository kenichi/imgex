package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ken/imgex/lib"
	"github.com/spf13/cobra"
)

var (
	username string
	password string
	registry string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "imgex",
	Short: "Docker image export tool without Docker daemon",
	Long: `imgex is a tool for extracting Docker image configurations and 
filesystems directly from registries without requiring a running Docker daemon.`,
}

var configCmd = &cobra.Command{
	Use:   "config <image-reference>",
	Short: "Extract image configuration (ENTRYPOINT, CMD, USER, etc.)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]
		
		var auth *lib.AuthConfig
		if username != "" || password != "" {
			auth = &lib.AuthConfig{
				Username: username,
				Password: password,
				Registry: registry,
			}
		}

		exporter := lib.NewImageExporter()
		config, err := exporter.GetImageConfig(imageRef, auth)
		if err != nil {
			return fmt.Errorf("failed to get image config: %w", err)
		}

		output, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Println(string(output))
		return nil
	},
}

var filesystemCmd = &cobra.Command{
	Use:   "filesystem <image-reference>",
	Short: "Export complete filesystem as tar archive",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef := args[0]
		outputPath, _ := cmd.Flags().GetString("output")
		
		var auth *lib.AuthConfig
		if username != "" || password != "" {
			auth = &lib.AuthConfig{
				Username: username,
				Password: password,
				Registry: registry,
			}
		}

		exporter := lib.NewImageExporter()
		
		if outputPath != "" {
			err := exporter.ExportImageFilesystem(imageRef, outputPath, auth)
			if err != nil {
				return fmt.Errorf("failed to export filesystem: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Filesystem exported to %s\n", outputPath)
		} else {
			err := exporter.ExportImageFilesystemToWriter(imageRef, os.Stdout, auth)
			if err != nil {
				return fmt.Errorf("failed to export filesystem: %w", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(filesystemCmd)

	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "", "Registry username")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Registry password")
	rootCmd.PersistentFlags().StringVarP(&registry, "registry", "r", "", "Registry URL")

	filesystemCmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")
}