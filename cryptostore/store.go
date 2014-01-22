package cryptostore

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

var b64 = base64.StdEncoding

func NewStore(root string) *Store {
	return &Store{root: root}
}

type Store struct {
	root string
}

func (store *Store) UserExists(login string) bool {
	return fileExists(store.userPath(login))
}

func (store *Store) userPath(login string) string {
	return store.root + "/users/" + login
}

func (store *Store) loadPublicKeyForUser(login string) (key *rsa.PublicKey, e error) {
	if !store.UserExists(login) {
		return nil, fmt.Errorf("user " + login + " does not exist")
	}
	rawPubKey, e := ioutil.ReadFile(store.userPath(login) + "/id_rsa.pub")
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

func (store *Store) storeFileForUser(login, filePath string, payload []byte, options *StoreOptions) (e error) {
	if options == nil {
		options = &StoreOptions{}
	}
	if options.Encrypt {
		pubKey, e := store.loadPublicKeyForUser(login)
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
	e = os.MkdirAll(path.Dir(filePath), 0700)
	if e != nil {
		return e
	}
	return ioutil.WriteFile(filePath, payload, 0600)
}

func hashPassword(pwd string) []byte {
	out := make([]byte, 32)
	for i, b := range sha256.Sum256([]byte(pwd)) {
		out[i] = b
	}
	return out
}

func (store *Store) Get(key string, login string, secret string) (b []byte, e error) {
	privateKey := &rsa.PrivateKey{}
	raw, e := ioutil.ReadFile(store.userPath(login) + "/id_rsa")
	if e != nil {
		return nil, e
	}
	decoded, e := b64.DecodeString(string(raw))
	if e != nil {
		return nil, e
	}
	hashedPassword := hashPassword(secret)
	crypter := newCrypter(hashedPassword)
	decrypted, e := crypter.Decrypt(decoded)
	if e != nil {
		return nil, e
	}
	e = json.Unmarshal([]byte(decrypted), privateKey)
	if e != nil {
		return nil, e
	}

	dir := store.storePath(login) + "/" + key
	secretKey, e := ioutil.ReadFile(dir + "/BLOB.key")
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

	blob, e := ioutil.ReadFile(dir + "/BLOB")
	if e != nil {
		return nil, e
	}
	decodedBlob, e := b64.DecodeString(string(blob))
	if e != nil {
		return nil, e
	}
	crypter = newCrypter(s)
	decryptedDecodedBlob, e := crypter.Decrypt(decodedBlob)
	if e != nil {
		return nil, e
	}
	return []byte(decryptedDecodedBlob), nil
}

func (store *Store) storePath(login string) string {
	return store.userPath(login) + "/data"
}

func (store *Store) Put(key string, value []byte, login string) error {
	if !store.UserExists(login) {
		return fmt.Errorf("user " + login + " does not exist")
	}
	secret := generateRandomKey()

	dir := store.storePath(login) + "/" + key

	e := store.storeFileForUser(login, dir+"/BLOB.key", secret, &StoreOptions{Encrypt: true, Encode: true})
	if e != nil {
		return e
	}
	crypter := newCrypter(secret)
	encryptedBlob, e := crypter.Encrypt(value)
	if e != nil {
		return e
	}
	encodedEncryptedBlob := b64.EncodeToString(encryptedBlob)
	return ioutil.WriteFile(dir+"/BLOB", []byte(encodedEncryptedBlob), 0600)
}

func (store *Store) Users() (users []*User, e error) {
	users = []*User{}
	matches, e := filepath.Glob(store.root + "/users/*")
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

const DefaultBits = 2048

func (store *Store) CreateUser(login, password string) (u *User, e error) {
	return store.createUserWithBits(login, password, DefaultBits)
}

// password needs to have a valid length
func (store *Store) createUserWithBits(login, password string, bits int) (u *User, e error) {
	if store.UserExists(login) {
		return nil, fmt.Errorf("user already exists")
	}
	user := &User{}
	dir := store.userPath(login)
	log.Printf("creating directory %s", dir)
	e = os.MkdirAll(dir, 0755)
	if e != nil {
		return nil, e
	}
	log.Printf("generating key with %d bits", DefaultBits)
	started := time.Now()
	key, e := rsa.GenerateKey(rand.Reader, bits)
	if e != nil {
		return nil, e
	}
	log.Printf("generated key in %.3f", time.Since(started).Seconds())
	b, e := json.Marshal(key.PublicKey)
	if e != nil {
		return nil, e
	}
	e = ioutil.WriteFile(dir+"/id_rsa.pub", b, 0600)
	if e != nil {
		return nil, e
	}
	crypter := newCrypter(hashPassword(password))
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
