package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Config
	}{
		{
			name: "defaults",
			env: map[string]string{
				"DATABASE_URL":         "",
				"SERVER_ADDR":          "",
				"ROLE":                 "",
				"REDIS_ADDR":           "",
				"RABBIT_URL":           "",
				"RABBIT_QUEUE":         "",
				"BATCH_SIZE":           "",
				"BATCH_TIMEOUT_MS":     "",
				"MATCHER_TICK_MS":      "",
				"FETCH_APPEALS_LIMIT":  "",
				"FETCH_MANAGERS_LIMIT": "",
			},
			want: Config{
				ConnectionString:   "",
				ServerAddr:         ":8080",
				BalancerRole:       "all",
				RedisAddr:          "redis:6379",
				RabbitURL:          "",
				RabbitQueue:        "balancer.events",
				BatchSize:          50,
				BatchTimeout:       time.Second,
				MatcherTick:        time.Second,
				FetchAppealsLimit:  50,
				FetchManagersLimit: 50,
			},
		},
		{
			name: "from env",
			env: map[string]string{
				"DATABASE_URL":         "postgres://user:pass@localhost/db",
				"SERVER_ADDR":          ":9090",
				"ROLE":                 "worker",
				"REDIS_ADDR":           "localhost:6379",
				"RABBIT_URL":           "amqp://guest:guest@localhost/",
				"RABBIT_QUEUE":         "events",
				"BATCH_SIZE":           "10",
				"BATCH_TIMEOUT_MS":     "500",
				"MATCHER_TICK_MS":      "2000",
				"FETCH_APPEALS_LIMIT":  "25",
				"FETCH_MANAGERS_LIMIT": "30",
			},
			want: Config{
				ConnectionString:   "postgres://user:pass@localhost/db",
				ServerAddr:         ":9090",
				BalancerRole:       "worker",
				RedisAddr:          "localhost:6379",
				RabbitURL:          "amqp://guest:guest@localhost/",
				RabbitQueue:        "events",
				BatchSize:          10,
				BatchTimeout:       500 * time.Millisecond,
				MatcherTick:        2 * time.Second,
				FetchAppealsLimit:  25,
				FetchManagersLimit: 30,
			},
		},
		{
			name: "invalid int fallback",
			env: map[string]string{
				"BATCH_SIZE": "not-a-number",
			},
			want: Config{
				BatchSize: 50,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			cfg := Load()
			require.NotNil(t, cfg)
			if tc.name == "invalid int fallback" {
				assert.Equal(t, tc.want.BatchSize, cfg.BatchSize)
				return
			}
			assert.Equal(t, tc.want.ConnectionString, cfg.ConnectionString)
			assert.Equal(t, tc.want.ServerAddr, cfg.ServerAddr)
			assert.Equal(t, tc.want.BalancerRole, cfg.BalancerRole)
			assert.Equal(t, tc.want.RedisAddr, cfg.RedisAddr)
			assert.Equal(t, tc.want.RabbitURL, cfg.RabbitURL)
			assert.Equal(t, tc.want.RabbitQueue, cfg.RabbitQueue)
			assert.Equal(t, tc.want.BatchSize, cfg.BatchSize)
			assert.Equal(t, tc.want.BatchTimeout, cfg.BatchTimeout)
			assert.Equal(t, tc.want.MatcherTick, cfg.MatcherTick)
			assert.Equal(t, tc.want.FetchAppealsLimit, cfg.FetchAppealsLimit)
			assert.Equal(t, tc.want.FetchManagersLimit, cfg.FetchManagersLimit)
		})
	}
}
