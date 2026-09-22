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
	LogLevel      string
	DBType        string
	SQLitePath    string
	AutoFailover  bool
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
	JWTSecret            string
	MidtransServerKey    string
	MidtransClientKey    string
	MidtransIsProduction bool
}

type AppConfig struct {
	Server struct {
		Port     string `mapstructure:"port"`
		Env      string `mapstructure:"env"`
		LogLevel string `mapstructure:"log_level"`
	} `mapstructure:"server"`
	Database struct {
		Type         string `mapstructure:"type"`
		SQLitePath   string `mapstructure:"sqlite_path"`
		AutoFailover bool   `mapstructure:"auto_failover"`
		Host         string `mapstructure:"host"`
		Port         string `mapstructure:"port"`
		User         string `mapstructure:"user"`
		Password     string `mapstructure:"password"`
		Name         string `mapstructure:"name"`
		SSLMode      string `mapstructure:"ssl_mode"`
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
	JWT struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"jwt"`
	Midtrans struct {
		ServerKey    string `mapstructure:"server_key"`
		ClientKey    string `mapstructure:"client_key"`
		IsProduction bool   `mapstructure:"is_production"`
	} `mapstructure:"midtrans"`
}

func Load() *Config {
	v := viper.New()

	// Default values
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.env", "development")
	v.SetDefault("server.log_level", "info")
	v.SetDefault("database.type", "postgres")
	v.SetDefault("database.sqlite_path", "./pos_local.db")
	v.SetDefault("database.auto_failover", true)
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "5432")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.name", "pos_db")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.secret", "super-secret-pos-jwt-key-2026")

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
	v.BindEnv("server.log_level", "LOG_LEVEL")
	v.BindEnv("database.type", "DB_TYPE")
	v.BindEnv("database.sqlite_path", "SQLITE_PATH")
	v.BindEnv("database.auto_failover", "AUTO_FAILOVER")
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
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("midtrans.server_key", "MIDTRANS_SERVER_KEY")
	v.BindEnv("midtrans.client_key", "MIDTRANS_CLIENT_KEY")
	v.BindEnv("midtrans.is_production", "MIDTRANS_IS_PRODUCTION")

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

	jwtSecret := appCfg.JWT.Secret
	if jwtSecret == "" {
		jwtSecret = "super-secret-pos-jwt-key-2026"
	}

	logLevel := appCfg.Server.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}

	dbType := appCfg.Database.Type
	if dbType == "" {
		dbType = "postgres"
	}
	sqlitePath := appCfg.Database.SQLitePath
	if sqlitePath == "" {
		sqlitePath = "./pos_local.db"
	}

	return &Config{
		Port:                 appCfg.Server.Port,
		Env:                  appCfg.Server.Env,
		LogLevel:             logLevel,
		DBType:               dbType,
		SQLitePath:           sqlitePath,
		AutoFailover:         appCfg.Database.AutoFailover,
		DBHost:               appCfg.Database.Host,
		DBPort:               appCfg.Database.Port,
		DBUser:               appCfg.Database.User,
		DBPassword:           appCfg.Database.Password,
		DBName:               appCfg.Database.Name,
		DBSslMode:            appCfg.Database.SSLMode,
		RedisAddr:            appCfg.Redis.Addr,
		RedisPassword:        appCfg.Redis.Password,
		RedisDB:              appCfg.Redis.DB,
		GeminiAPIKey:         appCfg.Vision.GeminiAPIKey,
		OpenAIAPIKey:         appCfg.Vision.OpenAIAPIKey,
		JWTSecret:            jwtSecret,
		MidtransServerKey:    appCfg.Midtrans.ServerKey,
		MidtransClientKey:    appCfg.Midtrans.ClientKey,
		MidtransIsProduction: appCfg.Midtrans.IsProduction,
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s&connect_timeout=3",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSslMode)
}

