package main

import (

	_ "github.com/go-sql-driver/mysql"
)

func loadStars(keyword string) []*Star {

	stars := getStarsFromCache(keyword)

	return stars

}