package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	// "github.com/gomodule/redigo/redis"
)

func htmlify(w http.ResponseWriter, r *http.Request, content string) string {
	if content == "" {
		return ""
	}
	rows, err := db.Query(`
		SELECT * FROM entry ORDER BY CHARACTER_LENGTH(keyword) DESC
	`)
	panicIf(err)
	entries := make([]*Entry, 0, 500)
	for rows.Next() {
		e := Entry{}
		err := rows.Scan(&e.ID, &e.AuthorID, &e.Keyword, &e.Description, &e.UpdatedAt, &e.CreatedAt)
		panicIf(err)
		entries = append(entries, &e)
	}
	rows.Close()

	keywords := make([]string, 0, 500)
	for _, entry := range entries {
		keywords = append(keywords, regexp.QuoteMeta(entry.Keyword))
	}
	re := regexp.MustCompile("(" + strings.Join(keywords, "|") + ")")
	kw2sha := make(map[string]string)
	content = re.ReplaceAllStringFunc(content, func(kw string) string {
		kw2sha[kw] = "isuda_" + fmt.Sprintf("%x", sha1.Sum([]byte(kw)))
		return kw2sha[kw]
	})
	content = html.EscapeString(content)
	for kw, hash := range kw2sha {
		u, err := r.URL.Parse(baseUrl.String() + "/keyword/" + pathURIEscape(kw))
		panicIf(err)
		link := fmt.Sprintf("<a href=\"%s\">%s</a>", u, html.EscapeString(kw))
		content = strings.Replace(content, hash, link, -1)
	}
	return strings.Replace(content, "\n", "<br />\n", -1)
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
