package log

import (
	"fmt"
)

func Error(format string, i ...interface{}) {
	Log("ERROR", format, i...)
}

func Info(format string, i ...interface{}) {
	Log("INFO ", format, i...)
}

func Log(tag, format string, i ...interface{}) {
	fmt.Printf(tag+" "+format+"\n", i...)
}
