package main

import (
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"time"
)

type ChannelInfo struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	UpdatedAt   time.Time `db:"updated_at"`
	CreatedAt   time.Time `db:"created_at"`
}

func queryChannelsOrderById() ([]ChannelInfo, error) {
	channels := []ChannelInfo{}
	err := db.Select(&channels, "SELECT id, name, description, updated_at, created_at FROM channel ORDER BY id")
	return channels, err
}

func getChannel(c echo.Context) error {
	user, err := ensureLogin(c)
	if user == nil {
		return err
	}
	cID, err := strconv.Atoi(c.Param("channel_id"))
	if err != nil {
		return err
	}
	channels := []ChannelInfo{}
	err = db.Select(&channels, "SELECT id, name, description, updated_at, created_at FROM channel ORDER BY id")
	if err != nil {
		return err
	}

	var desc string
	for _, ch := range channels {
		if ch.ID == int64(cID) {
			desc = ch.Description
			break
		}
	}
	return c.Render(http.StatusOK, "channel", map[string]interface{}{
		"ChannelID":   cID,
		"Channels":    channels,
		"User":        user,
		"Description": desc,
	})
}
