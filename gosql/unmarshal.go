package gosql

import (
	"fmt"
	"reflect"
	"strings"
)

type Rows interface {
	Columns() ([]string, error)
	Scan(...interface{}) error
	Next() bool
	Err() error
}

func UnmarshalRows(rows Rows, i interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("expected slice type, got %s", v.Kind())
	}
	limit := v.Len()
	elementType := v.Type().Elem()
	if elementType.Kind() == reflect.Ptr {
		elementType = elementType.Elem()
	}
	all := reflect.MakeSlice(v.Type(), 0, limit)
	for rows.Next() {
		i := reflect.New(elementType)
		e := UnmarshalRow(rows, i.Interface())
		if e != nil {
			return e
		}
		all = reflect.Append(all, i)
	}
	v.Set(all)
	return rows.Err()
}

func UnmarshalRow(rows Rows, i interface{}) error {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	fields := map[string]reflect.Value{}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		name := v.Type().Field(i).Tag.Get("sql")
		if name == "" {
			name = v.Type().Field(i).Tag.Get("json")
		}
		if name != "" {
			parts := strings.Split(name, ",")
			if len(parts) > 1 {
				name = parts[0]
			}
			fields[name] = field
		}
	}
	columns, e := rows.Columns()
	if e != nil {
		return e
	}
	out := []interface{}{}
	for _, c := range columns {
		var i interface{}
		field, ok := fields[c]
		if ok {
			i = field.Addr().Interface()
		} else {
			i = &i
		}
		out = append(out, i)
	}
	return rows.Scan(out...)
}
