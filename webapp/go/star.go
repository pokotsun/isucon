package main

import (
	"database/sql"
	"net/http"
	"time"
)

var (
	allStars = make([]Star, 0, 2000)
)

func loadStars(keyword string) []Star {
	stars := make([]Star, 0, 10)
	for _, star := range allStars {
		if star.Keyword == keyword {
			stars = append(stars, star)
		}
	}
	//rows, err := db.Query(`SELECT * FROM star WHERE keyword = ?`, keyword)
	//if err != nil && err != sql.ErrNoRows {
	//	panicIf(err)
	//	return stars
	//}

	//for rows.Next() {
	//	s := Star{}
	//	err := rows.Scan(&s.ID, &s.Keyword, &s.UserName, &s.CreatedAt)
	//	panicIf(err)
	//	stars = append(stars, s)
	//}
	//rows.Close()

	return stars
}

// GET /star
func starsHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.FormValue("keyword")
	stars := loadStars(keyword)
	re.JSON(w, http.StatusOK, map[string][]Star{
		"result": stars,
	})
}

// POST /star
func starsPostHandler(w http.ResponseWriter, r *http.Request) {
	keyword := r.FormValue("keyword")

	// check target keyword exist
	row := db.QueryRow(`SELECT COUNT(*) FROM entry WHERE keyword = ?`, keyword)
	var count int64
	err := row.Scan(&count)
	if err == sql.ErrNoRows || count == 0 {
		notFound(w)
		return
	}

	user := r.FormValue("user")
	newStar := Star{Keyword: keyword, UserName: user, CreatedAt: time.Now()}
	allStars = append(allStars, newStar)
	//_, err = db.Exec(`INSERT INTO star (keyword, user_name, created_at) VALUES (?, ?, NOW())`, keyword, user)
	//panicIf(err)

	re.JSON(w, http.StatusOK, map[string]string{"result": "ok"})
}
