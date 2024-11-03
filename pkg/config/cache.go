package config

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var GlobalCache *cache.Cache

func SetupCache() {
	GlobalCache = cache.New(5*time.Minute, 360*time.Minute)
}

func GetCache(key string) (interface{}, bool) {
	return GlobalCache.Get(key)
}

func SetCache(key string, value interface{}) {
	GlobalCache.Set(key, value, cache.DefaultExpiration)
}
