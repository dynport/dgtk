package main

import (
	"io"
)

type Resource interface {
	Store() error
	Tags() (map[string]string, error)
	Exists() bool
	Open() (io.Reader, int64, error)
	DockerSize() (int64, error)
	LoadResource(p string) ([]byte, error)
}
