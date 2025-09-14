#include <stdio.h>
#include <stdlib.h>
#include "dist/libimgex.h"

int main() {
    printf("Testing imgex C library...\n");
    
    // Test getting image config
    char* config_json = get_image_config_json("alpine:latest", "");
    if (config_json == NULL) {
        char* error = get_last_error();
        printf("Error getting config: %s\n", error ? error : "unknown error");
        if (error) free_string(error);
        return 1;
    }
    
    printf("Config JSON:\n%s\n", config_json);
    free_string(config_json);
    
    // Test filesystem export
    printf("\nTesting filesystem export...\n");
    int result = export_image_filesystem_to_file("alpine:latest", "/tmp/alpine_from_c.tar", "");
    if (result != 0) {
        char* error = get_last_error();
        printf("Error exporting filesystem: %s\n", error ? error : "unknown error");
        if (error) free_string(error);
        return 1;
    }
    
    printf("Filesystem exported successfully to /tmp/alpine_from_c.tar\n");
    return 0;
}