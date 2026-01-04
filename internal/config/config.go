package config

import (
  "encoding/json"
  "fmt"
  "os"
  "path/filepath"
)

const (
  ConfigDir  = ".cursor-bar"
  ConfigFile = "config.json"
)

type Config struct {
  Token  string `json:"token"`
  UserID string `json:"user_id"`
}

func GetConfigPath() (string, error) {
  home, err := os.UserHomeDir()
  if err != nil {
    return "", fmt.Errorf("failed to get home directory: %w", err)
  }
  return filepath.Join(home, ConfigDir, ConfigFile), nil
}

func Load() (*Config, error) {
  path, err := GetConfigPath()
  if err != nil {
    return nil, err
  }

  data, err := os.ReadFile(path)
  if err != nil {
    if os.IsNotExist(err) {
      return nil, nil
    }
    return nil, fmt.Errorf("failed to read config: %w", err)
  }

  var cfg Config
  if err := json.Unmarshal(data, &cfg); err != nil {
    return nil, fmt.Errorf("failed to parse config: %w", err)
  }

  return &cfg, nil
}

func Save(cfg *Config) error {
  path, err := GetConfigPath()
  if err != nil {
    return err
  }

  dir := filepath.Dir(path)
  if err := os.MkdirAll(dir, 0700); err != nil {
    return fmt.Errorf("failed to create config directory: %w", err)
  }

  data, err := json.MarshalIndent(cfg, "", "  ")
  if err != nil {
    return fmt.Errorf("failed to marshal config: %w", err)
  }

  if err := os.WriteFile(path, data, 0600); err != nil {
    return fmt.Errorf("failed to write config: %w", err)
  }

  return nil
}

func IsValid(cfg *Config) bool {
  return cfg != nil && cfg.Token != "" && cfg.UserID != ""
}
