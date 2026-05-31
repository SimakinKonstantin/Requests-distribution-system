package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ConnectionString string
	ServerAddr       string

	BalancerRole string

	RedisAddr   string
	RabbitURL   string
	RabbitQueue string

	BatchSize    int
	BatchTimeout time.Duration

	MatcherTick time.Duration

	FetchAppealsLimit  int
	FetchManagersLimit int
}

func Load() *Config {
	databaseURL := getEnv("DATABASE_URL", "")
	return &Config{
		ConnectionString: databaseURL,
		ServerAddr:       getEnv("SERVER_ADDR", ":8080"),

		BalancerRole: getEnv("ROLE", "all"),

		RedisAddr:   getEnv("REDIS_ADDR", "redis:6379"),
		RabbitURL:   getEnv("RABBIT_URL", ""),
		RabbitQueue: getEnv("RABBIT_QUEUE", "balancer.events"),

		BatchSize:    getEnvInt("BATCH_SIZE", 50),
		BatchTimeout: time.Duration(getEnvInt("BATCH_TIMEOUT_MS", 1000)) * time.Millisecond,

		MatcherTick: time.Duration(getEnvInt("MATCHER_TICK_MS", 1000)) * time.Millisecond,

		FetchAppealsLimit:  getEnvInt("FETCH_APPEALS_LIMIT", 50),
		FetchManagersLimit: getEnvInt("FETCH_MANAGERS_LIMIT", 50),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
