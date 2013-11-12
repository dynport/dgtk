package es

import (
	"reflect"
)

type Source map[string]interface{}

func (source *Source) Unmarshal(i interface{}) error {
	value := reflect.ValueOf(i)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	for k, v := range *source {
		field := value.FieldByName(k)
		if field != reflect.Zero(value.Type()) {
			newValue := reflect.ValueOf(v)
			if field.CanSet() {
				field.Set(newValue)
			}
		}
	}
	return nil
}
