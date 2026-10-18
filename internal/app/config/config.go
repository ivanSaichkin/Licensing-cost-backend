package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	Redis       RedisConfig
	JWT         JWTConfig
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type JWTConfig struct {
	Token     string
	ExpiresIn time.Duration
}

func NewConfig() (*Config, error) {
	configName := "config"
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	cfg.Redis.Host = os.Getenv("REDIS_HOST")
	cfg.Redis.Port, _ = strconv.Atoi(os.Getenv("REDIS_PORT"))
	cfg.Redis.User = os.Getenv("REDIS_USER")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	cfg.JWT.Token = os.Getenv("JWT_TOKEN")
	cfg.JWT.ExpiresIn, _ = time.ParseDuration(os.Getenv("JWT_EXPIRES_IN"))

	log.Info("config parsed")
	return cfg, nil
}
