package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the framework configuration
type Config struct {
	// Agent configuration
	Agent struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		MaxTokens   int    `json:"max_tokens"`
	} `json:"agent"`

	// LLM provider configuration
	Provider struct {
		Type        string                 `json:"type"`
		ModelName   string                 `json:"model_name"`
		APIKey      string                 `json:"api_key"`
		BaseURL     string                 `json:"base_url"`
		Parameters  map[string]interface{} `json:"parameters"`
	} `json:"provider"`

	// Memory configuration
	Memory struct {
		Type        string `json:"type"`
		Path        string `json:"path"`
		MaxSize     int    `json:"max_size"`
	} `json:"memory"`

	// Tool configuration
	Tools struct {
		PluginDir   string   `json:"plugin_dir"`
		EnabledTools []string `json:"enabled_tools"`
	} `json:"tools"`
}

// LoadConfig loads configuration from a file
func LoadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig saves configuration to a file
func (c *Config) SaveConfig(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	configFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	return encoder.Encode(c)
}
