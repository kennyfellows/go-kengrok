package utils

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func initRedisClient( addr string ) *redis.Client {
  redisClient = redis.NewClient(&redis.Options{
    Addr: addr,
  })

  ctx, cancel := context.WithTimeout( context.Background(), 5*time.Second )

  defer cancel()

  _, err := redisClient.Ping( ctx ).Result()

  if err != nil {
    log.Fatalf( "Failed to connect to Redis: %v", err )
  }

  log.Println("Successfully connected to Redis")

  return redisClient
}

func GetRedisClient() *redis.Client {
  if redisClient != nil {
    return redisClient
  }

  redisClient = initRedisClient("localhost:6379")

  return redisClient
}
