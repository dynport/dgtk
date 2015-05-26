package progress

import "fmt"

var (
	oneKb = int64(1024)
	oneMb = oneKb * 1024
	oneGb = oneMb * 1024
)

func SizePretty(raw int64) string {
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

func formatSize(p string, i int64, div int64) string {
	return fmt.Sprintf(p, float64(i)/float64(div))
}
