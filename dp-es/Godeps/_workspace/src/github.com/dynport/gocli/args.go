package gocli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	STRING  = "string"
	INTEGER = "int"
	BOOL    = "bool"
)

var (
	re = regexp.MustCompile("^([\\-]+.*)")
)

type FlagMap map[string]*Flag
type ArgumentMap map[string]*Argument

type Args struct {
	Args        []string
	Attributes  map[string][]string
	currentKey  string
	FlagMap     FlagMap
	ArgumentMap ArgumentMap
	Flags       []*Flag
}

func NewArgs(mapping FlagMap) *Args {
	a := &Args{}
	for key, flag := range mapping {
		flag.CliFlag = key
		a.RegisterFlag(flag)
	}
	return a
}

type Argument struct {
	Key      string
	Index    int
	Multiple bool
}

func (a *Args) KeyForCliKey(cliKey string) string {
	for _, flag := range a.Flags {
		if flag.CliFlag == cliKey {
			return flag.Key
		}
	}
	return ""
}

func (a *Args) RegisterArgs(args string) {
	if a.ArgumentMap == nil {
		a.ArgumentMap = ArgumentMap{}
	}
	for i, arg := range strings.Split(args, " ") {
		a.ArgumentMap[arg] = &Argument{Key: arg, Index: i}
	}
	return
}

func (a *Args) getWithDefaults(key string) []string {
	flag := a.FlagMap[key]
	if flag == nil {
		panic("no flag found for " + key)
	}
	values := a.Attributes[key]
	if len(values) == 0 {
		if flag.DefaultValue != "" {
			values = []string{flag.DefaultValue}
		}
	}
	return values
}

func (a *Args) AttributesMap() map[string]string {
	m := make(map[string]string)
	for k, flag := range a.FlagMap {
		key := flag.Key
		if key == "" {
			key = strings.TrimLeft(k, "-")
		}
		v := a.getWithDefaults(k)
		if len(v) == 1 {
			m[key] = v[0]
		} else if len(v) > 0 {
			m[key] = strings.Join(v, ",")
		}
	}
	return m
}

func (a *Args) RegisterFlag(flag *Flag) {
	if a.FlagMap == nil {
		a.FlagMap = make(map[string]*Flag)
	}
	a.FlagMap[flag.CliFlag] = flag
	a.Flags = append(a.Flags, flag)
}

func (a *Args) BuildStringFlag(key string, required bool, defaultValue, description string) *Flag {
	return &Flag{
		Type:         STRING,
		CliFlag:      key,
		Required:     required,
		DefaultValue: defaultValue,
		Description:  description,
	}
}

func (a *Args) RegisterString(cliKey string, key string, required bool, defaultValue, description string) {
	a.RegisterFlag(
		&Flag{
			Type:         STRING,
			Key:          key,
			CliFlag:      cliKey,
			Required:     required,
			DefaultValue: defaultValue,
			Description:  description,
		},
	)
}

func (a *Args) RegisterInt(cliKey string, key string, required bool, defaultValue int, description string) {
	a.RegisterFlag(
		&Flag{
			Type:         INTEGER,
			CliFlag:      cliKey,
			Key:          key,
			Required:     required,
			DefaultValue: strconv.Itoa(defaultValue),
			Description:  description,
		},
	)
}

func (a *Args) RegisterBool(cliKey string, key string, required bool, defaultValue bool, description string) {
	a.RegisterFlag(
		&Flag{
			Type:         BOOL,
			Key:          key,
			CliFlag:      cliKey,
			Required:     required,
			DefaultValue: strconv.FormatBool(required),
			Description:  description,
		},
	)
}

func (a *Args) Usage() string {
	table := NewTable()
	table.Separator = " "
	for _, flag := range a.Flags {
		table.AddStrings(flag.Usage())
	}
	return table.String()
}

func (a *Args) lookup(key string) (flags []*Flag) {
	for i := range a.Flags {
		arg := a.Flags[i]
		if arg.Matches(key) {
			flags = append(flags, arg)
		}
	}
	return flags
}

func (a *Args) Length() int {
	return len(a.Args)
}

func (a *Args) String(key string) {
	a.AddFlag(key, STRING)
}

func (a *Args) Bool(key string) {
	a.AddFlag(key, BOOL)
}

func (a *Args) AddFlag(key, value string) {
	a.RegisterFlag(&Flag{Type: value, CliFlag: key})
}

func (a *Args) AddAttribute(k, v string) {
	if a.Attributes == nil {
		a.Attributes = map[string][]string{}
	}
	a.Attributes[k] = append(a.Attributes[k], v)
}

func (a *Args) Parse(args []string) error {
	a.Args = make([]string, 0, 10)
	a.Attributes = make(map[string][]string)
	for _, arg := range args {
		if e := a.handleArg(arg); e != nil {
			return e
		}
	}
	return nil
}

func (a *Args) TypeOf(key string) (out string, e error) {
	flags := a.lookup(key)
	switch len(flags) {
	case 0:
		e = fmt.Errorf("no mapping defined for %s", key)
	case 1:
		out = flags[0].Type
	default:
		e = fmt.Errorf("mapping for %s not uniq", key)
	}
	return out, e
}

func (a *Args) handleArgFlag(flag string) error {
	if t, e := a.TypeOf(flag); e != nil {
		return e
	} else {
		switch t {
		case STRING, INTEGER:
			a.currentKey = flag
		case BOOL:
			a.AddAttribute(flag, "true")
		default:
			return fmt.Errorf("no mapping defined for %s", flag)
		}
	}
	return nil
}

func (a *Args) handleArg(arg string) error {
	if parts := re.FindStringSubmatch(arg); len(parts) == 2 {
		chunks := strings.Split(parts[1], "=")
		if len(chunks) == 2 {
			key, value := chunks[0], chunks[1]
			a.AddAttribute(key, value)
			return nil
		} else {
			if e := a.handleArgFlag(chunks[0]); e != nil {
				return e
			}
			return nil
		}
	} else if a.currentKey != "" {
		a.AddAttribute(a.currentKey, arg)
		a.currentKey = ""
		return nil
	}
	a.Args = append(a.Args, arg)
	return nil
}

func (a *Args) Get(key string) []string {
	return a.Attributes[key]
}

func (a *Args) GetInt(key string) (int, error) {
	s, e := a.GetString(key)
	if e != nil {
		return 0, e
	}
	return strconv.Atoi(s)
}

func (a *Args) MustGetInt(key string) int {
	i, e := a.GetInt(key)
	if e != nil {
		panic(e.Error())
	}
	return i
}

func (a *Args) MustGetString(key string) string {
	s, e := a.GetString(key)
	if e != nil {
		panic(e.Error())
	}
	return s
}

func (a *Args) GetString(key string) (string, error) {
	flag, ok := a.FlagMap[key]
	if !ok {
		return "", fmt.Errorf("no mapping defined for %s", key)
	}
	values := a.Get(key)
	if len(values) > 0 {
		return values[len(values)-1], nil
	}
	if flag.Required {
		return "", fmt.Errorf("flag %s is required", key)
	}
	return flag.DefaultValue, nil
}

func (a *Args) GetBool(key string) bool {
	args := a.Attributes[key]
	return (len(args) == 1 && args[0] == "true")
}
