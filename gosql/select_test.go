package gosql

import (
	"database/sql"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var cachedTx *sql.Tx

func cleanup() {
	if cachedTx != nil {
		cachedTx.Rollback()
		cachedTx = nil
	}
}

func testTx(t *testing.T) *sql.Tx {
	if cachedTx != nil {
		return cachedTx
	}
	e := func() error {
		db := testDb(t)
		var e error
		cachedTx, e = db.Begin()
		return e
	}()
	if e != nil {
		t.Fatal(e.Error())
	}
	return cachedTx
}

func prepareTx(t *testing.T, stmts ...string) *sql.Tx {
	tx := testTx(t)
	e := func() error {
		for _, s := range stmts {
			_, e := tx.Exec(s)
			if e != nil {
				return e
			}
		}
		return nil
	}()
	if e != nil {
		t.Fatal(e)
	}
	return tx
}

func TestSelect(t *testing.T) {

	type user struct {
		Id   int    `sql:"id"`
		Name string `sql:"name"`
	}

	Convey("Select struct", t, func() {
		defer cleanup()
		tx := prepareTx(t, "CREATE TABLE users (id integer, name varchar)", "INSERT INTO users (id, name) VALUES (1, 'hans'), (11, 'marek')")
		u := &user{}
		e := SelectStruct(tx, "SELECT id, name FROM users ORDER BY id", u)
		So(e, ShouldBeNil)
		So(u.Name, ShouldEqual, "hans")
		So(u.Id, ShouldEqual, 1)

		var users []*user
		e = SelectStructs(tx, "SELECT id, name FROM users ORDER BY id", &users)
		So(e, ShouldBeNil)
		So(len(users), ShouldEqual, 2)
		So(users[1].Name, ShouldEqual, "marek")
		So(users[1].Id, ShouldEqual, 11)
	})

	Convey("SelectInt", t, func() {
		defer cleanup()
		tx := prepareTx(t, "CREATE TABLE ints (id INTEGER)", "INSERT INTO ints VALUES (77), (88)")
		cnt, e := SelectInt(tx, "SELECT id FROM ints ORDER BY id")
		So(e, ShouldBeNil)
		So(cnt, ShouldEqual, 77)

		ints, e := SelectInts(tx, "SELECT id FROM ints ORDER BY id")
		So(e, ShouldBeNil)
		So(len(ints), ShouldEqual, 2)
		So(ints[0], ShouldEqual, 77)
		So(ints[1], ShouldEqual, 88)
	})

	Convey("SelectString", t, func() {
		tx := prepareTx(t, "CREATE TABLE names (name VARCHAR)", "INSERT INTO names (name) VALUES ('hans'), ('meyer')")
		name, e := SelectString(tx, "SELECT name FROM names ORDER BY name")
		So(e, ShouldBeNil)
		So(name, ShouldEqual, "hans")

		names, e := SelectStrings(tx, "SELECT name FROM names ORDER BY name")
		So(e, ShouldBeNil)
		So(len(names), ShouldEqual, 2)
		So(names[0], ShouldEqual, "hans")
		So(names[1], ShouldEqual, "meyer")
	})

	Convey("ScanMap", t, func() {
		tx := prepareTx(t, "CREATE TABLE users (id SERIAL NOT NULL, name VARCHAR NOT NULL)", "INSERT INTO users (name) VALUES ('Marek Mintal'), ('Hans Meyer')")

		m := map[string]interface{}{}
		rows, e := tx.Query("SELECT * from users ORDER BY id")
		So(e, ShouldBeNil)
		rows.Next()

		e = ScanMap(rows, m)
		So(e, ShouldBeNil)
		So(len(m), ShouldEqual, 2)
		So(m["name"], ShouldEqual, "Marek Mintal")
	})

}
