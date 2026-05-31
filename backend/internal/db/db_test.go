package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInvalidConnectionString(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "invalid connection string",
			dsn:  "postgres://invalid:invalid@127.0.0.1:1/nodb?connect_timeout=1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.dsn)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "db.Ping")
		})
	}
}
