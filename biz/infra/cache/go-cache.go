package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

func NewCache() *cache.Cache {
	c := cache.New(1*time.Hour, 2*time.Hour)
	return c
}
