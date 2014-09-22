package bsondecoder

import (
	"bufio"
	"bytes"
	"io"

	"labix.org/v2/mgo/bson"
)

func New(in io.Reader) *Decoder {
	return &Decoder{in: bufio.NewReader(in)}
}

type Decoder struct {
	in *bufio.Reader
}

func (d *Decoder) Decode(i interface{}) error {
	b, e := d.in.Peek(4)
	if e != nil {
		return e
	}
	buf := &bytes.Buffer{}
	_, e = io.CopyN(buf, d.in, int64(decodeint32(b)))
	if e != nil {
		return e
	}
	return bson.Unmarshal(buf.Bytes(), i)
}

func decodeint32(b []byte) int32 {
	return int32((uint32(b[0]) << 0) |
		(uint32(b[1]) << 8) |
		(uint32(b[2]) << 16) |
		(uint32(b[3]) << 24))
}
