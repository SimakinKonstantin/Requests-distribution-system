package main

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Role string

	PostgresDSN string

	RedisAddr string

	RabbitURL   string
	RabbitQueue string

	BatchSize    int
	BatchTimeout time.Duration

	MatcherTick time.Duration

	FetchAppealsLimit  int
	FetchManagersLimit int
}

func LoadConfig() Config {
	return Config{
		Role: getenv("ROLE", "all"),

		PostgresDSN: mustGetenv("POSTGRES_DSN"),
		RedisAddr:   getenv("REDIS_ADDR", "localhost:6379"),

		RabbitURL:   mustGetenv("RABBIT_URL"),
		RabbitQueue: getenv("RABBIT_QUEUE", "balancer.demo.events"),

		BatchSize:    getenvInt("BATCH_SIZE", 50),
		BatchTimeout: time.Duration(getenvInt("BATCH_TIMEOUT_MS", 1000)) * time.Millisecond,

		MatcherTick: time.Duration(getenvInt("MATCHER_TICK_MS", 1000)) * time.Millisecond,

		FetchAppealsLimit:  getenvInt("FETCH_APPEALS_LIMIT", 50),
		FetchManagersLimit: getenvInt("FETCH_MANAGERS_LIMIT", 50),
	}
}

func mustGetenv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing env: " + key)
	}
	return v
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func getenvInt(key string, def int) int {
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

