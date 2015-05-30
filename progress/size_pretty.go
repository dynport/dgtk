package progress

import "fmt"

var (
	oneKb float64 = 1024
	oneMb float64 = oneKb * 1024
	oneGb float64 = oneMb * 1024
)

func SizePretty(raw float64) string {
	f := raw
	if f < oneKb {
		return formatSize("%.0fB", f, 1)
	} else if f < oneMb {
		return formatSize("%.2fKB", f, oneKb)
	} else if f < oneGb {
		return formatSize("%.2fMB", f, oneMb)
	} else {
		return formatSize("%.2fGB", f, oneGb)
	}
}

func formatSize(p string, i, div float64) string {
	return fmt.Sprintf(p, i/div)
}
