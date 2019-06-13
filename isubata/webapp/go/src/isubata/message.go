package main

import (
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"time"
)

type Message struct {
	ID        int64     `db:"id"`
	ChannelID int64     `db:"channel_id"`
	UserID    int64     `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

// messageをゲット
func queryMessages(chanID, lastID int64) ([]Message, error) {
	msgs := []Message{}
	err := db.Select(&msgs, "SELECT * FROM message WHERE id > ? AND channel_id = ? ORDER BY id DESC LIMIT 100",
		lastID, chanID)
	return msgs, err
}

// Messageをuserといっしょに取ってくる
func queryMessagesWithUser(chanID, lastID int64) ([]Message, []User, error) {
	msgs := []Message{}
	users := []User{}

	rows, err := db.Query(
		"SELECT message.id, channel_id, user_id, content, message.created_at,"+
			" name, display_name, avatar_icon FROM message"+
			" INNER JOIN user ON message.user_id = user.id"+
			" WHERE message.id > ? AND channel_id = ? ORDER BY message.id DESC LIMIT 100",
		lastID, chanID)
	defer rows.Close()
	for rows.Next() {
		var m Message
		var u User
		if err := rows.Scan(
			&m.ID, &m.ChannelID, &m.UserID, &m.Content, &m.CreatedAt,
			&u.Name, &u.DisplayName, &u.AvatarIcon,
		); err != nil {
			return msgs, users, err
		}
		msgs = append(msgs, m)
		users = append(users, u)
	}
	return msgs, users, err
}

func addMessage(channelID, userID int64, content string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	res, err := tx.Exec(
		"INSERT INTO message (channel_id, user_id, content, created_at) VALUES (?, ?, ?, NOW())",
		channelID, userID, content)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	numMessages, found := GetNumMessagesFromCache(channelID)
	if !found {
		numMessages = 0
	}
	fmt.Println("UPDATE is REQUESTED!!")
	SetNumMessagesToCache(channelID, numMessages+1)
	//// channelにメッセージの総数を追加
	//if _, err = tx.Exec(
	//	"UPDATE channel SET num_messages = num_messages + 1 WHERE id = ?",
	//	channelID); err != nil {
	//	tx.Rollback()
	//	return 0, err
	//}
	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func getMessage(c echo.Context) error {
	userID := sessUserID(c)
	if userID == 0 {
		return c.NoContent(http.StatusForbidden)
	}

	chanID, err := strconv.ParseInt(c.QueryParam("channel_id"), 10, 64)
	if err != nil {
		return err
	}
	lastID, err := strconv.ParseInt(c.QueryParam("last_message_id"), 10, 64)
	if err != nil {
		return err
	}

	//messages, err := queryMessages(chanID, lastID)
	messages, users, err := queryMessagesWithUser(chanID, lastID)
	if err != nil {
		return err
	}

	response := make([]map[string]interface{}, 0)
	for i := len(messages) - 1; i >= 0; i-- {
		m := messages[i]
		u := users[i]
		r, err := jsonifyMessageWithUser(m, u)
		if err != nil {
			return err
		}
		response = append(response, r)
	}

	if len(messages) > 0 {
		_, err := db.Exec(
			"INSERT INTO haveread (user_id, channel_id, message_id, updated_at, created_at)"+
				" VALUES (?, ?, ?, NOW(), NOW())"+
				" ON DUPLICATE KEY UPDATE message_id = ?, updated_at = NOW()",
			userID, chanID, messages[0].ID, messages[0].ID)
		if err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, response)
}

func postMessage(c echo.Context) error {
	user, err := ensureLogin(c)
	if user == nil {
		return err
	}

	message := c.FormValue("message")
	if message == "" {
		return echo.ErrForbidden
	}

	var chanID int64
	if x, err := strconv.Atoi(c.FormValue("channel_id")); err != nil {
		return echo.ErrForbidden
	} else {
		chanID = int64(x)
	}

	if _, err := addMessage(chanID, user.ID, message); err != nil {
		return err
	}

	return c.NoContent(204)
}
