package bsondecoder

import (
	"bufio"
	"bytes"
	"io"

	"labix.org/v2/mgo/bson"
)

func NewUnbuffered(in io.Reader) *Decoder {
	return &Decoder{in: in}
}

func New(in io.Reader) *Decoder {
	return &Decoder{in: bufio.NewReader(in)}
}

type Decoder struct {
	in io.Reader
}

func (d *Decoder) Decode(i interface{}) error {
	// no longer use bufio.Reader.Peek because we do want to avoid buffering (to enable capturing of parts of bson)
	// read 4 bytes into buffer and save them in saved (so we can use them for reading later)
	buf, saved := &bytes.Buffer{}, &bytes.Buffer{}
	if _, err := io.CopyN(buf, io.TeeReader(d.in, saved), 4); err != nil {
		return err
	}
	toRead := int64(decodeint32(buf.Bytes()))

	bsonBuf := &bytes.Buffer{}
	if _, err := io.CopyN(bsonBuf, io.MultiReader(saved, d.in), toRead); err != nil {
		return err
	}
	return bson.Unmarshal(bsonBuf.Bytes(), i)
}

func decodeint32(b []byte) int32 {
	return int32((uint32(b[0]) << 0) |
		(uint32(b[1]) << 8) |
		(uint32(b[2]) << 16) |
		(uint32(b[3]) << 24))
}
