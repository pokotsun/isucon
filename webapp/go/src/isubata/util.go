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
	rows, err := db.Queryx("SELECT m.id, m.content, m.created_at, u.name, u.display_name, u.avatar_icon FROM message m INNER JOIN user u on u.id = m.user_id WHERE id > ? AND channel_id = ? ORDER BY id DESC LIMIT 100", lastID, chanID)
	for rows.Next() {
		var m Message
		var u User
		err = rows.Scan(&m.ID, &m.Content, &m.CreatedAt, &u.Name, &u.DisplayName, &u.AvatarIcon)
		if err != nil {
			return response, err
		}
		r := make(map[string]interface{})
		r["id"] = m.ID
		r["user"] = u
		r["date"] = m.CreatedAt.Format("2006/01/02 15:04:05")
		r["content"] = m.Content
		response = append(response, r)
	}
	return response, nil
}
