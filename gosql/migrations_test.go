package gosql

import (
	"database/sql"

	"testing"

	. "github.com/smartystreets/goconvey/convey"

	_ "github.com/lib/pq"
)

const dbUrl = "postgres://tobias@localhost/gosql_test?sslmode=disable"

func testDb(t *testing.T) (db *sql.DB) {
	if !runIntegrationTests {
		t.SkipNow()
		return nil
	}
	var e error
	e = func() error {
		db, e = sql.Open("postgres", dbUrl)
		if e != nil {
			t.Fatal(e)
		}
		return db.Ping()
	}()

	if e != nil {
		t.Fatal(e)
	}
	return db

}

func TestMigrations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	Convey("Migrations", t, func() {
		db := testDb(t)
		tx, e := db.Begin()
		So(e, ShouldBeNil)
		defer tx.Rollback()
		m := NewMigrator("CREATE TABLE users (id SERIAL PRIMARY KEY)", "SELECT 2")
		//m.Logger = log.New(os.Stdout, "", 0)
		So(m, ShouldNotBeNil)

		migs, e := m.migrations(tx)
		So(e, ShouldBeNil)
		So(len(migs), ShouldEqual, 2)

		So(migs[0].Statement, ShouldEqual, "CREATE TABLE users (id SERIAL PRIMARY KEY)")
		So(migs[0].Idx, ShouldEqual, 1)
		So(migs[0].Executed, ShouldEqual, false)

		So(migs[1].Statement, ShouldEqual, "SELECT 2")
		So(migs[1].Idx, ShouldEqual, 2)
		So(migs[1].Executed, ShouldEqual, false)

		cnt, e := migs.ExecuteUntil(tx, 1)
		So(e, ShouldBeNil)
		So(cnt, ShouldEqual, 1)

		cnt, e = migs.ExecuteUntil(tx, 1)
		So(e, ShouldBeNil)
		So(cnt, ShouldEqual, 0)

		migs, e = m.migrations(tx)
		So(len(migs), ShouldEqual, 2)
		So(migs[0].Executed, ShouldEqual, true)
		So(migs[1].Executed, ShouldEqual, false)

		cnt, e = migs.Execute(tx)
		So(e, ShouldBeNil)
		So(cnt, ShouldEqual, 1)
		So(migs[1].Executed, ShouldEqual, true)

		cnt, e = migs.Execute(tx)
		So(e, ShouldBeNil)
		So(cnt, ShouldEqual, 0)

		_, e = tx.Exec("INSERT INTO users (id) VALUES ($1)", 77)
		So(e, ShouldBeNil)

		cnt = 0
		e = tx.QueryRow("SELECT id FROM users LIMIT 1").Scan(&cnt)
		So(e, ShouldBeNil)
		So(cnt, ShouldEqual, 77)
	})
}
