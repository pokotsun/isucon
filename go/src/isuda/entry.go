package main

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
)

func htmlify(w http.ResponseWriter, r *http.Request, content string) string {
	if content == "" {
		return ""
	}
	rep_data, err := getReplacerFromCache()
	if err != nil {
		return ""
	}
	rep_data = append(rep_data, "\n")
	rep_data = append(rep_data, "<br />\n")
	replacer := strings.NewReplacer(rep_data...)
	return replacer.Replace(content)
}

func initReplacerToCache(r *http.Request) error {
	strs := []string{
		`&`, "&amp;",
		`'`, "&#39;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&#34;",
	}

	rows, err := db.Query(`
		SELECT keyword FROM entry ORDER BY CHARACTER_LENGTH(keyword) DESC
	`)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer rows.Close()
	key := REPLACER_KEY
	for i := 0; i < len(strs); i += 2 {
		data, _ := json.Marshal(strs[i])
		pushListDataToCache(key, data)
		data, _ = json.Marshal(strs[i+1])
		pushListDataToCache(key, data)
	}
	for rows.Next() {
		var keyword string
		err = rows.Scan(&keyword)
		if err != nil {
			fmt.Println(err)
			return err
		}
		data, _ := json.Marshal(keyword)
		err := pushListDataToCache(key, data)
		if err != nil {
			fmt.Println(err)
		}

		u, err := r.URL.Parse(baseUrl.String() + "/keyword/" + pathURIEscape(keyword))
		if err != nil {
			fmt.Println(err)
		}
		link := fmt.Sprintf("<a href=\"%s\">%s</a>", u, html.EscapeString(keyword))
		data, _ = json.Marshal(link)
		pushListDataToCache(key, data)
	}
	return nil
}

func pushReplacerToCache(keyword string, r *http.Request) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = conn.Do("MULTI")
	if err != nil {
		fmt.Println(err)
		return
	}

	data, _ := json.Marshal(keyword)
	err = pushListDataToCacheWithConnection(REPLACER_KEY, data, conn)
	if err != nil {
		fmt.Println(err)
	}

	u, err := r.URL.Parse(baseUrl.String() + "/keyword/" + pathURIEscape(keyword))
	if err != nil {
		fmt.Println(err)
	}
	link := fmt.Sprintf("<a href=\"%s\">%s</a>", u, html.EscapeString(keyword))
	data, _ = json.Marshal(link)
	err = pushListDataToCacheWithConnection(REPLACER_KEY, data, conn)
	if err != nil {
		fmt.Println(err)
	}
	conn.Do("Exec")
}

func removeReplacerFromCache(keyword string, r *http.Request) {
	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", redisHost, redisPort))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = conn.Do("MULTI")
	if err != nil {
		fmt.Println(err)
		return
	}
	data, _ := json.Marshal(keyword)
	removeListDataFromCacheWithConnection(REPLACER_KEY, data, conn)

	u, err := r.URL.Parse(baseUrl.String() + "/keyword/" + pathURIEscape(keyword))
	if err != nil {
		fmt.Println(err)
	}
	link := fmt.Sprintf("<a href=\"%s\">%s</a>", u, html.EscapeString(keyword))
	data, _ = json.Marshal(link)
	removeListDataFromCacheWithConnection(REPLACER_KEY, data, conn)
	conn.Do("Exec")
}

func getReplacerFromCache() ([]string, error) {
	data, err := getListDataFromCache(REPLACER_KEY)
	if err != nil {
		return nil, err
	}
	var rep_data []string
	err = json.Unmarshal(data, &rep_data)
	if err != nil {
		return nil, err
	}
	return rep_data, nil
}

func setHtmlTocache(entryID int, content string) {
	data, err := json.Marshal(content)
	if err != nil {
		fmt.Println(err)
	}
	setHashDataToCache(ENTRY_HTML, strconv.Itoa(entryID), data)
}

func getHtmlFromCache(entryID int) (string, error) {
	data, err := getHashDataFromCache(ENTRY_HTML, strconv.Itoa(entryID))
	if err != nil {
		return "", err
	}
	var str string
	err = json.Unmarshal(data, &str)
	if err != nil {
		return "", err
	}
	return str, nil
}
