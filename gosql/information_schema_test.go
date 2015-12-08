package gosql

import (
	"database/sql"
	"os"
	"sort"
	"testing"

	_ "github.com/lib/pq"
)

var runIntegrationTests = os.Getenv("RUN_INTEGRATION_TESTS") == "true"

func connect(t *testing.T) (db *sql.DB) {
	if !runIntegrationTests {
		t.SkipNow()
		return nil
	}
	err := func() error {
		var err error
		db, err = sql.Open("postgres", "postgres://localhost/gosql_test?sslmode=disable")
		if err != nil {
			return err
		}
		return db.Ping()
	}()
	if err != nil {
		t.Fatal("error connecting db", err)
	}
	return db
}

func setup(t *testing.T) *sql.DB {
	db := connect(t)
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS accounts (id SERIAL NOT NULL, name VARCHAR NOT NULL, age INTEGER)")
	if err != nil {
		t.Error("error creating accounts table", err)
	}
	return db

}

func TestColumns(t *testing.T) {
	db := setup(t)
	defer db.Close()
}

func TestTables(t *testing.T) {
	db := setup(t)
	defer db.Close()

	tables, err := Tables(db, PublicTables)
	if err != nil {
		t.Fatal("unexpected error calling Tables", err)
	}
	if len(tables) != 1 {
		t.Errorf("expected at 1, was %d", len(tables))
	}

	tables, err = Tables(db)
	if err != nil {
		t.Fatal("unexpected error calling Tables", err)
	}
	if len(tables) < 100 {
		t.Errorf("expected at least 100 tables, found %d", len(tables))
	}

	user := func() *Table {
		for _, tab := range tables {
			if tab.TableSchema == "public" && tab.TableName == "accounts" {
				return tab
			}
		}
		return nil
	}()
	if user == nil {
		t.Fatal("did not find accounts table")
	}
	tests := []struct {
		Description string
		Expected    interface{}
		Value       interface{}
	}{
		{"Name", "accounts", user.TableName},
		{"IsInsertableInto", "YES", user.IsInsertableInto},
		{"TableName", "BASE TABLE", user.TableType},
	}
	for _, tst := range tests {
		if tst.Expected != tst.Value {
			t.Errorf("expected value of %q to be %q, was %q", tst.Description, tst.Expected, tst.Value)
		}
	}
}

func TestReflect(t *testing.T) {
	db := connect(t)
	defer db.Close()
	names, err := columnNames(db, &Table{})
	if err != nil {
		t.Error("error getting columnNames", err)
	}
	if len(names) != 12 {
		t.Errorf("expected %d column, found %d", 12, len(names))
	}
	sort.Strings(names)

	for i, expected := range []string{"commit_action", "is_insertable_into"} {
		v := names[i]
		if v != expected {
			t.Errorf("expected value of index %d to be %q, was %q", i, expected, v)
		}
	}
}
