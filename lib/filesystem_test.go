package lib

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportImageFilesystemToWriter(t *testing.T) {
	exporter := NewImageExporter()
	
	var buf bytes.Buffer
	err := exporter.ExportImageFilesystemToWriter("alpine:latest", &buf, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if buf.Len() == 0 {
		t.Fatal("Expected filesystem data, got empty buffer")
	}
	
	data := buf.Bytes()
	if !bytes.Contains(data, []byte("bin/")) && !bytes.Contains(data, []byte("etc/")) {
		t.Error("Expected filesystem to contain common directories like bin/ or etc/")
	}
}

func TestExportImageFilesystem(t *testing.T) {
	exporter := NewImageExporter()
	
	tmpDir, err := os.MkdirTemp("", "imgex-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	outputPath := filepath.Join(tmpDir, "alpine.tar")
	err = exporter.ExportImageFilesystem("alpine:latest", outputPath, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Output file doesn't exist: %v", err)
	}
	
	if fileInfo.Size() == 0 {
		t.Fatal("Expected non-empty tar file")
	}
	
	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()
	
	header := make([]byte, 262)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		t.Fatalf("Failed to read header: %v", err)
	}
	
	if n == 0 {
		t.Fatal("Empty file")
	}
}

func TestExportImageFilesystem_InvalidImage(t *testing.T) {
	exporter := NewImageExporter()
	
	tmpDir, err := os.MkdirTemp("", "imgex-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	outputPath := filepath.Join(tmpDir, "invalid.tar")
	err = exporter.ExportImageFilesystem("invalid-image-name", outputPath, nil)
	if err == nil {
		t.Fatal("Expected error for invalid image name")
	}
}

func TestExportImageFilesystemToWriter_InvalidImage(t *testing.T) {
	exporter := NewImageExporter()
	
	var buf bytes.Buffer
	err := exporter.ExportImageFilesystemToWriter("invalid/image/name/with/too/many/slashes", &buf, nil)
	if err == nil {
		t.Fatal("Expected error for invalid image name")
	}
	
	if !strings.Contains(err.Error(), "failed to parse image reference") &&
		!strings.Contains(err.Error(), "failed to fetch image") {
		t.Errorf("Expected parse or fetch error, got %v", err)
	}
}