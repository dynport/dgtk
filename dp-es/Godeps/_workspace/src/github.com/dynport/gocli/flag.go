package gocli

import (
	"fmt"
	"strings"
)

type Flag struct {
	Type         string
	Key          string
	CliFlag      string
	Required     bool
	DefaultValue string
	Description  string
}

func (f *Flag) Matches(key string) bool {
	return strings.HasPrefix(f.CliFlag, key)
}

func (f *Flag) Usage() []string {
	parts := make([]string, 3)
	parts[0] = f.CliFlag
	if f.Required {
		parts[1] = "REQUIRED"
	} else {
		parts[1] = fmt.Sprintf("DEFAULT: %q", f.DefaultValue)
	}
	parts[2] = f.Description

	return parts
}
