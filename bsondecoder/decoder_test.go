package bsondecoder

import (
	"io"
	"os"
	"testing"

	"labix.org/v2/mgo/bson"
)

func TestScanner(t *testing.T) {
	f, err := os.Open("fixtures/albums.bson")
	if err != nil {
		t.Fatalf("error opening fixture: %s", err)
	}
	defer f.Close()

	scanner := New(f)
	albums := []*Album{}
	for {
		var a *Album
		if err = scanner.Decode(&a); err == io.EOF {
			break
		} else if err != nil {
			t.Error("error decoding album", err)
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
