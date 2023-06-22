package utils

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const REDIS_HOST string = "URL_REDIS_HOST"
const REDIS_PASSWORD string = "REDIS_PASSWORD"
const REDIS_DEFAULT_DB string = "REDIS_DEFAULT_DB"
const REDIS_CACHE_TTL string = "REDIS_CACHE_TTL"

const ERROR_CACHE_UNKNOWN string = "cache: unknown"
const ERROR_CACHE_NOTFOUND string = "cache: not_found"

type Cache struct {
	Client *redis.Client
	Ttl    time.Duration
}

func NewCache() *Cache {
	var instance Cache
	db, _ := strconv.Atoi(GetEnv(REDIS_DEFAULT_DB, "0"))
	instance.Client = redis.NewClient(&redis.Options{
		Addr:     GetEnv(REDIS_HOST, "localhost:6379"),
		Password: GetEnv(REDIS_PASSWORD, ""),
		DB:       db,
	})
	duration, _ := time.ParseDuration(GetEnv(REDIS_CACHE_TTL, "10m"))
	instance.Ttl = duration
	return &instance
}

func (cache Cache) Set(key, value string, ttl time.Duration) error {
	err := cache.Client.Set(context.TODO(), key, value, ttl)

	if err != nil {
		return errors.New(ERROR_CACHE_UNKNOWN)
	}
	return nil
}

func (cache Cache) Get(key string) (string, error) {
	val, err := cache.Client.Get(context.TODO(), key).Result()

	if err == redis.Nil {
		return "", errors.New(ERROR_CACHE_NOTFOUND)
	} else if err != nil {
		return "", errors.New(ERROR_CACHE_UNKNOWN)
	}
	return val, nil
}
