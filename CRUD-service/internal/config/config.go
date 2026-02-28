package config

import "os"

// Config holds the application configuration.
type Config struct {
	ConnectionString string
	ServerAddr       string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		ConnectionString: getEnv("DATABASE_URL", ""),
		ServerAddr:       getEnv("SERVER_ADDR", ":8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
