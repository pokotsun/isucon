package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

func loadStars(keyword string) []*Star {

	rows, err := db.Query(`SELECT * FROM star WHERE keyword = ?`, keyword)
	if err != nil && err != sql.ErrNoRows {
		panicIf(err)
	}

	stars := make([]*Star, 0, 10)
	for rows.Next() {
		s := Star{}
		err := rows.Scan(&s.ID, &s.Keyword, &s.UserName, &s.CreatedAt)
		panicIf(err)
		stars = append(stars, &s)
	}
	rows.Close()

	return stars

}