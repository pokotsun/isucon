package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

func loginRequired(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, err := getLoginUser(c); err != nil {
			return resError(c, "login_required", 401)
		}
		return next(c)
	}
}

type Renderer struct {
	templates *template.Template
}

func (r *Renderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}

var db *sql.DB

func main() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_DATABASE"),
	)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()
	funcs := template.FuncMap{
		"encode_json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	}
	e.Renderer = &Renderer{
		templates: template.Must(template.New("").Delims("[[", "]]").Funcs(funcs).ParseGlob("views/*.tmpl")),
	}
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{Output: os.Stderr}))
	e.Static("/", "public")
	e.GET("/", func(c echo.Context) error {
		events, err := getEvents(false)
		if err != nil {
			return err
		}
		for i, v := range events {
			events[i] = sanitizeEvent(v)
		}
		return c.Render(200, "index.tmpl", echo.Map{
			"events": events,
			"user":   c.Get("user"),
			"origin": c.Scheme() + "://" + c.Request().Host,
		})
	}, fillinUser)
	e.GET("/initialize", func(c echo.Context) error {
		cmd := exec.Command("../../db/init.sh")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			return nil
		}

		return c.NoContent(204)
	})
	e.POST("/api/users", func(c echo.Context) error {
		var params struct {
			Nickname  string `json:"nickname"`
			LoginName string `json:"login_name"`
			Password  string `json:"password"`
		}
		c.Bind(&params)

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		var user User
		if err := tx.QueryRow("SELECT * FROM users WHERE login_name = ?", params.LoginName).Scan(&user.ID, &user.LoginName, &user.Nickname, &user.PassHash); err != sql.ErrNoRows {
			tx.Rollback()
			if err == nil {
				return resError(c, "duplicated", 409)
			}
			return err
		}

		res, err := tx.Exec("INSERT INTO users (login_name, pass_hash, nickname) VALUES (?, SHA2(?, 256), ?)", params.LoginName, params.Password, params.Nickname)
		if err != nil {
			tx.Rollback()
			return resError(c, "", 0)
		}
		userID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return resError(c, "", 0)
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		return c.JSON(201, echo.Map{
			"id":       userID,
			"nickname": params.Nickname,
		})
	})
	e.GET("/api/users/:id", func(c echo.Context) error {
		var user User
		if err := db.QueryRow("SELECT id, nickname FROM users WHERE id = ?", c.Param("id")).Scan(&user.ID, &user.Nickname); err != nil {
			return err
		}

		loginUser, err := getLoginUser(c)
		if err != nil {
			return err
		}
		if user.ID != loginUser.ID {
			return resError(c, "forbidden", 403)
		}

		rows, err := db.Query("SELECT * FROM reservations r INNER JOIN events e ON e.id = r.event_id WHERE r.user_id = ? ORDER BY IFNULL(r.canceled_at, r.reserved_at) DESC", user.ID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var recentReservations []Reservation
		var recentEvents []*Event
		var eventIDs []int64
		var totalPrice int

		for rows.Next() {
			var reservation Reservation
			var event Event
			if err := rows.Scan(&reservation.ID, &reservation.EventID, &reservation.SheetID, &reservation.UserID, &reservation.ReservedAt, &reservation.CanceledAt, &event.ID, &event.Title, &event.PublicFg, &event.ClosedFg, &event.Price); err != nil {
				return err
			}
			if reservation.CanceledAt != nil {
				reservation.CanceledAtUnix = reservation.CanceledAt.Unix()
				totalPrice += int(event.Price)
			}

			if len(recentReservations) < 5 {
				sheet, ok := getSheetByID(reservation.SheetID)
				if ok < 0 {
					return errors.New("not found")
				}

				price := sheet.Price + event.Price

				reservation.Event = &event
				reservation.SheetRank = sheet.Rank
				reservation.SheetNum = sheet.Num
				reservation.Price = price
				reservation.ReservedAtUnix = reservation.ReservedAt.Unix()
				recentReservations = append(recentReservations, reservation)
			}

			if len(recentEvents) < 5 && !contains(eventIDs, event.ID) {
				e, err := getEventWithoutDetail(event, -1)
				if err != nil {
					return err
				}
				recentEvents = append(recentEvents, e)
				eventIDs = append(eventIDs, event.ID)
			}
		}

		if recentReservations == nil {
			recentReservations = make([]Reservation, 0)
		}
		if recentEvents == nil {
			recentEvents = make([]*Event, 0)
		}

		return c.JSON(200, echo.Map{
			"id":                  user.ID,
			"nickname":            user.Nickname,
			"recent_reservations": recentReservations,
			"total_price":         totalPrice,
			"recent_events":       recentEvents,
		})
	}, loginRequired)
	e.POST("/api/actions/login", func(c echo.Context) error {
		var params struct {
			LoginName string `json:"login_name"`
			Password  string `json:"password"`
		}
		c.Bind(&params)

		user := new(User)
		if err := db.QueryRow("SELECT * FROM users WHERE login_name = ?", params.LoginName).Scan(&user.ID, &user.LoginName, &user.Nickname, &user.PassHash); err != nil {
			if err == sql.ErrNoRows {
				return resError(c, "authentication_failed", 401)
			}
			return err
		}

		var passHash string
		if err := db.QueryRow("SELECT SHA2(?, 256)", params.Password).Scan(&passHash); err != nil {
			return err
		}
		if user.PassHash != passHash {
			return resError(c, "authentication_failed", 401)
		}

		sessSetUserID(c, user.ID)
		user, err = getLoginUser(c)
		if err != nil {
			return err
		}
		return c.JSON(200, user)
	})
	e.POST("/api/actions/logout", func(c echo.Context) error {
		sessDeleteUserID(c)
		return c.NoContent(204)
	}, loginRequired)
	e.GET("/api/events", func(c echo.Context) error {
		events, err := getEvents(true)
		if err != nil {
			return err
		}
		for i, v := range events {
			events[i] = sanitizeEvent(v)
		}
		return c.JSON(200, events)
	})
	e.GET("/api/events/:id", func(c echo.Context) error {
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return resError(c, "not_found", 404)
		}

		loginUserID := int64(-1)
		if user, err := getLoginUser(c); err == nil {
			loginUserID = user.ID
		}

		event, err := getEvent(eventID, loginUserID)
		if err != nil {
			if err == sql.ErrNoRows {
				return resError(c, "not_found", 404)
			}
			return err
		} else if !event.PublicFg {
			return resError(c, "not_found", 404)
		}
		return c.JSON(200, sanitizeEvent(event))
	})
	e.POST("/api/events/:id/actions/reserve", func(c echo.Context) error {
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return resError(c, "not_found", 404)
		}
		var params struct {
			Rank string `json:"sheet_rank"`
		}
		c.Bind(&params)

		user, err := getLoginUser(c)
		if err != nil {
			return err
		}

		event, err := getEvent(eventID, user.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return resError(c, "invalid_event", 404)
			}
			return err
		} else if !event.PublicFg {
			return resError(c, "invalid_event", 404)
		}

		if !validateRank(params.Rank) {
			return resError(c, "invalid_rank", 400)
		}

		var sheets []Sheet
		var reservationID int64

		rows, err := db.Query("SELECT * FROM sheets WHERE id NOT IN (SELECT sheet_id FROM reservations WHERE event_id = ? AND canceled_at IS NULL FOR UPDATE) AND `rank` = ?", event.ID, params.Rank)
		if err == sql.ErrNoRows {
			return resError(c, "sold_out", 409)
		}
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var sheet Sheet
			if err := rows.Scan(&sheet.ID, &sheet.Rank, &sheet.Num, &sheet.Price); err != nil {
				return err
			}
			sheets = append(sheets, sheet)
		}

		var sheet Sheet
		for {
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			rand.Seed(time.Now().UnixNano())
			sheet = sheets[rand.Intn(len(sheets))]
			res, err := tx.Exec("INSERT INTO reservations (event_id, sheet_id, user_id, reserved_at) VALUES (?, ?, ?, ?)", event.ID, sheet.ID, user.ID, time.Now().UTC().Format("2006-01-02 15:04:05.000000"))
			if err != nil {
				tx.Rollback()
				log.Println("re-try: rollback by", err)
				continue
			}
			reservationID, err = res.LastInsertId()
			if err != nil {
				tx.Rollback()
				log.Println("re-try: rollback by", err)
				continue
			}
			if err := tx.Commit(); err != nil {
				tx.Rollback()
				log.Println("re-try: rollback by", err)
				continue
			}

			break
		}
		return c.JSON(202, echo.Map{
			"id":         reservationID,
			"sheet_rank": params.Rank,
			"sheet_num":  sheet.Num,
		})
	}, loginRequired)
	e.DELETE("/api/events/:id/sheets/:rank/:num/reservation", func(c echo.Context) error {
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return resError(c, "not_found", 404)
		}
		rank := c.Param("rank")
		num := c.Param("num")

		user, err := getLoginUser(c)
		if err != nil {
			return err
		}

		event, err := getEvent(eventID, user.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return resError(c, "invalid_event", 404)
			}
			return err
		} else if !event.PublicFg {
			return resError(c, "invalid_event", 404)
		}

		if !validateRank(rank) {
			return resError(c, "invalid_rank", 404)
		}

		num64, _ := strconv.ParseInt(num, 10, 64)
		sheet, ok := getSheetByNumAndRank(num64, rank)
		if ok < 0 {
			return resError(c, "invalid_sheet", 404)
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		var reservation Reservation
		if err := tx.QueryRow("SELECT * FROM reservations WHERE event_id = ? AND sheet_id = ? AND canceled_at IS NULL GROUP BY event_id HAVING reserved_at = MIN(reserved_at) FOR UPDATE", event.ID, sheet.ID).Scan(&reservation.ID, &reservation.EventID, &reservation.SheetID, &reservation.UserID, &reservation.ReservedAt, &reservation.CanceledAt); err != nil {
			tx.Rollback()
			if err == sql.ErrNoRows {
				return resError(c, "not_reserved", 400)
			}
			return err
		}
		if reservation.UserID != user.ID {
			tx.Rollback()
			return resError(c, "not_permitted", 403)
		}

		if _, err := tx.Exec("UPDATE reservations SET canceled_at = ? WHERE id = ?", time.Now().UTC().Format("2006-01-02 15:04:05.000000"), reservation.ID); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return c.NoContent(204)
	}, loginRequired)
	e.GET("/admin/", func(c echo.Context) error {
		var events []*Event
		administrator := c.Get("administrator")
		if administrator != nil {
			var err error
			if events, err = getEvents(true); err != nil {
				return err
			}
		}
		return c.Render(200, "admin.tmpl", echo.Map{
			"events":        events,
			"administrator": administrator,
			"origin":        c.Scheme() + "://" + c.Request().Host,
		})
	}, fillinAdministrator)
	e.POST("/admin/api/actions/login", func(c echo.Context) error {
		var params struct {
			LoginName string `json:"login_name"`
			Password  string `json:"password"`
		}
		c.Bind(&params)

		administrator := new(Administrator)
		if err := db.QueryRow("SELECT * FROM administrators WHERE login_name = ?", params.LoginName).Scan(&administrator.ID, &administrator.LoginName, &administrator.Nickname, &administrator.PassHash); err != nil {
			if err == sql.ErrNoRows {
				return resError(c, "authentication_failed", 401)
			}
			return err
		}

		var passHash string
		if err := db.QueryRow("SELECT SHA2(?, 256)", params.Password).Scan(&passHash); err != nil {
			return err
		}
		if administrator.PassHash != passHash {
			return resError(c, "authentication_failed", 401)
		}

		sessSetAdministratorID(c, administrator.ID)
		administrator, err = getLoginAdministrator(c)
		if err != nil {
			return err
		}
		return c.JSON(200, administrator)
	})
	e.POST("/admin/api/actions/logout", func(c echo.Context) error {
		sessDeleteAdministratorID(c)
		return c.NoContent(204)
	}, adminLoginRequired)
	e.GET("/admin/api/events", func(c echo.Context) error {
		events, err := getEvents(true)
		if err != nil {
			return err
		}
		return c.JSON(200, events)
	}, adminLoginRequired)
	e.POST("/admin/api/events", func(c echo.Context) error {
		var params struct {
			Title  string `json:"title"`
			Public bool   `json:"public"`
			Price  int    `json:"price"`
		}
		c.Bind(&params)

		tx, err := db.Begin()
		if err != nil {
			return err
		}

		res, err := tx.Exec("INSERT INTO events (title, public_fg, closed_fg, price) VALUES (?, ?, 0, ?)", params.Title, params.Public, params.Price)
		if err != nil {
			tx.Rollback()
			return err
		}
		eventID, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		event, err := getEvent(eventID, -1)
		if err != nil {
			return err
		}
		return c.JSON(200, event)
	}, adminLoginRequired)
	e.GET("/admin/api/events/:id", func(c echo.Context) error {
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return resError(c, "not_found", 404)
		}
		event, err := getEvent(eventID, -1)
		if err != nil {
			if err == sql.ErrNoRows {
				return resError(c, "not_found", 404)
			}
			return err
		}
		return c.JSON(200, event)
	}, adminLoginRequired)
	e.POST("/admin/api/events/:id/actions/edit", func(c echo.Context) error {
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return resError(c, "not_found", 404)
		}

		var params struct {
			Public bool `json:"public"`
			Closed bool `json:"closed"`
		}
		c.Bind(&params)
		if params.Closed {
			params.Public = false
		}

		event, err := getEvent(eventID, -1)
		if err != nil {
			if err == sql.ErrNoRows {
				return resError(c, "not_found", 404)
			}
			return err
		}

		if event.ClosedFg {
			return resError(c, "cannot_edit_closed_event", 400)
		} else if event.PublicFg && params.Closed {
			return resError(c, "cannot_close_public_event", 400)
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE events SET public_fg = ?, closed_fg = ? WHERE id = ?", params.Public, params.Closed, event.ID); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}

		e, err := getEvent(eventID, -1)
		if err != nil {
			return err
		}
		c.JSON(200, e)
		return nil
	}, adminLoginRequired)
	e.GET("/admin/api/reports/events/:id/sales", func(c echo.Context) error {
		eventID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			return resError(c, "not_found", 404)
		}

		event, err := getEvent(eventID, -1)
		if err != nil {
			return err
		}

		rows, err := db.Query(`SELECT r.*, e.price FROM reservations r INNER JOIN events e ON e.id = r.event_id WHERE r.event_id = ? ORDER BY reserved_at ASC FOR UPDATE`, event.ID)
		if err != nil {
			return err
		}
		defer rows.Close()

		var reports []Report
		for rows.Next() {
			var reservation Reservation
			var sheet *Sheet
			if err := rows.Scan(&reservation.ID, &reservation.EventID, &reservation.SheetID, &reservation.UserID, &reservation.ReservedAt, &reservation.CanceledAt, &event.Price); err != nil {
				return err
			}

			sheet, ok := getSheetByID(reservation.SheetID)
			if ok < 0 {
				return resError(c, "not_found", 404)
			}

			report := Report{
				ReservationID: reservation.ID,
				EventID:       event.ID,
				Rank:          sheet.Rank,
				Num:           sheet.Num,
				UserID:        reservation.UserID,
				SoldAt:        reservation.ReservedAt.Format("2006-01-02T15:04:05.000000Z"),
				Price:         event.Price + sheet.Price,
			}
			if reservation.CanceledAt != nil {
				report.CanceledAt = reservation.CanceledAt.Format("2006-01-02T15:04:05.000000Z")
			}
			reports = append(reports, report)
		}
		return renderReportCSV(c, reports)
	}, adminLoginRequired)
	e.GET("/admin/api/reports/sales", func(c echo.Context) error {
		rows, err := db.Query("SELECT r.*, e.id, e.price FROM reservations r INNER JOIN events e ON e.id = r.event_id ORDER BY reserved_at ASC FOR UPDATE")
		if err != nil {
			return err
		}
		defer rows.Close()

		var reports []Report
		for rows.Next() {
			var reservation Reservation
			var sheet *Sheet
			var event Event

			if err := rows.Scan(&reservation.ID, &reservation.EventID, &reservation.SheetID, &reservation.UserID, &reservation.ReservedAt, &reservation.CanceledAt, &event.ID, &event.Price); err != nil {
				return err
			}

			sheet, ok := getSheetByID(reservation.SheetID)
			if ok < 0 {
				return resError(c, "not_found", 404)
			}

			report := Report{
				ReservationID: reservation.ID,
				EventID:       event.ID,
				Rank:          sheet.Rank,
				Num:           sheet.Num,
				UserID:        reservation.UserID,
				SoldAt:        reservation.ReservedAt.Format("2006-01-02T15:04:05.000000Z"),
				Price:         event.Price + sheet.Price,
			}
			if reservation.CanceledAt != nil {
				report.CanceledAt = reservation.CanceledAt.Format("2006-01-02T15:04:05.000000Z")
			}
			reports = append(reports, report)
		}
		return renderReportCSV(c, reports)
	}, adminLoginRequired)

	e.Start(":8080")
}

type Report struct {
	ReservationID int64
	EventID       int64
	Rank          string
	Num           int64
	UserID        int64
	SoldAt        string
	CanceledAt    string
	Price         int64
}

func renderReportCSV(c echo.Context, reports []Report) error {
	sort.Slice(reports, func(i, j int) bool { return strings.Compare(reports[i].SoldAt, reports[j].SoldAt) < 0 })

	body := bytes.NewBufferString("reservation_id,event_id,rank,num,price,user_id,sold_at,canceled_at\n")
	for _, v := range reports {
		body.WriteString(fmt.Sprintf("%d,%d,%s,%d,%d,%d,%s,%s\n",
			v.ReservationID, v.EventID, v.Rank, v.Num, v.Price, v.UserID, v.SoldAt, v.CanceledAt))
	}

	c.Response().Header().Set("Content-Type", `text/csv; charset=UTF-8`)
	c.Response().Header().Set("Content-Disposition", `attachment; filename="report.csv"`)
	_, err := io.Copy(c.Response(), body)
	return err
}

func resError(c echo.Context, e string, status int) error {
	if e == "" {
		e = "unknown"
	}
	if status < 100 {
		status = 500
	}
	return c.JSON(status, map[string]string{"error": e})
}
