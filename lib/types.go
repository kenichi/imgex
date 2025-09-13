package lib

import "io"

type ImageConfig struct {
	User       string            `json:"user"`
	Entrypoint []string          `json:"entrypoint"`
	Cmd        []string          `json:"cmd"`
	WorkingDir string            `json:"working_dir"`
	Env        []string          `json:"env"`
	Labels     map[string]string `json:"labels"`
}

type AuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Registry string `json:"registry"`
}

type ImageExporter interface {
	GetImageConfig(imageRef string, auth *AuthConfig) (*ImageConfig, error)
	ExportImageFilesystem(imageRef string, outputPath string, auth *AuthConfig) error
	ExportImageFilesystemToWriter(imageRef string, writer io.Writer, auth *AuthConfig) error
}