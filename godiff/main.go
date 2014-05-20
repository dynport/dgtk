package godiff

import (
	"fmt"
	"reflect"
	"strings"
)

type DiffResult struct {
	Diff string
}

func Diff(a, b interface{}) *DiffResult {
	switch ta := a.(type) {
	case string:
		return diffString(ta, b)
	case bool:
		return diffBool(ta, b)
	case int:
		return diffInt(ta, b)
	case float64:
		return diffFloat(ta, b)
	default:
		if a == nil {
			return diffNil(b)
		}
		v := reflect.ValueOf(a)
		switch v.Kind() {
		case reflect.Map:
			return diffMap(v, b)
		case reflect.Slice:
			return diffSlice(v, b)
		default:
			return &DiffResult{Diff: fmt.Sprintf("type %T not supported yet", a)}
		}
	}
}

func diffNil(b interface{}) *DiffResult {
	if b == nil {
		return nil
	}
	return &DiffResult{Diff: fmt.Sprintf("expected nil, got %v", b)}

}

func diffSlice(a reflect.Value, b interface{}) *DiffResult {
	if a.Kind() != reflect.Slice {
		panic("at least a must be of kind Slice")
	}
	vb := reflect.ValueOf(b)
	if vb.Kind() != reflect.Slice {
		return &DiffResult{Diff: fmt.Sprintf("%v != %v", a.Interface(), b)}
	}
	if a.Len() != vb.Len() {
		return &DiffResult{Diff: fmt.Sprintf("%v != %v", a.Interface(), b)}
	}

	diff := []string{}
	for i := 0; i < a.Len(); i++ {
		value := a.Index(i)
		d := Diff(value.Interface(), vb.Index(i).Interface())
		if d != nil {
			diff = append(diff, fmt.Sprintf("%d: %s", i, d.Diff))
		}
	}
	if len(diff) > 0 {
		return &DiffResult{Diff: strings.Join(diff, ", ")}
	}
	return nil
}

func diffMap(a reflect.Value, b interface{}) *DiffResult {
	if a.Kind() != reflect.Map {
		panic("at least a must be of kind Map")
	}
	vb := reflect.ValueOf(b)
	if vb.Kind() != reflect.Map {
		return &DiffResult{Diff: fmt.Sprintf("%v != %v", a.Interface(), b)}
	}
	keysChecked := map[interface{}]struct{}{}
	keysA := a.MapKeys()
	keysB := vb.MapKeys()

	diff := []string{}

	for _, k := range keysA {
		value := a.MapIndex(k).Interface()
		keysChecked[k.Interface()] = struct{}{}
		valueB := vb.MapIndex(k)
		if valueB.Kind() == reflect.Invalid {
			diff = append(diff, fmt.Sprintf("%v (value:%v) in a but not in b", value, k.Interface()))

		} else {
			d := Diff(value, valueB.Interface())
			if d != nil {
				diff = append(diff, fmt.Sprintf("%v -> %s", k.Interface(), d.Diff))
			}
		}
	}

	for _, k := range keysB {
		_, exists := keysChecked[k.Interface()]
		if !exists {
			v := vb.MapIndex(k)
			diff = append(diff, fmt.Sprintf("%v (value:%v) in b but not in a", v.Interface(), k.Interface()))
		}
	}

	if len(diff) > 0 {
		return &DiffResult{Diff: strings.Join(diff, "\n")}
	}

	return nil
}

func diffFloat(a float64, b interface{}) *DiffResult {
	if tb, ok := b.(float64); ok {
		if a == tb {
			return nil
		}
	}
	return &DiffResult{Diff: fmt.Sprintf("%v != %v", a, b)}
}

func diffBool(a bool, b interface{}) *DiffResult {
	if tb, ok := b.(bool); ok {
		if a == tb {
			return nil
		}
	}
	return &DiffResult{Diff: fmt.Sprintf("%v != %v", a, b)}
}

func diffInt(a int, b interface{}) *DiffResult {
	if tb, ok := b.(int); ok {
		if a == tb {
			return nil
		}
	}
	return &DiffResult{Diff: fmt.Sprintf("%v != %v", a, b)}
}

func diffStringSlice(a []string, b interface{}) *DiffResult {
	if tb, ok := b.([]string); ok {
		if len(a) == len(tb) {
			diff := []string{}
			for i, v := range a {
				if v != tb[i] {
					diff = append(diff, fmt.Sprintf("%d: %v != %v", i, v, tb[i]))
				}
			}
			if len(diff) == 0 {
				return nil
			}
			return &DiffResult{Diff: strings.Join(diff, "\n")}
		}
	}
	return &DiffResult{Diff: fmt.Sprintf("%v != %v", a, b)}
}

func diffString(a string, b interface{}) *DiffResult {
	if tb, ok := b.(string); ok {
		if a == tb {
			return nil
		}
	}
	return &DiffResult{Diff: fmt.Sprintf("%v != %v", a, b)}
}
