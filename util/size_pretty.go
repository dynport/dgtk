package util

import "fmt"

var (
	oneKb = 1024.0
	oneMb = oneKb * 1024.0
	oneGb = oneMb * 1024.0
)

func SizePretty(raw int64) string {
	f := float64(raw)
	if f < oneKb {
		return fmt.Sprintf("%.0f", f)
	} else if f < oneMb {
		return fmt.Sprintf("%.2fKB", f/oneKb)
	} else if f < oneGb {
		return fmt.Sprintf("%.2fMB", f/oneMb)
	} else {
		return fmt.Sprintf("%.2fGB", f/oneGb)
	}
}
