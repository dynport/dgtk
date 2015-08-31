package sequel

type QueryResult struct {
	Columns []string
	Rows    [][]interface{}
}

func QueryGeneric(con Con, q string, args ...interface{}) (*QueryResult, error) {
	rows, err := con.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	t := &QueryResult{Columns: cols}

	for rows.Next() {
		rowArgs := []interface{}{}
		ref := []*interface{}{}
		for n := 0; n < len(cols); n++ {
			var i interface{}
			rowArgs = append(rowArgs, &i)
			ref = append(ref, &i)
		}
		if err := rows.Scan(rowArgs...); err != nil {
			return nil, err
		}
		row := []interface{}{}
		for _, o := range ref {
			if o != nil {
				row = append(row, *o)
			} else {
				row = append(row, nil)
			}
		}
		t.Rows = append(t.Rows, row)
	}
	return t, rows.Err()
}
