package main

import(

)

// eventsの取得
func getEvents(all bool) ([]*Event, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Commit()

	// rows, err := tx.Query("SELECT * FROM events ORDER BY id ASC")
	rows, err := tx.Query("SELECT * FROM events")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Publicなものを消している
	var events []*Event
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.Title, &event.PublicFg, &event.ClosedFg, &event.Price); err != nil {
			return nil, err
		}
		// Privateかつallフラグが立っていない場合はスキップ
		if !all && !event.PublicFg {
			continue
		}
		assignedEvent, err := getEvent(&event, -1)
		if err != nil {
			return nil, err
		}
		for k := range assignedEvent.Sheets {
			assignedEvent.Sheets[k].Detail = nil
		}
		events = append(events, assignedEvent)
	}

	return events, nil
}

// Event情報だけをDBから取ってくる(予約情報とか席情報とか抜きで)
func getEventOnly(eventID int64) (*Event, error) {
	var event Event
	if err := db.QueryRow("SELECT * FROM events WHERE id = ?", 
		eventID).Scan(&event.ID, &event.Title, &event.PublicFg,
			 &event.ClosedFg, &event.Price); err != nil {
		return nil, err
	}
	return &event, nil
}

// 予約情報なども用意したevent情報を取ってくる
func getEvent(event *Event, loginUserID int64) (*Event, error) {
	event.Sheets = map[string]*Sheets {
		"S": &Sheets{},
		"A": &Sheets{},
		"B": &Sheets{},
		"C": &Sheets{},
	}

	// sheetの全件をセットしていく
	for i:= 1; i<=1000; i++ {
		sheet := getSheetFromID(int64(i))
		event.Sheets[sheet.Rank].Price = event.Price + sheet.Price
		event.Total++
		event.Sheets[sheet.Rank].Total++
		event.Sheets[sheet.Rank].Detail = append(event.Sheets[sheet.Rank].Detail, &sheet)
	}

	query := "SELECT * FROM reservations WHERE event_id = ? AND canceled_at IS NULL GROUP BY event_id, sheet_id HAVING reserved_at = MIN(reserved_at)"
	rows, err := db.Query(query, event.ID)
	if err != nil {
		return nil, err
	}
	event.Remains = 1000
	event.Sheets["S"].Remains = 50
	event.Sheets["A"].Remains = 150
	event.Sheets["B"].Remains = 300
	event.Sheets["C"].Remains = 500

	// 現状予約として有効なreservationだけ取得する
	for rows.Next() {
		var reservation Reservation
		if err := rows.Scan(&reservation.ID, &reservation.EventID,
			&reservation.SheetID, &reservation.UserID,
			&reservation.ReservedAt, &reservation.CanceledAt); err != nil {
				return nil, err
		}
		sheet := getSheetFromID(int64(reservation.SheetID))
		sheet.Mine = reservation.UserID == loginUserID
		sheet.Reserved = true
		sheet.ReservedAtUnix = reservation.ReservedAt.Unix()
		event.Remains--
		event.Sheets[sheet.Rank].Remains--
		event.Sheets[sheet.Rank].Detail[sheet.Num-1] = &sheet
	}
	return event, nil
}

// eventの取得(EventIDから)
func getEventByID(eventID, loginUserID int64) (*Event, error) {
	var event Event
	if err := db.QueryRow("SELECT * FROM events WHERE id = ?", 
		eventID).Scan(&event.ID, &event.Title, &event.PublicFg,
			 &event.ClosedFg, &event.Price); err != nil {
		return nil, err
	}

	return getEvent(&event, loginUserID)
}
