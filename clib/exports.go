package main

/*
#include <stdlib.h>

// Define the progress callback function pointer type
typedef void (*progress_callback_t)(int current, int total, const char* description);

// Helper function to call the callback from Go
static void call_progress_callback(progress_callback_t callback, int current, int total, const char* description) {
    if (callback != NULL) {
        callback(current, total, description);
    }
}
*/
import "C"
import (
	"encoding/json"
	"strings"
	"sync"
	"unsafe"

	"github.com/kenichi/imgex/lib"
)

var (
	lastError     string
	lastErrorLock sync.RWMutex
)

func setLastError(err error) {
	lastErrorLock.Lock()
	defer lastErrorLock.Unlock()
	if err != nil {
		lastError = err.Error()
	} else {
		lastError = ""
	}
}

func getLastErrorInternal() string {
	lastErrorLock.RLock()
	defer lastErrorLock.RUnlock()
	return lastError
}

//export get_image_config_json
func get_image_config_json(image_ref *C.char, auth_json *C.char) *C.char {
	imageRef := C.GoString(image_ref)
	authJSON := C.GoString(auth_json)

	var auth *lib.AuthConfig
	if authJSON != "" {
		auth = &lib.AuthConfig{}
		if err := json.Unmarshal([]byte(authJSON), auth); err != nil {
			setLastError(err)
			return nil
		}
	}

	exporter := lib.NewImageExporter()
	config, err := exporter.GetImageConfig(imageRef, auth)
	if err != nil {
		setLastError(err)
		return nil
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		setLastError(err)
		return nil
	}

	setLastError(nil)
	return C.CString(string(configJSON))
}

//export export_image_filesystem_to_file
func export_image_filesystem_to_file(image_ref *C.char, output_path *C.char, auth_json *C.char) C.int {
	imageRef := C.GoString(image_ref)
	outputPath := C.GoString(output_path)
	authJSON := C.GoString(auth_json)

	var auth *lib.AuthConfig
	if authJSON != "" {
		auth = &lib.AuthConfig{}
		if err := json.Unmarshal([]byte(authJSON), auth); err != nil {
			setLastError(err)
			return -1
		}
	}

	exporter := lib.NewImageExporter()
	err := exporter.ExportImageFilesystem(imageRef, outputPath, auth)
	if err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export free_string
func free_string(str *C.char) {
	C.free(unsafe.Pointer(str))
}

//export export_image_filesystem_with_options
func export_image_filesystem_with_options(image_ref *C.char, output_path *C.char, auth_json *C.char, compress C.int, progress_callback unsafe.Pointer) C.int {
	imageRef := C.GoString(image_ref)
	outputPath := C.GoString(output_path)
	authJSON := C.GoString(auth_json)

	var auth *lib.AuthConfig
	if authJSON != "" {
		auth = &lib.AuthConfig{}
		if err := json.Unmarshal([]byte(authJSON), auth); err != nil {
			setLastError(err)
			return -1
		}
	}

	// Set up export options
	opts := &lib.ExportOptions{
		Compress: compress != 0,
	}

	// Append .gz extension if compression is enabled and not already present
	if compress != 0 && !strings.HasSuffix(outputPath, ".gz") {
		outputPath += ".gz"
	}

	// Set up progress callback if provided
	if progress_callback != nil {
		opts.Progress = func(current, total int, description string) {
			// Convert to C types and call the callback
			cDescription := C.CString(description)
			defer C.free(unsafe.Pointer(cDescription))

			C.call_progress_callback(
				C.progress_callback_t(progress_callback),
				C.int(current),
				C.int(total),
				cDescription,
			)
		}
	}

	exporter := lib.NewImageExporter()
	err := exporter.ExportImageFilesystemWithOptions(imageRef, outputPath, auth, opts)
	if err != nil {
		setLastError(err)
		return -1
	}

	setLastError(nil)
	return 0
}

//export get_version
func get_version() *C.char {
	return C.CString(lib.Version)
}

//export get_description
func get_description() *C.char {
	return C.CString(lib.Description)
}

//export get_last_error
func get_last_error() *C.char {
	errMsg := getLastErrorInternal()
	if errMsg == "" {
		return nil
	}
	return C.CString(errMsg)
}


// Required for CGO
func main() {}
