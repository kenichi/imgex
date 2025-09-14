// Package lib provides Docker image export functionality without requiring a Docker daemon.
//
// This package enables extraction of Docker image configurations and filesystems
// directly from container registries using the Docker Registry API. It supports
// both public and private registries with authentication.
package lib

import "io"

// Version information for imgex
const (
	// Version is the current version of imgex
	Version = "0.1.2"

	// Description is a short description of imgex
	Description = "Docker image export tool without Docker daemon"
)

// ImageConfig represents the configuration of a Docker image.
// This structure contains the essential configuration fields that define
// how a container should be run, extracted from the image manifest.
type ImageConfig struct {
	// User specifies the username or UID which the process in the container should run as.
	// Empty string means root user.
	User string `json:"user"`

	// Entrypoint defines a list of arguments to use as the command to execute when the container starts.
	// If nil, the default entrypoint from the base image is used.
	Entrypoint []string `json:"entrypoint"`

	// Cmd provides defaults for an executing container. These defaults can include an executable,
	// or they can omit the executable, in which case you must specify an ENTRYPOINT instruction as well.
	Cmd []string `json:"cmd"`

	// WorkingDir sets the working directory for any RUN, CMD, ENTRYPOINT, COPY and ADD instructions
	// that follow it in the Dockerfile.
	WorkingDir string `json:"working_dir"`

	// Env is a list of environment variables to set in the container.
	// Each entry should be in the format "KEY=VALUE".
	Env []string `json:"env"`

	// Labels contains metadata for the image as key-value pairs.
	// These are typically used for organization, licensing, and other descriptive information.
	Labels map[string]string `json:"labels"`
}

// AuthConfig contains authentication credentials for accessing private registries.
// All fields are optional - if no authentication is provided, the system will
// attempt to use default credentials from the Docker credential store.
type AuthConfig struct {
	// Username for registry authentication.
	Username string `json:"username"`

	// Password for registry authentication.
	Password string `json:"password"`

	// Registry URL. If empty, authentication applies to Docker Hub.
	Registry string `json:"registry"`
}

// ProgressCallback is called during export operations to report progress.
// Parameters: current step, total steps, description of current operation
type ProgressCallback func(current, total int, description string)

// ExportOptions contains options for filesystem export operations
type ExportOptions struct {
	// Compress enables gzip compression of the output tar (creates .tar.gz)
	Compress bool

	// Progress callback for reporting export progress
	Progress ProgressCallback
}

// ImageExporter defines the interface for extracting Docker image data.
// Implementations of this interface can retrieve image configurations and
// export complete filesystems without requiring a Docker daemon.
type ImageExporter interface {
	// GetImageConfig retrieves the configuration of a Docker image from a registry.
	// The imageRef should be in standard Docker format (e.g., "nginx:latest" or "registry.com/image:tag").
	// Returns the image configuration or an error if the image cannot be found or accessed.
	GetImageConfig(imageRef string, auth *AuthConfig) (*ImageConfig, error)

	// ExportImageFilesystem exports the complete filesystem of a Docker image to a tar file.
	// The resulting tar file is equivalent to what 'docker export' would produce.
	// The outputPath specifies where to write the tar file.
	ExportImageFilesystem(imageRef string, outputPath string, auth *AuthConfig) error

	// ExportImageFilesystemToWriter exports the complete filesystem of a Docker image to an io.Writer.
	// This allows streaming the tar data directly without creating intermediate files.
	// The writer receives the tar data as it's being generated.
	ExportImageFilesystemToWriter(imageRef string, writer io.Writer, auth *AuthConfig) error

	// ExportImageFilesystemWithOptions exports with additional options like compression and progress
	ExportImageFilesystemWithOptions(imageRef string, outputPath string, auth *AuthConfig, opts *ExportOptions) error

	// ExportImageFilesystemToWriterWithOptions exports to writer with additional options
	ExportImageFilesystemToWriterWithOptions(imageRef string, writer io.Writer, auth *AuthConfig, opts *ExportOptions) error
}
