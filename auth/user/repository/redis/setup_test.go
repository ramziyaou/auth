package redis

import (
	"log"
	"os"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
)

var (
	client *redis.Client
)

var (
	key = "key"
	val = "val"
)

func TestMain(m *testing.M) {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	client = redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	code := m.Run()
	os.Exit(code)
}
