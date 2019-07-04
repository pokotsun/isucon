package main

import (
	"github.com/patricmn/go-cache"
	"strconv"
	"time"
)

const (
	STAR_NUM_KEY = "STAR_NUM-"
)

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

func GetStarNumFromCache(keyword string) (int, bool) {
	key := STAR_NUM_KEY + keyword
	data_i, found := getDataFromCache(key)
	var num int = -1
	if found {
		num, _ = data_i.(int)
	}
	return num, found
}
