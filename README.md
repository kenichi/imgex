# imgex - Docker Image Export Tool

[![CI](https://github.com/kenichi/imgex/workflows/CI/badge.svg)](https://github.com/kenichi/imgex/actions)
[![License](https://img.shields.io/github/license/kenichi/imgex)](LICENSE)
[![Release](https://img.shields.io/github/v/release/kenichi/imgex)](https://github.com/kenichi/imgex/releases)

A Go library and CLI tool for extracting Docker image configurations and filesystems directly from registries without requiring a running Docker daemon or Linux kernel.

## 🌟 Features

- **🚫 No Docker daemon required** - Works on any system with network access
- **🌍 Multi-platform support** - Linux, macOS, Windows, FreeBSD, illumos/Solaris
- **🔐 Registry support** - Docker Hub, private registries, authentication
- **📦 Multiple interfaces** - Go library, CLI tool, and C library for language bindings

## 🚀 Installation

<!-- ### Pre-built Binaries (Recommended) -->

<!-- Download pre-built binaries from the [releases page](https://github.com/kenichi/imgex/releases) for: -->
<!-- - **Linux**: x86_64, ARM64 -->
<!-- - **macOS**: Intel, Apple Silicon (M1/M2) -->
<!-- - **Windows**: x86_64, ARM64 -->
<!-- - **FreeBSD**: x86_64 -->
<!-- - **illumos/Solaris**: x86_64 -->

### Build from Source

Requirements: Go 1.21 or later

```bash
# Clone and build
git clone https://github.com/kenichi/imgex.git
cd imgex
make build
# Binary will be in dist/imgex

# Or install directly
go install github.com/kenichi/imgex/cmd/imgex@latest
```

### C Library for Native Integration

```bash
# Build C libraries (both static and shared)
make clib
# Libraries will be in dist/libimgex.{so,a,h}
```

## Usage

### CLI Tool

```bash
# Get image configuration
./dist/imgex config nginx:latest

# Export filesystem to stdout
./dist/imgex filesystem alpine:latest > alpine.tar

# Export filesystem to file
./dist/imgex filesystem --output nginx.tar nginx:alpine

# With authentication
./dist/imgex --username user --password pass config private-registry.com/image:tag
```

### C Library

```c
#include "dist/libimgex.h"
#include <stdio.h>

int main() {
    // Get config as JSON
    char* config = get_image_config_json("alpine:latest", "");
    printf("Config: %s\n", config);
    free_string(config);

    // Export filesystem
    int result = export_image_filesystem_to_file("alpine:latest", "/tmp/alpine.tar", "");
    if (result != 0) {
        char* error = get_last_error();
        printf("Error: %s\n", error);
        free_string(error);
    }

    return 0;
}

// Compile with:
// gcc -o myapp myapp.c -Ldist -limgex -lpthread -ldl
// LD_LIBRARY_PATH=dist ./myapp
```
