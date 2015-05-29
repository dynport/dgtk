package gocli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Router struct {
	Actions   map[string]*Action
	Separator string
	Writer    io.Writer
}

func NewRouter(mapping map[string]*Action) *Router {
	router := &Router{}
	for path, action := range mapping {
		router.Register(path, action)
	}
	return router
}

func (cli *Router) Register(path string, action *Action) {
	if cli.Actions == nil {
		cli.Actions = make(map[string]*Action)
	}
	cli.Actions[path] = action
}

func (router *Router) matchKey(patterns []string, key string) bool {
	keyParts := strings.Split(key, "/")
	for i, pattern := range patterns {
		if i > (len(keyParts) - 1) {
			return false
		}
		if !strings.HasPrefix(keyParts[i], pattern) {
			return false
		}
	}
	return true
}

func (router *Router) Search(patterns []string) map[string]*Action {
	actions := make(map[string]*Action)
	for key, action := range router.Actions {
		if router.matchKey(patterns, key) {
			actions[key] = action
		}
	}
	return actions
}

func (cli *Router) Usage() string {
	keys := []string{}
	for key := range cli.Actions {
		keys = append(keys, key)
	}
	return cli.UsageForKeys(keys, "")
}

func (cli *Router) UsageForKeys(keys []string, pattern string) string {
	sort.Strings(keys)
	table := NewTable()
	if cli.Separator != "" {
		table.Separator = cli.Separator
	}
	maxParts := 0
	selected := []string{}
	for _, key := range keys {
		partsCount := len(strings.Split(key, "/"))
		if partsCount > maxParts {
			maxParts = partsCount
		}
		selected = append(selected, key)
	}
	for _, key := range selected {
		parts := strings.Split(key, "/")
		action := cli.Actions[key]

		// fill up parts
		for i := (maxParts - len(parts)); i > 0; i-- {
			parts = append(parts, "")
		}

		parts = append(parts, action.Usage, action.Description)
		table.AddStrings(parts)
		if action.Args != nil {
			usage := action.Args.Usage()
			if usage != "" {
				lines := strings.Split(usage, "\n")
				for _, line := range lines {
					usageParts := []string{}
					for j := 0; j < 3; j++ {
						usageParts = append(usageParts, "")
					}
					current := append(usageParts, line)
					table.AddStrings(current)
				}
			}
		}
	}
	out := []string{"USAGE"}
	out = append(out, table.String())
	return strings.Join(out, "\n")
}

func AddActionUsage(parts []string, table *Table, action *Action) {
	parts = append(parts, action.Usage, action.Description)
	table.AddStrings(parts)
	if action.Args != nil {
		usage := action.Args.Usage()
		if usage != "" {
			lines := strings.Split(usage, "\n")
			for _, line := range lines {
				usageParts := []string{}
				for j := 0; j < 3; j++ {
					usageParts = append(usageParts, "")
				}
				current := append(usageParts, line)
				table.AddStrings(current)
			}
		}
	}
}

func (router *Router) printActionUsage(parts []string, action *Action, message interface{}) {
	table := NewTable()
	fmt.Println("ERROR:", message)
	AddActionUsage(parts, table, action)
	router.Println(table.String())
	os.Exit(1)
}

func (router *Router) Println(a ...interface{}) {
	writer := router.Writer
	if writer == nil {
		writer = os.Stdout
	}
	fmt.Fprintln(writer, a...)
}

func (cli *Router) Handle(raw []string) error {
	for i := len(raw); i > 0; i-- {
		parts := raw[1:i]
		actions := cli.Search(parts)
		switch len(actions) {
		case 0:
			continue
		case 1:
			var action *Action
			for k, a := range actions {
				parts = strings.Split(k, "/")
				action = a
			}
			if os.Getenv("DEBUG") != "true" {
				defer func(parts []string, action *Action) {
					if r := recover(); r != nil {
						cli.printActionUsage(parts, action, r)
					}
				}(parts, action)
			}
			args := action.Args
			if args == nil {
				args = &Args{}
			}
			e := args.Parse(raw[i:])
			if e == nil {
				e = action.Handler(args)
			}
			if e != nil {
				cli.printActionUsage(parts, action, e.Error())
			}
			return nil
		default:
			// multiple actions count => print help for them
			keys := []string{}
			for key, _ := range actions {
				keys = append(keys, key)
			}
			cli.Println(cli.UsageForKeys(keys, ""))
			return nil

		}
	}
	cli.Println(cli.Usage())
	return nil
}
