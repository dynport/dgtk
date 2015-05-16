package sequel

import (
	"database/sql"
	"errors"
)

type IterateOpt struct {
	ExtraSQL string
	Args     []interface{}
}

func NewIterateOpt(q string, args ...interface{}) *IterateOpt {
	return &IterateOpt{ExtraSQL: q, Args: args}
}

func IterateQuery(con Con, q string, opt *IterateOpt) (*sql.Rows, error) {
	args := []interface{}{}
	if opt != nil {
		if opt.ExtraSQL == "" {
			return nil, errors.New("ExtraSQL must not be blank")
		}
		q += " " + opt.ExtraSQL
		args = append(args, opt.Args...)
	}
	rows, err := con.Query(q, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
