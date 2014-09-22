package gosql

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

const createMigrationsSql = `
  CREATE TABLE migrations (idx INTEGER PRIMARY KEY NOT NULL, md5 UUID NOT NULL, statement VARCHAR NOT NULL, created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL)
`

func NewMigrator(migrations ...interface{}) *Migrator {
	return &Migrator{steps: migrations}
}

type Migrator struct {
	Logger *log.Logger
	steps  []interface{}
}

type migrations []*Migration

func (list migrations) Execute(tx Dbi) (int, error) {
	if len(list) == 0 {
		return 0, fmt.Errorf("nothing to execute")
	}
	return list.ExecuteUntil(tx, list[len(list)-1].Idx)
}

func (list migrations) ExecuteUntil(tx Dbi, step int) (int, error) {
	_, e := setupMigrations(tx)
	if e != nil {
		return 0, e
	}
	cnt := 0
	for _, m := range list {
		if m.Idx <= step {
			if m.Executed {
				m.log("SKIP")
			} else {
				m.log("EXECUTE")
				if _, e := m.Execute(tx); e != nil {
					return cnt, e
				}
				m.Executed = true
				cnt++
			}
		}
	}
	return cnt, nil
}

func (migrator *Migrator) migrations(tx Dbi) (migrations, error) {
	out := []*Migration{}

	_, e := setupMigrations(tx)
	if e != nil {
		return nil, e
	}
	for idx, i := range migrator.steps {
		m, e := newMigration(idx+1, i)
		if e != nil {
			return nil, e
		}
		m.Logger = migrator.Logger
		var cs string
		if e = tx.QueryRow("SELECT md5 FROM migrations where idx = $1", m.Idx).Scan(&cs); e != nil {
			if e != sql.ErrNoRows {
				return nil, e
			}
		} else {
			cs = strings.Replace(cs, "-", "", -1)
			if cs == m.checksum() {
				m.Executed = true
			} else {
				return nil, fmt.Errorf("migration %d (new checksum %s) already exists with checksum %q", m.Idx, cs, m.checksum())
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func (list *Migrator) Execute(tx Dbi) error {
	migs, e := list.migrations(tx)
	if e != nil {
		return e
	}
	_, e = migs.Execute(tx)
	return e
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
	Logger    *log.Logger
	Executed  bool
}

const errorMigrationsDoesNotExist = `pq: relation "migrations" does not exist`

func setupMigrations(tx Dbi) (sql.Result, error) {
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

func (m *Migration) Execute(tx Dbi) (sql.Result, error) {
	res, e := tx.Exec(m.Statement)
	if e != nil {
		return nil, e
	}
	_, e = tx.Exec("INSERT INTO migrations (idx, md5, statement, created_at) VALUES ($1, $2, $3, $4)", m.Idx, m.checksum(), m.Statement, time.Now().UTC().Format(time.RFC3339Nano))
	return res, e
}
