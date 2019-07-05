package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Songmu/strrand"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
)

const (
	sessionName   = "isuda_session"
	sessionSecret = "tonymoris"
)

var (
	isutarEndpoint string
	isupamEndpoint string

	baseUrl *url.URL
	db      *sql.DB
	re      *render.Render
	store   *sessions.CookieStore

	totalEntries int

	errInvalidUser = errors.New("Invalid User")
)

func setName(w http.ResponseWriter, r *http.Request) error {
	session := getSession(w, r)
	userID, ok := session.Values["user_id"]
	if !ok {
		return nil
	}
	setContext(r, "user_id", userID)
	userName, ok := session.Values["user_name"]
	setContext(r, "user_name", userName)
	return nil
}

func authenticate(w http.ResponseWriter, r *http.Request) error {
	if u := getContext(r, "user_id"); u != nil {
		return nil
	}
	return errInvalidUser
}

func initializeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := db.Exec(`DELETE FROM entry WHERE id > 7101`)
	panicIf(err)

	_, err = db.Exec("TRUNCATE star")
	panicIf(err)

	// at first, cache replacer strings
	replacerStrings := getReplacerStringsForHtmlify(r)
	SetHtmlifyReplacerStringsToCache(replacerStrings)

	//entries := GetAllEntries(w, r)
	//SetEntriesToCache(entries)

	row := db.QueryRow(`SELECT COUNT(*) FROM entry`)
	err = row.Scan(&totalEntries)
	if err != nil && err != sql.ErrNoRows {
		panicIf(err)
	}

	re.JSON(w, http.StatusOK, map[string]string{"result": "ok"})
}

//func GetAllEntries(w http.ResponseWriter, r *http.Request) []Entry {
//	entries, found := GetEntriesFromCache()
//	if !found {
//		rows, err := db.Query(fmt.Sprintf(
//			"SELECT id, author_id, keyword, description, updated_at, created_at FROM entry " +
//				"ORDER BY updated_at DESC",
//		))
//		if err != nil && err != sql.ErrNoRows {
//			panicIf(err)
//		}
//		entries = make([]Entry, 0, 10000)
//		replacer := getReplacerForHtmlify(r)
//		for rows.Next() {
//			e := Entry{}
//			err := rows.Scan(&e.ID, &e.AuthorID, &e.Keyword, &e.Description, &e.UpdatedAt, &e.CreatedAt)
//			panicIf(err)
//			e.Html = htmlifyWithReplacer(w, r, e.Description, replacer)
//			e.Stars = loadStars(e.Keyword)
//			entries = append(entries, e)
//		}
//		rows.Close()
//	}
//	return entries
//
//}

func GetEntriesPerPage(perPage, page int, w http.ResponseWriter, r *http.Request) []Entry {
	entries, found := GetEntriesPerPageFromCache(page)
	if !found {
		rows, err := db.Query(fmt.Sprintf(
			"SELECT id, author_id, keyword, description, updated_at, created_at FROM entry "+
				"ORDER BY updated_at DESC LIMIT %d OFFSET %d",
			perPage, perPage*(page-1),
		))
		if err != nil && err != sql.ErrNoRows {
			panicIf(err)
		}

		entries = make([]Entry, 0, 10)
		replacer := getReplacerForHtmlify(r)
		for rows.Next() {
			e := Entry{}
			err := rows.Scan(&e.ID, &e.AuthorID, &e.Keyword, &e.Description, &e.UpdatedAt, &e.CreatedAt)
			panicIf(err)
			e.Html = htmlifyWithReplacer(w, r, e.Description, replacer)
			e.Stars = loadStars(e.Keyword)
			entries = append(entries, e)
		}
		rows.Close()
	}
	return entries
}

// GET /
func topHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}

	perPage := 10
	p := r.URL.Query().Get("page")
	if p == "" {
		p = "1"
	}
	page, _ := strconv.Atoi(p)

	entries := GetEntriesPerPage(perPage, page, w, r)
	SetEntriesPerPageToCache(page, entries)

	lastPage := int(math.Ceil(float64(totalEntries) / float64(perPage)))
	pages := make([]int, 0, 10)
	start := int(math.Max(float64(1), float64(page-5)))
	end := int(math.Min(float64(lastPage), float64(page+5)))
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	re.HTML(w, http.StatusOK, "index", struct {
		Context  context.Context
		Entries  []Entry
		Page     int
		LastPage int
		Pages    []int
	}{
		r.Context(), entries, page, lastPage, pages,
	})
}

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	notFound(w)
}

