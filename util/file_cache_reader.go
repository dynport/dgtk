package util

import (
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

type GzipFileCloser struct {
	reader io.Reader
	gz     *gzip.Reader
}

func (f *GzipFileCloser) Close() error {
	errors := []string{}
	if f.gz != nil {
		e := f.gz.Close()
		if e != nil {
			errors = append(errors, e.Error())
		}
	}
	if f.reader != nil {
		if rc, ok := f.reader.(io.ReadCloser); ok {
			e := rc.Close()
			if e != nil {
				errors = append(errors, e.Error())
			}
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, ", "))
	}
	return nil
}

func (f *GzipFileCloser) Read(b []byte) (int, error) {
	if f.gz != nil {
		return f.gz.Read(b)
	} else if f.reader != nil {
		return f.reader.Read(b)
	}
	return 0, fmt.Errorf("needs to be opened first")
}
