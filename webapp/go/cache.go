package main

import (
	"github.com/patrickmn/go-cache"
	"strconv"
	"strings"
	"time"
)

const (
	STAR_NUM_KEY             = "STAR_NUM-"
	HTMLIFY_REPLACER_STRINGS = "HTMLIFY_REPLACER_KEY"
	KEYWORD_HTML_KEY         = "KEYWORD_HTML_KEY-"
)

var (
	cacheKeywordCount = 0
	cache_            = cache.New(5*time.Hour, 10*time.Hour)
)

func getDataFromCache(key string) (interface{}, bool) {
	data_i, found := cache_.Get(key)
	return data_i, found
}

func setData(key string, value interface{}) {
	cache_.Set(key, value, cache.NoExpiration)
}

func deleteData(key string) {
	cache_.Delete(key)
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

/*************************/
/* HtmlifyReplacerString */
/*************************/
func GetHtmlifyReplacerStringFromCache() (*strings.Replacer, bool) {
	key := HTMLIFY_REPLACER_STRINGS
	data_i, found := getDataFromCache(key)
	if found {
		r, _ := data_i.(strings.Replacer)
		return &r, found
	} else {
		return strings.NewReplacer(), found
	}
}

func SetHtmlifyReplacerStringToCache(r *strings.Replacer) {
	key := HTMLIFY_REPLACER_STRINGS
	setData(key, *r)
}

func DeleteHtmlifyReplacerStringFromCache() {
	key := HTMLIFY_REPLACER_STRINGS
	deleteData(key)
}

/***********************/
/* Keyword Linked HTML */
/***********************/
func GetKeywordHtmlFromCache(entryID int) (string, bool) {
	key := KEYWORD_HTML_KEY + strconv.Itoa(entryID)
	data_i, found := getDataFromCache(key)
	var html string = ""
	if found {
		html, _ = data_i.(string)
	}
	return html, found
}

func SetKeywordHtmlToCache(entryID int, html string) {
	key := KEYWORD_HTML_KEY + strconv.Itoa(entryID)
	setData(key, html)
}

func DeleteKeywordHtmlFromCache(entryID int) {
	key := KEYWORD_HTML_KEY + strconv.Itoa(entryID)
	deleteData(key)
}
