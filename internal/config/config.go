package config

import (
	"fmt"
	"os"
)

// Config holds our verified environment properties.
type Config struct {
	OllamaURL string
	ModelName string
	TargetNS  string
	ReportDir string
	SlackURL  string
}

// Load reads values from the environment and ensures critical keys exist.
func Load() (*Config, error) {
	cfg := &Config{
		OllamaURL: getEnv("OLLAMA_URL", "http://localhost:11434"),
		ModelName: getEnv("MODEL_NAME", "qwen3:8b"),
		TargetNS:  getEnv("TARGET_NAMESPACE", "kubemind"),
		ReportDir: getEnv("REPORT_DIR", "./reports"),
		SlackURL:  getEnv("SLACK_WEBHOOK_URL", ""), // 🔍 FIXED: Crucial variable binding mapped here
	}

	// Simple structural validation sanity check
	if cfg.OllamaURL == "" {
		return nil, fmt.Errorf("OLLAMA_URL configuration property is empty")
	}

	return cfg, nil
}

// getEnv is an internal helper function (lowercase = private to this file)
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}