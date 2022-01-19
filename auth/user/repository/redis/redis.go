package redis

import (
	"auth/myerrors"
	"auth/user/repository"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type redisCacheInterface struct {
	redisConn *redis.Client
}

func (r *redisCacheInterface) InsertToken(IIN, token string, refreshTtl time.Duration) error {
	return r.redisConn.Set(IIN, token, refreshTtl).Err()
}

func (r *redisCacheInterface) FindToken(IIN, token string) (string, error) {
	value, err := r.redisConn.Get(IIN).Result()
	if err == redis.Nil {
		return "", myerrors.ErrRefreshNotFound
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func NewRedisCacheInterface() (repository.CacheInterface, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	fmt.Println(pong)
	return &redisCacheInterface{redisConn: client}, nil
}
