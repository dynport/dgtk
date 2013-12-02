package dockerclient

import (
	"io"
)

func sendLineToWriter(w io.Writer, p []byte) (e error) {
	if w == nil {
		return nil
	}
	i := 0
	for i < len(p) {
		n, e := w.Write(p[i:])
		if e != nil {
			return e
		}
		i += n
	}
	return nil
}
