package cli2

import (
	"fmt"
	"reflect"
	"strconv"
)

type option struct {
	field    string
	isFlag   bool
	desc     string
	short    string
	long     string
	required bool
	value    string
}

// Reflect the gathered information into the concrete action instance.
func (o *option) reflectTo(value reflect.Value) (e error) {
	if o.value == "" {
		if o.required {
			return fmt.Errorf("option %q is required but not set", o.field)
		}
		return nil
	}

	field := value.FieldByName(o.field)

	switch field.Kind() {
	case reflect.String:
		field.SetString(o.value)
	case reflect.Int:
		i, e := strconv.Atoi(o.value)
		if e != nil {
			return e
		}
		field.SetInt(int64(i))
	case reflect.Bool:
		field.SetBool(o.value == "true")
	default:
		return fmt.Errorf("invalid type %q", field.Type().String())
	}
	return nil
}

func (o *option) description() string {
	desc := "    "
	desc += o.shortDescription(" ")
	desc += fmt.Sprintf("%-*s", 30-len(desc), " ") + o.desc
	return desc
}

func (o *option) shortDescription(sep string) (desc string) {
	if o.short != "" {
		desc += "-" + o.short
	}
	if o.long != "" {
		if o.short != "" {
			desc += sep
		}
		desc += "--" + o.long
	}

	if !o.isFlag {
		desc += " <" + o.field + ">"
	}

	return desc
}

func (a *action) createOption(field reflect.StructField, value reflect.Value, tagMap map[string]string) (e error) {
	if e := validateTagMap(tagMap, "type", "desc", "short", "long", "required", "default"); e != nil {
		return fmt.Errorf("[option:%s] %s", field.Name, e.Error())
	}
	opt := &option{field: field.Name}

	if field.Type.Kind() == reflect.Bool {
		opt.isFlag = true
	}

	opt.short, e = handleShortIdentifier(tagMap)
	if e != nil {
		return e
	}

	opt.long, e = handleLongIdentifier(tagMap)
	if e != nil {
		return e
	}

	opt.required, e = handleRequired(tagMap)
	if e != nil {
		return e
	}

	if opt.required && opt.isFlag {
		return fmt.Errorf(`field %q is a flag and required, that doesn't make much sense`, field.Name)
	}

	opt.value, e = handleDefault(field, tagMap)
	if e != nil {
		return fmt.Errorf(`wrong value for "default" tag: %s`, e.Error())
	}

	opt.desc = handleDescription(tagMap)

	if opt.short == "" && opt.long == "" {
		return fmt.Errorf("option %q has neither long nor short accessor set", field.Name)
	}
	if opt.short != "" {
		a.params[opt.short] = opt
	}
	if opt.long != "" {
		a.params[opt.long] = opt
	}
	a.opts = append(a.opts, opt)
	return nil
}
