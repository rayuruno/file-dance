package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"gitlab.com/proctorexam/go/env"
)

var (
	redisUrl = env.Fetch("REDIS_URL", "redis://localhost:6379")
	rctx     = context.Background()
	rdb      *redis.Client
)

func init() {
	rop, err := redis.ParseURL(redisUrl)
	if err != nil {
		log.Fatal(err)
	}
	rdb = redis.NewClient(rop)
}

func userExists(user string) bool {
	return rdb.Exists(rctx, proxyKey(user)).Val() == 1
}

func setPassword(user, password string) (bool, error) {
	if len(password) < 1 {
		return false, fmt.Errorf("password cannot be blank")
	}
	phash, err := hashPassword(password)
	if err != nil {
		return false, err
	}
	return rdb.SetNX(rctx, passwordKey(user), phash, 0).Val(), nil
}

func getPassword(user string) string {
	return rdb.Get(rctx, passwordKey(user)).Val()
}

func setProxyAddress(user, addr string) error {
	return rdb.Set(rctx, proxyKey(user), addr, 0).Err()
}

func removeProxyAddress(user string) error {
	return rdb.Del(rctx, proxyKey(user)).Err()
}

func getProxyAddress(user string) string {
	return rdb.Get(rctx, proxyKey(user)).Val()
}

func proxyKey(user string) string {
	return "file.dance:proxy:" + user
}

func passwordKey(user string) string {
	return "file.dance:password:" + user
}
