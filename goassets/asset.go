package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type Asset struct {
	Path  string
	Name  string
	Bytes string
}

func (asset *Asset) Load() error {
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	f, e := os.Open(asset.Path)
	if e != nil {
		return e
	}
	defer f.Close()
	i, e := io.Copy(gz, f)
	log("wrote %s (%d bytes)", asset.Path, i)
	gz.Flush()
	gz.Close()
	if e != nil {
		return e
	}
	list := make([]string, 0, len(buf.Bytes()))
	for _, b := range buf.Bytes() {
		list = append(list, fmt.Sprintf("0x%x", b))
	}
	buffer := makeLineBuffer()
	asset.Name = path.Base(asset.Path)
	for _, b := range list {
		buffer = append(buffer, b)
		if len(buffer) == BYTE_LENGTH {
			asset.Bytes += strings.Join(buffer, ",") + ",\n"
			buffer = makeLineBuffer()
		}
	}
	if len(buffer) > 0 {
		asset.Bytes += strings.Join(buffer, ",") + ",\n"
	}
	return nil
}
