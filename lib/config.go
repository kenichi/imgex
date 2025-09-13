package lib

import (
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type imageExporter struct{}

func NewImageExporter() ImageExporter {
	return &imageExporter{}
}

func (e *imageExporter) GetImageConfig(imageRef string, auth *AuthConfig) (*ImageConfig, error) {
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("failed to parse image reference %s: %w", imageRef, err)
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
		return nil, fmt.Errorf("failed to fetch image %s: %w", imageRef, err)
	}

	configFile, err := image.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get config file: %w", err)
	}

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

