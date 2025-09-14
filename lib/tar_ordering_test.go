package lib

import (
	"archive/tar"
	"bytes"
	"testing"
	"time"
)

func TestTarOrderingForExtraction(t *testing.T) {
	// Create a mock filesystem with ordering challenges:
	// - A symlink that points to a file
	// - Directories that need to exist before files in them
	// - Files that need to exist before links to them
	filesystem := map[string]*fileEntry{
		// A symlink to a file (should come AFTER the target file)
		"link_to_file": {
			header: &tar.Header{
				Name:     "link_to_file",
				Typeflag: tar.TypeSymlink,
				Linkname: "target_file",
				Mode:     0644,
				ModTime:  time.Unix(0, 0),
			},
			data: nil,
		},
		// The target file (should come BEFORE the symlink)
		"target_file": {
			header: &tar.Header{
				Name:     "target_file",
				Typeflag: tar.TypeReg,
				Size:     12, // "file content" is 12 bytes
				Mode:     0644,
				ModTime:  time.Unix(0, 0),
			},
			data: []byte("file content"),
		},
		// A file in a subdirectory (directory should come first)
		"subdir/nested_file": {
			header: &tar.Header{
				Name:     "subdir/nested_file",
				Typeflag: tar.TypeReg,
				Size:     6, // "nested" is 6 bytes
				Mode:     0644,
				ModTime:  time.Unix(0, 0),
			},
			data: []byte("nested"),
		},
		// The subdirectory (should come BEFORE files in it)
		"subdir/": {
			header: &tar.Header{
				Name:     "subdir/",
				Typeflag: tar.TypeDir,
				Mode:     0755,
				ModTime:  time.Unix(0, 0),
			},
			data: nil,
		},
	}

	exporter := &imageExporter{}
	var buf bytes.Buffer

	// Export the filesystem to tar
	err := exporter.writeFilesystemTar(filesystem, &buf)
	if err != nil {
		t.Fatalf("Failed to write filesystem tar: %v", err)
	}

	// Parse the tar to check ordering
	tarReader := tar.NewReader(&buf)
	var order []string
	var types []byte

	for {
		header, err := tarReader.Next()
		if err != nil {
			break // End of tar
		}
		order = append(order, header.Name)
		types = append(types, header.Typeflag)
	}

	// Verify that we have all entries
	if len(order) != 4 {
		t.Fatalf("Expected 4 entries, got %d: %v", len(order), order)
	}

	// Check ordering constraints for proper extraction:
	// 1. Directory "subdir/" should come before "subdir/nested_file"
	subdirIdx := indexOf(order, "subdir/")
	nestedFileIdx := indexOf(order, "subdir/nested_file")
	if subdirIdx == -1 || nestedFileIdx == -1 {
		t.Fatalf("Missing directory or nested file in tar")
	}
	if subdirIdx > nestedFileIdx {
		t.Errorf("Directory 'subdir/' (index %d) should come before 'subdir/nested_file' (index %d)",
			subdirIdx, nestedFileIdx)
		t.Logf("Actual order: %v", order)
		t.Logf("Actual types: %v", types)
	}

	// 2. Target file should come before symlink to it
	targetFileIdx := indexOf(order, "target_file")
	symlinkIdx := indexOf(order, "link_to_file")
	if targetFileIdx == -1 || symlinkIdx == -1 {
		t.Fatalf("Missing target file or symlink in tar")
	}
	if targetFileIdx > symlinkIdx {
		t.Errorf("Target file 'target_file' (index %d) should come before symlink 'link_to_file' (index %d)",
			targetFileIdx, symlinkIdx)
		t.Logf("Actual order: %v", order)
		t.Logf("Actual types: %v", types)
	}

	// 3. Generally, directories should come first, then files, then links
	firstLinkIdx := -1
	lastDirIdx := -1
	lastFileIdx := -1

	for i, typ := range types {
		switch typ {
		case tar.TypeDir:
			lastDirIdx = i
		case tar.TypeReg:
			lastFileIdx = i
		case tar.TypeSymlink, tar.TypeLink:
			if firstLinkIdx == -1 {
				firstLinkIdx = i
			}
		}
	}

	if lastDirIdx != -1 && lastFileIdx != -1 && lastDirIdx > lastFileIdx {
		t.Errorf("Directories should generally come before files, but found directory at %d after file at %d",
			lastDirIdx, lastFileIdx)
	}

	if firstLinkIdx != -1 && lastFileIdx != -1 && firstLinkIdx < lastFileIdx {
		t.Errorf("Links should generally come after files, but found link at %d before file at %d",
			firstLinkIdx, lastFileIdx)
	}
}

// Helper function to find index of string in slice
func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}