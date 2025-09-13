package lib

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// imageExporter is the concrete implementation of ImageExporter interface.
// It provides methods to extract Docker image configurations from container registries.
type imageExporter struct{}

// NewImageExporter creates a new instance of ImageExporter.
// This is the primary entry point for creating an image exporter that can
// interact with Docker registries to extract image configurations and filesystems.
func NewImageExporter() ImageExporter {
	return &imageExporter{}
}

// GetImageConfig retrieves the configuration of a Docker image from a registry.
//
// This method fetches the image manifest and configuration blob from the registry
// without downloading the actual layer data. It's much faster than a full image pull
// and only requires network access to the registry.
//
// Parameters:
//   - imageRef: Docker image reference (e.g., "nginx:latest", "registry.com/org/image:v1.0")
//   - auth: Optional authentication configuration for private registries
//
// Returns:
//   - *ImageConfig: The image configuration containing entrypoint, command, environment, etc.
//   - error: Any error encountered during the operation
//
// Example:
//
//	exporter := NewImageExporter()
//	config, err := exporter.GetImageConfig("nginx:alpine", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Entrypoint: %v\n", config.Entrypoint)
func (e *imageExporter) GetImageConfig(imageRef string, auth *AuthConfig) (*ImageConfig, error) {
	// Parse the image reference to ensure it's valid and extract registry/repository information
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}

	// Configure authentication for registry access
	var authOption remote.Option
	if auth != nil {
		// Use provided credentials for private registries
		authOption = remote.WithAuth(&authn.Basic{
			Username: auth.Username,
			Password: auth.Password,
		})
	} else {
		// Fall back to system keychain (Docker credentials, etc.)
		authOption = remote.WithAuthFromKeychain(authn.DefaultKeychain)
	}

	// Fetch the image metadata from the registry
	// This downloads the manifest and config blob but not the layer data
	image, err := remote.Image(ref, authOption)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image %s: %w", imageRef, err)
	}

	// Extract the configuration file from the image
	configFile, err := image.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get config file: %w", err)
	}

	// Convert the registry config format to our simplified format
	config := &ImageConfig{
		User:       configFile.Config.User,
		Entrypoint: configFile.Config.Entrypoint,
		Cmd:        configFile.Config.Cmd,
		WorkingDir: configFile.Config.WorkingDir,
		Env:        configFile.Config.Env,
		Labels:     configFile.Config.Labels,
	}

	return config, nil
}
