package main

import (
	"time"
)

type Message struct {
	ID        int64     `db:"id"`
	ChannelID int64     `db:"channel_id"`
	UserID    int64     `db:"user_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
}

func addMessage(channelID, userID int64, content string) (int64, error) {
	res, err := db.Exec(
		"INSERT INTO message (channel_id, user_id, content, created_at) VALUES (?, ?, ?, NOW())",
		channelID, userID, content)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func queryMessages(chanID, lastID int64) ([]Message, error) {
	msgs := []Message{}
	err := db.Select(&msgs, "SELECT * FROM message WHERE id > ? AND channel_id = ? ORDER BY id DESC LIMIT 100",
		lastID, chanID)
	return msgs, err
}

func jsonifyMessage(m Message) (map[string]interface{}, error) {
	u := User{}
	err := db.Get(&u, "SELECT name, display_name, avatar_icon FROM user WHERE id = ?",
		m.UserID)
	if err != nil {
		return nil, err
	}

	r := make(map[string]interface{})
	r["id"] = m.ID
	r["user"] = u
	r["date"] = m.CreatedAt.Format("2006/01/02 15:04:05")
	r["content"] = m.Content
	return r, nil
}
