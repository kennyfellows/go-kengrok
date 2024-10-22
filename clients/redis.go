package clients

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProxyMapper interface {
  Has(ctx context.Context, key string) bool
  Set(ctx context.Context, key string, val any) (bool, error)
  Get(ctx context.Context, key string) any
}

type RedisClient struct {
  client *redis.Client
}

func (r *RedisClient) Has(ctx context.Context, key string) bool {
  return true
}

func (r *RedisClient) Set(ctx context.Context, key string, val any)(bool, error) {
  return true, nil
}


func (r *RedisClient) Get(ctx context.Context, key string) any {
  return true
}

var client *RedisClient

func newClient() {

}

func GetClient() {

}


/* var redisClient *redis.Client */
/*  */
/* func initRedisClient( addr string ) *redis.Client { */
  /* redisClient = redis.NewClient(&redis.Options{ */
    /* Addr: addr, */
  /* }) */
/*  */
  /* ctx, cancel := context.WithTimeout( context.Background(), 5*time.Second ) */
/*  */
  /* defer cancel() */
/*  */
  /* _, err := redisClient.Ping( ctx ).Result() */
/*  */
  /* if err != nil { */
    /* log.Fatalf( "Failed to connect to Redis: %v", err ) */
  /* } */
/*  */
  /* log.Println("Successfully connected to Redis") */
/*  */
  /* return redisClient */
/* } */
/*  */
/* func GetRedisClient() *redis.Client { */
  /* if redisClient != nil { */
    /* return redisClient */
  /* } */
/*  */
  /* redisClient = initRedisClient("localhost:6379") */
/*  */
  /* return redisClient */
/* } */
