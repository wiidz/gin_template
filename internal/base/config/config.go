package config

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type HTTPConfig struct {
	IP   string `mapstructure:"ip"`
	Port string `mapstructure:"port"`
}

type HTTPMultiConfig struct {
	Client  HTTPConfig `mapstructure:"client"`
	Console HTTPConfig `mapstructure:"console"`
}

type DBConfig struct {
	DSN         string `mapstructure:"dsn"`
	AutoMigrate bool   `mapstructure:"autoMigrate"`
}

type AppConfig struct {
	Env   string          `mapstructure:"env"`
	HTTP  HTTPConfig      `mapstructure:"http"`
	HTTP2 HTTPMultiConfig `mapstructure:"http2"`
	DB    DBConfig        `mapstructure:"db"`
}

var C AppConfig

func Init() {
	// Load .env if present
	_ = godotenv.Load()

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("env", "dev")
	viper.SetDefault("http.ip", "0.0.0.0")
	viper.SetDefault("http.port", "8080")
	viper.SetDefault("http2.client.ip", "0.0.0.0")
	viper.SetDefault("http2.client.port", "8080")
	viper.SetDefault("http2.console.ip", "0.0.0.0")
	viper.SetDefault("http2.console.port", "8081")
	viper.SetDefault("db.dsn", "")
	viper.SetDefault("db.autoMigrate", false)

	if err := viper.ReadInConfig(); err != nil {
		// optional
		log.Printf("config: using defaults/env, no config file: %v", err)
	}
	if err := viper.Unmarshal(&C); err != nil {
		log.Fatalf("config unmarshal error: %v", err)
	}
}
