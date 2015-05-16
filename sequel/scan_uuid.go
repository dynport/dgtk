package sequel

import (
	"fmt"
	"strings"
)

func ScanUUID(s *string) *uuidScanner {
	return &uuidScanner{ref: s}
}

type uuidScanner struct {
	ref *string
}

func (u *uuidScanner) Scan(i interface{}) error {
	switch c := i.(type) {
	case []byte:
		s := string(c)
		*u.ref = strings.Replace(s, "-", "", -1)
		return nil
	default:
		return fmt.Errorf("type %T not supported when scanning UUIDs, expected []byte", i)
	}
}
