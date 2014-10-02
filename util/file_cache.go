package util

import (
	"compress/gzip"
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
}

type Source interface {
	Open() (io.Reader, error)
	Size() (int64, error)
}

func (f *FileCache) Open() (io.ReadCloser, error) {
	root := f.Root
	if root == "" {
		root = os.ExpandEnv("$HOME/.cache")
	}
	localPath := root + "/" + f.Key
	tmpPath := localPath + ".tmp"
	reader, e := os.Open(localPath)

	r := &GzipFileCloser{}
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
			return nil, e
		}
		dbg.Printf("opening local path %q", localPath)
		reader, e = os.Open(localPath)
	} else if e != nil {
		return nil, e
	} else {
		dbg.Printf("using cached reader from %q", localPath)
	}
	r.reader = reader
	if f.SrcCompressed {
		r.gz, e = gzip.NewReader(r.reader)
		if e != nil {
			return nil, e
		}
	}
	return r, nil
}
