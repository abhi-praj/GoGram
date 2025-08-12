package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigSingleton(t *testing.T) {
	instance = nil
	
	config1 := GetInstance()
	if config1 == nil {
		t.Fatal("GetInstance() returned nil")
	}
	
	config2 := GetInstance()
	if config2 == nil {
		t.Fatal("GetInstance() returned nil on second call")
	}
	
	if config1 != config2 {
		t.Fatal("Singleton pattern failed - different instances returned")
	}
}

func TestConfigDefaults(t *testing.T) {
	instance = nil
	
	tempDir := t.TempDir()
	
	originalUserHomeDir := os.UserHomeDir
	os.UserHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { os.UserHomeDir = originalUserHomeDir }()
	
	config := GetInstance()
	
	if config.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", config.Language)
	}
	
	if !config.Chat.Colors {
		t.Error("Expected chat colors to be true")
	}
	
	if config.Chat.Layout != "compact" {
		t.Errorf("Expected chat layout 'compact', got '%s'", config.Chat.Layout)
	}
	
	if config.Advanced.DebugMode {
		t.Error("Expected debug mode to be false")
	}
	
	if config.Advanced.GeorgistCredits != 627 {
		t.Errorf("Expected georgist credits 627, got %d", config.Advanced.GeorgistCredits)
	}
}

func TestConfigGetSet(t *testing.T) {
	instance = nil
	
	tempDir := t.TempDir()
	
	originalUserHomeDir := os.UserHomeDir
	os.UserHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { os.UserHomeDir = originalUserHomeDir }()
	
	config := GetInstance()
	
	config.Set("test.key", "test_value")
	
	value := config.Get("test.key", "default")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%v'", value)
	}
	
	config.Set("nested.deep.key", 42)
	
	nestedValue := config.Get("nested.deep.key", 0)
	if nestedValue != 42 {
		t.Errorf("Expected 42, got %v", nestedValue)
	}
	
	defaultValue := config.Get("nonexistent.key", "default_value")
	if defaultValue != "default_value" {
		t.Errorf("Expected 'default_value', got '%v'", defaultValue)
	}
}

func TestConfigGetString(t *testing.T) {
	instance = nil
	
	tempDir := t.TempDir()
	
	originalUserHomeDir := os.UserHomeDir
	os.UserHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { os.UserHomeDir = originalUserHomeDir }()
	
	config := GetInstance()
	
	config.Set("string.test", "hello")
	
	value := config.GetString("string.test", "default")
	if value != "hello" {
		t.Errorf("Expected 'hello', got '%s'", value)
	}
	
	defaultValue := config.GetString("nonexistent", "default_string")
	if defaultValue != "default_string" {
		t.Errorf("Expected 'default_string', got '%s'", defaultValue)
	}
}

func TestConfigGetBool(t *testing.T) {
	instance = nil
	
	tempDir := t.TempDir()
	
	originalUserHomeDir := os.UserHomeDir
	os.UserHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { os.UserHomeDir = originalUserHomeDir }()
	
	config := GetInstance()
	
	config.Set("bool.test", true)
	
	value := config.GetBool("bool.test", false)
	if !value {
		t.Error("Expected true, got false")
	}
	
	defaultValue := config.GetBool("nonexistent", true)
	if !defaultValue {
		t.Error("Expected true, got false")
	}
}

func TestConfigGetInt(t *testing.T) {
	instance = nil
	
	tempDir := t.TempDir()
	
	originalUserHomeDir := os.UserHomeDir
	os.UserHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { os.UserHomeDir = originalUserHomeDir }()
	
	config := GetInstance()
	
	config.Set("int.test", 123)
	
	value := config.GetInt("int.test", 0)
	if value != 123 {
		t.Errorf("Expected 123, got %d", value)
	}
	
	defaultValue := config.GetInt("nonexistent", 999)
	if defaultValue != 999 {
		t.Errorf("Expected 999, got %d", defaultValue)
	}
}

func TestConfigReset(t *testing.T) {
	instance = nil
	
	tempDir := t.TempDir()
	
	originalUserHomeDir := os.UserHomeDir
	os.UserHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { os.UserHomeDir = originalUserHomeDir }()
	
	config := GetInstance()
	
	config.Set("custom.key", "custom_value")
	config.Set("custom.number", 456)
	
	if config.Get("custom.key", "") != "custom_value" {
		t.Error("Custom value not set correctly")
	}
	
	config.Reset()
	
	if config.Get("custom.key", "") != "" {
		t.Error("Custom value not reset")
	}
	
	if config.Language != "en" {
		t.Error("Default language not restored")
	}
}

func TestConfigFileCreation(t *testing.T) {
	instance = nil
	
	tempDir := t.TempDir()
	
	originalUserHomeDir := os.UserHomeDir
	os.UserHomeDir = func() (string, error) {
		return tempDir, nil
	}
	defer func() { os.UserHomeDir = originalUserHomeDir }()
	
	config := GetInstance()
	
	configFile := config.GetConfigFile()
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}
	
	configDir := config.GetConfigDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Fatal("Config directory was not created")
	}
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	
	if len(data) == 0 {
		t.Fatal("Config file is empty")
	}
}
