package gosql

import (
	"database/sql"
	"fmt"
)

func SelectStruct(db Dbi, q string, i interface{}) error {
	rows, e := db.Query(q)
	if e != nil {
		return e
	}
	defer rows.Close()
	ok := rows.Next()
	if !ok {
		return fmt.Errorf("expected to find 1 row, found 0")
	}
	return UnmarshalRow(rows, i)
}

func SelectStructs(db Dbi, q string, i interface{}) error {
	rows, e := db.Query(q)
	if e != nil {
		return e
	}
	defer rows.Close()
	return UnmarshalRows(rows, i)
}

func SelectInt(db Dbi, q string) (int, error) {
	var i int
	return i, db.QueryRow(q).Scan(&i)
}

func SelectString(db Dbi, q string) (string, error) {
	var v string
	return v, db.QueryRow(q).Scan(&v)
}

func SelectInts(db Dbi, q string) ([]int, error) {
	out := []int{}

	e := selectRaw(db, q, func(rows *sql.Rows) error {
		var i int
		e := rows.Scan(&i)
		if e != nil {
			return e
		}
		out = append(out, i)
		return nil
	})
	return out, e
}

func SelectStrings(db Dbi, q string) ([]string, error) {
	out := []string{}

	e := selectRaw(db, q, func(rows *sql.Rows) error {
		var v string
		e := rows.Scan(&v)
		if e != nil {
			return e
		}
		out = append(out, v)
		return nil
	})
	return out, e
}

func selectRaw(db Dbi, q string, c func(*sql.Rows) error) error {
	rows, e := db.Query(q)
	if e != nil {
		return e
	}
	defer rows.Close()
	for rows.Next() {
		e = c(rows)
		if e != nil {
			return e
		}
	}
	return rows.Err()
}
