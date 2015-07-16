package tskip

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type ti interface {
	Error(i ...interface{})
}

func Errorf(t ti, skip int, m string, args ...interface{}) {
	doError(t, 0+skip, fmt.Sprintf(m, args...))
}

func Error(t ti, skip int, m string) {
	doError(t, 0+skip, m)
}

func doError(t ti, skip int, m string) {
	_, file, line, _ := runtime.Caller(1 + skip)
	if tr := os.Getenv("TEST_ROOT"); tr != "" {
		file = strings.TrimPrefix(file, tr+"/")
	}
	t.Error("\r        " + file + ":" + strconv.Itoa(line) + ": " + m)
}
