package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port          string
	Env           string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSslMode     string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	GeminiAPIKey  string
	OpenAIAPIKey  string
}

type AppConfig struct {
	Server struct {
		Port string `mapstructure:"port"`
		Env  string `mapstructure:"env"`
	} `mapstructure:"server"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Name     string `mapstructure:"name"`
		SSLMode  string `mapstructure:"ssl_mode"`
	} `mapstructure:"database"`
	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`
	Vision struct {
		GeminiAPIKey string `mapstructure:"gemini_api_key"`
		OpenAIAPIKey string `mapstructure:"openai_api_key"`
	} `mapstructure:"vision"`
}

func Load() *Config {
	v := viper.New()

	// Default values
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.env", "development")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "pos_db")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Viper config search paths
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath("./backend/config")
	v.AddConfigPath("../config")
	v.AddConfigPath("../../config")
	v.AddConfigPath(".")

	// Bind environment variables explicitly for seamless overrides
	v.BindEnv("server.port", "PORT")
	v.BindEnv("server.env", "ENV")
	v.BindEnv("database.host", "DB_HOST")
	v.BindEnv("database.port", "DB_PORT")
	v.BindEnv("database.user", "DB_USER")
	v.BindEnv("database.password", "DB_PASSWORD")
	v.BindEnv("database.name", "DB_NAME")
	v.BindEnv("database.ssl_mode", "DB_SSLMODE")
	v.BindEnv("redis.addr", "REDIS_ADDR")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.db", "REDIS_DB")
	v.BindEnv("vision.gemini_api_key", "GEMINI_API_KEY")
	v.BindEnv("vision.openai_api_key", "OPENAI_API_KEY")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// Suppress verbose log in tests if config file isn't present
	} else {
		log.Printf("Viper loaded configuration file: %s", v.ConfigFileUsed())
	}

	var appCfg AppConfig
	if err := v.Unmarshal(&appCfg); err != nil {
		log.Printf("Warning: Viper unmarshal error: %v", err)
	}

	return &Config{
		Port:          appCfg.Server.Port,
		Env:           appCfg.Server.Env,
		DBHost:        appCfg.Database.Host,
		DBPort:        appCfg.Database.Port,
		DBUser:        appCfg.Database.User,
		DBPassword:    appCfg.Database.Password,
		DBName:        appCfg.Database.Name,
		DBSslMode:     appCfg.Database.SSLMode,
		RedisAddr:     appCfg.Redis.Addr,
		RedisPassword: appCfg.Redis.Password,
		RedisDB:       appCfg.Redis.DB,
		GeminiAPIKey:  appCfg.Vision.GeminiAPIKey,
		OpenAIAPIKey:  appCfg.Vision.OpenAIAPIKey,
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSslMode)
}
