package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
)

func loadStars(keyword string) []*Star {
	v := url.Values{}
	v.Set("keyword", keyword)
	resp, err := http.Get(fmt.Sprintf("%s/stars", isutarEndpoint) + "?" + v.Encode())
	panicIf(err)
	defer resp.Body.Close()

	var data struct {
		Result []*Star `json:result`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	panicIf(err)
	return data.Result
}