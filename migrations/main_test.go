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

func TestRun(t *testing.T) {
	if os.Getenv("TEST_WITH_DB") != "true" {
		t.SkipNow()
	}
	migs := New(
		"CREATE TABLE users (id SERIAL NOT NULL PRIMARY KEY, name VARCHAR NOT NULL)",
		insertUsers,
	)

	db, err := sql.Open("postgres", "postgres://localhost/dgtk_migrations?sslmode=disable")
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
	defer tx.Rollback()

	err = migs.ExecuteTx(tx)
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

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"cnt", 1, cnt},
		{"name", "Linux", name},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}

func TestMigrations(t *testing.T) {
	migs, err := New("CREATE TABLE users (id SERIAL NOT NULL PRIMARY KEY, email VARCHAR)", migFunc).migrations()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		Name     string
		Expected interface{}
		Value    interface{}
	}{
		{"len(migs)", 2, len(migs)},
		{"migs[0].Idx", 1, migs[0].Idx},
		{"migs[0].Statement", "CREATE TABLE users (id SERIAL NOT NULL PRIMARY KEY, email VARCHAR)", migs[0].Statement},
		{"migs[1].Idx", 2, migs[1].Idx},
		{"migs[1].Statement", "github.com/dynport/dgtk/migrations.migFunc", migs[1].Statement},
	}

	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected %s to be %#v, was %#v", tst.Name, tst.Expected, tst.Value)
		}
	}
}

func migFunc(tx Con) error {
	return nil
}
