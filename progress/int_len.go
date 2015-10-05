package progress

import "math"

func IntLen(i int) int {
	if i == 0 {
		return 1
	} else if i < 0 {
		return IntLen(int(math.Abs(float64(i)))) + 1
	}
	return int(math.Ceil(math.Log10(float64(i + 1))))
}
