package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

//redis计数
func redisCount() {
	redisClient := redis.NewClient(&redis.Options{
		Addr:         redisAddress,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     100,
		PoolTimeout:  30 * time.Second,
	})
	for name := range ch {
		redisClient.Incr(fmt.Sprintf("ratelimit:%v:%v", name, time.Now().Format("20060102")))
	}
}
