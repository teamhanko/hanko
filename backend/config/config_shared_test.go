package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisConfigDialAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		wantAddr string
		wantDB   int
		wantErr  bool
	}{
		{
			name:     "host and port without database",
			address:  "localhost:6379",
			wantAddr: "localhost:6379",
			wantDB:   0,
		},
		{
			name:     "host and port with database",
			address:  "localhost:6379/9",
			wantAddr: "localhost:6379",
			wantDB:   9,
		},
		{
			name:     "host without port with database",
			address:  "redis/3",
			wantAddr: "redis",
			wantDB:   3,
		},
		{
			name:     "explicit database zero",
			address:  "localhost:6379/0",
			wantAddr: "localhost:6379",
			wantDB:   0,
		},
		{
			name:     "trailing slash means unspecified",
			address:  "localhost:6379/",
			wantAddr: "localhost:6379",
			wantDB:   0,
		},
		{
			name:     "ipv6 host without database",
			address:  "[::1]:6379",
			wantAddr: "[::1]:6379",
			wantDB:   0,
		},
		{
			name:     "ipv6 host with database",
			address:  "[::1]:6379/4",
			wantAddr: "[::1]:6379",
			wantDB:   4,
		},
		{
			name:    "non numeric database",
			address: "localhost:6379/nine",
			wantErr: true,
		},
		{
			name:    "negative database",
			address: "localhost:6379/-1",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := RedisConfig{Address: test.address}

			address, database, err := cfg.DialAddress()

			if test.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantAddr, address)
			assert.Equal(t, test.wantDB, database)
		})
	}
}
