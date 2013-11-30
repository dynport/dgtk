package cryptostore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func NewCrypter(secret string) *Crypter {
	return &Crypter{Secret: secret}
}

type Crypter struct {
	Secret string
}

func (crypter *Crypter) Key() []byte {
	return []byte(crypter.Secret)
}

func (crypter *Crypter) Cipher() (c cipher.Block, e error) {
	return aes.NewCipher(crypter.Key())
}

func (crypter *Crypter) Encrypt(plaintext []byte) (b []byte, e error) {
	bl, e := crypter.Cipher()
	if e != nil {
		return b, e
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return b, e
	}

	stream := cipher.NewCFBEncrypter(bl, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func (crypter *Crypter) Decrypt(ciphertext []byte) (s string, e error) {
	bl, e := crypter.Cipher()
	if e != nil {
		return s, e
	}
	if len(ciphertext) < aes.BlockSize {
		return s, fmt.Errorf("ciphertext too short (was %d)", len(ciphertext))
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(bl, iv)

	stream.XORKeyStream(ciphertext, ciphertext)
	return fmt.Sprintf("%s", ciphertext), nil
}