// POST /keyword
func keywordPostHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}
	if err := authenticate(w, r); err != nil {
		forbidden(w)
		return
	}

	keyword := r.FormValue("keyword")
	if keyword == "" {
		badRequest(w)
		return
	}
	keywordLength := getStrLength(keyword)
	userID := getContext(r, "user_id").(int)
	description := r.FormValue("description")

	if isSpamContents(description) || isSpamContents(keyword) {
		http.Error(w, "SPAM!", http.StatusBadRequest)
		return
	}
	_, err := db.Exec(`
		INSERT INTO entry (author_id, keyword, description, created_at, updated_at, keyword_length)
		VALUES (?, ?, ?, NOW(), NOW(), ?)
		ON DUPLICATE KEY UPDATE
		author_id = ?, keyword = ?, description = ?, updated_at = NOW(), keyword_length = ?
	`, userID, keyword, description, keywordLength,
		userID, keyword, description, keywordLength)
	panicIf(err)

	// update replacer strings
	replacerStrings := getReplacerStringsForHtmlify(r)
	keywordLink := GetKeywordLink(keyword, r)
	replacerStrings = append(replacerStrings, keyword)
	replacerStrings = append(replacerStrings, keywordLink)
	SetHtmlifyReplacerStringsToCache(replacerStrings)

	DeleteEntriesPerPageFromCache(1)

	totalEntries++

	http.Redirect(w, r, "/", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}

	re.HTML(w, http.StatusOK, "authenticate", struct {
		Context context.Context
		Action  string
	}{
		r.Context(), "login",
	})
}

// POST /login authenticate user is valid
func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	row := db.QueryRow(`SELECT * FROM user WHERE name = ?`, name)
	user := User{}
	err := row.Scan(&user.ID, &user.Name, &user.Salt, &user.Password, &user.CreatedAt)
	if err == sql.ErrNoRows || user.Password != fmt.Sprintf("%x", sha1.Sum([]byte(user.Salt+r.FormValue("password")))) {
		forbidden(w)
		return
	}
	panicIf(err)
	session := getSession(w, r)
	session.Values["user_id"] = user.ID
	session.Values["user_name"] = user.Name
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(w, r)
	session.Options = &sessions.Options{MaxAge: -1}
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}

	re.HTML(w, http.StatusOK, "authenticate", struct {
		Context context.Context
		Action  string
	}{
		r.Context(), "register",
	})
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pw := r.FormValue("password")
	if name == "" || pw == "" {
		badRequest(w)
		return
	}
	userID := register(name, pw)
	session := getSession(w, r)
	session.Values["user_id"] = userID
	session.Values["user_name"] = name
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func register(user string, pass string) int64 {
	salt, err := strrand.RandomString(`....................`)
	panicIf(err)
	res, err := db.Exec(`INSERT INTO user (name, salt, password, created_at) VALUES (?, ?, ?, NOW())`,
		user, salt, fmt.Sprintf("%x", sha1.Sum([]byte(salt+pass))))
	panicIf(err)
	lastInsertID, _ := res.LastInsertId()
	return lastInsertID
}

// GET /keyword/{keyword}
func keywordByKeywordHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}

	keywordEncoded := mux.Vars(r)["keyword"]
	keyword := decodeUriString(keywordEncoded)
	row := db.QueryRow(`SELECT * FROM entry WHERE keyword = ?`, keyword)
	e := Entry{}
	err := row.Scan(&e.ID, &e.AuthorID, &e.Keyword, &e.Description, &e.UpdatedAt, &e.CreatedAt, &e.KeywordLength)
	if err == sql.ErrNoRows {
		notFound(w)
		return
	}
	e.Html = htmlify(w, r, e.Description)
	e.Stars = loadStars(e.Keyword)

	re.HTML(w, http.StatusOK, "keyword", struct {
		Context context.Context
		Entry   Entry
	}{
		r.Context(), e,
	})
}

func keywordByKeywordDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if err := setName(w, r); err != nil {
		forbidden(w)
		return
	}
	if err := authenticate(w, r); err != nil {
		forbidden(w)
		return
	}

	keyword := mux.Vars(r)["keyword"]
	if keyword == "" {
		badRequest(w)
		return
	}
	if r.FormValue("delete") == "" {
		badRequest(w)
		return
	}
	row := db.QueryRow(`SELECT * FROM entry WHERE keyword = ?`, keyword)
	e := Entry{}
	err := row.Scan(&e.ID, &e.AuthorID, &e.Keyword, &e.Description, &e.UpdatedAt, &e.CreatedAt, &e.KeywordLength)
	if err == sql.ErrNoRows {
		notFound(w)
		return
	}
	_, err = db.Exec(`DELETE FROM entry WHERE keyword = ?`, keyword)
	panicIf(err)
	http.Redirect(w, r, "/", http.StatusFound)
}

