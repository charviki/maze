package config

import "github.com/charviki/maze/fabrication/cradle/configutil"

// Config holds all configuration for The Forge service.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	LLM      LLMConfig      `yaml:"llm"`
	File     FileConfig     `yaml:"file"`
}

// ServerConfig holds HTTP and gRPC server configuration.
type ServerConfig struct {
	configutil.ServerConfig `yaml:",inline"`
	GRPCAddr                string `yaml:"grpc_addr"`
}

// DatabaseConfig holds PostgreSQL connection parameters.
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// LLMConfig holds LLM provider settings for AI features.
type LLMConfig struct {
	Provider string `yaml:"provider"`
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"`
	Model    string `yaml:"model"`
}

// FileConfig holds file storage path configuration.
type FileConfig struct {
	StoragePath string `yaml:"storage_path"`
}

// LoadFromExe loads configuration from a YAML file next to the executable and applies environment variable overrides.
func LoadFromExe(filename ...string) (*Config, error) {
	var cfg Config
	if _, err := configutil.LoadFromExe(&cfg, filename...); err != nil {
		return nil, err
	}
	if err := configutil.ApplyEnvOverrides("THE_FORGE", &cfg); err != nil {
		return nil, err
	}
	validate(&cfg)
	return &cfg, nil
}

func validate(cfg *Config) {
	if cfg.Server.ListenAddr == "" {
		cfg.Server.ListenAddr = ":8080"
	}
	if cfg.Server.GRPCAddr == "" {
		cfg.Server.GRPCAddr = ":9090"
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.Name == "" {
		cfg.Database.Name = "maze_forge"
	}
	if cfg.LLM.Provider == "" {
		cfg.LLM.Provider = "openai"
	}
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "gpt-4o"
	}
	if cfg.File.StoragePath == "" {
		cfg.File.StoragePath = "/tmp/the-forge/files"
	}
}
