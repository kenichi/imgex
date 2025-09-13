package lib

import (
	"fmt"
	"io"
	"os"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

func (e *imageExporter) ExportImageFilesystem(imageRef string, outputPath string, auth *AuthConfig) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer file.Close()

	return e.ExportImageFilesystemToWriter(imageRef, file, auth)
}

func (e *imageExporter) ExportImageFilesystemToWriter(imageRef string, writer io.Writer, auth *AuthConfig) error {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
	}

	var authOption remote.Option
	if auth != nil {
		authOption = remote.WithAuth(&authn.Basic{
			Username: auth.Username,
			Password: auth.Password,
		})
	} else {
		authOption = remote.WithAuthFromKeychain(authn.DefaultKeychain)
	}

	image, err := remote.Image(ref, authOption)
	if err != nil {
		return fmt.Errorf("failed to fetch image %s: %w", imageRef, err)
	}

	err = tarball.Write(ref, image, writer)
	if err != nil {
		return fmt.Errorf("failed to write image as tarball: %w", err)
	}

	return nil
}