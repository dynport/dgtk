package cryptostore

import (
	"crypto/rand"
	"io"
	"os"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func GenerateRandomKey() []byte {
	key := make([]byte, 32)
	_, e := io.ReadFull(rand.Reader, key)
	if e != nil {
		panic("unable to generate random key: " + e.Error())
	}
	return key
}
