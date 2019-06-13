package main

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
)

const (
	MESSAGE_NUM_KEY = "NUM_MESSAGE-"
)

var (
	messageNumCache = cache.New(50*time.Minute, 100*time.Minute)
)

// get num messages from cache
func GetNumMessagesFromCache(chID int64) (int64, bool) {
	key := MESSAGE_NUM_KEY + strconv.FormatInt(chID, 10)
	num_i, found := messageNumCache.Get(key)
	if found {
		num, _ := num_i.(int64)
		fmt.Println("ON_GETNUM_FROMCACHE-" + key + ": " + strconv.FormatInt(num, 10))
		return num, true
	} else {
		return -1, false
	}
}

// set num messages to cache
func SetNumMessagesToCache(chID int64, value int64) {
	key := MESSAGE_NUM_KEY + strconv.FormatInt(chID, 10)
	fmt.Println("ON_SETNUM_FROMCACHE-" + key + ": " + strconv.FormatInt(value, 10))
	messageNumCache.Set(key, value, cache.DefaultExpiration)
}
