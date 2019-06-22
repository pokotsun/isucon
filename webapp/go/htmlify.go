package main

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
)

// get keyword-1, link-1, ..., keyword-n, link-n string list
func getReplaceKeywordAndLink(r *http.Request) []string {
	rows, err := db.Query(`
		SELECT keyword FROM entry ORDER BY CHARACTER_LENGTH(keyword) DESC
	`)
	panicIf(err)
	keywords := make([]string, 0, 500)
	for rows.Next() {
		var keyword string
		err := rows.Scan(&keyword)
		panicIf(err)

		keyword = regexp.QuoteMeta(keyword)
		u, err := r.URL.Parse(baseUrl.String() + "/keyword/" + pathURIEscape(keyword))
		panicIf(err)
		link := fmt.Sprintf("<a href=\"%s\">%s</a>", u, html.EscapeString(keyword))
		keywords = append(keywords, keyword)
		keywords = append(keywords, link)
	}
	rows.Close()
	return keywords
}

//TODO ここがN+1の根源
func htmlify(w http.ResponseWriter, r *http.Request, content string) string {
	if content == "" {
		return ""
	}

	//rows, err := db.Query(`
	//SELECT keyword FROM entry ORDER BY CHARACTER_LENGTH(keyword) DESC
	//`)
	//panicIf(err)
	//keywords := make([]string, 0, 500)
	//for rows.Next() {
	//var keyword string
	//err := rows.Scan(&keyword)
	//panicIf(err)
	//
	//keyword = regexp.QuoteMeta(keyword)
	//u, err := r.URL.Parse(baseUrl.String() + "/keyword/" + pathURIEscape(keyword))
	//panicIf(err)
	//link := fmt.Sprintf("<a href=\"%s\">%s</a>", u, html.EscapeString(keyword))
	//keywords = append(keywords, keyword)
	//keywords = append(keywords, link)
	//}
	//rows.Close()
	keywords := getReplaceKeywordAndLink(r)

	replacer := strings.NewReplacer(keywords...)
	content = replacer.Replace(content)

	return strings.Replace(content, "\n", "<br />\n", -1)
}
