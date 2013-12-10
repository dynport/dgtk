package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/dynport/dgtk/log"
	"io"
	"os"
	"path"
	"strings"
)

type Asset struct {
	Path  string
	Key   string
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
	log.Debug("loading %s (%d bytes)", asset.Path, i)
	gz.Flush()
	gz.Close()
	if e != nil {
		return e
	}
	list := make([]string, 0, len(buf.Bytes()))
	for _, b := range buf.Bytes() {
		list = append(list, fmt.Sprintf("0x%x", b))
	}
	log.Debug("length of list is %d", len(list))
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
	log.Debug("asset has %d bytes", len(asset.Bytes))
	return nil
}
