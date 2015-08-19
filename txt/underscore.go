package txt

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	charMinorA = "a"
	charMinorZ = "z"
	charMajorA = "A"
	charMajorZ = "Z"
	char0      = "0"
	char9      = "9"
)

func Underscore(s string) string {
	out := &bytes.Buffer{}
	parts := strings.Split(s, "")
	var last string
	allMajor := true
	for i, c := range parts {
		next := ""
		if len(parts) > i+1 {
			next = parts[i+1]
		}
		if isMajor(c) {
			if isMajor(last) {
				if next != "" && !isMajor(next) {
					fmt.Fprint(out, "_")
				}
				fmt.Fprint(out, strings.ToLower(c))
			} else {
				p := "_"
				if i == 0 || allMajor {
					p = ""
				}
				fmt.Fprintf(out, p+strings.ToLower(c))
			}
		} else {
			if allMajor {
				allMajor = false
			}
			fmt.Fprintf(out, c)
		}
		last = c
	}
	return out.String()
}

func SnakeCase(s string) string {
	out := &bytes.Buffer{}
	parts := strings.Split(s, "")
	var last string
	for _, c := range parts {
		if last == "_" {
			if (charMinorA <= c && c <= charMinorZ) || (charMajorA <= c && c <= charMajorZ) {
				fmt.Fprintf(out, strings.ToUpper(c))
			}
		} else if c != "_" {
			fmt.Fprintf(out, strings.ToLower(c))
		}
		last = c
	}
	o := out.String()
	if strings.HasSuffix(o, "Id") {
		return o[0:len(o)-2] + "ID"
	}
	return o
}

func CamelCase(s string) string {
	return strings.Title(SnakeCase(s))
}

func isMajor(c string) bool {
	return isNumeric(c) || (charMajorA <= c && c <= charMajorZ)
}

func isNumeric(c string) bool {
	return char0 <= c && c <= char9
}
