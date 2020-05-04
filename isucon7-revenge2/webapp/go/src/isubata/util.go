package main

import ()

func tAdd(a, b int64) int64 {
	return a + b
}

func tRange(a, b int64) []int64 {
	r := make([]int64, b-a+1)
	for i := int64(0); i <= (b - a); i++ {
		r[i] = a + i
	}
	return r
}

func jsonifyMessageWith(chanID, lastID int64) ([]map[string]interface{}, error) {
	response := make([]map[string]interface{}, 0, 100)
	rows, err := db.Queryx("SELECT m.id, m.content, m.created_at, u.name, u.display_name, u.avatar_icon FROM message m INNER JOIN user u on u.id = m.user_id WHERE m.id > ? AND m.channel_id = ? ORDER BY m.id DESC LIMIT 100", lastID, chanID)
	if err != nil {
		return response, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Message
		var u User
		err = rows.Scan(&m.ID, &m.Content, &m.CreatedAt, &u.Name, &u.DisplayName, &u.AvatarIcon)
		if err != nil {
			return response, err
		}
		r := makeResponse(u, m)
		response = append(response, r)
	}
	return response, nil
}

func jsonifyMessageWithChannel(chID, limit, offset int64) ([]map[string]interface{}, error) {
	response := make([]map[string]interface{}, 0, 100)
	rows, err := db.Queryx("SELECT m.id, m.content, m.created_at, u.name, u.display_name, u.avatar_icon FROM message m INNER JOIN user u on u.id = m.user_id WHERE m.channel_id = ? ORDER BY m.id DESC LIMIT ? OFFSET ?", chID, limit, offset)
	if err != nil {
		return response, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Message
		var u User
		err = rows.Scan(&m.ID, &m.Content, &m.CreatedAt, &u.Name, &u.DisplayName, &u.AvatarIcon)
		if err != nil {
			return response, err
		}
		r := makeResponse(u, m)
		response = append(response, r)
	}
	return response, nil
}

func makeResponse(u User, m Message) map[string]interface{} {
	r := make(map[string]interface{})
	r["id"] = m.ID
	r["user"] = u
	r["date"] = m.CreatedAt.Format("2006/01/02 15:04:05")
	r["content"] = m.Content
	return r
}

func reverseJsonifiedMessage(mjson []map[string]interface{}) []map[string]interface{} {
	response := make([]map[string]interface{}, 0, 100)
	for i := len(mjson) - 1; i >= 0; i-- {
		response = append(response, mjson[i])
	}
	return response
}
