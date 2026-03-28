package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	ConfigDirName  = ".nvidia-nim"
	ConfigFileName = "config"
	ConfigFileExt  = "yaml"

	KeyAPIKey       = "api_key"
	KeyDefaultModel = "default_model"
	KeyMaxTokens    = "max_tokens"
	KeyTemperature  = "temperature"
	KeyStreamMode   = "stream"

	DefaultModel       = "meta/llama3-70b-instruct"
	DefaultMaxTokens   = 1024
	DefaultTemperature = 0.7
	DefaultStream      = true
)

type Config struct {
	APIKey       string  `mapstructure:"api_key"`
	DefaultModel string  `mapstructure:"default_model"`
	MaxTokens    int     `mapstructure:"max_tokens"`
	Temperature  float64 `mapstructure:"temperature"`
	Stream       bool    `mapstructure:"stream"`
}

func Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("impossible de trouver le répertoire home : %w", err)
	}

	configDir := filepath.Join(homeDir, ConfigDirName)

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("impossible de créer le répertoire de config : %w", err)
	}

	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileExt)
	viper.AddConfigPath(configDir)

	viper.SetDefault(KeyDefaultModel, DefaultModel)
	viper.SetDefault(KeyMaxTokens, DefaultMaxTokens)
	viper.SetDefault(KeyTemperature, DefaultTemperature)
	viper.SetDefault(KeyStreamMode, DefaultStream)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("erreur lecture config : %w", err)
		}
	}

	return nil
}

func SaveAPIKey(apiKey string) error {
	viper.Set(KeyAPIKey, apiKey)
	return writeConfig()
}

func GetAPIKey() string {
	if envKey := os.Getenv("NVIDIA_API_KEY"); envKey != "" {
		return envKey
	}
	return viper.GetString(KeyAPIKey)
}

func HasAPIKey() bool {
	return GetAPIKey() != ""
}

func SaveModel(model string) error {
	viper.Set(KeyDefaultModel, model)
	return writeConfig()
}

func GetModel() string {
	return viper.GetString(KeyDefaultModel)
}

func GetMaxTokens() int {
	return viper.GetInt(KeyMaxTokens)
}

func GetTemperature() float64 {
	return viper.GetFloat64(KeyTemperature)
}

func IsStreamEnabled() bool {
	return viper.GetBool(KeyStreamMode)
}

func GetAll() (*Config, error) {
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("erreur décodage config : %w", err)
	}
	return cfg, nil
}

func SaveAll(cfg *Config) error {
	viper.Set(KeyAPIKey, cfg.APIKey)
	viper.Set(KeyDefaultModel, cfg.DefaultModel)
	viper.Set(KeyMaxTokens, cfg.MaxTokens)
	viper.Set(KeyTemperature, cfg.Temperature)
	viper.Set(KeyStreamMode, cfg.Stream)
	return writeConfig()
}

func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ConfigDirName, ConfigFileName+"."+ConfigFileExt)
}

func writeConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configPath := filepath.Join(homeDir, ConfigDirName, ConfigFileName+"."+ConfigFileExt)
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("impossible d'écrire la config : %w", err)
	}
	return nil
}
