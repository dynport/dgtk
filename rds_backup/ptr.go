package main

func p2s(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func p2i64(s *int64) int64 {
	if s == nil {
		return 0
	}
	return *s
}

func s2p(s string) *string {
	return &s
}
