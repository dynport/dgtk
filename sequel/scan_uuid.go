package sequel

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type UUID string

func (u *UUID) Scan(src interface{}) error {
	if b, ok := src.([]byte); !ok {
		return fmt.Errorf("src needs to be []byte")
	} else {
		*u = UUID(strings.Replace(string(b), "-", "", -1))
		return nil
	}
}

func (u UUID) Value() (driver.Value, error) {
	return string(u), nil
}

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
