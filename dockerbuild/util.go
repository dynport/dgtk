package dockerbuild

import (
	"archive/tar"
	"fmt"
	"github.com/dynport/dgtk/dockerclient"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var (
	started = time.Now()
	now     = time.Date(2013, 11, 18, 12, 27, 0, 0, time.UTC)
)

func callback(s *dockerclient.JSONMessage) {
	if s.Stream != "" {
		debug(strings.TrimSpace(s.Stream))
	} else if s.Status != "" {
		out := strings.TrimSpace(s.Status)
		if s.Progress != nil {
			if s.Progress.Total > 0 {
				out += fmt.Sprintf(" %.1f", s.Progress.Current*100.0/s.Progress.Total)
			}
		}
		debug(out)

	} else if s.Error != nil {
		debug(fmt.Sprintf("%d: %s", s.Error.Code, s.Error.Message))
	} else {
		debug(fmt.Sprintf("%v", s))
	}

}

func writeFilesToArchive(files map[string][]byte, w *tar.Writer) error {
	for name, content := range files {
		e := writeFileToArchive(name, content, w)
		if e != nil {
			return e
		}
	}
	return nil
}

func writeFileToArchive(name string, content []byte, w *tar.Writer) error {
	e := w.WriteHeader(&tar.Header{Size: int64(len(content)), Name: name, Mode: 0644, ModTime: now})
	if e != nil {
		return e
	}
	_, e = w.Write(content)
	if e != nil {
		return e
	}
	return nil
}

func debug(message string) {
	log.Println(fmt.Sprintf("%7s %s", fmt.Sprintf("%.03f", time.Since(started).Seconds()), message))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func mustRead(p string) []byte {
	b, e := ioutil.ReadFile(p)
	if e != nil {
		panic(e.Error())
	}
	return b
}
