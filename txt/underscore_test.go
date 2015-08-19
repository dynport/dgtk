package txt

import "testing"

func TestUnderscore(t *testing.T) {
	tests := []struct {
		in       string
		expected string
	}{
		{"id", "id"},
		{"user", "user"},
		{"ID", "id"},
		{"MD5", "md5"},
		{"GTIN", "gtin"},
		{"userId", "user_id"},
		{"GTINValid", "gtin_valid"},
		{"userID", "user_id"},
		{"GTINAndProductPresent", "gtin_and_product_present"},
		//{"CategoryIDs", "category_ids"},
	}

	for _, tst := range tests {
		v := Underscore(tst.in)
		if tst.expected != v {
			t.Errorf("expected %s to be be %#v in snakeCase was %#v", tst.in, tst.expected, v)
		}
	}
}

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		in       string
		expected string
	}{
		{"id", "id"},
		{"user", "user"},
		{"ID", "id"},
		{"user_id", "userID"},
	}

	for _, tst := range tests {
		v := SnakeCase(tst.in)
		if tst.expected != v {
			t.Errorf("expected %s to be be %#v in snakeCase was %#v", tst.in, tst.expected, v)
		}
	}
}
