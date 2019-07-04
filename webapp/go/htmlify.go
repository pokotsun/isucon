package main

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
)

// keyword-1, link-1, ..., keyword-n, link-n string list to Replacer
func getReplacerForHtmlify(r *http.Request) *strings.Replacer {
	//rows, err := db.Query(`
	//	SELECT keyword FROM entry ORDER BY CHARACTER_LENGTH(keyword) DESC
	//`)
	replacer, found := GetHtmlifyReplacerFromCache()
	if !found {
		rows, err := db.Query(`
		SELECT keyword FROM entry ORDER BY keyword_length DESC
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
		replacer = strings.NewReplacer(keywords...)

		SetHtmlifyReplacerToCache(replacer) // cache Replacer
	}
	return replacer
}

func htmlify(w http.ResponseWriter, r *http.Request, content string) string {
	replacer := getReplacerForHtmlify(r)
	return htmlifyWithReplacer(w, r, content, replacer)
}

func htmlifyWithReplacer(w http.ResponseWriter, r *http.Request, content string, replacer *strings.Replacer) string {
	if content == "" {
		return ""
	}
	content = replacer.Replace(content)

	return strings.Replace(content, "\n", "<br />\n", -1)
}
