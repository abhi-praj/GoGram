package config

import (
	"strings"
	"testing"
)

func TestGetInstance(t *testing.T) {
	// Clear any existing instance
	instance = nil

	// Test singleton pattern
	cfg1 := GetInstance()
	cfg2 := GetInstance()

	if cfg1 != cfg2 {
		t.Error("GetInstance() should return the same instance")
	}

	if cfg1 == nil {
		t.Error("GetInstance() returned nil")
	}
}

func TestGetConfigDir(t *testing.T) {
	cfg := GetInstance()
	configDir := cfg.GetConfigDir()

	if configDir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	// Should contain .instagram-cli
	if !strings.HasSuffix(configDir, ".instagram-cli") {
		t.Errorf("Expected config dir to end with .instagram-cli, got: %s", configDir)
	}
}

func TestGetConfigFile(t *testing.T) {
	cfg := GetInstance()
	configFile := cfg.GetConfigFile()

	if configFile == "" {
		t.Error("GetConfigFile() returned empty string")
	}

	// Should end with config.yaml
	if !strings.HasSuffix(configFile, "config.yaml") {
		t.Errorf("Expected config file to end with config.yaml, got: %s", configFile)
	}
}

func TestGetDefaultValue(t *testing.T) {
	cfg := GetInstance()

	// Test getting a default value
	value := cfg.Get("language", "default")
	if value != "en" {
		t.Errorf("Expected language to be 'en', got: %v", value)
	}
}

func TestSetAndGet(t *testing.T) {
	cfg := GetInstance()

	// Test setting and getting a value
	testKey := "test.key"
	testValue := "test_value"

	err := cfg.Set(testKey, testValue)
	if err != nil {
		t.Errorf("Failed to set config: %v", err)
	}

	// Get the value back
	retrievedValue := cfg.Get(testKey, nil)
	if retrievedValue != testValue {
		t.Errorf("Expected %s, got %v", testValue, retrievedValue)
	}

	// Clean up
	cfg.Set(testKey, nil)
}
