package sequel

import (
	"encoding/json"
	"fmt"
)

// to be used with rows.Scan(scanJSON(&u.Tags))
func ScanJSON(i interface{}) *jsonScanner {
	return &jsonScanner{ref: i}
}

type jsonScanner struct {
	ref interface{}
}

func (s *jsonScanner) Scan(src interface{}) error {
	if src == nil {
		return nil
	}
	b, ok := src.([]byte)
	if !ok {
		return fmt.Errorf("src for JSON needs to be []byte %T", src)
	}
	return json.Unmarshal(b, &s.ref)
}
