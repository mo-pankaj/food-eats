package db

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	redisClient *redis.Client
)

func InitRedisPool(address string, maxIdle, maxActive int) {
	redisClient = redis.NewClient(&redis.Options{
		MaxIdleConns:    maxIdle,
		MaxActiveConns:  maxActive,
		ConnMaxIdleTime: 1 * time.Hour,
		ConnMaxLifetime: 24 * time.Hour,
		Addr:            address,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
}

func RedisConnFromPool() (*redis.Conn, error) {
	conn := redisClient.Conn()
	if conn == nil {
		return nil, errors.New("error getting redis conn from redis pool")
	}
	return conn, nil
}

func Close(conn *redis.Conn) error {
	if conn != nil {
		err := conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
