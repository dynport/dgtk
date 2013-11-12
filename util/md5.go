package util

import (
	"crypto/md5"
	"fmt"
	"io"
)

func MD5String(s string) string {
	m := md5.New()
	io.WriteString(m, s)
	return fmt.Sprintf("%x", m.Sum(nil))
}
