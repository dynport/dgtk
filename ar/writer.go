package ar

import (
	"fmt"
	"io"
	"strconv"
	"time"
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

type Writer struct {
	writer  io.Writer
	written int
}

type Header struct {
	Name     string
	Modified time.Time
	Owner    int
	Group    int
	FileMode int
	Size     int
}

func (w *Writer) Write(b []byte) (int, error) {
	i, e := w.writer.Write(b)
	w.written += i
	return i, e
}

func (w *Writer) WriteHeader(header *Header) error {
	if w.written == 0 {
		_, e := io.WriteString(w, "!<arch>\n")
		if e != nil {
			return e
		}
	}
	if w.written%2 != 0 {
		io.WriteString(w, "\n")
	}
	header.Modified = time.Now()
	_, e := io.WriteString(w, fmt.Sprintf("%-16s%-12s%-6s%-6s%-8s%-10s",
		header.Name+"/",
		strconv.FormatInt(header.Modified.UTC().Unix(), 10),
		"0",
		"0",
		"100644",
		strconv.Itoa(header.Size),
	),
	)
	if e != nil {
		return e
	}
	_, e = w.Write([]byte{0x60, 0x0A})
	return e
}
