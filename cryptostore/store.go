package cryptostore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

var b64 = base64.StdEncoding

func NewStore(root string) *Store {
	return &Store{Root: root}
}

type Store struct {
	Root string
}

func (store *Store) UserExist(login string) bool {
	return fileExists(store.UserPath(login))
}

func (store *Store) UserPath(login string) string {
	return store.Root + "/users/" + login
}

func (store *Store) LoadPublicKeyForUser(login string) (key *rsa.PublicKey, e error) {
	if !store.UserExist(login) {
		return nil, fmt.Errorf("user " + login + " does not exist")
	}
	rawPubKey, e := ioutil.ReadFile(store.UserPath(login) + "/id_rsa.pub")
	if e != nil {
		return nil, e
	}
	pubKey := &rsa.PublicKey{}
	e = json.Unmarshal(rawPubKey, pubKey)
	if e != nil {
		return nil, e
	}
	return pubKey, nil
}

func (store *Store) StoreFileForUser(login, name string, payload []byte, options *StoreOptions) (e error) {
	if options == nil {
		options = &StoreOptions{}
	}
	if options.Encrypt {
		pubKey, e := store.LoadPublicKeyForUser(login)
		if e != nil {
			return e
		}

		payload, e = rsa.EncryptOAEP(sha1.New(), rand.Reader, pubKey, payload, nil)
		if e != nil {
			return e
		}
	}
	if options.Encode {
		payload = []byte(b64.EncodeToString(payload))
	}
	return ioutil.WriteFile(store.UserPath(login)+"/"+name, payload, 0600)
}

func (store *Store) Read(login string, secret string) (b []byte, e error) {
	privateKey := &rsa.PrivateKey{}
	raw, e := ioutil.ReadFile(store.UserPath(login) + "/id_rsa")
	if e != nil {
		return nil, e
	}
	decoded, e := b64.DecodeString(string(raw))
	if e != nil {
		return nil, e
	}
	crypter := NewCrypter(secret)
	decrypted, e := crypter.Decrypt(decoded)
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal([]byte(decrypted), privateKey)
	if e != nil {
		return nil, e
	}

	secretKey, e := ioutil.ReadFile(store.UserPath(login) + "/BLOB.key")
	if e != nil {
		return nil, e
	}
	decodedSecretKey, e := b64.DecodeString(string(secretKey))
	if e != nil {
		return nil, e
	}

	s, e := rsa.DecryptOAEP(sha1.New(), rand.Reader, privateKey, decodedSecretKey, nil)
	if e != nil {
		return nil, e
	}

	blob, e := ioutil.ReadFile(store.UserPath(login) + "/BLOB")
	if e != nil {
		return nil, e
	}
	decodedBlob, e := b64.DecodeString(string(blob))
	if e != nil {
		return nil, e
	}
	crypter = NewCrypter(string(s))
	decryptedDecodedBlob, e := crypter.Decrypt(decodedBlob)
	if e != nil {
		return nil, e
	}
	return []byte(decryptedDecodedBlob), nil
}

func (store *Store) Store(blob []byte, login string) error {
	if !store.UserExist(login) {
		return fmt.Errorf("user " + login + " does not exist")
	}
	key := GenerateRandomKey()

	e := store.StoreFileForUser(login, "BLOB.key", key, &StoreOptions{Encrypt: true, Encode: true})
	if e != nil {
		return e
	}
	crypter := NewCrypter(string(key))
	encryptedBlob, e := crypter.Encrypt(blob)
	if e != nil {
		return e
	}
	encodedEncryptedBlob := b64.EncodeToString(encryptedBlob)
	return ioutil.WriteFile(store.UserPath(login)+"/BLOB", []byte(encodedEncryptedBlob), 0600)
}

func (store *Store) Users() (users []*User, e error) {
	users = []*User{}
	matches, e := filepath.Glob(store.Root + "/users/*")
	if e != nil {
		return nil, e
	}
	for _, p := range matches {
		stat, e := os.Stat(p)
		if e == nil {
			if stat.IsDir() {
				users = append(users, &User{Login: path.Base(p)})
			}
		}
	}
	return users, nil
}

const DefaultBits = 4096

func (store *Store) CreateUser(login, password string) (u *User, e error) {
	return store.CreateUserWithBits(login, password, DefaultBits)
}

// password needs to have a valid length
func (store *Store) CreateUserWithBits(login, password string, bits int) (u *User, e error) {
	user := &User{}
	dir := store.UserPath(login)
	e = os.MkdirAll(dir, 0755)
	if e != nil {
		return nil, e
	}
	key, e := rsa.GenerateKey(rand.Reader, bits)
	if e != nil {
		return nil, e
	}
	b, e := json.Marshal(key.PublicKey)
	if e != nil {
		return nil, e
	}
	e = ioutil.WriteFile(dir+"/id_rsa.pub", b, 0600)
	if e != nil {
		return nil, e
	}
	crypter := NewCrypter(password)
	b, e = json.Marshal(key)
	if e != nil {
		os.Remove(dir + "/id_rsa.pub")
		return nil, e
	}
	ct, e := crypter.Encrypt(b)
	if e != nil {
		os.Remove(dir + "/id_rsa.pub")
		return nil, e
	}
	dst := []byte(b64.EncodeToString(ct))
	e = ioutil.WriteFile(dir+"/id_rsa", dst, 0600)
	if e != nil {
		os.Remove(dir + "/id_rsa.pub")
		return nil, e
	}
	return user, nil
}
