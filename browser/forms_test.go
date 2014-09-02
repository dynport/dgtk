package browser

import (
	"io/ioutil"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLogin(t *testing.T) {
	Convey("parse file", t, func() {
		b, e := ioutil.ReadFile("fixtures/login_bahn.html")
		So(e, ShouldBeNil)
		So(b, ShouldNotBeNil)

		forms, e := loadForms("http://test.xx/a/test", b)
		So(e, ShouldBeNil)
		So(len(forms), ShouldEqual, 1)

		form := forms[0]
		So(form.Method, ShouldEqual, "post")
		So(form.Action, ShouldEqual, "http://test.xx/wlan/start.do;jsessionid=AD69644BB0438DDDC8A9371D73DEB2E4.P1")

		So(len(form.Inputs), ShouldEqual, 9)

		input := form.Inputs[0]

		So(input.Type, ShouldEqual, "submit")
		So(input.Name, ShouldEqual, "f_login_submit")
		So(input.Value, ShouldEqual, "")

		input = form.Inputs[3]
		So(input.Value, ShouldEqual, "t-mobile.net")
	})

	Convey("parse checked", t, func() {
		f := `
		<form>
		<input type='checkbox' name='terms' />
		<input type='newsletter' name='terms' checked/>
		</form>
		`

		forms, e := loadForms("http://127.0.0.1", []byte(f))
		So(e, ShouldBeNil)
		So(len(forms), ShouldEqual, 1)

		form := forms[0]
		inputs := form.Inputs
		So(len(inputs), ShouldEqual, 2)

		first := inputs[0]
		second := inputs[1]

		So(first, ShouldNotBeNil)
		So(first.Checked, ShouldEqual, false)
		So(second, ShouldNotBeNil)
		So(second.Checked, ShouldEqual, true)
	})
}
