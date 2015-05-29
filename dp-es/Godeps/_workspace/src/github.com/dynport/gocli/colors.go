package gocli

import (
	"fmt"
	"strings"
)

const (
	BLACK = iota
	RED
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	WHITE
)

func Black(s string) string {
	return Colorize(BLACK, s)
}

func Red(s string) string {
	return Colorize(RED, s)
}

func Green(s string) string {
	return Colorize(GREEN, s)
}

func Yellow(s string) string {
	return Colorize(YELLOW, s)
}

func Blue(s string) string {
	return Colorize(BLUE, s)
}

func Magenta(s string) string {
	return Colorize(MAGENTA, s)
}

func Cyan(s string) string {
	return Colorize(CYAN, s)
}

func White(s string) string {
	return Colorize(WHITE, s)
}

func Colorize(c int, s string) (r string) {
	if strings.TrimSpace(s) == "" {
		return s
	}
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", c, s)
}
