package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// our default config values
var DefaultConfig = map[string]interface{}{
	"language": "en",
	"login": map[string]interface{}{
		"default_username": nil,
		"current_username": nil,
	},
	"chat": map[string]interface{}{
		"layout": "compact",
		"colors": true,
	},
	"scheduling": map[string]interface{}{
		"default_schedule_duration": "01:00",
	},
	"privacy": map[string]interface{}{
		"invisible_mode": false,
	},
	"advanced": map[string]interface{}{
		"debug_mode":       false,
		"georgist_credits": 627,
	},
}

// Config represents the configuration manager (shoutout 207)
type Config struct {
	configDir  string
	configFile string
	viper      *viper.Viper
}

var instance *Config

// GetInstance returns the singleton instance of config
func GetInstance() *Config {
	if instance == nil {
		instance = &Config{}
		instance.initialize()
	}
	return instance
}

// initialize sets up the config
func (c *Config) initialize() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Failed to get user home directory: %v", err))
	}

	c.configDir = filepath.Join(homeDir, ".instagram-cli")
	c.configFile = filepath.Join(c.configDir, "config.yaml")

	// setting our defaults for derived paths
	if advanced, ok := DefaultConfig["advanced"].(map[string]interface{}); ok {
		advanced["data_dir"] = c.configDir
		advanced["users_dir"] = filepath.Join(c.configDir, "users")
		advanced["cache_dir"] = filepath.Join(c.configDir, "cache")
		advanced["media_dir"] = filepath.Join(c.configDir, "media")
		advanced["generated_dir"] = filepath.Join(c.configDir, "generated")
	}

	c.loadConfig()
}

// loadConfig loads configuration from file or creates default if not exists
func (c *Config) loadConfig() {
	if err := os.MkdirAll(c.configDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create config directory: %v", err))
	}

	c.viper = viper.New()
	c.viper.SetConfigFile(c.configFile)
	c.viper.SetConfigType("yaml")

	for key, value := range DefaultConfig {
		c.viper.SetDefault(key, value)
	}

	if err := c.viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			c.saveConfig(DefaultConfig)
		} else {
			fmt.Printf("Warning: Error reading config file: %v\n", err)
		}
	}
}

// saveConfig saves configuration to file
func (c *Config) saveConfig(config map[string]interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(c.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return c.viper.ReadInConfig()
}

// Get retrieves a config value by key
func (c *Config) Get(key string, defaultValue interface{}) interface{} {
	value := c.viper.Get(key)
	if value != nil {
		return value
	}

	// Try to get from default config
	keys := strings.Split(key, ".")
	var current interface{} = DefaultConfig

	for _, k := range keys {
		if currentMap, ok := current.(map[string]interface{}); ok {
			if val, exists := currentMap[k]; exists {
				current = val
			} else {
				fmt.Printf("Warning: Config key '%s' not found in config.yaml file, using default value: %v\n", key, defaultValue)
				return defaultValue
			}
		} else {
			fmt.Printf("Warning: Config key '%s' not found in config.yaml file, using default value: %v\n", key, defaultValue)
			return defaultValue
		}
	}

	return current
}

// Set sets a configuration value by key
func (c *Config) Set(key string, value interface{}) error {
	keys := strings.Split(key, ".")

	currentConfig := make(map[string]interface{})
	if err := c.viper.Unmarshal(&currentConfig); err != nil {
		return fmt.Errorf("failed to unmarshal current config: %v", err)
	}

	current := currentConfig
	for i, k := range keys[:len(keys)-1] {
		if current[k] == nil {
			current[k] = make(map[string]interface{})
		}

		if currentMap, ok := current[k].(map[string]interface{}); ok {
			current = currentMap
		} else {
			return fmt.Errorf("key '%s' is not a map", strings.Join(keys[:i+1], "."))
		}
	}

	current[keys[len(keys)-1]] = value

	return c.saveConfig(currentConfig)
}

// List returns all configuration values as flattened key-val pairs
func (c *Config) List() []KeyValue {
	var result []KeyValue
	c.flattenMap("", c.viper.AllSettings(), &result)
	return result
}

// just a key-val pair
type KeyValue struct {
	Key   string
	Value interface{}
}

// ideg flattenMap i just looked at GOrilla docs for this
func (c *Config) flattenMap(prefix string, m map[string]interface{}, result *[]KeyValue) {
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}

		if nestedMap, ok := v.(map[string]interface{}); ok {
			c.flattenMap(key, nestedMap, result)
		} else {
			*result = append(*result, KeyValue{Key: key, Value: v})
		}
	}
}

// Reload reloads configuration from file
func (c *Config) Reload() error {
	return c.viper.ReadInConfig()
}

// GetConfigFile returns the path to the config file
func (c *Config) GetConfigFile() string {
	return c.configFile
}

// GetConfigDir returns the path to the config directory
func (c *Config) GetConfigDir() string {
	return c.configDir
}
