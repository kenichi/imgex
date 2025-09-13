package lib

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestIntegration_PublicImages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testCases := []struct {
		name     string
		imageRef string
	}{
		{"alpine", "alpine:latest"},
		{"nginx", "nginx:alpine"},
		{"busybox", "busybox:latest"},
		{"hello-world", "hello-world:latest"},
	}

	exporter := NewImageExporter()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing with image: %s", tc.imageRef)

			config, err := exporter.GetImageConfig(tc.imageRef, nil)
			if err != nil {
				t.Fatalf("Failed to get config for %s: %v", tc.imageRef, err)
			}

			if config == nil {
				t.Fatalf("Config is nil for %s", tc.imageRef)
			}

			t.Logf("Config for %s: User=%s, Cmd=%v, Entrypoint=%v", 
				tc.imageRef, config.User, config.Cmd, config.Entrypoint)

			var buf bytes.Buffer
			err = exporter.ExportImageFilesystemToWriter(tc.imageRef, &buf, nil)
			if err != nil {
				t.Fatalf("Failed to export filesystem for %s: %v", tc.imageRef, err)
			}

			if buf.Len() == 0 {
				t.Fatalf("Empty filesystem export for %s", tc.imageRef)
			}

			t.Logf("Exported filesystem for %s: %d bytes", tc.imageRef, buf.Len())

			tmpDir, err := os.MkdirTemp("", "imgex-integration-")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			outputPath := filepath.Join(tmpDir, tc.name+".tar")
			err = exporter.ExportImageFilesystem(tc.imageRef, outputPath, nil)
			if err != nil {
				t.Fatalf("Failed to export to file for %s: %v", tc.imageRef, err)
			}

			fileInfo, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("Output file doesn't exist for %s: %v", tc.imageRef, err)
			}

			if fileInfo.Size() == 0 {
				t.Fatalf("Empty output file for %s", tc.imageRef)
			}

			t.Logf("File export for %s: %d bytes", tc.imageRef, fileInfo.Size())
		})
	}
}

func TestIntegration_ConfigValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	testCases := []struct {
		name           string
		imageRef       string
		expectedCmd    []string
		expectedEnv    bool
		expectedUser   string
	}{
		{
			name:        "alpine",
			imageRef:    "alpine:latest",
			expectedCmd: []string{"/bin/sh"},
			expectedEnv: true,
			expectedUser: "",
		},
		{
			name:        "nginx-alpine", 
			imageRef:    "nginx:alpine",
			expectedCmd: []string{"nginx", "-g", "daemon off;"},
			expectedEnv: true,
			expectedUser: "",
		},
	}

	exporter := NewImageExporter()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := exporter.GetImageConfig(tc.imageRef, nil)
			if err != nil {
				t.Fatalf("Failed to get config: %v", err)
			}

			if len(config.Cmd) != len(tc.expectedCmd) {
				t.Errorf("Expected Cmd %v, got %v", tc.expectedCmd, config.Cmd)
			} else {
				for i, expected := range tc.expectedCmd {
					if config.Cmd[i] != expected {
						t.Errorf("Expected Cmd[%d] = %s, got %s", i, expected, config.Cmd[i])
					}
				}
			}

			if tc.expectedEnv && len(config.Env) == 0 {
				t.Errorf("Expected environment variables, got none")
			}

			if config.User != tc.expectedUser {
				t.Errorf("Expected User %s, got %s", tc.expectedUser, config.User)
			}
		})
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	exporter := NewImageExporter()

	t.Run("nonexistent-image", func(t *testing.T) {
		_, err := exporter.GetImageConfig("nonexistent/image:tag", nil)
		if err == nil {
			t.Fatal("Expected error for nonexistent image")
		}
		t.Logf("Got expected error: %v", err)
	})

	t.Run("invalid-reference", func(t *testing.T) {
		_, err := exporter.GetImageConfig("invalid/image/name/with/too/many/parts", nil)
		if err == nil {
			t.Fatal("Expected error for invalid reference")
		}
		t.Logf("Got expected error: %v", err)
	})
}