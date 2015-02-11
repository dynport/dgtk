package migrations

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

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

func (list Migrations) ExecuteTx(tx *sql.Tx) error {
	_, e := list.setup(tx)
	if e != nil {
		return e
	}

	return func() error {
		for idx, i := range list.steps {
			m, e := newMigration(idx+1, i)
			if e != nil {
				return e
			}
			m.Logger = list.Logger
			_, e = m.Execute(tx)
			if e != nil {
				return e
			}
		}
		return nil
	}()
}

func newMigration(idx int, statement interface{}) (*Migration, error) {
	m := &Migration{Idx: idx}

	switch casted := statement.(type) {
	case string:
		m.Statement = casted
	case fmt.Stringer:
		m.Statement = casted.String()
	default:
		return nil, fmt.Errorf("type %T not supported", casted)
	}
	return m, nil
}

type Migration struct {
	Idx       int
	Statement string
	Logger    logger
}

const errorMigrationsDoesNotExist = `pq: relation "migrations" does not exist`

func (list Migrations) setup(tx *sql.Tx) (sql.Result, error) {
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

func (m *Migration) log(t string) {
	if m.Logger != nil {
		out := []string{}
		lines := strings.Split(strings.TrimSpace(m.Statement), "\n")
		for _, l := range lines {
			out = append(out, strings.TrimSpace(l))
		}
		m.Logger.Printf(t+": migration %d %q %q", m.Idx, m.checksum(), strings.Join(strings.Fields(strings.Join(out, " ")), " "))
	}
}

func (m *Migration) checksum() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(m.Statement)))
}

func (m *Migration) Execute(tx *sql.Tx) (sql.Result, error) {
	rows, e := tx.Query("SELECT md5 FROM migrations where idx = $1", m.Idx)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var cs string
		e = rows.Scan(&cs)
		if e != nil {
			return nil, e
		}
		cs = strings.Replace(cs, "-", "", -1)
		if cs == m.checksum() {
			m.log("SKIP")
			return nil, nil
		} else {
			return nil, fmt.Errorf("migration %d (new checksum %s) already exists with checksum %q", m.Idx, cs, m.checksum())
		}
	}
	m.log("EXEC")
	res, e := tx.Exec(m.Statement)
	if e != nil {
		return nil, e
	}
	_, e = tx.Exec("INSERT INTO migrations (idx, md5, statement, created_at) VALUES ($1, $2, $3, $4)", m.Idx, m.checksum(), m.Statement, time.Now().UTC().Format(time.RFC3339Nano))
	if e != nil {
		return nil, e
	}
	return res, e
}
