package cli

import (
	"fmt"
	"github.com/dynport/dgtk/tagparse"
	"log"
	"reflect"
	"strings"
)

type action struct {
	path        string             // Path used for the routing.
	params      map[string]*option // Mapping of flags and options (short and long) to according value.
	opts        []*option          // The options available for the action.
	args        []*argument        // List of arguments accepted.
	runner      Runner             // Who's connected to the action.
	description string             // Description of the action.
	value       reflect.Value
}

// Register an action for the given path with the given runner.
func newAction(path string, r Runner, desc string) (act *action, e error) {

	act = &action{
		path:        path,
		runner:      r,
		params:      map[string]*option{},
		description: desc}

	// Inject the "help" option (handled specially).
	helpOption := &option{field: "Help", short: "h", long: "help", isFlag: true, desc: "show help for action"}
	act.opts = append(act.opts, helpOption)
	act.params["h"] = helpOption
	act.params["help"] = helpOption

	if e := act.reflect(); e != nil {
		return nil, e
	}
	return act, nil
}

// Method to reflect on the action's runner type and determine the according options and arguments.
func (a *action) reflect() (e error) {
	v := reflect.ValueOf(a.runner)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	a.value = v
	e = a.reflectRecurse(v)
	if e != nil {
		e = fmt.Errorf("%s: %s", v.Type().Name(), e)
	}
	return e
}

func (a *action) reflectRecurse(value reflect.Value) (e error) {
	if !value.IsValid() {
		// ignore invalid stuff
		return nil
	}

	v := reflect.ValueOf(value.Interface())
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v = reflect.New(v.Type().Elem())
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		value := v.Field(i)

		if field.PkgPath != "" { // Unexported field have a pkg path set.
			continue // Ignore unexported fields.
		}

		if field.Anonymous {
			e = a.reflectRecurse(reflect.ValueOf(value.Interface()))
			if e != nil {
				return e
			}
			continue
		}

		e = a.handleField(field, value)
		if e != nil {
			return e
		}
	}
	return nil
}

func (a *action) handleField(field reflect.StructField, value reflect.Value) (e error) {
	tagMap, e := tagparse.Parse(field, "cli")
	if e != nil {
		return fmt.Errorf("failed to parse tag for field %q: %s", field.Name, e)
	}

	if len(tagMap) == 0 {
		return nil
	}

	switch tagMap["type"] {
	case "arg":
		if e = a.createArgument(field, value, tagMap); e != nil {
			return e
		}
	case "opt":
		if e = a.createOption(field, value, tagMap); e != nil {
			return e
		}
	default:
		if tagMap["type"] == "" {
			return fmt.Errorf("tag for field %q has no type set", field.Name)
		}
		return fmt.Errorf("tag for field %q has unknown type %q", field.Name, tagMap["type"])
	}
	return nil
}

func (a *action) parseArgs(params []string) (e error) {
	argIdx := 0
	for idx := 0; idx < len(params); idx++ {
		value := params[idx]
		switch {
		case strings.Contains(value, " "): // Must be an arg!
			if argIdx, e = a.handleArgs(value, argIdx); e != nil {
				return e
			}
		case strings.HasPrefix(value, "--"):
			idx, e = a.handleParams(value[2:], params, idx)
			if e != nil {
				return e
			}
		case strings.HasPrefix(value, "-"):
			idx, e = a.handleParams(value[1:], params, idx)
			if e != nil {
				return e
			}
		default:
			if argIdx, e = a.handleArgs(value, argIdx); e != nil {
				return e
			}
		}
	}
	return a.reflectIntoRunner()
}

func (a *action) handleArgs(value string, index int) (int, error) {
	if arg := a.argumentForPosition(index); arg != nil {
		arg.setValue(value)
		return index + 1, nil
	}
	return -1, fmt.Errorf("too many arguments given")
}

func (a *action) handleParams(paramName string, args []string, idx int) (int, error) {
	// Keep that on top, as this is some special sort of handling. Required to make help appear in usage description,
	// but not be injected to deep.
	if paramName == "h" || paramName == "help" {
		return -1, ErrorHelpRequested
	}

	option, found := a.params[paramName]
	if !found {
		return -1, fmt.Errorf("unknown parameter found: %q", paramName)
	}

	if option.isFlag {
		if option.value == "" || option.value == "false" {
			option.value = "true"
		}
	} else {
		if idx+1 > len(args) {
			return -1, fmt.Errorf("missing value for option %q!", option.field)
		}
		option.value = args[idx+1]
		idx += 1
	}
	return idx, nil
}

// Use reflection to set values of the runner, if the action was called with a matching route.
func (a *action) reflectIntoRunner() (e error) {
	if e = a.reflectOptions(); e != nil {
		return e
	}
	if e = a.reflectArguments(); e != nil {
		return e
	}
	return nil
}

func (a *action) reflectOptions() (e error) {
	for _, option := range a.opts {
		if e = option.reflectTo(a.value); e != nil {
			return e
		}
	}
	return nil
}

func (a *action) reflectArguments() (e error) {
	for _, arg := range a.args {
		if e = arg.reflectTo(a.value); e != nil {
			return e
		}
	}
	return nil
}

func (a *action) showHelp() {
	a.showShortHelp()
	if a.description != "" {
		log.Print("  ", a.description)
	}

	optsAvailable := false
	if len(a.opts) > 0 {
		optsAvailable = true
		log.Print("  OPTIONS")
		for _, opt := range a.opts {
			log.Print(opt.description())
		}
	}
	if len(a.args) > 0 {
		if optsAvailable {
			log.Println()
		}
		log.Print("  ARGUMENTS")
		for _, arg := range a.args {
			log.Print(arg.description())
		}
	}
	log.Println()
}

func (a *action) showShortHelp() {
	line := strings.Replace(a.path, "/", " ", -1) + " "
	for i := range a.opts {
		line += "[" + a.opts[i].shortDescription("|") + "] "
	}
	for _, arg := range a.args {
		line += arg.shortDescription()
		line += " "
	}
	log.Print(line)
}

func (a *action) showTabularHelp(t *table) {
	oDesc := make([]string, len(a.opts))
	aDesc := make([]string, len(a.args))
	for i := range a.opts {
		if a.opts[i].required {
			oDesc[i] = "[" + a.opts[i].shortDescription("|") + "]"
		}
	}
	for i := range a.args {
		aDesc[i] = a.args[i].shortDescription()
	}
	t.addRow(
		row{
			strings.Replace(a.path, "/", " ", -1),
			strings.Join(oDesc, " "),
			strings.Join(aDesc, " ")})
}
