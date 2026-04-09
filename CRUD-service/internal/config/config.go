package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration.
type Config struct {
	// CRUD-service
	ConnectionString string
	ServerAddr       string

	// Balancer subsystem
	BalancerRole string

	// Balancer uses the same Postgres DB as the CRUD-service (DATABASE_URL),
	// but can be overridden with POSTGRES_DSN.
	BalancerPostgresDSN string

	RedisAddr   string
	RabbitURL   string
	RabbitQueue string

	BatchSize    int
	BatchTimeout time.Duration

	MatcherTick time.Duration

	FetchAppealsLimit  int
	FetchManagersLimit int
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	databaseURL := getEnv("DATABASE_URL", "")
	// Balancer may reuse the same DB; override with POSTGRES_DSN if needed.
	postgresDSN := getEnv("POSTGRES_DSN", databaseURL)

	return &Config{
		ConnectionString: databaseURL,
		ServerAddr:       getEnv("SERVER_ADDR", ":8080"),

		BalancerRole:        getEnv("ROLE", "all"),
		BalancerPostgresDSN: postgresDSN,

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
