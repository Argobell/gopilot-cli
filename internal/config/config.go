package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// RetryConfig 重试配置
type RetryConfig struct {
	Enabled         bool    `yaml:"enabled"`
	MaxRetries      int     `yaml:"max_retries"`
	InitialDelay    float64 `yaml:"initial_delay"`
	MaxDelay        float64 `yaml:"max_delay"`
	ExponentialBase float64 `yaml:"exponential_base"`
}

// LLMConfig LLM 配置
type LLMConfig struct {
	APIKey  string      `yaml:"api_key"`
	APIBase string      `yaml:"api_base"`
	Model   string      `yaml:"model"`
	Retry   RetryConfig `yaml:"retry"`
}

// AgentConfig Agent 配置
type AgentConfig struct {
	MaxSteps         int    `yaml:"max_steps"`
	WorkspaceDir     string `yaml:"workspace_dir"`
	SystemPromptPath string `yaml:"system_prompt_path"`
	TokenLimit       int    `yaml:"token_limit"`
}

// Config 主配置
type Config struct {
	LLM   LLMConfig   `yaml:"llm"`
	Agent AgentConfig `yaml:"agent"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		LLM: LLMConfig{
			APIBase: "http://localhost:8080",
			Model:   "gpt-oss",
			Retry: RetryConfig{
				Enabled:         true,
				MaxRetries:      3,
				InitialDelay:    1.0,
				MaxDelay:        60.0,
				ExponentialBase: 2.0,
			},
		},
		Agent: AgentConfig{
			MaxSteps:     50,
			WorkspaceDir: "./workspace",
			TokenLimit:   80000,
		},
	}
}

// LoadFromFile 从 YAML 文件加载配置
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
