package web

import (
	"bytes"
	"html/template"
	"reflect"
)

func clone(p interface{}) interface{} {
	value := reflect.ValueOf(p)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return reflect.New(value.Type()).Interface()
}

func render(b []byte, funcs template.FuncMap, i interface{}) ([]byte, error) {
	tpl, e := template.New(string(b)).Funcs(funcs).Parse(string(b))
	if e != nil {
		return nil, e
	}
	buf := &bytes.Buffer{}
	e = tpl.Execute(buf, i)
	return buf.Bytes(), e
}
