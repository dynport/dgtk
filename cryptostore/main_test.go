package cryptostore

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	log.SetFlags(0)
	// buf := &bytes.Buffer{}
	// log.SetOutput(buf)
}

func TestEncryptAndDecrypt(t *testing.T) {
	secret := []byte("cei6je9aig2ahzi8eiyau2oP8feeKie7")
	crypter := newCrypter(secret)
	text := "this is secret"
	encrypted, e := crypter.Encrypt([]byte(text))
	if e != nil {
		t.Errorf("error encrypting %v", e)
	}
	if string(encrypted) == text {
		t.Error("encrypted should not equal text")
	}
	var v, ex interface{}
	v = fmt.Sprintf("%T", encrypted)
	ex = "[]uint8"
	if ex != v {
		t.Errorf("expected encrypted to be %#v, was %#v", ex, v)
	}
	decrypted, e := crypter.Decrypt(encrypted)
	if e != nil {
		t.Errorf("error decrypting: %v", e)
	}
	v = string(decrypted)
	ex = text
	if ex != v {
		t.Errorf("expected decrypted to be %#v, was %#v", ex, v)
	}
}

func TestEncryptKeyNotLongEnough(t *testing.T) {
	crypter := newCrypter([]byte("test"))
	_, e := crypter.Cipher()
	if e == nil {
		t.Errorf("error should not be nil")
	}
	var v, ex interface{}
	v = e.Error()
	ex = "crypto/aes: invalid key size 4"
	if ex != v {
		t.Errorf("expected error to be %#v, was %#v", ex, v)
	}
}

const (
	TestStorePath = "./tmp/store"
	userSecret    = "sososecret123456"
)

var blob = []byte("this is a test")

func setup() (*Store, error) {
	storePath, e := filepath.Abs(TestStorePath)
	if e != nil {
		return nil, e
	}
	os.RemoveAll(storePath)
	return NewStore(storePath), nil
}

func createUser(store *Store) error {
	_, e := store.createUserWithBits("user1", userSecret, 1024)
	return e
}

func TestStoreCreateUser(t *testing.T) {
	store, err := setup()
	if err != nil {
		t.Fatal("error setting up", err)
	}

	users, e := store.Users()
	if e != nil {
		t.Errorf("error iterating users: %v", e)
	}
	if len(users) != 0 {
		t.Errorf("expected to get 0 users, got %v", len(users))
	}

	e = createUser(store)
	if e != nil {
		t.Errorf("error creating user with bits %v", e)
	}
	se := []string{
		"./tmp/store/users/user1",
		"./tmp/store/users/user1/id_rsa.pub",
		"./tmp/store/users/user1/id_rsa",
	}

	for _, s := range se {
		_, err := os.Stat(s)
		if err != nil {
			t.Errorf("expected %v to exist but got error %v", s, err)
		}
	}

	users, e = store.Users()
	if e != nil {
		t.Errorf("error iterating users %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected to find 1 user, found %d", len(users))
	}
	if users[0].Login != "user1" {
		t.Errorf("expected first login to be %v. was %v", "user1", users[0].Login)
	}
}

func TestPutAndGetBlob(t *testing.T) {
	store, err := setup()
	if err != nil {
		t.Fatal("error setting up", err)
	}
	err = createUser(store)
	if err != nil {
		t.Fatal("error creating user", err)
	}
	err = store.Put("first", blob, "user1")
	if err != nil {
		t.Fatal("error putting blob:", err)
	}
	b, err := store.Get("first", "user1", userSecret)
	if err != nil {
		t.Errorf("error getting from store: %v", err)
	}
	var v, ex interface{}
	v = string(b)
	ex = "this is a test"
	if ex != v {
		t.Errorf("expected value of blob to be %#v, was %#v", ex, v)
	}
}
