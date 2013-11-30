package cryptostore

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptAndDecrypt(t *testing.T) {
	secret := "cei6je9aig2ahzi8eiyau2oP8feeKie7"
	crypter := NewCrypter(secret)
	Convey("Encrypt and Decrypt", t, func() {
		text := "this is secret"
		encrypted, e := crypter.Encrypt([]byte(text))
		So(e, ShouldBeNil)
		So(encrypted, ShouldNotEqual, text)
		So(fmt.Sprintf("%T", encrypted), ShouldEqual, "[]uint8")
		decrypted, e := crypter.Decrypt(encrypted)
		So(e, ShouldBeNil)
		So(string(decrypted), ShouldEqual, text)
	})

	Convey("Key not long enough", t, func() {
		crypter = NewCrypter("test")
		_, e := crypter.Cipher()
		So(e, ShouldNotBeNil)
		So(e.Error(), ShouldEqual, "crypto/aes: invalid key size 4")
	})
}

func ShouldExist(actual interface{}, expected ...interface{}) string {
	switch s := actual.(type) {
	case string:
		_, e := os.Stat(s)
		if e != nil {
			return s + " does not exist"
		}
		return ""
	}
	return fmt.Sprintf("actual needs to be string but was %T", actual)
}

const TestStorePath = "./tmp/store"

func TestStore(t *testing.T) {
	userSecret := "sososecret123456"
	blob := []byte("this is a test")
	storePath, e := filepath.Abs(TestStorePath)
	if e != nil {
		t.Fatal(e.Error())
	}
	os.RemoveAll(storePath)
	store := NewStore(storePath)

	Convey("Store", t, func() {
		Convey("Create User", func() {
			So(store, ShouldNotBeNil)
			users, e := store.Users()
			So(e, ShouldBeNil)
			So(len(users), ShouldEqual, 0)

			user, e := store.CreateUserWithBits("user1", userSecret, 1024)
			So(e, ShouldBeNil)
			So(user, ShouldNotBeNil)
			So("./tmp/store/users/user1", ShouldExist)
			So("./tmp/store/users/user1/id_rsa.pub", ShouldExist)
			So("./tmp/store/users/user1/id_rsa", ShouldExist)

			users, e = store.Users()
			So(e, ShouldBeNil)
			So(len(users), ShouldEqual, 1)
			So(users[0].Login, ShouldEqual, "user1")
		})

		Convey("Store BLOB", func() {
			So(store.Store(blob, "user1"), ShouldBeNil)
			So("./tmp/store/users/user1/BLOB", ShouldExist)
			So("./tmp/store/users/user1/BLOB.key", ShouldExist)
		})

		Convey("Read BLOB", func() {
			b, e := store.Read("user1", userSecret)
			So(e, ShouldBeNil)
			So(b, ShouldNotBeNil)
			So(string(b), ShouldEqual, "this is a test")
		})
	})
}
