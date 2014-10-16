package gosql

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Valuer interface {
	Values() Map
}
type JsonType string

type Map map[string]interface{}

type QueryOption struct {
	IfNotExists bool
}

func IfNotExists(o *QueryOption) {
	o.IfNotExists = true
}

func (m Map) CreateTableStatement(name string, opts ...func(*QueryOption)) (string, error) {
	opt := &QueryOption{}
	for _, o := range opts {
		o(opt)
	}
	cols, e := m.Columns()
	if e != nil {
		return "", e
	}
	names := []string{}
	for _, c := range cols {
		dbType := ""
		switch t := c.Value.(type) {
		case JsonType:
			dbType = "JSON"
		case string:
			dbType = "VARCHAR"
		case int:
			dbType = "INTEGER"
		case time.Time:
			dbType = "TIMESTAMP WITH TIME ZONE"
		default:
			_ = t
			return "", fmt.Errorf("unable to handle type %T", c.Value)
		}
		names = append(names, c.Name+" "+dbType)
	}
	prefix := "CREATE TABLE "
	if opt.IfNotExists {
		prefix += "IF NOT EXISTS "
	}
	return prefix + name + " (" + strings.Join(names, ", ") + ")", nil

}

func (m Map) InsertStatement(tableName string) (string, []interface{}, error) {
	cols, err := m.Columns()
	if err != nil {
		return "", nil, err
	}
	names := []string{}
	idxs := []string{}
	args := []interface{}{}
	for i, c := range cols {
		names = append(names, c.Name)
		idxs = append(idxs, fmt.Sprintf("$%d", i+1))
		args = append(args, c.Value)
	}
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(names, ", "), strings.Join(idxs, ", ")), args, nil
}

func (m Map) Columns(names ...string) (ColumnValues, error) {
	includer := func(name string) bool {
		return true
	}
	if len(names) > 0 {
		m := map[string]struct{}{}
		for _, n := range names {
			m[n] = struct{}{}
		}
		includer = func(name string) bool {
			_, ok := m[name]
			return ok
		}
	}
	out := ColumnValues{}
	for k, v := range m {
		if includer(k) {
			out = append(out, &ColumnValue{Name: k, Value: v})
		}
	}
	sort.Sort(out)
	return out, nil
}

type ColumnValues []*ColumnValue

func (list ColumnValues) Len() int {
	return len(list)
}

func (list ColumnValues) Swap(a, b int) {
	list[a], list[b] = list[b], list[a]
}

func (list ColumnValues) Less(a, b int) bool {
	return list[a].Name < list[b].Name
}

func (list ColumnValues) Names() []string {
	out := []string{}
	for _, c := range list {
		out = append(out, c.Name)
	}
	return out
}

func (list ColumnValues) Values() []interface{} {
	out := []interface{}{}
	for _, c := range list {
		out = append(out, c.Value)
	}
	return out
}

type ColumnValue struct {
	Name  string
	Value interface{}
}
