package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var defaultConfig = map[string]interface{}{
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
		"debug_mode": false,
		"georgist_credits": 627,
	},
}

type Config struct {
	Language    string                 `yaml:"language"`
	Login       LoginConfig            `yaml:"login"`
	Chat        ChatConfig             `yaml:"chat"`
	Scheduling  SchedulingConfig       `yaml:"scheduling"`
	Privacy     PrivacyConfig          `yaml:"privacy"`
	Advanced    AdvancedConfig         `yaml:"advanced"`
	configDir   string
	configFile  string
	rawConfig   map[string]interface{}
}

type LoginConfig struct {
	DefaultUsername *string `yaml:"default_username"`
	CurrentUsername *string `yaml:"current_username"`
}

type ChatConfig struct {
	Layout string `yaml:"layout"`
	Colors bool   `yaml:"colors"`
}

type SchedulingConfig struct {
	DefaultScheduleDuration string `yaml:"default_schedule_duration"`
}

type PrivacyConfig struct {
	InvisibleMode bool `yaml:"invisible_mode"`
}

type AdvancedConfig struct {
	DebugMode        bool   `yaml:"debug_mode"`
	DataDir          string `yaml:"data_dir"`
	UsersDir         string `yaml:"users_dir"`
	CacheDir         string `yaml:"cache_dir"`
	MediaDir         string `yaml:"media_dir"`
	GeneratedDir     string `yaml:"generated_dir"`
	GeorgistCredits  int    `yaml:"georgist_credits"`
}

var instance *Config

func GetInstance() *Config {
	if instance == nil {
		instance = &Config{}
		instance.initialize()
	}
	return instance
}

func (c *Config) initialize() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	defaultDataDir := filepath.Join(homeDir, ".instagram-cli")
	
	defaultConfig["advanced"].(map[string]interface{})["data_dir"] = defaultDataDir
	defaultConfig["advanced"].(map[string]interface{})["users_dir"] = filepath.Join(defaultDataDir, "users")
	defaultConfig["advanced"].(map[string]interface{})["cache_dir"] = filepath.Join(defaultDataDir, "cache")
	defaultConfig["advanced"].(map[string]interface{})["media_dir"] = filepath.Join(defaultDataDir, "media")
	defaultConfig["advanced"].(map[string]interface{})["generated_dir"] = filepath.Join(defaultDataDir, "generated")

	c.configDir = defaultDataDir
	c.configFile = filepath.Join(defaultDataDir, "config.yaml")
	c.loadConfig()
}

func (c *Config) loadConfig() {
	if err := os.MkdirAll(c.configDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create config directory: %v\n", err)
		return
	}

	if _, err := os.Stat(c.configFile); os.IsNotExist(err) {
		c.saveConfig(defaultConfig)
		c.rawConfig = defaultConfig
		c.unmarshalConfig()
		return
	}

	data, err := os.ReadFile(c.configFile)
	if err != nil {
		fmt.Printf("Warning: Could not read config file: %v\n", err)
		c.rawConfig = defaultConfig
		c.unmarshalConfig()
		return
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("Warning: Could not parse config file: %v\n", err)
		c.rawConfig = defaultConfig
		c.unmarshalConfig()
		return
	}

	c.rawConfig = mergeConfigs(defaultConfig, config)
	c.unmarshalConfig()
}

func (c *Config) unmarshalConfig() {
	data, err := yaml.Marshal(c.rawConfig)
	if err != nil {
		fmt.Printf("Warning: Could not marshal config: %v\n", err)
		return
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		fmt.Printf("Warning: Could not unmarshal config: %v\n", err)
	}
}

func (c *Config) saveConfig(config map[string]interface{}) {
	data, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Error marshaling config: %v\n", err)
		return
	}

	if err := os.WriteFile(c.configFile, data, 0644); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
	}
}

func mergeConfigs(defaults, user map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for k, v := range defaults {
		result[k] = v
	}
	
	for k, v := range user {
		if userMap, ok := v.(map[string]interface{}); ok {
			if defaultMap, exists := result[k].(map[string]interface{}); exists {
				result[k] = mergeConfigs(defaultMap, userMap)
			} else {
				result[k] = v
			}
		} else {
			result[k] = v
		}
	}
	
	return result
}

func (c *Config) Get(key string, defaultValue interface{}) interface{} {
	keys := strings.Split(key, ".")
	current := c.rawConfig

	for _, k := range keys {
		if val, exists := current[k]; exists {
			if mapVal, ok := val.(map[string]interface{}); ok {
				current = mapVal
			} else {
				return val
			}
		} else {
			defaultVal := getFromMap(defaultConfig, keys)
			if defaultVal != nil {
				fmt.Printf("Warning: Config key '%s' not found in config.yaml file, using default value: %v\n", key, defaultVal)
				return defaultVal
			}
			return defaultValue
		}
	}

	return current
}

func getFromMap(config map[string]interface{}, keys []string) interface{} {
	current := config
	for _, k := range keys {
		if val, exists := current[k]; exists {
			if mapVal, ok := val.(map[string]interface{}); ok {
				current = mapVal
			} else {
				return val
			}
		} else {
			return nil
		}
	}
	return current
}

func (c *Config) Set(key string, value interface{}) {
	keys := strings.Split(key, ".")
	current := c.rawConfig

	for _, k := range keys[:len(keys)-1] {
		if val, exists := current[k]; exists {
			if mapVal, ok := val.(map[string]interface{}); ok {
				current = mapVal
			} else {
				current[k] = make(map[string]interface{})
				current = current[k].(map[string]interface{})
			}
		} else {
			current[k] = make(map[string]interface{})
			current = current[k].(map[string]interface{})
		}
	}

	current[keys[len(keys)-1]] = value

	c.saveConfig(c.rawConfig)
	
	c.unmarshalConfig()
}

func (c *Config) GetString(key string, defaultValue string) string {
	if val := c.Get(key, defaultValue); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (c *Config) GetBool(key string, defaultValue bool) bool {
	if val := c.Get(key, defaultValue); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

func (c *Config) GetInt(key string, defaultValue int) int {
	if val := c.Get(key, defaultValue); val != nil {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return defaultValue
}

func (c *Config) GetConfigDir() string {
	return c.configDir
}

func (c *Config) GetConfigFile() string {
	return c.configFile
}

func (c *Config) Reset() {
	c.rawConfig = defaultConfig
	c.saveConfig(c.rawConfig)
	c.unmarshalConfig()
}
