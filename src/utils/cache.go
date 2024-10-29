package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type CacheQrValue struct {
	Id     string
	Ref    string
	ApiKey string
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
	err := cache.Client.Set(context.TODO(), key, value, ttl).Err()

	if err != nil {
		return err
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

func (cache Cache) Delete(key string) error {
	_, err := cache.Client.Del(context.TODO(), key).Result()

	if err != nil {
		return err
	}
	return nil
}

func (cache Cache) QrKey(id string) string {
	return fmt.Sprintf("QR-%s", id)
}

func (cache Cache) QrValue(id, ref, apiKey string) string {
	item := CacheQrValue{Id: id, Ref: ref, ApiKey: apiKey}
	data, _ := json.Marshal(item)
	return string(data)
}

func CacheQrValueFromString(data string) (*CacheQrValue, error) {
	var value CacheQrValue
	err := json.Unmarshal([]byte(data), &value)
	if err != nil {
		return nil, err
	}
	return &value, nil
}
