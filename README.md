# imgex - Docker Image Export Tool

A Go library and CLI tool for extracting Docker image configurations and filesystems directly from registries **without requiring a running Docker daemon**.

## Features

- **No Docker daemon required** - Works on any system with network access
- **Cross-platform** - Works on Windows, macOS, Linux  
- **Registry support** - Works with Docker Hub, private registries, authentication
- **Multiple interfaces** - Go library, CLI tool, and C library for language bindings
- **TDD-developed** - Comprehensive test suite with integration tests

## Installation

### CLI Tool

```bash
go build -o imgex ./cmd/imgex
```

### Go Library

```bash
go get github.com/ken/imgex/lib
```

### C Library

```bash
# Shared library
go build -buildmode=c-shared -o libimgex.so ./clib

# Static library  
go build -buildmode=c-archive -o libimgex.a ./clib
```

## Usage

### CLI Tool

```bash
# Get image configuration
./imgex config nginx:latest

# Export filesystem to stdout
./imgex filesystem alpine:latest > alpine.tar

# Export filesystem to file
./imgex filesystem --output nginx.tar nginx:alpine

# With authentication
./imgex --username user --password pass config private-registry.com/image:tag
```

### Go Library

```go
package main

import (
    "fmt"
    "github.com/ken/imgex/lib"
)

func main() {
    exporter := lib.NewImageExporter()
    
    // Get image config
    config, err := exporter.GetImageConfig("nginx:latest", nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Entrypoint: %v\n", config.Entrypoint)
    fmt.Printf("Cmd: %v\n", config.Cmd)
    
    // Export filesystem
    err = exporter.ExportImageFilesystem("alpine:latest", "/tmp/alpine.tar", nil)
    if err != nil {
        panic(err)
    }
}
```

### C Library

```c
#include "libimgex.h"
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
```

## Testing

```bash
# Unit tests
go test ./lib

# Integration tests (requires network)
go test ./lib -v -run TestIntegration

# All tests
go test ./lib -v
```

## Architecture

- `lib/` - Core Go library with image export functionality
- `cmd/imgex/` - CLI interface using Cobra framework  
- `clib/` - C-compatible exports for language bindings
- Comprehensive test suite including unit and integration tests
- Uses `google/go-containerregistry` for efficient registry communication

## Elixir NIF Integration

The C library can be used with Elixir NIFs for seamless integration:

```elixir
defmodule DockerImageExport do
  @on_load :load_nifs

  def load_nifs do
    :erlang.load_nif('./priv/libimgex', 0)
  end

  def get_image_config(image_ref, auth \\ nil) do
    auth_json = if auth, do: Jason.encode!(auth), else: ""
    case get_image_config_nif(image_ref, auth_json) do
      {:ok, json} -> Jason.decode(json)
      error -> error
    end
  end

  # NIF implementations...
end
```

## Benefits vs Docker Export

1. **No Docker daemon required** - Works without Docker installation
2. **Faster** - Direct registry communication, no container creation
3. **Cross-platform** - Works on any OS with network access  
4. **Integration-friendly** - Library interface for embedding
5. **Language interop** - C bindings for use in other languages

## License

MIT License