package main

import (
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
)

func GetKeywordLink(keyword string, r *http.Request) string {
	keyword = regexp.QuoteMeta(keyword)
	u, err := r.URL.Parse(baseUrl.String() + "/keyword/" + pathURIEscape(keyword))
	panicIf(err)
	link := fmt.Sprintf("<a href=\"%s\">%s</a>", u, html.EscapeString(keyword))
	return link
}

func getReplacerStringsForHtmlify(r *http.Request) []string {
	replacerStrings, found := GetHtmlifyReplacerStringsFromCache()
	if !found {
		rows, err := db.Query(`
			SELECT keyword FROM entry ORDER BY keyword_length DESC
		`)
		panicIf(err)
		replacerStrings = make([]string, 0, 20000)
		for rows.Next() {
			var keyword string
			err := rows.Scan(&keyword)
			panicIf(err)

			link := GetKeywordLink(keyword, r)
			replacerStrings = append(replacerStrings, keyword)
			replacerStrings = append(replacerStrings, link)
		}
		rows.Close()
	}

	return replacerStrings
}

func getReplacerForHtmlify(r *http.Request) *strings.Replacer {
	replacerStrings := getReplacerStringsForHtmlify(r)
	replacer := strings.NewReplacer(replacerStrings...)

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
