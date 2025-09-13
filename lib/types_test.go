package lib

import (
	"encoding/json"
	"testing"
)

func TestImageConfig_JSON(t *testing.T) {
	config := ImageConfig{
		User:       "www-data",
		Entrypoint: []string{"/entrypoint.sh"},
		Cmd:        []string{"nginx", "-g", "daemon off;"},
		WorkingDir: "/var/www",
		Env:        []string{"PATH=/usr/local/sbin:/usr/local/bin", "HOME=/root"},
		Labels:     map[string]string{"version": "1.0", "maintainer": "test"},
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal ImageConfig: %v", err)
	}

	var unmarshaled ImageConfig
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ImageConfig: %v", err)
	}

	if unmarshaled.User != config.User {
		t.Errorf("Expected User %s, got %s", config.User, unmarshaled.User)
	}
	if len(unmarshaled.Entrypoint) != len(config.Entrypoint) {
		t.Errorf("Expected Entrypoint length %d, got %d", len(config.Entrypoint), len(unmarshaled.Entrypoint))
	}
	if len(unmarshaled.Cmd) != len(config.Cmd) {
		t.Errorf("Expected Cmd length %d, got %d", len(config.Cmd), len(unmarshaled.Cmd))
	}
}

func TestAuthConfig_JSON(t *testing.T) {
	auth := AuthConfig{
		Username: "testuser",
		Password: "testpass",
		Registry: "registry.example.com",
	}

	jsonData, err := json.Marshal(auth)
	if err != nil {
		t.Fatalf("Failed to marshal AuthConfig: %v", err)
	}

	var unmarshaled AuthConfig
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal AuthConfig: %v", err)
	}

	if unmarshaled.Username != auth.Username {
		t.Errorf("Expected Username %s, got %s", auth.Username, unmarshaled.Username)
	}
	if unmarshaled.Password != auth.Password {
		t.Errorf("Expected Password %s, got %s", auth.Password, unmarshaled.Password)
	}
	if unmarshaled.Registry != auth.Registry {
		t.Errorf("Expected Registry %s, got %s", auth.Registry, unmarshaled.Registry)
	}
}

func TestImageConfig_DefaultValues(t *testing.T) {
	config := ImageConfig{}

	if config.User != "" {
		t.Errorf("Expected empty User, got %s", config.User)
	}
	if config.Entrypoint != nil {
		t.Errorf("Expected nil Entrypoint, got %v", config.Entrypoint)
	}
	if config.Cmd != nil {
		t.Errorf("Expected nil Cmd, got %v", config.Cmd)
	}
	if config.Labels != nil {
		t.Errorf("Expected nil Labels, got %v", config.Labels)
	}
}