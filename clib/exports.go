package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"sync"
	"unsafe"

	"github.com/ken/imgex/lib"
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
