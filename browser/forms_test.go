package browser

import (
	"io/ioutil"
	"testing"
)

func TestLoginFromFilr(t *testing.T) {
	b, err := ioutil.ReadFile("fixtures/login_bahn.html")
	if err != nil {
		t.Fatal("error loading fixture", err)
	}

	forms, err := loadForms("http://test.xx/a/test", b)
	if err != nil {
		t.Fatal("error loading forms", err)
	}
	if len(forms) != 1 {
		t.Errorf("expected to fund 1 form, found %d", len(forms))
	}

	form := forms[0]

	if form.Method != "post" {
		t.Errorf("expected first method to be post, was %s", form.Method)
	}
	var ex interface{} = "http://test.xx/wlan/start.do;jsessionid=AD69644BB0438DDDC8A9371D73DEB2E4.P1"
	if form.Action != ex {
		t.Errorf("expected Action to eq %q, was %q", ex, form.Action)
	}

	if len(form.Inputs) != 9 {
		t.Errorf("expected first form to have 9 inputs, was %d", len(form.Inputs))
	}

	input := form.Inputs[0]

	if input.Type != "submit" {
		t.Errorf("expected first input type to be submit, was %q", input.Type)
	}

	if input.Name != "f_login_submit" {
		t.Errorf("expected first name to be f_login_submit, was %q", input.Name)
	}

	if input.Value != "" {
		t.Errorf("expected value of first input to be blank, was %q", input.Value)
	}

	input = form.Inputs[3]

	if input.Value != "t-mobile.net" {
		t.Errorf("expected value of 4th input to eq t-mobile.net, was %q", input.Value)
	}
}

func TestLoginFormChecked(t *testing.T) {
	f := `
		<form>
		<input type='checkbox' name='terms' />
		<input type='newsletter' name='terms' checked/>
		</form>
		`

	forms, err := loadForms("http://127.0.0.1", []byte(f))
	if err != nil {
		t.Fatal("error loading forms", err)
	}
	if len(forms) != 1 {
		t.Fatalf("expected to find 1 forms, found %d", len(forms))
	}

	form := forms[0]
	inputs := form.Inputs
	if len(inputs) != 2 {
		t.Fatalf("expected to find 2 inputs, found %d", len(inputs))
	}

	first := inputs[0]
	second := inputs[1]

	if first == nil {
		t.Fatalf("expected first to not be nil")
	}

	if first.Checked != false {
		t.Fatalf("expected first input to not be checked")
	}

	if second == nil {
		t.Fatalf("expected second to not be nil but it was")
	}
	if second.Checked != true {
		t.Errorf("expected second to be checked")
	}
}
