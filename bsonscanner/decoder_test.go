package bsondecoder

import (
	"io"
	"os"
	"os/exec"
	"testing"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	. "github.com/smartystreets/goconvey/convey"
)

func TestScanner(t *testing.T) {
	Convey("Scanner", t, func() {
		dir, e := setup()
		So(e, ShouldBeNil)
		f, e := os.Open(dir + "/albums.bson")
		So(e, ShouldBeNil)
		defer f.Close()

		scanner := New(f)
		So(scanner, ShouldNotBeNil)

		albums := []*Album{}
		for {
			var a *Album
			e = scanner.Decode(&a)
			if e == io.EOF {
				break
			}
			So(e, ShouldBeNil)
			albums = append(albums, a)
		}
		So(len(albums), ShouldEqual, 3)
		first := albums[0]
		So(first.Artist, ShouldEqual, "Mos Def")
		So(first.Title, ShouldEqual, "Black on Both Sides")
	})
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

func setup() (string, error) {
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
			So(e, ShouldBeNil)
		}

		cnt, e := db.C(colName).Find(nil).Count()
		So(e, ShouldBeNil)
		So(cnt, ShouldEqual, 3)

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
