package main

import (
	"github.com/patricmn/go-cache"
	"strconv"
	"time"
)

const ()

var (
	cache_ = cache.New(5*time.Hour, 10*time.Hour)
)

func getDataFromCache(key string) (Interface, bool) {
	data_i, found := cache_.Get(key)
	return data_i, found
}

func setData(key string, value Interface) {
	cache_.Set(key, value, cache.NoExpiration)
}
