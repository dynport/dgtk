package gosql

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSelect(t *testing.T) {
	Convey("ScanMap", t, func() {
		db := testDb(t)
		tx, e := db.Begin()
		So(e, ShouldBeNil)
		defer tx.Rollback()

		_, e = tx.Exec("CREATE TABLE users (id SERIAL NOT NULL, name VARCHAR NOT NULL)")
		So(e, ShouldBeNil)

		for _, n := range []string{"Marek Mintal", "Hans Meyer"} {
			_, e = tx.Exec("INSERT INTO users (name) VALUES ($1)", n)
			So(e, ShouldBeNil)
		}

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
