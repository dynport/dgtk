package bsondecoder

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func TestScanner(t *testing.T) {
	dir, e := setup(t)
	if e != nil {
		t.Fatal("error setting up")
	}
	f, e := os.Open(dir + "/albums.bson")
	if e != nil {
		t.Fatal("error opening fixture")
	}
	defer f.Close()

	scanner := New(f)
	albums := []*Album{}
	for {
		var a *Album
		e = scanner.Decode(&a)
		if e == io.EOF {
			break
		}
		if e != nil {
			t.Error("error decoding album", e)
		}
		albums = append(albums, a)
	}
	if len(albums) != 3 {
		t.Errorf("expectet len(albums) to eq 3, was %v", len(albums))
	}
	first := albums[0]
	var ex, v interface{}
	ex = "Mos Def"
	v = first.Artist
	if ex != v {
		t.Errorf("expected first.Artist to be %#v, was %#v", ex, v)
	}

	ex = "Black on Both Sides"
	v = first.Title

	if ex != v {
		t.Errorf("expected first.Title to be %#v, was %#v", ex, v)
	}
}

type Album struct {
	Id     bson.ObjectId `bson:"_id"`
	Artist string        `bson:"artist"`
	Title  string        `bson:"title"`
}

const (
	colName = "albums"
	dbName  = "bsonscanner-test"
)

func setup(t *testing.T) (string, error) {
	e := func() error {
		ses, e := mgo.Dial("127.0.0.1")
		if e != nil {
			return e
		}
		defer ses.Close()

		db := ses.DB(dbName)
		e = db.DropDatabase()
		if e != nil {
			return e
		}
		albums := []*Album{
			{Artist: "Mos Def", Title: "Black on Both Sides"},
			{Artist: "Gang Starr", Title: "Full Clip"},
			{Artist: "A Tribe Called Quest", Title: "Low End Theory"},
		}

		for _, album := range albums {
			album.Id = bson.NewObjectId()
			e := db.C(colName).Insert(album)
			if e != nil {
				t.Fatal("error writing album", e)
			}
		}

		cnt, e := db.C(colName).Find(nil).Count()
		if e != nil {
			t.Fatal("error counting albums", e)
		}
		if cnt != 3 {
			t.Errorf("expected cnt to be 3, was %v", cnt)
		}
		if e = os.RemoveAll("tmp"); e != nil {
			return e
		}

		if e = os.MkdirAll("tmp", 0755); e != nil {
			return e
		}

		c := exec.Command("mongodump", "-d", dbName, "-o", ".")
		c.Dir = "tmp"
		c.Stderr = os.Stderr
		return c.Run()
	}()
	if e != nil {
		return "", e
	}
	return "tmp/" + dbName, nil
}
