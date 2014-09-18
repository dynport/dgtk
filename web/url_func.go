package web

import (
	"fmt"
	"strings"
)

type UrlFunc func(...interface{}) string

func parseUrlFunc(name string) func(i ...interface{}) string {
	parts := strings.Split(name, "/")
	variables := 0
	funcs := []func([]interface{}) string{}
	i := 0
	for _, p := range parts {
		if strings.HasPrefix(p, ":") {
			idx := variables
			funcs = append(funcs, func(atts []interface{}) string { return fmt.Sprint(atts[idx]) })
			variables++
		} else {
			idx := i
			funcs = append(funcs, func([]interface{}) string { return parts[idx] })
		}
		i++
	}

	return func(i ...interface{}) string {
		if variables != len(i) {
			panic(fmt.Sprintf("wrong number of attributes. expected %d, got %d", variables, len(i)))

		}
		out := []string{}
		for _, f := range funcs {
			out = append(out, f(i))
		}
		return strings.Join(out, "/")
	}
}
