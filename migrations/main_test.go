package migrations

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func insertUsers(con Con) error {
	_, err := con.Exec("INSERT INTO users (name) VALUES ($1)", "Linux")
	return err
}

func databaseURL() string {
	env := os.Getenv("TEST_DATABSE_URL")
	if env != "" {
		return env
	}
	return "postgres://localhost/dgtk_migrations?sslmode=disable"
}

func testConnect(t *testing.T) *sql.Tx {
	db, err := sql.Open("postgres", databaseURL())
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func TestRun(t *testing.T) {
	if os.Getenv("TEST_WITH_DB") != "true" {
		t.SkipNow()
	}
	migs := New(
		"CREATE TABLE users (id SERIAL NOT NULL PRIMARY KEY, name VARCHAR NOT NULL)",
		insertUsers,
	)

	tx := testConnect(t)
	defer tx.Rollback()

	err := migs.ExecuteTx(tx)
	if err != nil {
		t.Errorf("error running migrations: %s", err)
	}

	var cnt int
	if err := tx.QueryRow("SELECT COUNT(1) FROM users").Scan(&cnt); err != nil {
		t.Fatal(err)
	}
	var name string
	if err := tx.QueryRow("SELECT name FROM users").Scan(&name); err != nil {
		t.Fatal(err)
	}

	all, err := migs.All(tx)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[int]struct{ Has, Want interface{} }{
		1: {cnt, 1},
		2: {name, "Linux"},
		3: {len(all), 2},
		4: {all[0].Executed, true},
		5: {all[1].Executed, true},
	}
	for i, tc := range tests {
		if tc.Has != tc.Want {
			t.Errorf("%d: want=%#v has=%#v", i, tc.Want, tc.Has)
		}
	}

}

func TestMigrations(t *testing.T) {
	if os.Getenv("TEST_WITH_DB") != "true" {
		t.SkipNow()
	}
	tx := testConnect(t)
	defer tx.Rollback()
	migs := New("CREATE TABLE users (id SERIAL NOT NULL PRIMARY KEY, email VARCHAR)", migFunc)
	_, err := migs.setup(tx)
	if err != nil {
		t.Fatal(err)
	}
	all, err := migs.All(tx)
	if err != nil {
		t.Fatal(err)
	}

	tests := map[int]struct{ Has, Want interface{} }{
		1: {len(all), 2},
		2: {all[0].Idx, 1},
		3: {all[0].Statement, "CREATE TABLE users (id SERIAL NOT NULL PRIMARY KEY, email VARCHAR)"},
		4: {all[1].Idx, 2},
		5: {all[1].Statement, "github.com/dynport/dgtk/migrations.migFunc"},
		6: {all[0].Executed, false},
	}
	for i, tc := range tests {
		if tc.Has != tc.Want {
			t.Errorf("%d: want=%#v has=%#v", i, tc.Want, tc.Has)
		}
	}
}

func migFunc(tx Con) error {
	return nil
}
