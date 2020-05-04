package main

import (
	"database/sql"
)

// sheetIDからsheet情報を取得
func getSheetFromID(id int64) Sheet {
	switch {
	case (0 < id && id <= 50):
		return Sheet{ ID: id, Rank:"S", Num: id, Price: 5000 }
	case (50 < id && id <=200):
		return Sheet{ ID: id, Rank:"A", Num: id-50, Price: 3000 }
	case (200 < id && id <= 500):
		return Sheet{ ID: id, Rank:"B", Num: id-200, Price: 1000 }
	case (500 < id && id <= 1000):
		return Sheet{ ID: id, Rank:"C", Num: id-500, Price: 0 }
	default:
		return Sheet{ ID: -1, Rank:"INVALID", Num:-1, Price:0}
	}
} 

// SheetIDをRankとNumから取得
func getSheetIdFromRankAndNum(rank string, num int64) (int64, error) {
	switch rank {
	case "S":
		if 0 < num && num <=50 {
			return num, nil
		}
	case "A":
		if 0 < num && num <= 150 {
			return num + 50, nil
		}
	case "B":
		if 0 < num && num <= 300 {
			return num + 200, nil
		}
	case "C":
		if 0 < num && num <= 500 {
			return num + 500, nil
		}
	}	
	return -1, sql.ErrNoRows
}