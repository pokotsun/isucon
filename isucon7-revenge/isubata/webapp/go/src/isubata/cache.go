package main

import (
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
)

const (
	MESSAGE_NUM_KEY = "NUM_MESSAGE-"
	HAVEREAD_KEY    = "HAVEREAD-"
)

var (
	dataCache = cache.New(5*time.Minute, 10*time.Minute)
)

// get num messages from cache
func GetNumMessagesFromCache(chID int64) (int64, bool) {
	key := MESSAGE_NUM_KEY + strconv.FormatInt(chID, 10)
	num_i, found := dataCache.Get(key)
	if found {
		num, _ := num_i.(int64)
		return num, true
	} else {
		return -1, false
	}
}

// set num messages to cache
func SetNumMessagesToCache(chID int64, value int64) {
	key := MESSAGE_NUM_KEY + strconv.FormatInt(chID, 10)
	dataCache.Set(key, value, cache.NoExpiration)
}

func GetHavereadFromCache(uID, chID int64) (int64, bool) {
	key := HAVEREAD_KEY + strconv.FormatInt(uID, 10) + "-" + strconv.FormatInt(chID, 10)
	num_i, found := dataCache.Get(key)
	if found {
		num, _ := num_i.(int64)
		return num, true
	} else {
		return -1, false
	}
}

func SetHavereadToCache(uID, chID int64, value int64) {
	key := HAVEREAD_KEY + strconv.FormatInt(uID, 10) + "-" + strconv.FormatInt(chID, 10)
	dataCache.Set(key, value, cache.NoExpiration)
}
