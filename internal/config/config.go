package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Environment string        `mapstructure:"ENVIRONMENT"`
	Server      ServerConfig  `mapstructure:"SERVER"`
	Database    DatabaseConfig `mapstructure:"DATABASE"`
	JWT         JWTConfig     `mapstructure:"JWT"`
}

type ServerConfig struct {
	Port int `mapstructure:"PORT"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
	DBName   string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"SSL_MODE"`
}

type JWTConfig struct {
	Secret string `mapstructure:"SECRET"`
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("ENVIRONMENT", "development")
	viper.SetDefault("SERVER.PORT", 8080)
	viper.SetDefault("DATABASE.HOST", "localhost")
	viper.SetDefault("DATABASE.PORT", 5432)
	viper.SetDefault("DATABASE.USER", "postgres")
	viper.SetDefault("DATABASE.PASSWORD", "password")
	viper.SetDefault("DATABASE.DB_NAME", "simple_api")
	viper.SetDefault("DATABASE.SSL_MODE", "disable")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
