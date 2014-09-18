package gosql

import "sort"

type Map map[string]interface{}

func (m Map) Columns(names ...string) (Columns, error) {
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
	out := Columns{}
	for k, v := range m {
		if includer(k) {
			out = append(out, &Column{Name: k, Value: v})
		}
	}
	sort.Sort(out)
	return out, nil
}

type Columns []*Column

func (list Columns) Len() int {
	return len(list)
}

func (list Columns) Swap(a, b int) {
	list[a], list[b] = list[b], list[a]
}

func (list Columns) Less(a, b int) bool {
	return list[a].Name < list[b].Name
}

func (list Columns) Names() []string {
	out := []string{}
	for _, c := range list {
		out = append(out, c.Name)
	}
	return out
}

func (list Columns) Values() []interface{} {
	out := []interface{}{}
	for _, c := range list {
		out = append(out, c.Value)
	}
	return out
}

type Column struct {
	Name  string
	Value interface{}
}
