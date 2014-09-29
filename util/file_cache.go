package util

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/dynport/dgtk/stats"
)

type FileCache struct {
	Root          string
	Key           string
	Source        Source
	SrcCompressed bool

	opened bool
	reader io.Reader
	gz     *gzip.Reader
}

type Source interface {
	Open() (io.Reader, error)
	Size() (int64, error)
}

func (f *FileCache) Close() error {
	es := []string{}
	if f.gz != nil {
		e := f.gz.Close()
		if e != nil {
			es = append(es, e.Error())
		}
	}
	if f.reader != nil {
		if rc, ok := f.reader.(io.ReadCloser); ok {
			e := rc.Close()
			if e != nil {
				es = append(es, e.Error())
			}
		}
	}
	return nil
}

func (f *FileCache) Read(b []byte) (int, error) {
	if !f.opened {
		e := f.open()
		if e != nil {
			return 0, e
		}
		f.opened = true
	}
	if f.gz != nil {
		return f.gz.Read(b)
	} else if f.reader != nil {
		return f.reader.Read(b)
	}
	return 0, fmt.Errorf("needs to be opened first")
}

func (f *FileCache) open() error {
	root := f.Root
	if root == "" {
		root = os.ExpandEnv("$HOME/.cache")
	}
	localPath := root + "/" + f.Key
	tmpPath := localPath + ".tmp"

	var e error

	reader, e := os.Open(localPath)

	if os.IsNotExist(e) {
		e = func() error {
			started := time.Now()
			defer func() {
				logger.Printf("cached in %.06f", time.Since(started).Seconds())
			}()
			dir := path.Dir(tmpPath)
			dbg.Printf("creating %s", dir)
			e = os.MkdirAll(dir, 0755)
			if e != nil {
				return e
			}

			dbg.Printf("caching to file %q", tmpPath)
			file, e := os.OpenFile(tmpPath, os.O_EXCL|os.O_RDWR|os.O_CREATE, 0644)
			if e != nil {
				return e
			}

			src, e := f.Source.Open()
			if e != nil {
				return e
			}

			size, e := f.Source.Size()
			if e != nil {
				return e
			}

			var r io.Reader = src

			if size > 0 {
				w := &stats.ProgressWriter{Total: size}
				defer w.Close()
				r = io.TeeReader(r, w)
			}
			i, e := io.Copy(file, r)
			if e != nil {
				return e
			}
			dbg.Printf("copied %d to %q", i, tmpPath)
			return os.Rename(tmpPath, localPath)
		}()
		if e != nil {
			return e
		}
		reader, e = os.Open(localPath)
	} else if e != nil {
		return e
	} else {
		dbg.Printf("using cached reader from %q", localPath)
	}
	if e != nil {
		return e
	}
	f.reader = reader
	if f.SrcCompressed {
		f.gz, e = gzip.NewReader(f.reader)
		if e != nil {
			return e
		}
	}
	f.opened = true
	return nil
}
