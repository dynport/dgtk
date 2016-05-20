package main

func b2p(b bool) *bool {
	return &b
}

func i2s(in *int) int {
	if in == nil {
		return 0
	}
	return *in
}

func p2i(in int) *int {
	return &in
}

func i642s(in *int64) int64 {
	if in == nil {
		return 0
	}
	return *in
}

func p2i64(in int64) *int64 {
	return &in
}

func p2s(in *string) string {
	if in == nil {
		return ""
	}
	return *in
}

func s2p(in string) *string {
	return &in
}
