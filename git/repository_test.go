package git

import (
	"os"
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestRepository(t *testing.T) {
	Convey("Repository", t, func() {
		So(1, ShouldEqual, 1)
		repo := &Repository{Origin: "git@github.com:test/it.git"}
		So(repo.cachePath(), ShouldEqual, os.ExpandEnv("$HOME/.dgtk/cache/git_repositories/github.com/test/it.git"))
	})
}
