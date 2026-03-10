package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	ForgeUsername string `mapstructure:"forge_username"`
	Author        string `mapstructure:"author"`
	License       string `mapstructure:"license"`
	ForgeToken    string `mapstructure:"forge_token"`
}

func Load(path string) (Config, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, err
		}
		path = filepath.Join(home, ".config", "jig", "config.toml")
	}

	viper.SetConfigFile(path)
	viper.SetEnvPrefix("JIG")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return Config{}, err
		}
	}

	config := Config{}
	err := viper.Unmarshal(&config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
