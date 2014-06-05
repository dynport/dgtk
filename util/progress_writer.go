package util

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type ProgressWriter struct {
	Writer  io.Writer
	last    string
	written int
}

func (w *ProgressWriter) Write(b []byte) (int, error) {
	if w.Writer == nil {
		w.Writer = os.Stdout
	}
	if len(w.last) > 0 {
		fmt.Fprintf(w.Writer, "\r%s", strings.Repeat(" ", len(w.last)))
	}
	i, e := fmt.Fprintf(w.Writer, "\r"+string(b))
	w.written++
	w.last = string(b)
	return i, e
}

func (w *ProgressWriter) Close() error {
	if w.written > 0 && w.Writer != nil {
		_, e := fmt.Fprint(w.Writer, "\n")
		return e
	}
	return nil
}
