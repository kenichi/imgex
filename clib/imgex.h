#ifndef IMGEX_H
#define IMGEX_H

#ifdef __cplusplus
extern "C" {
#endif

// Progress callback function type
// Called during filesystem export to report progress
// Parameters: current step (0-based), total steps, description of current operation
typedef void (*progress_callback_t)(int current, int total, const char* description);

// Get image configuration as JSON string
// Returns: JSON string (must be freed with free_string), or NULL on error
// Parameters:
//   image_ref: Docker image reference (e.g., "nginx:latest")
//   auth_json: Authentication JSON or NULL/empty for default auth
char* get_image_config_json(const char* image_ref, const char* auth_json);

// Export image filesystem to file (basic version)
// Returns: 0 on success, -1 on error
// Parameters:
//   image_ref: Docker image reference
//   output_path: File path to write tar archive
//   auth_json: Authentication JSON or NULL/empty for default auth
int export_image_filesystem_to_file(const char* image_ref, const char* output_path, const char* auth_json);

// Export image filesystem to file with options
// Returns: 0 on success, -1 on error
// Parameters:
//   image_ref: Docker image reference
//   output_path: File path to write archive (.gz extension added if compress=1 and not present)
//   auth_json: Authentication JSON or NULL/empty for default auth
//   compress: 1 to enable gzip compression, 0 to disable
//   progress_callback: Function pointer for progress updates, or NULL to disable
int export_image_filesystem_with_options(const char* image_ref, const char* output_path,
                                        const char* auth_json, int compress,
                                        progress_callback_t progress_callback);

// Get library version string
// Returns: Version string (must be freed with free_string)
char* get_version(void);

// Get library description string
// Returns: Description string (must be freed with free_string)
char* get_description(void);

// Get last error message
// Returns: Error string (must be freed with free_string), or NULL if no error
char* get_last_error(void);

// Free string returned by library functions
// Parameters:
//   str: String to free (returned by other functions)
void free_string(char* str);

#ifdef __cplusplus
}
#endif

#endif // IMGEX_H