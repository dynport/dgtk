package gosql

import (
	"database/sql"
	"fmt"
)

func ScanMap(rows *sql.Rows, m map[string]interface{}) error {
	if m == nil {
		return fmt.Errorf("map must be initizlied")
	}
	vals := []*sqlScanner{}
	ints := []interface{}{}
	cols, e := rows.Columns()
	if e != nil {
		return e
	}
	for _ = range cols {
		s := &sqlScanner{}
		vals = append(vals, s)
		ints = append(ints, s)
	}
	e = rows.Scan(ints...)
	if e != nil {
		return e
	}
	for i, n := range cols {
		m[n] = vals[i].value
	}
	return nil
}

type sqlScanner struct {
	value interface{}
}

func (m *sqlScanner) Scan(i interface{}) error {
	switch v := i.(type) {
	case []uint8:
		m.value = string(v)
	default:
		m.value = i
	}
	return nil
}
