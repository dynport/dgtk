package migrations

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Con interface {
	QueryRow(query string, args ...interface{}) *sql.Row
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Prepare(query string) (*sql.Stmt, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

const createMigrationsSql = `
  CREATE TABLE migrations (idx INTEGER PRIMARY KEY NOT NULL, md5 UUID NOT NULL, statement VARCHAR NOT NULL, created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL)
`

func New(migrations ...interface{}) *Migrations {
	return &Migrations{steps: migrations}
}

type logger interface {
	Printf(string, ...interface{})
}

type Migrations struct {
	Logger logger
	steps  []interface{}
}

func (list Migrations) Execute(db *sql.DB) error {
	tx, e := db.Begin()
	if e != nil {
		return e
	}

	e = list.ExecuteTx(tx)

	if e != nil {
		tx.Rollback()
		return e
	}
	return tx.Commit()
}

type Tx interface {
	Con
	Commit() error
	Rollback() error
}

func (list Migrations) ExecuteTx(tx Tx) error {
	started := time.Now()
	if _, err := list.setup(tx); err != nil {
		return err
	}

	migrations, err := list.All(tx)
	if err != nil {
		return err
	}
	for _, m := range migrations {
		if err := m.Execute(tx); err != nil {
			return err
		}
	}
	if list.Logger != nil {
		list.Logger.Printf("migrated in %.06f", time.Since(started).Seconds())
	}
	return nil
}

func (list Migrations) loadMigrationStatus(tx Tx) (m map[int]*Migration, err error) {
	m = map[int]*Migration{}
	q := "SELECT idx, md5, statement FROM migrations ORDER BY idx"
	rows, err := tx.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		s := &Migration{}
		var cs string
		if err := rows.Scan(&s.Idx, &cs, &s.Statement); err != nil {
			return nil, err
		}
		s.MD5 = strings.Replace(cs, "-", "", -1)
		if _, ok := m[s.Idx]; ok {
			return nil, fmt.Errorf("there are two migrations with idx=%d", s.Idx)
		}
		m[s.Idx] = s
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

type migrationStatus struct {
	IDX       int
	Statement string
	MD5       string
}

func (list Migrations) All(tx Tx) (out []*Migration, err error) {
	status, err := list.loadMigrationStatus(tx)
	if err != nil {
		return nil, err
	}
	for idx, i := range list.steps {
		if m, err := newMigration(idx+1, i); err != nil {
			return nil, err
		} else {
			m.Logger = list.Logger
			m.MD5 = strings.Replace(m.Statement, "-", "", -1)
			executed, ok := status[m.Idx]
			if ok {
				if m.Statement != executed.Statement {
					return nil, fmt.Errorf("MIGRATION MISMATCH:\n<<<<<<< code migration %d\n%q\n=======\n%q\n>>>>>>> db migration\n", m.Idx, m.Statement, executed.Statement)
				}
				m.Executed = true
			}
			out = append(out, m)
		}
	}
	return out, nil
}

func newMigration(idx int, statement interface{}) (*Migration, error) {
	m := &Migration{Idx: idx}

	switch casted := statement.(type) {
	case string:
		m.Statement = casted
	case fmt.Stringer:
		m.Statement = casted.String()
	case func(Con) error:
		m.Statement = runtime.FuncForPC(reflect.ValueOf(statement).Pointer()).Name()
		m.Func = casted
	default:
		return nil, fmt.Errorf("type %T not supported", casted)
	}
	return m, nil
}

type Migration struct {
	Idx       int
	Statement string
	Func      func(Con) error
	Logger    logger
	MD5       string
	Executed  bool
}

func (list Migrations) setup(tx Tx) (sql.Result, error) {
	row := tx.QueryRow("SELECT COUNT(1) FROM pg_tables WHERE schemaname = $1 AND tablename = $2", "public", "migrations")
	var cnt int
	e := row.Scan(&cnt)
	if e != nil {
		return nil, e
	}

	if cnt == 0 {
		return tx.Exec(createMigrationsSql)
	}
	return nil, nil
}

func (m *Migration) log(t string, dur time.Duration) {
	if m.Logger != nil {
		out := []string{}
		lines := strings.Split(strings.TrimSpace(m.Statement), "\n")
		for _, l := range lines {
			out = append(out, strings.TrimSpace(l))
		}
		msg := fmt.Sprintf("%s: migration %d %q %q", t, m.Idx, m.checksum(), strings.Join(strings.Fields(strings.Join(out, " ")), " "))
		if dur != 0 {
			msg += fmt.Sprintf(" [%.06f]", dur.Seconds())
		}
		m.Logger.Printf(msg)
	}
}

func (m *Migration) checksum() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(m.Statement)))
}

func (m *Migration) Execute(tx Tx) error {
	rows, err := tx.Query("SELECT md5, statement FROM migrations where idx = $1", m.Idx)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cs, statement string
		if err := rows.Scan(&cs, &statement); err != nil {
			return err
		}
		cs = strings.Replace(cs, "-", "", -1)
		if statement == m.Statement {
			m.log("SKIP", 0)
			return nil
		} else {
			return fmt.Errorf("MIGRATION MISMATCH:\n<<<<<<< code migration %d\n%q\n=======\n%q\n>>>>>>> db migration\n", m.Idx, m.Statement, statement)
		}
	}
	started := time.Now()
	if m.Func != nil {
		if err := m.Func(tx); err != nil {
			return err
		}
	} else {
		if _, err := tx.Exec(m.Statement); err != nil {
			return err
		}
	}
	_, err = tx.Exec("INSERT INTO migrations (idx, md5, statement, created_at) VALUES ($1, $2, $3, $4)", m.Idx, m.checksum(), m.Statement, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return err
	}
	m.log("EXEC", time.Since(started))
	return err
}
