package redis

import (
	"testing"
	"time"

	"github.com/elliotchance/redismock"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

func TestSet(t *testing.T) {
	exp := time.Duration(0)

	mock := redismock.NewNiceMock(client)
	mock.On("Set", key, val, exp).Return(redis.NewStatusResult("", nil))

	r := &redisCacheInterface{client}
	err := r.InsertToken(key, val, exp)
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	mock := redismock.NewNiceMock(client)
	mock.On("Get", key).Return(redis.NewStringResult(val, nil))

	r := &redisCacheInterface{client}
	res, err := r.FindToken(key, val)
	assert.NoError(t, err)
	assert.Equal(t, val, res)
}
