package main

import "testing"

func TestValidToken(t *testing.T) {
	valid := []string{
		"123456",
		"112233",
	}
	for _, tok := range valid {
		if !validToken(tok) {
			t.Errorf("expected token %q to be valid", tok)
		}
	}
	notValid := []string{
		"123 456",
		"12345",
		"12345a",
	}
	for _, tok := range notValid {
		if validToken(tok) {
			t.Errorf("expected token %q NOT to be valid", tok)
		}
	}

}
