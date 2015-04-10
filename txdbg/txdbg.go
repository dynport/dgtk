// Debug SQL transactions inside tests
//
package txdbg

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"text/template"
	"time"
)

// StartServer starts a new http server for the given transaction
//
// The server shuts down when a) the user presses the "quit" button or b) on the first db error (as we can not recover from that error inside the transaction)
func StartServer(tx *sql.Tx) (waitForClose chan struct{}, address string) {
	cc := make(chan struct{})
	var s *httptest.Server
	handler := func(w http.ResponseWriter, r *http.Request) {
		doClose := func(err error) {
			msg := "Quit"
			if err != nil {
				msg = "Abort due to error: " + err.Error()
			}
			io.WriteString(w, msg)
			close(cc)
		}
		err := func() error {
			err := r.ParseForm()
			if err != nil {
				return err
			}
			if r.Form.Get("quit") == "true" {
				doClose(nil)
				return nil
			}
			ctx := &debugContext{
				Query: r.Form.Get("query"),
			}
			if ctx.Query != "" {
				ctx.Table, err = loadQuery(tx, ctx.Query)
				if err != nil {
					doClose(err)
					return nil
				}
			}
			tpl, err := template.New("index").Parse(txDebugTpl)
			if err != nil {
				return err
			}
			buf := &bytes.Buffer{}
			err = tpl.Execute(buf, ctx)
			if err != nil {
				return err
			}
			io.Copy(w, buf)
			return nil
		}()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	s = httptest.NewServer(http.HandlerFunc(handler))
	return cc, s.URL
}

func loadQuery(tx *sql.Tx, q string, args ...interface{}) (*table, error) {
	rows, err := tx.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	t := &table{Header: names}
	for rows.Next() {
		ints := []*interface{}{}
		refs := []interface{}{}
		for i := 0; i < len(names); i++ {
			var i interface{}
			refs = append(refs, &i)
			ints = append(ints, &i)
		}
		err = rows.Scan(refs...)
		if err != nil {
			return nil, err
		}
		r := row{}
		for _, v := range ints {
			r = append(r, valueToString(*v))
		}
		t.Rows = append(t.Rows, r)
	}
	return t, rows.Err()
}

type debugContext struct {
	Query string
	Table *table
}

type table struct {
	Header []string
	Rows   []row
}

type row []string

const txDebugTpl = `
<html>
<body>

<form method="get" id="query_form">
<textarea id="query_text" style="width:100%;height:300px" name="query">{{ .Query }}</textarea>
<input type="submit" value="Send" />
</form>

<form method="post">
<input type="submit" value="Quit" />
<input type="hidden" name="quit" value="true" />
</form>

{{ with .Table }}
	<table>
	<tr>{{ range .Header }}<th>{{ . }}{{ end }}</tr>
	{{ range .Rows }}
		<tr>{{ range . }}<td>{{ . }}{{ end }}
	{{ end }}
	</table>
{{ end }}

<script>
var el = document.getElementById("query_text");
if (el != null) {
	el.select();
}
</script>
`

func valueToString(i interface{}) string {
	switch c := i.(type) {
	case []uint8:
		return string(c)
	case time.Time:
		return c.UTC().Format("2006-01-02T15:04:05")
	default:
		return fmt.Sprint(i)
	}
}
