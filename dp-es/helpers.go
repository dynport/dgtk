package main

import "strings"

func normalizeIndexAddress(in string) string {
	if strings.HasPrefix(in, "http") {
		return in
	}
	return "http://" + in
}
