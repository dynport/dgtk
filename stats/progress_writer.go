package stats

import (
	"fmt"
	"math"
	"strings"
)

type ProgressWriter struct {
	Total    int
	Written  int
	lastLine string
}

func (p *ProgressWriter) Write(b []byte) (int, error) {
	p.Written += len(b)
	perc := float64(p.Written) / float64(p.Total)

	ballsCount := 64

	full := int(math.Ceil(float64(ballsCount) * perc))
	empty := ballsCount - full

	prefix := "[" + strings.Repeat("*", full) + strings.Repeat("-", empty) + "]"

	line := fmt.Sprintf("%s (%.02f%%)", prefix, 100.0*perc)
	fmt.Print("\r" + strings.Repeat(" ", len(p.lastLine)) + "\r" + line)
	p.lastLine = line
	return len(b), nil
}

func (p *ProgressWriter) Close() error {
	fmt.Printf("\n")
	return nil
}
