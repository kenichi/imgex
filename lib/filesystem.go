package lib

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// ExportImageFilesystem exports the complete filesystem of a Docker image to a tar file.
//
// This method downloads all layers of the specified image and reconstructs the complete
// filesystem, writing it as a tar archive to the specified output path. The resulting
// tar file contains the flattened filesystem equivalent to what 'docker export' produces.
//
// Parameters:
//   - imageRef: Docker image reference (e.g., "nginx:latest", "registry.com/org/image:v1.0")
//   - outputPath: Local filesystem path where the tar file should be written
//   - auth: Optional authentication configuration for private registries
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	exporter := NewImageExporter()
//	err := exporter.ExportImageFilesystem("alpine:latest", "/tmp/alpine.tar", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (e *imageExporter) ExportImageFilesystem(imageRef string, outputPath string, auth *AuthConfig) error {
	// Create the output file with proper permissions
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer func() {
		// Ensure file is closed even if export fails
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close output file: %w", closeErr)
		}
	}()

	// Delegate to the writer-based implementation for consistency
	return e.ExportImageFilesystemToWriter(imageRef, file, auth)
}

// ExportImageFilesystemToWriter exports the complete filesystem of a Docker image to an io.Writer.
//
// This method downloads all image layers, applies them in order to reconstruct the complete
// flattened filesystem, and writes the result as a tar archive. The output is equivalent to
// what 'docker export' produces - a single tar containing the final filesystem state with
// all layers applied and merged.
//
// The process involves:
// 1. Fetching all image layers from the registry
// 2. Extracting and applying each layer in sequence
// 3. Building a final filesystem state with proper whiteout handling
// 4. Writing the flattened result as a tar archive
//
// Parameters:
//   - imageRef: Docker image reference (e.g., "nginx:latest", "registry.com/org/image:v1.0")
//   - writer: Destination for the tar data stream
//   - auth: Optional authentication configuration for private registries
//
// Returns:
//   - error: Any error encountered during the operation
//
// Example:
//
//	exporter := NewImageExporter()
//	var buf bytes.Buffer
//	err := exporter.ExportImageFilesystemToWriter("alpine:latest", &buf, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// buf now contains the complete flattened filesystem as tar data
func (e *imageExporter) ExportImageFilesystemToWriter(imageRef string, writer io.Writer, auth *AuthConfig) error {
	// Parse and validate the image reference
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
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

	// Fetch the complete image from the registry
	// This downloads all layers and metadata needed for filesystem reconstruction
	image, err := remote.Image(ref, authOption)
	if err != nil {
		return fmt.Errorf("failed to fetch image %s: %w", imageRef, err)
	}

	// Get the ordered list of layers from the image
	layers, err := image.Layers()
	if err != nil {
		return fmt.Errorf("failed to get image layers: %w", err)
	}

	// Apply all layers to build the final filesystem state
	// This creates a map representing the flattened filesystem
	filesystem, err := e.applyLayers(layers)
	if err != nil {
		return fmt.Errorf("failed to apply layers: %w", err)
	}

	// Write the flattened filesystem as a tar archive
	err = e.writeFilesystemTar(filesystem, writer)
	if err != nil {
		return fmt.Errorf("failed to write filesystem tar: %w", err)
	}

	return nil
}

// fileEntry represents a single file or directory in the flattened filesystem
type fileEntry struct {
	header *tar.Header // tar header with metadata (name, mode, size, etc.)
	data   []byte      // file content data (empty for directories)
}

// applyLayers processes all image layers in order and builds the final filesystem state.
// It handles Docker layer application rules including whiteout files for deletions.
func (e *imageExporter) applyLayers(layers []v1.Layer) (map[string]*fileEntry, error) {
	filesystem := make(map[string]*fileEntry)

	for i, layer := range layers {
		// Get the layer content as a tar stream
		layerReader, err := layer.Uncompressed()
		if err != nil {
			return nil, fmt.Errorf("failed to get layer %d content: %w", i, err)
		}
		defer layerReader.Close()

		// Process the layer tar stream
		tarReader := tar.NewReader(layerReader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to read layer %d tar: %w", i, err)
			}

			// Handle whiteout files (Docker layer deletion mechanism)
			if e.isWhiteoutFile(header.Name) {
				e.handleWhiteout(filesystem, header.Name)
				continue
			}

			// Read file data for regular files
			var data []byte
			if header.Typeflag == tar.TypeReg {
				data = make([]byte, header.Size)
				_, err = io.ReadFull(tarReader, data)
				if err != nil {
					return nil, fmt.Errorf("failed to read file data: %w", err)
				}
			}

			// Clean the path and add to filesystem
			cleanPath := e.cleanPath(header.Name)
			filesystem[cleanPath] = &fileEntry{
				header: header,
				data:   data,
			}
		}
	}

	return filesystem, nil
}

// writeFilesystemTar writes the flattened filesystem map as a tar archive
func (e *imageExporter) writeFilesystemTar(filesystem map[string]*fileEntry, writer io.Writer) error {
	tarWriter := tar.NewWriter(writer)
	defer tarWriter.Close()

	// Write each file/directory in the filesystem
	for _, entry := range filesystem {
		// Update header timestamps for consistency and format compatibility
		entry.header.ModTime = time.Unix(0, 0)
		// Clear unsupported fields for USTAR format
		entry.header.AccessTime = time.Time{}
		entry.header.ChangeTime = time.Time{}

		// Write the header
		err := tarWriter.WriteHeader(entry.header)
		if err != nil {
			return fmt.Errorf("failed to write header for %s: %w", entry.header.Name, err)
		}

		// Write file data for regular files
		if entry.header.Typeflag == tar.TypeReg && len(entry.data) > 0 {
			_, err = tarWriter.Write(entry.data)
			if err != nil {
				return fmt.Errorf("failed to write data for %s: %w", entry.header.Name, err)
			}
		}
	}

	return nil
}

// isWhiteoutFile checks if a file is a Docker whiteout file used for deletions
func (e *imageExporter) isWhiteoutFile(filename string) bool {
	base := path.Base(filename)
	return strings.HasPrefix(base, ".wh.")
}

// handleWhiteout processes a whiteout file by removing the target from the filesystem
func (e *imageExporter) handleWhiteout(filesystem map[string]*fileEntry, whiteoutPath string) {
	dir := path.Dir(whiteoutPath)
	base := path.Base(whiteoutPath)

	if base == ".wh..wh..opq" {
		// Opaque whiteout - remove all files in this directory
		prefix := dir + "/"
		if dir == "." {
			prefix = ""
		}

		for filePath := range filesystem {
			if strings.HasPrefix(filePath, prefix) {
				delete(filesystem, filePath)
			}
		}
	} else if strings.HasPrefix(base, ".wh.") {
		// Regular whiteout - remove the specific file/directory
		target := path.Join(dir, strings.TrimPrefix(base, ".wh."))
		target = e.cleanPath(target)

		// Remove the target file and any files under it (if it's a directory)
		delete(filesystem, target)
		prefix := target + "/"
		for filePath := range filesystem {
			if strings.HasPrefix(filePath, prefix) {
				delete(filesystem, filePath)
			}
		}
	}
}

// cleanPath normalizes a file path for consistent handling
func (e *imageExporter) cleanPath(filePath string) string {
	// Remove leading slash to make paths relative
	cleaned := strings.TrimPrefix(filePath, "/")

	// Handle root directory case
	if cleaned == "" {
		return "."
	}

	return cleaned
}
