package stats

import (
	"fmt"
	"sort"
	"strings"
)

type Map map[string]*Value

func (m Map) String() string {
	return m.ReversedValues().String()
}

func (m Map) Sum() (sum int) {
	for _, v := range m {
		sum += v.Value
	}
	return sum

}

func (m Map) ReversedValues() Values {
	v := m.Values()
	sort.Sort(v)
	return v
}

func (m Map) SortedValues() Values {
	v := m.Values()
	sort.Sort(v)
	return v
}

func (m Map) Values() Values {
	values := make(Values, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func (m Map) Inc(key string) int {
	return m.IncBy(key, 1)
}

func (m Map) IncBy(key string, value int) int {
	if m[key] == nil {
		m[key] = &Value{Key: key}
	}
	m[key].Value += value
	return m[key].Value
}

type Values []*Value

func (v Values) String() string {
	out := []string{}
	for _, k := range v {
		out = append(out, fmt.Sprintf("%s: %d", k.Key, k.Value))
	}
	return strings.Join(out, ",")
}

func (list Values) TopN(n int) Values {
	out := make(Values, 0, n)
	for _, v := range list {
		out = append(out, v)
		if len(out) >= n {
			break
		}
	}
	return out
}

func (list Values) Len() int {
	return len(list)
}

func (list Values) Swap(a, b int) {
	list[a], list[b] = list[b], list[a]
}

func (list Values) Less(a, b int) bool {
	diff := list[a].Value - list[b].Value
	switch {
	case diff < 0:
		return false
	case diff > 0:
		return true
	default:
		return list[a].Key < list[b].Key
	}
}

type Value struct {
	Key   string
	Value int
}
