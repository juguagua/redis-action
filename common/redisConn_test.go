package common

import (
	"context"
	"testing"

	"redis-practice"

	"github.com/stretchr/testify/assert"
)

func TestConnectRedis(t *testing.T) {
	ctx := context.Background()

	conn := ConnectRedis(ctx, &RedisConf{
		Addr:     redis_practice.Addr,
		Password: redis_practice.Password,
		DB:       redis_practice.DB,
	})
	assert.NotNil(t, conn, "conn should not be nil")
}
