package txdbg

import (
	"database/sql"
	"log"
	"os/exec"
)

func ExampleStartServer() {
	// dummy database setup
	db, err := sql.Open("postgres", "postgres://127.0.0.1/template1")
	if err != nil {
		log.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()

	waitForClose, addr := StartServer(tx)
	openURL(addr)
	<-waitForClose // blocks until the QUIT button is pressed or first error occurs on tx
}

// openURL tries to open the url with open, xdg-open, etc.
func openURL(u string) {
	for _, s := range []string{"open", "xdg-open"} {
		p, err := exec.LookPath(s)
		if err == nil {
			err := exec.Command(p, u).Run()
			if err == nil {
				return
			}
		}
	}
	log.Printf("please open %q", u)
	return
}