func isSpamContents(content string) bool {
	v := url.Values{}
	v.Set("content", content)
	resp, err := http.PostForm(isupamEndpoint, v)
	panicIf(err)
	defer resp.Body.Close()

	var data struct {
		Valid bool `json:valid`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	panicIf(err)
	return !data.Valid
}

func getContext(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

func setContext(r *http.Request, key, val interface{}) {
	if val == nil {
		return
	}

	r2 := r.WithContext(context.WithValue(r.Context(), key, val))
	*r = *r2
}

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, _ := store.Get(r, sessionName)
	return session
}

func main() {
	host := os.Getenv("ISUDA_DB_HOST")
	if host == "" {
		host = "localhost"
	}
	portstr := os.Getenv("ISUDA_DB_PORT")
	if portstr == "" {
		portstr = "3306"
	}
	port, err := strconv.Atoi(portstr)
	if err != nil {
		log.Fatalf("Failed to read DB port number from an environment variable ISUDA_DB_PORT.\nError: %s", err.Error())
	}
	user := os.Getenv("ISUDA_DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("ISUDA_DB_PASSWORD")
	dbname := os.Getenv("ISUDA_DB_NAME")
	if dbname == "" {
		dbname = "isuda"
	}

	db, err = sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?loc=Local&parseTime=true",
		user, password, host, port, dbname,
	))
	if err != nil {
		log.Fatalf("Failed to connect to DB: %s.", err.Error())
	}
	db.Exec("SET SESSION sql_mode='TRADITIONAL,NO_AUTO_VALUE_ON_ZERO,ONLY_FULL_GROUP_BY'")
	db.Exec("SET NAMES utf8mb4")

	isutarEndpoint = os.Getenv("ISUTAR_ORIGIN")
	if isutarEndpoint == "" {
		isutarEndpoint = "http://localhost:5001"
	}
	isupamEndpoint = os.Getenv("ISUPAM_ORIGIN")
	if isupamEndpoint == "" {
		isupamEndpoint = "http://localhost:5050"
	}

	store = sessions.NewCookieStore([]byte(sessionSecret))

	re = render.New(render.Options{
		Directory: "views",
		Funcs: []template.FuncMap{
			{
				"url_for": func(path string) string {
					return baseUrl.String() + path
				},
				"title": func(s string) string {
					return strings.Title(s)
				},
				"raw": func(text string) template.HTML {
					return template.HTML(text)
				},
				"add": func(a, b int) int { return a + b },
				"sub": func(a, b int) int { return a - b },
				"entry_with_ctx": func(entry Entry, ctx context.Context) *EntryWithCtx {
					return &EntryWithCtx{Context: ctx, Entry: entry}
				},
			},
		},
	})

	r := mux.NewRouter()
	r.UseEncodedPath()
	r.HandleFunc("/", myHandler(topHandler))
	r.HandleFunc("/initialize", myHandler(initializeHandler)).Methods("GET")
	r.HandleFunc("/robots.txt", myHandler(robotsHandler))
	r.HandleFunc("/keyword", myHandler(keywordPostHandler)).Methods("POST")

	l := r.PathPrefix("/login").Subrouter()
	l.Methods("GET").HandlerFunc(myHandler(loginHandler))
	l.Methods("POST").HandlerFunc(myHandler(loginPostHandler))
	r.HandleFunc("/logout", myHandler(logoutHandler))

	g := r.PathPrefix("/register").Subrouter()
	g.Methods("GET").HandlerFunc(myHandler(registerHandler))
	g.Methods("POST").HandlerFunc(myHandler(registerPostHandler))

	k := r.PathPrefix("/keyword/{keyword}").Subrouter()
	k.Methods("GET").HandlerFunc(myHandler(keywordByKeywordHandler))
	k.Methods("POST").HandlerFunc(myHandler(keywordByKeywordDeleteHandler))

	s := r.PathPrefix("/stars").Subrouter()
	s.Methods("GET").HandlerFunc(myHandler(starsHandler))
	s.Methods("POST").HandlerFunc(myHandler(starsPostHandler))

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))
	log.Fatal(http.ListenAndServe(":5000", r))
}
