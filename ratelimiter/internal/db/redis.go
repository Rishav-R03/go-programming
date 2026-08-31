package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitRedis(ctx context.Context, add string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         add,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := rdb.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis failed: %v", err)
	}
	return rdb, nil
}

func LoadLuaScript(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Unable to read script: %v", err)
	}
	return string(content), nil
}

