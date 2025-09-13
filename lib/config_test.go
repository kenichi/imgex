package lib

import (
	"strings"
	"testing"
)

func TestGetImageConfig_ValidImage(t *testing.T) {
	exporter := NewImageExporter()

	config, err := exporter.GetImageConfig("nginx:alpine", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to not be nil")
	}
}

func TestGetImageConfig_InvalidImage(t *testing.T) {
	exporter := NewImageExporter()

	_, err := exporter.GetImageConfig("invalid-image-name", nil)
	if err == nil {
		t.Fatal("Expected error for invalid image name")
	}
}

func TestGetImageConfig_WithAuth(t *testing.T) {
	exporter := NewImageExporter()
	auth := &AuthConfig{
		Username: "testuser",
		Password: "testpass",
		Registry: "registry.example.com",
	}

	_, err := exporter.GetImageConfig("registry.example.com/private/image:latest", auth)

	if err != nil && !strings.Contains(err.Error(), "authentication") &&
		!strings.Contains(err.Error(), "unauthorized") &&
		!strings.Contains(err.Error(), "not found") &&
		!strings.Contains(err.Error(), "no such host") &&
		!strings.Contains(err.Error(), "dial tcp") {
		t.Errorf("Expected authentication-related error, not found, or network error, got %v", err)
	}
}

func TestGetImageConfig_PublicRegistry(t *testing.T) {
	exporter := NewImageExporter()

	config, err := exporter.GetImageConfig("alpine:latest", nil)
	if err != nil {
		t.Fatalf("Expected no error for public image, got %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to not be nil")
	}

	if config.User == "" && len(config.Env) == 0 {
		t.Log("Config appears minimal for alpine image (expected)")
	}
}
